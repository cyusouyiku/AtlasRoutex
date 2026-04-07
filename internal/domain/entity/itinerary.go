package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidDayNumber     = errors.New("invalid day number")
	ErrInvalidItineraryName = errors.New("itinerary name cannot be empty")
	ErrInvalidDateRange     = errors.New("end date must be after start date")
	ErrEmptyItinerary       = errors.New("itinerary must have at least one day")
	ErrItineraryNotFound    = errors.New("itinerary not found")
)

// generateID 生成唯一ID
func generateID() string {
	return uuid.New().String()
}

// ItineraryStatus 行程状态
type ItineraryStatus string

const (
	ItineraryStatusDraft     ItineraryStatus = "draft"     // 草稿
	ItineraryStatusPlanned   ItineraryStatus = "planned"   // 已规划
	ItineraryStatusConfirmed ItineraryStatus = "confirmed" // 已确认
	ItineraryStatusOngoing   ItineraryStatus = "ongoing"   // 进行中
	ItineraryStatusCompleted ItineraryStatus = "completed" // 已完成
	ItineraryStatusCancelled ItineraryStatus = "cancelled" // 已取消
)

// Itinerary 行程实体
type Itinerary struct {
	ID          string               `json:"id"`
	UserID      string               `json:"user_id"`     // 用户ID
	Name        string               `json:"name"`        // 行程名称
	Description string               `json:"description"` // 行程描述
	Status      ItineraryStatus      `json:"status"`      // 行程状态
	StartDate   time.Time            `json:"start_date"`  // 开始日期
	EndDate     time.Time            `json:"end_date"`    // 结束日期
	DayCount    int                  `json:"day_count"`   // 天数
	Days        []*ItineraryDay      `json:"days"`        // 按天分组的行程
	Budget      *ItineraryBudget     `json:"budget"`      // 预算信息
	Statistics  *ItineraryStatistics `json:"statistics"`  // 行程统计
	Constraints []Constraint         `json:"constraints"` // 约束条件
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	PublishedAt *time.Time           `json:"published_at"`
}

// ItineraryDay 行程中的每一天
type ItineraryDay struct {
	DayNumber   int              `json:"day_number"`  // 第几天（1开始）
	Date        time.Time        `json:"date"`        // 日期
	Attractions []*DayAttraction `json:"attractions"` // 该天的景点列表
	Meals       []*DayMeal       `json:"meals"`       // 该天的餐饮
	Hotel       *DayHotel        `json:"hotel"`       // 住宿信息
	Notes       string           `json:"notes"`       // 备注
	Statistics  *DayStatistics   `json:"statistics"`  // 该天的统计信息
}

// DayAttraction 行程中某一天的景点
type DayAttraction struct {
	ID             string          `json:"id"`             // 景点ID
	POI            *POI            `json:"poi"`            // POI信息
	StartTime      time.Time       `json:"start_time"`     // 开始时间
	EndTime        time.Time       `json:"end_time"`       // 结束时间
	StayDuration   int             `json:"stay_duration"`  // 停留时间（分钟）
	Order          int             `json:"order"`          // 访问顺序
	Cost           float64         `json:"cost"`           // 该景点花费
	Transportation *Transportation `json:"transportation"` // 交通信息
	Notes          string          `json:"notes"`          // 备注
}

// DayMeal 行程中某一天的餐饮
type DayMeal struct {
	ID         string    `json:"id"`
	MealType   string    `json:"meal_type"`  // 餐饮类型（早餐、午餐、晚餐）
	Restaurant *POI      `json:"restaurant"` // 餐厅信息
	Time       time.Time `json:"time"`       // 就餐时间
	Cost       float64   `json:"cost"`       // 花费
	Notes      string    `json:"notes"`      // 备注
}

// DayHotel 行程中某一天的住宿
type DayHotel struct {
	ID           string    `json:"id"`
	HotelName    string    `json:"hotel_name"`
	Address      string    `json:"address"`
	CheckInTime  time.Time `json:"check_in_time"`
	CheckOutTime time.Time `json:"check_out_time"`
	Cost         float64   `json:"cost"`
	RoomType     string    `json:"room_type"`
	Notes        string    `json:"notes"`
}

// Transportation 交通信息
type Transportation struct {
	ID            string  `json:"id"`
	Type          string  `json:"type"`           // 交通方式（步行、地铁、公交、出租车等）
	StartLocation string  `json:"start_location"` // 起点
	EndLocation   string  `json:"end_location"`   // 终点
	Distance      float64 `json:"distance"`       // 距离（公里）
	Duration      int     `json:"duration"`       // 耗时（分钟）
	Cost          float64 `json:"cost"`           // 花费
}

// ItineraryBudget 行程预算
type ItineraryBudget struct {
	TotalBudget float64           `json:"total_budget"` // 总预算
	TotalCost   float64           `json:"total_cost"`   // 总花费
	Currency    string            `json:"currency"`     // 货币
	Categories  []*BudgetCategory `json:"categories"`   // 分类预算
	Remaining   float64           `json:"remaining"`    // 剩余预算
}

// BudgetCategory 预算分类
type BudgetCategory struct {
	Category   string  `json:"category"`   // 分类名称（住宿、餐饮、交通、景点门票等）
	Budget     float64 `json:"budget"`     // 预算
	Spent      float64 `json:"spent"`      // 已花费
	Percentage float64 `json:"percentage"` // 占比
}

// ItineraryStatistics 行程统计信息
type ItineraryStatistics struct {
	TotalDistance       float64 `json:"total_distance"`        // 总行程距离（公里）
	TotalWalkingTime    int     `json:"total_walking_time"`    // 总步行时间（分钟）
	TotalAttractionTime int     `json:"total_attraction_time"` // 总游览时间（分钟）
	AverageScore        float64 `json:"average_score"`         // 行程平均评分
	PlaceCount          int     `json:"place_count"`           // 访问地点总数
	RestDayCount        int     `json:"rest_day_count"`        // 休息天数
}

// DayStatistics 每天的统计信息
type DayStatistics struct {
	WalkingDistance float64 `json:"walking_distance"` // 该天行走距离（公里）
	WalkingTime     int     `json:"walking_time"`     // 该天步行时间（分钟）
	PlaceCount      int     `json:"place_count"`      // 该天访问地点数
	DailyCost       float64 `json:"daily_cost"`       // 该天支出
	AttractionTime  int     `json:"attraction_time"`  // 该天游览时间（分钟）
	ElevationGain   float64 `json:"elevation_gain"`   // 爬升高度（米）
	IsRest          bool    `json:"is_rest"`          // 是否为休息日
}

// NewItinerary 创建新行程
func NewItinerary(userID, name string, startDate, endDate time.Time) *Itinerary {
	dayCount := int(endDate.Sub(startDate).Hours()/24) + 1
	if dayCount < 1 {
		dayCount = 1
	}
	return &Itinerary{
		ID:        generateID(),
		UserID:    userID,
		Name:      name,
		Status:    ItineraryStatusDraft,
		StartDate: startDate,
		EndDate:   endDate,
		DayCount:  dayCount,
		Days:      make([]*ItineraryDay, 0),
		Budget: &ItineraryBudget{
			Currency:   "CNY",
			Categories: make([]*BudgetCategory, 0),
		},
		Statistics:  &ItineraryStatistics{},
		Constraints: make([]Constraint, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// AddDay 添加一天的行程
func (it *Itinerary) AddDay(day *ItineraryDay) {
	it.Days = append(it.Days, day)
}

// AddAttraction 向指定的日期添加景点
func (it *Itinerary) AddAttraction(dayNumber int, attraction *DayAttraction) error {
	if dayNumber > len(it.Days) || dayNumber < 1 {
		return ErrInvalidDayNumber
	}
	it.Days[dayNumber-1].Attractions = append(it.Days[dayNumber-1].Attractions, attraction)
	return nil
}

// CalculateTotalCost 计算总花费
func (it *Itinerary) CalculateTotalCost() float64 {
	total := 0.0
	for _, day := range it.Days {
		for _, attr := range day.Attractions {
			total += attr.Cost
		}
		for _, meal := range day.Meals {
			total += meal.Cost
		}
		if day.Hotel != nil {
			total += day.Hotel.Cost
		}
	}
	it.Budget.TotalCost = total
	it.Budget.Remaining = it.Budget.TotalBudget - total
	return total
}

// CalculateTotalDistance 计算总距离
func (it *Itinerary) CalculateTotalDistance() float64 {
	total := 0.0
	for _, day := range it.Days {
		for _, attr := range day.Attractions {
			if attr.Transportation != nil {
				total += attr.Transportation.Distance
			}
		}
	}
	it.Statistics.TotalDistance = total
	return total
}

// UpdateStatus 更新行程状态
func (it *Itinerary) UpdateStatus(status ItineraryStatus) {
	it.Status = status
	it.UpdatedAt = time.Now()
	if status == ItineraryStatusConfirmed {
		now := time.Now()
		it.PublishedAt = &now
	}
}

// Validate 验证行程数据的合法性
func (it *Itinerary) Validate() error {
	if it.Name == "" {
		return ErrInvalidItineraryName
	}
	if it.EndDate.Before(it.StartDate) {
		return ErrInvalidDateRange
	}
	if len(it.Days) == 0 {
		return ErrEmptyItinerary
	}
	return nil
}

// NewItineraryDay 创建新的一天的行程
func NewItineraryDay(dayNumber int, date time.Time) *ItineraryDay {
	return &ItineraryDay{
		DayNumber:   dayNumber,
		Date:        date,
		Attractions: make([]*DayAttraction, 0),
		Meals:       make([]*DayMeal, 0),
		Statistics:  &DayStatistics{},
	}
}
