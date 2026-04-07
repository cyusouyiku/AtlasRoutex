package repository

import (
	"context"

	"atlas-routex/internal/domain/entity"
)

// POIRepository POI仓储接口
type POIRepository interface {
	// ========== 基础 CRUD ==========

	// FindByID 根据ID查询POI
	FindByID(ctx context.Context, id string) (*entity.POI, error)

	// FindByIDs 批量查询POI
	FindByIDs(ctx context.Context, ids []string) ([]*entity.POI, error)

	// Save 保存POI（新增或更新）
	Save(ctx context.Context, poi *entity.POI) error

	// Update 更新POI
	Update(ctx context.Context, poi *entity.POI) error

	// Delete 软删除POI
	Delete(ctx context.Context, id string) error

	// ========== 分类查询（统一方法，用参数区分） ==========

	// FindByCategory 根据分类查询POI列表
	FindByCategory(ctx context.Context, category entity.Category, city string, limit, offset int) ([]*entity.POI, error)

	// FindByCity 根据城市查询POI
	FindByCity(ctx context.Context, city string, limit, offset int) ([]*entity.POI, error)

	// FindByTags 根据标签查询POI
	FindByTags(ctx context.Context, tags []string, city string, limit int) ([]*entity.POI, error)

	// FindNearby 查询附近的POI（基于地理位置）
	FindNearby(ctx context.Context, lat, lng float64, radius int, category *entity.Category, limit int) ([]*entity.POI, error)

	// ========== 高级查询 ==========

	// FindByQuery 复杂条件查询
	FindByQuery(ctx context.Context, query *POIQuery) ([]*entity.POI, error)

	// ========== 热门推荐 ==========

	// GetPopularPOIs 获取热门POI
	GetPopularPOIs(ctx context.Context, city string, category *entity.Category, limit int) ([]*entity.POI, error)

	// GetTopRatedPOIs 获取高分POI
	GetTopRatedPOIs(ctx context.Context, city string, category *entity.Category, limit int) ([]*entity.POI, error)

	// ========== 距离矩阵 ==========

	// GetDistanceMatrix 获取POI之间的距离矩阵
	GetDistanceMatrix(ctx context.Context, poiIDs []string) (*DistanceMatrix, error)

	// ========== 统计 ==========

	// CountByCity 统计城市POI数量
	CountByCity(ctx context.Context, city string) (int, error)

	// CountByCategory 统计分类POI数量
	CountByCategory(ctx context.Context, category entity.Category, city string) (int, error)

	// ========== 存在性检查 ==========

	// ExistsByID 检查POI是否存在
	ExistsByID(ctx context.Context, id string) (bool, error)
}

// ========== 查询条件定义 ==========

// POIQuery POI查询条件
type POIQuery struct {
	// 基础筛选
	IDs         []string         `json:"ids"`          // POI ID列表
	City        *string          `json:"city"`         // 城市
	Category    *entity.Category `json:"category"`     // 分类
	SubCategory *string          `json:"sub_category"` // 子分类

	// 模糊搜索
	Keyword string `json:"keyword"` // 关键词（搜索名称、地址）

	// 范围筛选
	MinRating  *float64           `json:"min_rating"`  // 最低评分
	MaxPrice   *float64           `json:"max_price"`   // 最高价格
	PriceLevel *entity.PriceLevel `json:"price_level"` // 价格等级

	// 标签筛选
	Tags    []string `json:"tags"`     // 标签（任意匹配）
	TagsAll []string `json:"tags_all"` // 标签（全部匹配）

	// 地理位置筛选
	Nearby *GeoQuery `json:"nearby"` // 附近查询

	// 分页排序
	Limit     int    `json:"limit"`      // 返回数量（默认 20）
	Offset    int    `json:"offset"`     // 分页偏移
	SortBy    string `json:"sort_by"`    // 排序字段：rating, popularity, price
	SortOrder string `json:"sort_order"` // asc / desc（默认 desc）
}

// GeoQuery 地理位置查询条件
type GeoQuery struct {
	Lat    float64 `json:"lat"`    // 纬度
	Lng    float64 `json:"lng"`    // 经度
	Radius int     `json:"radius"` // 半径，单位：公里
}

// DistanceMatrix 距离矩阵
type DistanceMatrix struct {
	// Matrix[from][to] = 距离（公里）
	Matrix map[string]map[string]float64 `json:"matrix"`
	// Duration[from][to] = 时间（分钟）
	Duration map[string]map[string]int `json:"duration"`
}

// ========== 分页结果 ==========

// POIPage POI分页结果
type POIPage struct {
	POIs       []*entity.POI `json:"pois"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

// ========== 查询构建器（可选） ==========

// POIQueryBuilder 查询条件构建器
type POIQueryBuilder struct {
	query *POIQuery
}

// NewPOIQueryBuilder 创建查询构建器
func NewPOIQueryBuilder() *POIQueryBuilder {
	return &POIQueryBuilder{
		query: &POIQuery{
			Limit:     20,
			SortBy:    "rating",
			SortOrder: "desc",
		},
	}
}

// WithCity 按城市筛选
func (b *POIQueryBuilder) WithCity(city string) *POIQueryBuilder {
	b.query.City = &city
	return b
}

// WithCategory 按分类筛选
func (b *POIQueryBuilder) WithCategory(category entity.Category) *POIQueryBuilder {
	b.query.Category = &category
	return b
}

// WithKeyword 关键词搜索
func (b *POIQueryBuilder) WithKeyword(keyword string) *POIQueryBuilder {
	b.query.Keyword = keyword
	return b
}

// WithMinRating 最低评分
func (b *POIQueryBuilder) WithMinRating(rating float64) *POIQueryBuilder {
	b.query.MinRating = &rating
	return b
}

// WithPriceLevel 价格等级
func (b *POIQueryBuilder) WithPriceLevel(level entity.PriceLevel) *POIQueryBuilder {
	b.query.PriceLevel = &level
	return b
}

// WithTags 按标签筛选（任意匹配）
func (b *POIQueryBuilder) WithTags(tags []string) *POIQueryBuilder {
	b.query.Tags = tags
	return b
}

// WithNearby 附近查询
func (b *POIQueryBuilder) WithNearby(lat, lng float64, radius int) *POIQueryBuilder {
	b.query.Nearby = &GeoQuery{
		Lat:    lat,
		Lng:    lng,
		Radius: radius,
	}
	return b
}

// WithPagination 分页
func (b *POIQueryBuilder) WithPagination(page, size int) *POIQueryBuilder {
	b.query.Offset = (page - 1) * size
	b.query.Limit = size
	return b
}

// WithSort 排序
func (b *POIQueryBuilder) WithSort(field, order string) *POIQueryBuilder {
	b.query.SortBy = field
	b.query.SortOrder = order
	return b
}

// Build 构建查询条件
func (b *POIQueryBuilder) Build() *POIQuery {
	return b.query
}
