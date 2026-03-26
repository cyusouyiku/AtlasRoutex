package planner

import (
	"fmt"
	"time"

	"atlas-routex/internal/domain/entity"
)


// 应用层命令（用例入参，可与 HTTP/gRPC 解耦）

// PlanInput 规划命令。
type PlanInput struct {
	UserID           string
	ItineraryName    string
	Description      string
	City             string
	StartDate        time.Time
	EndDate          time.Time
	TotalBudget      float64
	Currency         string
	Tags             []string
	Constraints      []entity.Constraint
	MaxCandidatePOIs int
}

// Validate 规划命令校验（后续可将规则委托给 pkg/utils/validator）。
func (in *PlanInput) Validate() error {
	if in == nil {
		return ErrInvalidPlanInput
	}
	if in.ItineraryName == "" || in.City == "" {
		return ErrInvalidPlanInput
	}
	if in.EndDate.Before(in.StartDate) {
		return ErrInvalidPlanInput
	}
	if in.MaxCandidatePOIs < 0 {
		return ErrInvalidPlanInput
	}
	return nil
}

// PatchKind 变更类型。
type PatchKind string

const (
	PatchReplacePOI       PatchKind = "replace_poi"
	PatchReschedule       PatchKind = "reschedule"
	PatchRemoveAttraction PatchKind = "remove_attraction"
)

// AttractionPatch 对某一天某一活动的修改意图。
type AttractionPatch struct {
	Kind           PatchKind
	DayNumber      int
	AttractionID   string
	NewPOIID       string
	NewStart       *time.Time
	NewEnd         *time.Time
	NewStayMinutes *int
}

// Validate 单条补丁校验。
func (p *AttractionPatch) Validate() error {
	if p == nil {
		return ErrInvalidAdjustInput
	}
	if p.DayNumber < 1 {
		return ErrInvalidAdjustInput
	}
	if p.AttractionID == "" {
		return ErrInvalidAdjustInput
	}
	switch p.Kind {
	case PatchReplacePOI:
		if p.NewPOIID == "" {
			return ErrInvalidAdjustInput
		}
	case PatchReschedule, PatchRemoveAttraction:
	default:
		return fmt.Errorf("%w: unknown patch kind %q", ErrInvalidAdjustInput, p.Kind)
	}
	return nil
}

// AdjustInput 调整命令。
type AdjustInput struct {
	ItineraryID         string
	UserID              string
	ExpectedUpdatedAt   *time.Time
	Patches             []AttractionPatch
	FullReplan          bool
	ReplanMaxCandidates int
}

// Validate 调整命令校验。
func (in *AdjustInput) Validate() error {
	if in == nil || in.ItineraryID == "" {
		return ErrInvalidAdjustInput
	}
	if in.ReplanMaxCandidates < 0 {
		return ErrInvalidAdjustInput
	}
	for i := range in.Patches {
		if err := in.Patches[i].Validate(); err != nil {
			return fmt.Errorf("patches[%d]: %w", i, err)
		}
	}
	return nil
}


// 传输 DTO（JSON 等协议层使用，避免直接暴露实体）


// DestinationDTO 目的地（城市/区域）。
type DestinationDTO struct {
	City   string `json:"city"`
	Region string `json:"region,omitempty"`
}

// BudgetDTO 预算的可序列化形式。
type BudgetDTO struct {
	Total    float64 `json:"total"`
	Currency string  `json:"currency,omitempty"`
}

// PreferencesDTO 偏好标签与节奏（非领域 UserPreferences 全量，仅规划常用字段）。
type PreferencesDTO struct {
	Tags []string `json:"tags,omitempty"`
	Pace string   `json:"pace,omitempty"`
}

// ConstraintDTO 约束的可序列化投影，与 entity.Constraint 可互转。
type ConstraintDTO struct {
	ID          string `json:"id,omitempty"`
	Type        string `json:"type"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Priority    int    `json:"priority"`
	IsHard      bool   `json:"is_hard"`
}

// ToEntity 转为领域约束（时间字段由仓储/领域在落库时补全）。
func (c ConstraintDTO) ToEntity() entity.Constraint {
	return entity.Constraint{
		ID:          c.ID,
		Type:        entity.ConstraintType(c.Type),
		Name:        c.Name,
		Description: c.Description,
		Priority:    c.Priority,
		IsHard:      c.IsHard,
	}
}

// ConstraintFromEntity 从实体生成 DTO。
func ConstraintFromEntity(c entity.Constraint) ConstraintDTO {
	return ConstraintDTO{
		ID:          c.ID,
		Type:        string(c.Type),
		Name:        c.Name,
		Description: c.Description,
		Priority:    c.Priority,
		IsHard:      c.IsHard,
	}
}

// PlanCreateRequest 创建/规划行程的 API 形态入参。
type PlanCreateRequest struct {
	UserID            string          `json:"user_id"`
	ItineraryName     string          `json:"itinerary_name"`
	Description       string          `json:"description,omitempty"`
	Destination       DestinationDTO  `json:"destination"`
	StartDate         time.Time       `json:"start_date"`
	EndDate           time.Time       `json:"end_date"`
	Budget            *BudgetDTO      `json:"budget,omitempty"`
	Preferences       *PreferencesDTO `json:"preferences,omitempty"`
	Constraints       []ConstraintDTO `json:"constraints,omitempty"`
	MaxCandidatePOIs  int             `json:"max_candidate_pois,omitempty"`
}

// ValidateRequest 校验 HTTP 层入参。
func (r *PlanCreateRequest) ValidateRequest() error {
	if r == nil {
		return ErrInvalidPlanInput
	}
	if r.ItineraryName == "" || r.Destination.City == "" {
		return ErrInvalidPlanInput
	}
	if r.EndDate.Before(r.StartDate) {
		return ErrInvalidPlanInput
	}
	if r.MaxCandidatePOIs < 0 {
		return ErrInvalidPlanInput
	}
	if r.Budget != nil && r.Budget.Total < 0 {
		return ErrInvalidPlanInput
	}
	return nil
}

// ToPlanInput 转为用例命令。
func (r *PlanCreateRequest) ToPlanInput() (*PlanInput, error) {
	if err := r.ValidateRequest(); err != nil {
		return nil, err
	}
	in := &PlanInput{
		UserID:           r.UserID,
		ItineraryName:    r.ItineraryName,
		Description:      r.Description,
		City:             r.Destination.City,
		StartDate:        r.StartDate,
		EndDate:          r.EndDate,
		MaxCandidatePOIs: r.MaxCandidatePOIs,
	}
	if r.Budget != nil {
		in.TotalBudget = r.Budget.Total
		in.Currency = r.Budget.Currency
	}
	if r.Preferences != nil {
		in.Tags = append([]string(nil), r.Preferences.Tags...)
	}
	if len(r.Constraints) > 0 {
		in.Constraints = make([]entity.Constraint, 0, len(r.Constraints))
		for _, c := range r.Constraints {
			in.Constraints = append(in.Constraints, c.ToEntity())
		}
	}
	return in, nil
}

// PatchDTO 补丁的 API 形态（与 AttractionPatch 字段对齐，便于 json 绑定）。
type PatchDTO struct {
	Kind           PatchKind  `json:"kind"`
	DayNumber      int        `json:"day_number"`
	AttractionID   string     `json:"attraction_id"`
	NewPOIID       string     `json:"new_poi_id,omitempty"`
	NewStart       *time.Time `json:"new_start,omitempty"`
	NewEnd         *time.Time `json:"new_end,omitempty"`
	NewStayMinutes *int       `json:"new_stay_minutes,omitempty"`
}

// ToAttractionPatch 转为用例补丁。
func (d PatchDTO) ToAttractionPatch() AttractionPatch {
	return AttractionPatch{
		Kind:           d.Kind,
		DayNumber:      d.DayNumber,
		AttractionID:   d.AttractionID,
		NewPOIID:       d.NewPOIID,
		NewStart:       d.NewStart,
		NewEnd:         d.NewEnd,
		NewStayMinutes: d.NewStayMinutes,
	}
}

// AdjustRequest 调整行程的 API 形态入参。
type AdjustRequest struct {
	ItineraryID         string     `json:"itinerary_id"`
	UserID              string     `json:"user_id,omitempty"`
	ExpectedUpdatedAt   *time.Time `json:"expected_updated_at,omitempty"`
	Patches             []PatchDTO `json:"patches"`
	FullReplan          bool       `json:"full_replan"`
	ReplanMaxCandidates int        `json:"replan_max_candidates,omitempty"`
}

// ValidateRequest 校验 API 入参。
func (r *AdjustRequest) ValidateRequest() error {
	if r == nil || r.ItineraryID == "" {
		return ErrInvalidAdjustInput
	}
	if r.ReplanMaxCandidates < 0 {
		return ErrInvalidAdjustInput
	}
	for i := range r.Patches {
		p := r.Patches[i].ToAttractionPatch()
		if err := p.Validate(); err != nil {
			return fmt.Errorf("patches[%d]: %w", i, err)
		}
	}
	return nil
}

// ToAdjustInput 转为用例命令。
func (r *AdjustRequest) ToAdjustInput() (*AdjustInput, error) {
	if err := r.ValidateRequest(); err != nil {
		return nil, err
	}
	patches := make([]AttractionPatch, len(r.Patches))
	for i := range r.Patches {
		patches[i] = r.Patches[i].ToAttractionPatch()
	}
	return &AdjustInput{
		ItineraryID:         r.ItineraryID,
		UserID:              r.UserID,
		ExpectedUpdatedAt:   r.ExpectedUpdatedAt,
		Patches:             patches,
		FullReplan:          r.FullReplan,
		ReplanMaxCandidates: r.ReplanMaxCandidates,
	}, nil
}

// ItinerarySummaryDTO 行程摘要响应（避免把整个实体图返回给前端）。
type ItinerarySummaryDTO struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Description string    `json:"description,omitempty"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	DayCount    int       `json:"day_count"`
	TotalCost   float64   `json:"total_cost,omitempty"`
	Currency    string    `json:"currency,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ItineraryToSummaryDTO 从实体生成摘要 DTO。
func ItineraryToSummaryDTO(it *entity.Itinerary) ItinerarySummaryDTO {
	if it == nil {
		return ItinerarySummaryDTO{}
	}
	out := ItinerarySummaryDTO{
		ID:          it.ID,
		UserID:      it.UserID,
		Name:        it.Name,
		Status:      string(it.Status),
		Description: it.Description,
		StartDate:   it.StartDate,
		EndDate:     it.EndDate,
		DayCount:    it.DayCount,
		UpdatedAt:   it.UpdatedAt,
	}
	if it.Budget != nil {
		out.TotalCost = it.Budget.TotalCost
		out.Currency = it.Budget.Currency
	}
	return out
}
