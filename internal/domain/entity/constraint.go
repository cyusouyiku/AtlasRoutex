package entity

import (
	"time"
)

// ConstraintType 约束类型枚举
type ConstraintType string

const (
	// 时间约束
	ConstraintTypeOperatingHours ConstraintType = "operating_hours"      // 营业时间
	ConstraintTypeStayDuration   ConstraintType = "stay_duration"        // 停留时间
	ConstraintTypeMealTime       ConstraintType = "meal_time"            // 用餐时间

	// 空间约束
	ConstraintTypeDistance ConstraintType = "distance"      // 距离限制
	ConstraintTypeArea     ConstraintType = "area"          // 地区限制
	ConstraintTypeLocation ConstraintType = "location"      // 位置约束

	// 预算约束
	ConstraintTypeTotalBudget    ConstraintType = "total_budget"    // 总预算
	ConstraintTypeCategoryBudget ConstraintType = "category_budget" // 分类预算

	// 体力约束
	ConstraintTypeWalkingDistance ConstraintType = "walking_distance" // 行走距离
	ConstraintTypeClimbingHeight  ConstraintType = "climbing_height"  // 爬升高度
	ConstraintTypeDailyActivity   ConstraintType = "daily_activity"   // 日活动强度

	// 用户偏好约束
	ConstraintTypeUserPreference ConstraintType = "user_preference" // 用户偏好
	ConstraintTypeAgeLimit       ConstraintType = "age_limit"       // 年龄限制
)

// Constraint 约束基础实体
type Constraint struct {
	ID          string        `json:"id"`
	Type        ConstraintType `json:"type"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Priority    int           `json:"priority"`     // 优先级：1-10，数字越大优先级越高
	IsHard      bool          `json:"is_hard"`      // 是否为硬约束（必须满足）
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// TimeConstraint 时间约束
type TimeConstraint struct {
	*Constraint
	StartTime time.Time `json:"start_time"` // 约束开始时间
	EndTime   time.Time `json:"end_time"`   // 约束结束时间
	Duration  int       `json:"duration"`   // 持续时间（分钟）
}

// BudgetConstraint 预算约束
type BudgetConstraint struct {
	*Constraint
	MaxBudget    float64 `json:"max_budget"`    // 最大预算
	MinBudget    float64 `json:"min_budget"`    // 最小预算
	Currency     string  `json:"currency"`      // 货币单位
	Category     string  `json:"category"`      // 预算类别（住宿、餐饮、交通等）
}

// DistanceConstraint 距离约束
type DistanceConstraint struct {
	*Constraint
	MaxDistance float64 `json:"max_distance"` // 最大距离（公里）
	MinDistance float64 `json:"min_distance"` // 最小距离（公里）
	Unit        string  `json:"unit"`         // 单位（km, m等）
}

// PhysicalConstraint 体力约束
type PhysicalConstraint struct {
	*Constraint
	MaxWalkingDistance float64 `json:"max_walking_distance"` // 最大行走距离（公里）
	MaxClimbingHeight  float64 `json:"max_climbing_height"`  // 最大爬升高度（米）
	ActivityLevel      string  `json:"activity_level"`       // 活动强度（低、中、高）
}

// PreferenceConstraint 偏好约束
type PreferenceConstraint struct {
	*Constraint
	UserID       string   `json:"user_id"`       // 用户ID
	PreferredPOI []string `json:"preferred_poi"` // 偏好景点列表
	AvoidPOI     []string `json:"avoid_poi"`     // 回避景点列表
	Tags         []string `json:"tags"`          // 偏好标签
}

// ConstraintValidator 约束验证器
type ConstraintValidator struct {
	constraints []Constraint
}

// NewConstraintValidator 创建新的约束验证器
func NewConstraintValidator() *ConstraintValidator {
	return &ConstraintValidator{
		constraints: make([]Constraint, 0),
	}
}

// AddConstraint 添加约束
func (cv *ConstraintValidator) AddConstraint(c Constraint) {
	cv.constraints = append(cv.constraints, c)
}

// Validate 验证约束
func (cv *ConstraintValidator) Validate(target interface{}) bool {
	// 验证逻辑依赖于具体的约束和目标对象
	// 这是一个扩展点，具体实现在应用层
	for _, c := range cv.constraints {
		if c.IsHard {
			// 硬约束必须满足
			continue
		}
	}
	return true
}

