package repository

import (
	"context"
	"time"
	
	"atlas-routex/internal/domain/entity"
)

// ItineraryRepository 行程仓储接口
// 定义领域层需要的行程数据访问能力
type ItineraryRepository interface {
	// ========== 基础 CRUD ==========
	
	// FindByID 根据 ID 查询行程
	FindByID(ctx context.Context, id string) (*entity.Itinerary, error)
	
	// FindByUserID 查询用户的所有行程
	FindByUserID(ctx context.Context, userID string) ([]*entity.Itinerary, error)
	
	// FindByName 根据行程名称模糊搜索
	FindByName(ctx context.Context, name string) ([]*entity.Itinerary, error)
	
	// Save 保存行程（新增）
	Save(ctx context.Context, itinerary *entity.Itinerary) error
	
	// Update 更新行程
	Update(ctx context.Context, itinerary *entity.Itinerary) error
	
	// Delete 软删除行程
	Delete(ctx context.Context, id string) error
	
	// HardDelete 物理删除行程（慎用）
	HardDelete(ctx context.Context, id string) error
	
	// ========== 业务查询 ==========
	
	// FindByStatus 根据状态查询行程
	FindByStatus(ctx context.Context, status entity.ItineraryStatus) ([]*entity.Itinerary, error)
	
	// FindByUserIDAndStatus 查询用户的特定状态行程
	FindByUserIDAndStatus(ctx context.Context, userID string, status entity.ItineraryStatus) ([]*entity.Itinerary, error)
	
	// FindByDateRange 查询指定日期范围内的行程
	FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.Itinerary, error)
	
	// FindByUserIDAndDateRange 查询用户在指定日期范围内的行程
	FindByUserIDAndDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]*entity.Itinerary, error)
	
	// FindByBudgetRange 查询指定预算范围内的行程
	FindByBudgetRange(ctx context.Context, minBudget, maxBudget float64) ([]*entity.Itinerary, error)
	
	// FindByCity 查询指定城市的行程
	FindByCity(ctx context.Context, city string, limit int) ([]*entity.Itinerary, error)
	
	// FindByAttraction 查询包含指定景点的行程
	FindByAttraction(ctx context.Context, attractionID string) ([]*entity.Itinerary, error)
	
	// FindRecentItineraries 查询最近创建的行程（按创建时间倒序）
	FindRecentItineraries(ctx context.Context, limit int) ([]*entity.Itinerary, error)
	
	// FindPopularItineraries 查询热门行程（按收藏数/分享数排序）
	FindPopularItineraries(ctx context.Context, limit int) ([]*entity.Itinerary, error)
	
	// FindByConstraints 根据约束条件查询行程（高级查询）
	FindByConstraints(ctx context.Context, constraints *ItineraryQuery) ([]*entity.Itinerary, error)
	
	// ========== 统计 ==========
	
	// CountByUserID 统计用户的行程数量
	CountByUserID(ctx context.Context, userID string) (int, error)
	
	// CountByCity 统计指定城市的行程数量
	CountByCity(ctx context.Context, city string) (int, error)
	
	// CountByStatus 统计指定状态的行程数量
	CountByStatus(ctx context.Context, status entity.ItineraryStatus) (int, error)
	
	// CountByDateRange 统计指定日期范围内的行程数量
	CountByDateRange(ctx context.Context, startDate, endDate time.Time) (int, error)
	
	// ========== 存在性检查 ==========
	
	// ExistsByUserIDAndName 检查用户是否已有同名行程
	ExistsByUserIDAndName(ctx context.Context, userID, name string) (bool, error)
	
	// ========== 批量操作 ==========
	
	// BatchUpdateStatus 批量更新行程状态
	BatchUpdateStatus(ctx context.Context, ids []string, status entity.ItineraryStatus) error
	
	// BatchDelete 批量软删除行程
	BatchDelete(ctx context.Context, ids []string) error
}

// ========== 查询条件定义 ==========

// ItineraryStatus 状态类型别名（方便使用）
type ItineraryStatus = entity.ItineraryStatus

// ItineraryQuery 行程查询条件
type ItineraryQuery struct {
	// 基础筛选
	IDs        []string               `json:"ids"`         // 行程ID列表
	UserID     *string                `json:"user_id"`     // 用户ID（精确匹配）
	City       *string                `json:"city"`        // 城市（精确匹配）
	Status     *entity.ItineraryStatus `json:"status"`     // 状态（精确匹配）
	
	// 模糊搜索
	Keyword    string                 `json:"keyword"`     // 关键词（搜索行程名称、城市）
	
	// 范围筛选
	MinBudget  *float64               `json:"min_budget"`  // 最小预算
	MaxBudget  *float64               `json:"max_budget"`  // 最大预算
	MinDays    *int                   `json:"min_days"`    // 最少天数
	MaxDays    *int                   `json:"max_days"`    // 最多天数
	MinWalking *int                   `json:"min_walking"` // 最小步行距离
	MaxWalking *int                   `json:"max_walking"` // 最大步行距离
	
	// 时间筛选
	CreatedAfter  *time.Time          `json:"created_after"`  // 创建时间之后
	CreatedBefore *time.Time          `json:"created_before"` // 创建时间之前
	StartDateAfter  *time.Time        `json:"start_date_after"`  // 行程开始日期之后
	StartDateBefore *time.Time        `json:"start_date_before"` // 行程开始日期之前
	
	// 内容筛选
	ContainsAttraction string         `json:"contains_attraction"` // 包含的景点ID
	ContainsTag        string         `json:"contains_tag"`        // 包含的标签
	
	// 分页排序
	Limit      int                    `json:"limit"`        // 返回数量（默认 20）
	Offset     int                    `json:"offset"`       // 分页偏移
	SortBy     string                 `json:"sort_by"`      // 排序字段：created_at, total_cost, total_days
	SortOrder  string                 `json:"sort_order"`   // asc / desc（默认 desc）
}

// ItineraryQueryBuilder 查询条件构建器
type ItineraryQueryBuilder struct {
	query *ItineraryQuery
}

// NewItineraryQueryBuilder 创建查询构建器
func NewItineraryQueryBuilder() *ItineraryQueryBuilder {
	return &ItineraryQueryBuilder{
		query: &ItineraryQuery{
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
		},
	}
}

// WithUserID 按用户ID筛选
func (b *ItineraryQueryBuilder) WithUserID(userID string) *ItineraryQueryBuilder {
	b.query.UserID = &userID
	return b
}

// WithCity 按城市筛选
func (b *ItineraryQueryBuilder) WithCity(city string) *ItineraryQueryBuilder {
	b.query.City = &city
	return b
}

// WithStatus 按状态筛选
func (b *ItineraryQueryBuilder) WithStatus(status entity.ItineraryStatus) *ItineraryQueryBuilder {
	b.query.Status = &status
	return b
}

// WithKeyword 关键词搜索
func (b *ItineraryQueryBuilder) WithKeyword(keyword string) *ItineraryQueryBuilder {
	b.query.Keyword = keyword
	return b
}

// WithBudgetRange 预算范围
func (b *ItineraryQueryBuilder) WithBudgetRange(min, max float64) *ItineraryQueryBuilder {
	b.query.MinBudget = &min
	b.query.MaxBudget = &max
	return b
}

// WithDaysRange 天数范围
func (b *ItineraryQueryBuilder) WithDaysRange(min, max int) *ItineraryQueryBuilder {
	b.query.MinDays = &min
	b.query.MaxDays = &max
	return b
}

// WithWalkingRange 步行距离范围
func (b *ItineraryQueryBuilder) WithWalkingRange(min, max int) *ItineraryQueryBuilder {
	b.query.MinWalking = &min
	b.query.MaxWalking = &max
	return b
}

// WithCreatedAfter 创建时间之后
func (b *ItineraryQueryBuilder) WithCreatedAfter(t time.Time) *ItineraryQueryBuilder {
	b.query.CreatedAfter = &t
	return b
}

// WithStartDateRange 行程开始日期范围
func (b *ItineraryQueryBuilder) WithStartDateRange(start, end time.Time) *ItineraryQueryBuilder {
	b.query.StartDateAfter = &start
	b.query.StartDateBefore = &end
	return b
}

// WithContainsAttraction 包含指定景点
func (b *ItineraryQueryBuilder) WithContainsAttraction(attractionID string) *ItineraryQueryBuilder {
	b.query.ContainsAttraction = attractionID
	return b
}

// WithPagination 分页
func (b *ItineraryQueryBuilder) WithPagination(page, size int) *ItineraryQueryBuilder {
	b.query.Offset = (page - 1) * size
	b.query.Limit = size
	return b
}

// WithSort 排序
func (b *ItineraryQueryBuilder) WithSort(field, order string) *ItineraryQueryBuilder {
	b.query.SortBy = field
	b.query.SortOrder = order
	return b
}

// Build 构建查询条件
func (b *ItineraryQueryBuilder) Build() *ItineraryQuery {
	return b.query
}

// ========== 分页结果 ==========

// ItineraryPage 行程分页结果
type ItineraryPage struct {
	Itineraries []*entity.Itinerary `json:"itineraries"`
	Total       int64               `json:"total"`
	Page        int                 `json:"page"`
	PageSize    int                 `json:"page_size"`
	TotalPages  int                 `json:"total_pages"`
}
