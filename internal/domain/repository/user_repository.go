package repository

import (
	"context"
	"errors"
	"atlas-routex/internal/domain/entity"
)

type UserRepository interface {
	//基础CRUD

	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByIDs(ctx context.Context, ids []string) ([]*entity.User, error)
	Save(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error

	//业务查询
	// FindByRole 根据角色查询用户列表
	FindByRole(ctx context.Context, role entity.UserRole, limit, offset int) ([]*entity.User, error)
	
	// FindByStatus 根据状态查询用户列表
	FindByStatus(ctx context.Context, status entity.UserStatus, limit, offset int) ([]*entity.User, error)
	
	// FindActiveUsers 查询活跃用户（最近7天有登录）
	FindActiveUsers(ctx context.Context, limit int) ([]*entity.User, error)
	
	// FindByPreferences 根据偏好查询用户（用于推荐相似用户）
	FindByPreferences(ctx context.Context, preferences entity.UserPreferences, limit int) ([]*entity.User, error)

	
	// CountByRole 统计指定角色的用户数量
	CountByRole(ctx context.Context, role entity.UserRole) (int, error)
	
	// CountByStatus 统计指定状态的用户数量
	CountByStatus(ctx context.Context, status entity.UserStatus) (int, error)
	
	// CountActiveLastDays 统计最近 N 天活跃的用户数
	CountActiveLastDays(ctx context.Context, days int) (int, error)
	
	
	// ExistsByEmail 检查邮箱是否已被使用
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	
	// ExistsByPhone 检查手机号是否已被使用
	ExistsByPhone(ctx context.Context, phone string) (bool, error)
	
	
	// BatchUpdateStatus 批量更新用户状态
	BatchUpdateStatus(ctx context.Context, ids []string, status entity.UserStatus) error
	
	// BatchUpdateRole 批量更新用户角色
	BatchUpdateRole(ctx context.Context, ids []string, role entity.UserRole) error
}


//查询条件定义
type UserQuery struct {
	// 基础筛选
	IDs      []string           `json:"ids"`       // 用户ID列表
	Role     *entity.UserRole   `json:"role"`      // 角色（精确匹配）
	Status   *entity.UserStatus `json:"status"`    // 状态（精确匹配）
	
	// 模糊搜索
	Keyword  string             `json:"keyword"`   // 关键词（搜索姓名、邮箱）
	
	// 范围筛选
	MinAge   int                `json:"min_age"`   // 最小年龄
	MaxAge   int                `json:"max_age"`   // 最大年龄
	MinItineraryCount int       `json:"min_itinerary_count"` // 最小行程数
	MinTotalDistance  int       `json:"min_total_distance"`  // 最小旅行距离
	
	// 偏好筛选
	PreferredCategories []string `json:"preferred_categories"` // 偏好的景点类型
	
	// 时间筛选
	CreatedAfter  *string `json:"created_after"`   // 注册之后
	CreatedBefore *string `json:"created_before"`  // 注册之前
	LastLoginAfter *string `json:"last_login_after"` // 最后登录之后
	
	// 分页排序
	Limit      int    `json:"limit"`       // 返回数量（默认 20）
	Offset     int    `json:"offset"`      // 分页偏移
	SortBy     string `json:"sort_by"`     // 排序字段：created_at, last_login_at, itinerary_count
	SortOrder  string `json:"sort_order"`  // asc / desc（默认 desc）
}

// UserQueryBuilder 查询条件构建器（可选，用于灵活构建复杂查询）
type UserQueryBuilder struct {
	query *UserQuery
}

// NewUserQueryBuilder 创建查询构建器
func NewUserQueryBuilder() *UserQueryBuilder {
	return &UserQueryBuilder{
		query: &UserQuery{
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
		},
	}
}

// WithRole 按角色筛选
func (b *UserQueryBuilder) WithRole(role entity.UserRole) *UserQueryBuilder {
	b.query.Role = &role
	return b
}

// WithStatus 按状态筛选
func (b *UserQueryBuilder) WithStatus(status entity.UserStatus) *UserQueryBuilder {
	b.query.Status = &status
	return b
}

// WithKeyword 关键词搜索
func (b *UserQueryBuilder) WithKeyword(keyword string) *UserQueryBuilder {
	b.query.Keyword = keyword
	return b
}

// WithAgeRange 年龄范围
func (b *UserQueryBuilder) WithAgeRange(min, max int) *UserQueryBuilder {
	b.query.MinAge = min
	b.query.MaxAge = max
	return b
}

// WithMinItineraryCount 最小行程数
func (b *UserQueryBuilder) WithMinItineraryCount(count int) *UserQueryBuilder {
	b.query.MinItineraryCount = count
	return b
}

// WithPreferredCategories 按偏好分类筛选
func (b *UserQueryBuilder) WithPreferredCategories(categories []string) *UserQueryBuilder {
	b.query.PreferredCategories = categories
	return b
}

// WithPagination 分页
func (b *UserQueryBuilder) WithPagination(page, size int) *UserQueryBuilder {
	b.query.Offset = (page - 1) * size
	b.query.Limit = size
	return b
}

// WithSort 排序
func (b *UserQueryBuilder) WithSort(field, order string) *UserQueryBuilder {
	b.query.SortBy = field
	b.query.SortOrder = order
	return b
}

// Build 构建查询条件
func (b *UserQueryBuilder) Build() *UserQuery {
	return b.query
}


// UserPage 用户分页结果
type UserPage struct {
	Users      []*entity.User `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}
