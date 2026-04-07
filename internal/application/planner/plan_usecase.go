//大概的思路就是：输入校验，查询用户，拉取候选POI，如果无候选的话报错，然后计算天数+预算可行性检验，组装求解器输入，然后调用求解器，构建行程实体，领域校验，最后持久化行程并返回结果

package planner

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"atlas-routex/internal/domain/entity"
	"atlas-routex/internal/domain/repository"
)

// 业务语义错误，供handler映射HTTP状态或者错误码
var (
	ErrInvalidPlanInput  = errors.New("planner: invalid plan input")
	ErrUserNotFound      = errors.New("planner: user not found")
	ErrNoCandidatePOIs   = errors.New("planner: no candidate pois")
	ErrBudgetInfeasible  = errors.New("planner: budget infeasible for given candidates")
	ErrSolverTimeout     = errors.New("planner: solver deadline exceeded")
	ErrEmptySolverResult = errors.New("planner: solver returned empty schedule")
)

// SolverInput 是应用层和算法层之间的DTO对象，用于将领域数据传递给求解器。
type SolverInput struct {
	Candidates  []*entity.POI
	StartDate   time.Time
	EndDate     time.Time
	DayCount    int
	Constraints []entity.Constraint
	TotalBudget float64
	Currency    string
}

// SolverOutput 求解器给出的按天 POI 顺序（POI ID）。
type SolverOutput struct {
	Days [][]string
}

// PlanUsecase 新建/规划行程并持久化。
type PlanUsecase struct {
	users       repository.UserRepository
	pois        repository.POIRepository
	itineraries repository.ItineraryRepository
	solver      TripSolver
}

// NewPlanUsecase 构造用例。
func NewPlanUsecase(
	users repository.UserRepository,
	pois repository.POIRepository,
	itineraries repository.ItineraryRepository,
	solver TripSolver,
) *PlanUsecase {
	return &PlanUsecase{
		users:       users,
		pois:        pois,
		itineraries: itineraries,
		solver:      solver,
	}
}

// Execute 从输入编排：校验 → 用户 → 候选 POI → 求解 → 填充 Itinerary → Save。
func (uc *PlanUsecase) Execute(ctx context.Context, in *PlanInput) (*entity.Itinerary, error) {
	if err := validatePlanInput(in); err != nil {
		return nil, err
	}
	if uc.solver == nil {
		return nil, fmt.Errorf("planner: trip solver is nil")
	}

	if in.UserID != "" {
		u, err := uc.users.FindByID(ctx, in.UserID)
		if err != nil {
			return nil, err
		}
		if u == nil {
			return nil, ErrUserNotFound
		}
	}

	limit := in.MaxCandidatePOIs
	if limit <= 0 {
		limit = 80
	}

	candidates, err := uc.fetchCandidates(ctx, in, limit)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, ErrNoCandidatePOIs
	}

	dayCount := daysInclusive(in.StartDate, in.EndDate)
	if dayCount <= 0 {
		return nil, ErrInvalidPlanInput
	}

	if in.TotalBudget > 0 && !minSpendFeasible(candidates, dayCount, in.TotalBudget) {
		return nil, ErrBudgetInfeasible
	}

	currency := in.Currency
	if currency == "" {
		currency = "CNY"
	}

	solverIn := &SolverInput{
		Candidates:  candidates,
		StartDate:   in.StartDate,
		EndDate:     in.EndDate,
		DayCount:    dayCount,
		Constraints: in.Constraints,
		TotalBudget: in.TotalBudget,
		Currency:    currency,
	}

	out, err := uc.solver.Solve(ctx, solverIn)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, ErrSolverTimeout
		}
		if errors.Is(err, ErrBudgetInfeasible) {
			return nil, ErrBudgetInfeasible
		}
		return nil, err
	}
	if out == nil || len(out.Days) == 0 {
		return nil, ErrEmptySolverResult
	}

	it := buildItineraryFromSolver(in, out, candidates)
	if err := it.Validate(); err != nil {
		return nil, fmt.Errorf("planner: built itinerary invalid: %w", err)
	}

	if err := uc.itineraries.Save(ctx, it); err != nil {
		return nil, err
	}
	return it, nil
}

func validatePlanInput(in *PlanInput) error {
	if in == nil {
		return ErrInvalidPlanInput
	}
	if in.ItineraryName == "" || in.City == "" {
		return ErrInvalidPlanInput
	}
	if in.EndDate.Before(in.StartDate) {
		return ErrInvalidPlanInput
	}
	return nil
}

func daysInclusive(start, end time.Time) int {
	// 按日历日计算，包含开始日与结束日。
	d := int(end.Sub(start).Hours()/24) + 1
	if d < 1 {
		return 1
	}
	return d
}

func (uc *PlanUsecase) fetchCandidates(ctx context.Context, in *PlanInput, limit int) ([]*entity.POI, error) {
	qb := repository.NewPOIQueryBuilder().WithCity(in.City).WithPagination(1, limit)
	if len(in.Tags) > 0 {
		qb = qb.WithTags(in.Tags)
	}
	q := qb.Build()
	return uc.pois.FindByQuery(ctx, q)
}

// minSpendFeasible：粗判——每天至少去一个点时的最低门票/人均之和是否已超过总预算。
func minSpendFeasible(pois []*entity.POI, dayCount int, budget float64) bool {
	prices := make([]float64, 0, len(pois))
	for _, p := range pois {
		prices = append(prices, estimateVisitCost(p))
	}
	sort.Float64s(prices)
	n := dayCount
	if n > len(prices) {
		n = len(prices)
	}
	var sum float64
	for i := 0; i < n; i++ {
		sum += prices[i]
	}
	return sum <= budget
}

func estimateVisitCost(p *entity.POI) float64 {
	switch p.Category {
	case entity.CategoryRestaurant:
		if p.AvgPrice > 0 {
			return p.AvgPrice
		}
		return p.TicketPrice
	default:
		if p.TicketPrice > 0 {
			return p.TicketPrice
		}
		return p.AvgPrice
	}
}

func buildItineraryFromSolver(in *PlanInput, out *SolverOutput, candidates []*entity.POI) *entity.Itinerary {
	byID := make(map[string]*entity.POI, len(candidates))
	for _, p := range candidates {
		byID[p.ID] = p
	}

	dayCount := daysInclusive(in.StartDate, in.EndDate)
	it := entity.NewItinerary(in.UserID, in.ItineraryName, in.StartDate, in.EndDate)
	it.DayCount = dayCount
	it.Description = in.Description
	it.Constraints = append([]entity.Constraint(nil), in.Constraints...)
	if in.TotalBudget > 0 {
		it.Budget.TotalBudget = in.TotalBudget
		if in.Currency != "" {
			it.Budget.Currency = in.Currency
		}
	}
	for d := 1; d <= dayCount; d++ {
		date := in.StartDate.AddDate(0, 0, d-1)
		it.AddDay(entity.NewItineraryDay(d, date))
	}

	for i, poiIDs := range out.Days {
		dayNumber := i + 1
		if dayNumber > len(it.Days) {
			break
		}
		slot := time.Date(
			it.Days[i].Date.Year(), it.Days[i].Date.Month(), it.Days[i].Date.Day(),
			9, 0, 0, 0, it.Days[i].Date.Location(),
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
			_ = it.AddAttraction(dayNumber, attr)
			slot = end
		}
	}

	it.CalculateTotalCost()
	it.CalculateTotalDistance()
	it.UpdateStatus(entity.ItineraryStatusPlanned)
	return it
}
