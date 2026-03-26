package planner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"atlas-routex/internal/domain/entity"
	"atlas-routex/internal/domain/repository"
)

// 调整行程相关错误。
var (
	ErrInvalidAdjustInput    = errors.New("planner: invalid adjust input")
	ErrItineraryNotFound     = errors.New("planner: itinerary not found")
	ErrAdjustVersionConflict = errors.New("planner: itinerary was modified by another request")
	ErrAdjustForbidden       = errors.New("planner: user cannot modify this itinerary")
	ErrAttractionNotFound    = errors.New("planner: attraction not found on given day")
	ErrReplacementPOINotFound = errors.New("planner: replacement poi not found")
)

// PatchKind 变更类型。
type PatchKind string

const (
	PatchReplacePOI       PatchKind = "replace_poi"
	PatchReschedule       PatchKind = "reschedule"
	PatchRemoveAttraction PatchKind = "remove_attraction"
)

// AttractionPatch 对某一天某一活动的修改意图。
type AttractionPatch struct {
	Kind         PatchKind
	DayNumber    int    // 1-based，与 entity.ItineraryDay 一致
	AttractionID string // DayAttraction.ID

	// ReplacePOI：替换为新的 POI（仓储拉取）。
	NewPOIID string

	// Reschedule：直接改写时段；若只填其一可由调用方保证与 StayDuration 一致。
	NewStart       *time.Time
	NewEnd         *time.Time
	NewStayMinutes *int
}

// AdjustInput 调整命令。
type AdjustInput struct {
	ItineraryID string
	// UserID 为空则跳过归属校验；非空时必须与行程 UserID 一致。
	UserID string
	// ExpectedUpdatedAt 乐观锁：非空时须与当前行程 UpdatedAt 一致（API 层可传客户端上次读到的值）。
	ExpectedUpdatedAt *time.Time

	Patches []AttractionPatch

	// FullReplan 为 true 时，在应用补丁后按城市候选集调用 TripSolver 全量重排天与顺序；
	// 为 false 时仅应用补丁并重算费用/统计（适合只改时段）。
	FullReplan bool

	// ReplanMaxCandidates 全量重算时拉取 POI 上限，0 表示默认 80。
	ReplanMaxCandidates int
}

// AdjustUsecase 行程调整用例：加载 → 并发占位 → 应用补丁 → 可选全量重算 → Update。
type AdjustUsecase struct {
	pois        repository.POIRepository
	itineraries repository.ItineraryRepository
	solver      TripSolver
}

// NewAdjustUsecase 构造调整用例。solver 在 FullReplan 时必须非空。
func NewAdjustUsecase(
	pois repository.POIRepository,
	itineraries repository.ItineraryRepository,
	solver TripSolver,
) *AdjustUsecase {
	return &AdjustUsecase{
		pois:        pois,
		itineraries: itineraries,
		solver:      solver,
	}
}

// Execute 执行调整并持久化（Update）。
func (uc *AdjustUsecase) Execute(ctx context.Context, in *AdjustInput) (*entity.Itinerary, error) {
	if err := validateAdjustInput(in); err != nil {
		return nil, err
	}

	it, err := uc.itineraries.FindByID(ctx, in.ItineraryID)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, ErrItineraryNotFound
	}

	if in.UserID != "" && it.UserID != in.UserID {
		return nil, ErrAdjustForbidden
	}

	if in.ExpectedUpdatedAt != nil && !it.UpdatedAt.Equal(*in.ExpectedUpdatedAt) {
		return nil, ErrAdjustVersionConflict
	}

	for i := range in.Patches {
		if err := uc.applyPatch(ctx, it, &in.Patches[i]); err != nil {
			return nil, err
		}
	}

	if in.FullReplan {
		if uc.solver == nil {
			return nil, fmt.Errorf("planner: trip solver is required for full replan")
		}
		if err := uc.runFullReplan(ctx, it, in.ReplanMaxCandidates); err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return nil, ErrSolverTimeout
			}
			if errors.Is(err, ErrBudgetInfeasible) {
				return nil, ErrBudgetInfeasible
			}
			return nil, err
		}
	}

	it.CalculateTotalCost()
	it.CalculateTotalDistance()

	if err := uc.itineraries.Update(ctx, it); err != nil {
		return nil, err
	}
	return it, nil
}

func validateAdjustInput(in *AdjustInput) error {
	if in == nil || in.ItineraryID == "" {
		return ErrInvalidAdjustInput
	}
	return nil
}

func (uc *AdjustUsecase) applyPatch(ctx context.Context, it *entity.Itinerary, p *AttractionPatch) error {
	switch p.Kind {
	case PatchReplacePOI:
		if p.NewPOIID == "" {
			return ErrInvalidAdjustInput
		}
		attr, err := findAttraction(it, p.DayNumber, p.AttractionID)
		if err != nil {
			return err
		}
		poi, err := uc.pois.FindByID(ctx, p.NewPOIID)
		if err != nil {
			return err
		}
		if poi == nil {
			return ErrReplacementPOINotFound
		}
		attr.POI = poi
		attr.Cost = estimateVisitCost(poi)
		if p.NewStayMinutes != nil && *p.NewStayMinutes > 0 {
			attr.StayDuration = *p.NewStayMinutes
			attr.EndTime = attr.StartTime.Add(time.Duration(attr.StayDuration) * time.Minute)
		}

	case PatchReschedule:
		attr, err := findAttraction(it, p.DayNumber, p.AttractionID)
		if err != nil {
			return err
		}
		if p.NewStart != nil {
			attr.StartTime = *p.NewStart
		}
		if p.NewEnd != nil {
			attr.EndTime = *p.NewEnd
		}
		if p.NewStayMinutes != nil && *p.NewStayMinutes > 0 {
			attr.StayDuration = *p.NewStayMinutes
			attr.EndTime = attr.StartTime.Add(time.Duration(attr.StayDuration) * time.Minute)
		} else if !attr.EndTime.After(attr.StartTime) {
			d := attr.StayDuration
			if d <= 0 {
				d = 90
			}
			attr.StayDuration = d
			attr.EndTime = attr.StartTime.Add(time.Duration(d) * time.Minute)
		} else {
			attr.StayDuration = int(attr.EndTime.Sub(attr.StartTime).Minutes())
		}

	case PatchRemoveAttraction:
		if err := removeAttraction(it, p.DayNumber, p.AttractionID); err != nil {
			return err
		}

	default:
		return fmt.Errorf("planner: unknown patch kind %q", p.Kind)
	}
	return nil
}

func findAttraction(it *entity.Itinerary, dayNumber int, attractionID string) (*entity.DayAttraction, error) {
	if dayNumber < 1 || dayNumber > len(it.Days) {
		return nil, entity.ErrInvalidDayNumber
	}
	day := it.Days[dayNumber-1]
	for _, a := range day.Attractions {
		if a != nil && a.ID == attractionID {
			return a, nil
		}
	}
	return nil, ErrAttractionNotFound
}

func removeAttraction(it *entity.Itinerary, dayNumber int, attractionID string) error {
	if dayNumber < 1 || dayNumber > len(it.Days) {
		return entity.ErrInvalidDayNumber
	}
	day := it.Days[dayNumber-1]
	out := day.Attractions[:0]
	for _, a := range day.Attractions {
		if a == nil || a.ID != attractionID {
			out = append(out, a)
		}
	}
	if len(out) == len(day.Attractions) {
		return ErrAttractionNotFound
	}
	day.Attractions = out
	for i, a := range day.Attractions {
		if a != nil {
			a.Order = i
		}
	}
	return nil
}

func (uc *AdjustUsecase) runFullReplan(ctx context.Context, it *entity.Itinerary, maxCandidates int) error {
	city := deriveCityFromItinerary(it)
	if city == "" {
		return fmt.Errorf("planner: cannot derive city for replan")
	}

	limit := maxCandidates
	if limit <= 0 {
		limit = 80
	}
	q := repository.NewPOIQueryBuilder().WithCity(city).WithPagination(1, limit).Build()
	candidates, err := uc.pois.FindByQuery(ctx, q)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return ErrNoCandidatePOIs
	}

	dayCount := daysInclusive(it.StartDate, it.EndDate)
	totalBudget := 0.0
	if it.Budget != nil {
		totalBudget = it.Budget.TotalBudget
	}
	currency := "CNY"
	if it.Budget != nil && it.Budget.Currency != "" {
		currency = it.Budget.Currency
	}

	if totalBudget > 0 && !minSpendFeasible(candidates, dayCount, totalBudget) {
		return ErrBudgetInfeasible
	}

	solverIn := &SolverInput{
		Candidates:  candidates,
		StartDate:   it.StartDate,
		EndDate:     it.EndDate,
		DayCount:    dayCount,
		Constraints: append([]entity.Constraint(nil), it.Constraints...),
		TotalBudget: totalBudget,
		Currency:    currency,
	}

	out, err := uc.solver.Solve(ctx, solverIn)
	if err != nil {
		return err
	}
	if out == nil || len(out.Days) == 0 {
		return ErrEmptySolverResult
	}

	rebuildItineraryDaysFromSolver(it, out, candidates)
	return nil
}

func deriveCityFromItinerary(it *entity.Itinerary) string {
	for _, d := range it.Days {
		for _, a := range d.Attractions {
			if a != nil && a.POI != nil && a.POI.City != "" {
				return a.POI.City
			}
		}
	}
	return ""
}

// rebuildItineraryDaysFromSolver 保留行程 ID/用户/名称/日期/元数据，仅替换 Days 与统计字段。
func rebuildItineraryDaysFromSolver(it *entity.Itinerary, out *SolverOutput, candidates []*entity.POI) {
	byID := make(map[string]*entity.POI, len(candidates))
	for _, p := range candidates {
		byID[p.ID] = p
	}

	it.Days = make([]*entity.ItineraryDay, 0)
	dayCount := daysInclusive(it.StartDate, it.EndDate)
	for d := 1; d <= dayCount; d++ {
		date := it.StartDate.AddDate(0, 0, d-1)
		it.AddDay(entity.NewItineraryDay(d, date))
	}

	for i, poiIDs := range out.Days {
		dayNumber := i + 1
		if dayNumber > len(it.Days) {
			break
		}
		day := it.Days[i]
		slot := time.Date(
			day.Date.Year(), day.Date.Month(), day.Date.Day(),
			9, 0, 0, 0, day.Date.Location(),
		)
		for order, id := range poiIDs {
			p := byID[id]
			if p == nil {
				continue
			}
			stay := p.Duration
			if stay <= 0 {
				stay = p.AvgStayTime
			}
			if stay <= 0 {
				stay = 90
			}
			start := slot
			end := start.Add(time.Duration(stay) * time.Minute)
			attr := &entity.DayAttraction{
				ID:           uuid.New().String(),
				POI:          p,
				StartTime:    start,
				EndTime:      end,
				StayDuration: stay,
				Order:        order,
				Cost:         estimateVisitCost(p),
			}
			day.Attractions = append(day.Attractions, attr)
			slot = end
		}
	}

	// 保持与 NewItinerary 语义接近
	it.DayCount = dayCount
	it.CalculateTotalCost()
	it.CalculateTotalDistance()
}
