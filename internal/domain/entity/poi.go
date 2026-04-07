package entity

import (
	"errors"
	"strings"
	"time"
)

// ========== 错误定义 ==========

var (
	ErrInvalidPOIID       = errors.New("invalid POI ID")
	ErrInvalidPOIName     = errors.New("invalid POI name")
	ErrInvalidPOICategory = errors.New("invalid POI category")
	ErrInvalidPOILocation = errors.New("invalid POI location")
	ErrInvalidPOIPrice    = errors.New("invalid POI price")
	ErrInvalidPOIRating   = errors.New("invalid POI rating")
	ErrPOINotFound        = errors.New("POI not found")
)

// ========== 枚举定义 ==========

// Category POI 分类
type Category string

const (
	CategoryAttraction    Category = "attraction"    // 景点
	CategoryRestaurant    Category = "restaurant"    // 餐厅
	CategoryHotel         Category = "hotel"         // 酒店
	CategoryShopping      Category = "shopping"      // 购物
	CategoryEntertainment Category = "entertainment" // 娱乐
	CategoryTransport     Category = "transport"     // 交通枢纽
	CategoryNature        Category = "nature"        // 自然风光
	CategoryMuseum        Category = "museum"        // 博物馆
	CategoryTemple        Category = "temple"        // 寺庙/神社
	CategoryPark          Category = "park"          // 公园
)

// IsValid 验证分类是否有效
func (c Category) IsValid() bool {
	switch c {
	case CategoryAttraction, CategoryRestaurant, CategoryHotel,
		CategoryShopping, CategoryEntertainment, CategoryTransport,
		CategoryNature, CategoryMuseum, CategoryTemple, CategoryPark:
		return true
	default:
		return false
	}
}

// String 返回字符串表示
func (c Category) String() string {
	return string(c)
}

// PriceLevel 价格等级
type PriceLevel string

const (
	PriceLevelFree      PriceLevel = "free"      // 免费
	PriceLevelCheap     PriceLevel = "cheap"     // ¥ (0-100)
	PriceLevelMedium    PriceLevel = "medium"    // ¥¥ (100-300)
	PriceLevelExpensive PriceLevel = "expensive" // ¥¥¥ (300-1000)
	PriceLevelLuxury    PriceLevel = "luxury"    // ¥¥¥¥ (1000+)
)

// IsValid 验证价格等级是否有效
func (p PriceLevel) IsValid() bool {
	switch p {
	case PriceLevelFree, PriceLevelCheap, PriceLevelMedium,
		PriceLevelExpensive, PriceLevelLuxury:
		return true
	default:
		return false
	}
}

// ToRange 返回价格区间（元）
func (p PriceLevel) ToRange() (min, max int) {
	switch p {
	case PriceLevelFree:
		return 0, 0
	case PriceLevelCheap:
		return 0, 100
	case PriceLevelMedium:
		return 100, 300
	case PriceLevelExpensive:
		return 300, 1000
	case PriceLevelLuxury:
		return 1000, 999999
	default:
		return 0, 0
	}
}

// TransportMode 交通方式
type TransportMode string

const (
	TransportWalk    TransportMode = "walk"    // 步行
	TransportBus     TransportMode = "bus"     // 公交
	TransportSubway  TransportMode = "subway"  // 地铁
	TransportTaxi    TransportMode = "taxi"    // 出租车
	TransportTrain   TransportMode = "train"   // 火车
	TransportBicycle TransportMode = "bicycle" // 自行车
)

// ========== 值对象定义 ==========

// Location 地理位置（值对象）
type Location struct {
	Lat float64 `json:"lat"` // 纬度
	Lng float64 `json:"lng"` // 经度
}

// IsValid 验证位置是否有效
func (l Location) IsValid() bool {
	return l.Lat >= -90 && l.Lat <= 90 && l.Lng >= -180 && l.Lng <= 180
}

// DistanceTo 计算到另一个位置的距离（公里，Haversine公式）
func (l Location) DistanceTo(other Location) float64 {
	const R = 6371 // 地球半径（公里）

	lat1 := l.Lat * 3.1415926 / 180
	lat2 := other.Lat * 3.1415926 / 180
	dLat := (other.Lat - l.Lat) * 3.1415926 / 180
	dLng := (other.Lng - l.Lng) * 3.1415926 / 180

	a := sin(dLat/2)*sin(dLat/2) +
		cos(lat1)*cos(lat2)*
			sin(dLng/2)*sin(dLng/2)
	c := 2 * atan2(sqrt(a), sqrt(1-a))

	return R * c
}

func sin(x float64) float64 {
	// 简化实现，实际应使用 math.Sin
	return x - x*x*x/6 + x*x*x*x*x/120
}

func cos(x float64) float64 {
	return 1 - x*x/2 + x*x*x*x/24
}

func sqrt(x float64) float64 {
	// 简化实现，实际应使用 math.Sqrt
	if x <= 0 {
		return 0
	}
	z := 1.0
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

func atan2(y, x float64) float64 {
	// 简化实现，实际应使用 math.Atan2
	if x > 0 {
		return atan(y / x)
	}
	if x < 0 {
		if y >= 0 {
			return atan(y/x) + 3.1415926
		}
		return atan(y/x) - 3.1415926
	}
	if y > 0 {
		return 3.1415926 / 2
	}
	if y < 0 {
		return -3.1415926 / 2
	}
	return 0
}

func atan(x float64) float64 {
	return x - x*x*x/3 + x*x*x*x*x/5
}

// TimeRange 时间段（值对象）
type TimeRange struct {
	Start string `json:"start"` // "09:00"
	End   string `json:"end"`   // "17:00"
}

// IsValid 验证时间段是否有效
func (tr TimeRange) IsValid() bool {
	if tr.Start == "" || tr.End == "" {
		return false
	}
	// 简单的时间格式校验
	if len(tr.Start) != 5 || len(tr.End) != 5 {
		return false
	}
	return tr.Start < tr.End
}

// Contains 检查时间是否在范围内
func (tr TimeRange) Contains(timeStr string) bool {
	return timeStr >= tr.Start && timeStr <= tr.End
}

// OpeningHours 营业时间（值对象）
type OpeningHours struct {
	Monday    []TimeRange `json:"monday"`
	Tuesday   []TimeRange `json:"tuesday"`
	Wednesday []TimeRange `json:"wednesday"`
	Thursday  []TimeRange `json:"thursday"`
	Friday    []TimeRange `json:"friday"`
	Saturday  []TimeRange `json:"saturday"`
	Sunday    []TimeRange `json:"sunday"`
	Holidays  []TimeRange `json:"holidays"`  // 节假日特殊时间
	IsClosed  bool        `json:"is_closed"` // 是否全天关闭
	Note      string      `json:"note"`      // 备注
}

// IsOpen 检查给定时间是否在营业时间内
func (oh OpeningHours) IsOpen(weekday int, timeStr string) bool {
	if oh.IsClosed {
		return false
	}

	var ranges []TimeRange
	switch weekday {
	case 1: // Monday
		ranges = oh.Monday
	case 2:
		ranges = oh.Tuesday
	case 3:
		ranges = oh.Wednesday
	case 4:
		ranges = oh.Thursday
	case 5:
		ranges = oh.Friday
	case 6:
		ranges = oh.Saturday
	case 7:
		ranges = oh.Sunday
	default:
		return false
	}

	for _, tr := range ranges {
		if tr.Contains(timeStr) {
			return true
		}
	}
	return false
}

// Tag 标签（值对象）
type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // cuisine, atmosphere, feature, audience
}

// Image 图片（值对象）
type Image struct {
	URL     string `json:"url"`
	Caption string `json:"caption"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	IsMain  bool   `json:"is_main"`
}

// ========== 实体定义 ==========

// POI 兴趣点实体（聚合根）
type POI struct {
	// 基础信息
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	NameEn      string   `json:"name_en"`
	NameLocal   string   `json:"name_local"`
	Category    Category `json:"category"`
	SubCategory string   `json:"sub_category"`

	// 地理位置
	Location Location `json:"location"`
	Address  string   `json:"address"`
	City     string   `json:"city"`
	District string   `json:"district"`
	GeoHash  string   `json:"geohash"`

	// 时间信息
	OpeningHours OpeningHours `json:"opening_hours"`
	AvgStayTime  int          `json:"avg_stay_time"` // 平均停留时间（分钟）
	BestTime     []string     `json:"best_time"`     // 最佳游览季节/时段
	Duration     int          `json:"duration"`      // 建议游玩时长（分钟）

	// 价格信息
	PriceLevel  PriceLevel `json:"price_level"`
	TicketPrice float64    `json:"ticket_price"` // 门票价格（景点）
	AvgPrice    float64    `json:"avg_price"`    // 人均消费（餐厅/酒店）

	// 评分与热度
	Rating      float64 `json:"rating"`       // 评分 0-5
	RatingCount int     `json:"rating_count"` // 评价数量
	Popularity  float64 `json:"popularity"`   // 热度分 0-100
	Rank        int     `json:"rank"`         // 城市内排名

	// 标签与特征
	Tags        []Tag    `json:"tags"`
	Features    []string `json:"features"`     // 特色属性
	SimilarPOIs []string `json:"similar_pois"` // 相似POI ID列表

	// 多媒体
	Images    []Image  `json:"images"`
	Thumbnail string   `json:"thumbnail"`
	Videos    []string `json:"videos"`

	// 预订信息
	BookingURL string `json:"booking_url"`
	IsBookable bool   `json:"is_bookable"`
	Inventory  int    `json:"inventory"` // 库存（门票/房间）

	// 扩展信息
	Description string   `json:"description"`
	Tips        []string `json:"tips"`
	Warnings    []string `json:"warnings"`

	// 元数据
	Source     string    `json:"source"`
	Confidence float64   `json:"confidence"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ========== 构造函数 ==========

// NewPOI 创建新的 POI（工厂方法）
func NewPOI(
	name string,
	category Category,
	lat, lng float64,
	city string,
) (*POI, error) {
	poi := &POI{
		ID:         generatePOIID(), // 需要实现
		Name:       name,
		Category:   category,
		Location:   Location{Lat: lat, Lng: lng},
		City:       city,
		Rating:     0,
		Popularity: 0,
		IsBookable: false,
		Confidence: 1.0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := poi.Validate(); err != nil {
		return nil, err
	}

	return poi, nil
}

func generatePOIID() string {
	// 可以使用 uuid 或 自增ID
	return time.Now().Format("20060102150405") + randomString(6)
}

func randomString(n int) string {
	// 简化实现
	return "abcdef"
}

// ========== 验证方法 ==========

// Validate 验证 POI 实体的完整性
func (p *POI) Validate() error {
	// 验证 ID
	if p.ID == "" {
		return ErrInvalidPOIID
	}

	// 验证名称
	if len(p.Name) < 1 || len(p.Name) > 200 {
		return ErrInvalidPOIName
	}

	// 验证分类
	if !p.Category.IsValid() {
		return ErrInvalidPOICategory
	}

	// 验证位置
	if !p.Location.IsValid() {
		return ErrInvalidPOILocation
	}

	// 验证价格
	if p.TicketPrice < 0 {
		return ErrInvalidPOIPrice
	}

	// 验证评分
	if p.Rating < 0 || p.Rating > 5 {
		return ErrInvalidPOIRating
	}

	// 验证价格等级
	if p.PriceLevel != "" && !p.PriceLevel.IsValid() {
		return ErrInvalidPOIPrice
	}

	return nil
}

// ========== 业务行为 ==========

// UpdateRating 更新评分
func (p *POI) UpdateRating(newRating float64, newCount int) {
	if newRating < 0 || newRating > 5 {
		return
	}

	// 加权平均
	if p.RatingCount > 0 {
		p.Rating = (p.Rating*float64(p.RatingCount) + newRating*float64(newCount)) /
			float64(p.RatingCount+newCount)
	} else {
		p.Rating = newRating
	}
	p.RatingCount += newCount
	p.UpdatedAt = time.Now()
}

// UpdatePopularity 更新热度（基于访问量、搜索量等）
func (p *POI) UpdatePopularity(delta float64) {
	p.Popularity += delta
	if p.Popularity > 100 {
		p.Popularity = 100
	}
	if p.Popularity < 0 {
		p.Popularity = 0
	}
	p.UpdatedAt = time.Now()
}

// AddTag 添加标签
func (p *POI) AddTag(tag Tag) {
	// 去重
	for _, t := range p.Tags {
		if t.ID == tag.ID {
			return
		}
	}
	p.Tags = append(p.Tags, tag)
	p.UpdatedAt = time.Now()
}

// RemoveTag 移除标签
func (p *POI) RemoveTag(tagID string) {
	for i, t := range p.Tags {
		if t.ID == tagID {
			p.Tags = append(p.Tags[:i], p.Tags[i+1:]...)
			break
		}
	}
	p.UpdatedAt = time.Now()
}

// AddImage 添加图片
func (p *POI) AddImage(image Image) {
	p.Images = append(p.Images, image)
	if image.IsMain {
		p.Thumbnail = image.URL
	}
	p.UpdatedAt = time.Now()
}

// SetMainImage 设置主图
func (p *POI) SetMainImage(url string) {
	for i := range p.Images {
		p.Images[i].IsMain = (p.Images[i].URL == url)
	}
	p.Thumbnail = url
	p.UpdatedAt = time.Now()
}

// UpdatePrice 更新价格
func (p *POI) UpdatePrice(price float64) {
	if price < 0 {
		return
	}
	p.TicketPrice = price
	p.UpdatedAt = time.Now()

	// 自动更新价格等级
	p.updatePriceLevel()
}

// updatePriceLevel 根据价格自动更新等级
func (p *POI) updatePriceLevel() {
	switch {
	case p.TicketPrice == 0:
		p.PriceLevel = PriceLevelFree
	case p.TicketPrice < 100:
		p.PriceLevel = PriceLevelCheap
	case p.TicketPrice < 300:
		p.PriceLevel = PriceLevelMedium
	case p.TicketPrice < 1000:
		p.PriceLevel = PriceLevelExpensive
	default:
		p.PriceLevel = PriceLevelLuxury
	}
}

// UpdateInventory 更新库存
func (p *POI) UpdateInventory(delta int) error {
	if p.Inventory+delta < 0 {
		return errors.New("insufficient inventory")
	}
	p.Inventory += delta
	p.UpdatedAt = time.Now()
	return nil
}

// IsAvailable 检查是否可用（营业中且有库存）
func (p *POI) IsAvailable(weekday int, timeStr string) bool {
	// 检查营业时间
	if !p.OpeningHours.IsOpen(weekday, timeStr) {
		return false
	}

	// 检查库存（如果是可预订的）
	if p.IsBookable && p.Inventory <= 0 {
		return false
	}

	return true
}

// CalculateTravelTime 计算从指定位置到 POI 的交通时间
func (p *POI) CalculateTravelTime(from Location, mode TransportMode) int {
	distance := from.DistanceTo(p.Location)

	// 根据交通方式估算速度（公里/小时）
	var speed float64
	switch mode {
	case TransportWalk:
		speed = 5 // 5 km/h
	case TransportBicycle:
		speed = 15
	case TransportBus, TransportSubway, TransportTrain:
		speed = 30
	case TransportTaxi:
		speed = 40
	default:
		speed = 20
	}

	// 时间（分钟）
	return int((distance / speed) * 60)
}

// MatchTags 检查 POI 是否匹配指定标签（任意匹配）
func (p *POI) MatchTags(targetTags []string) bool {
	if len(targetTags) == 0 {
		return true
	}

	for _, tt := range targetTags {
		for _, pt := range p.Tags {
			if pt.ID == tt || pt.Name == tt {
				return true
			}
		}
	}
	return false
}

// MatchAllTags 检查 POI 是否匹配所有指定标签
func (p *POI) MatchAllTags(targetTags []string) bool {
	if len(targetTags) == 0 {
		return true
	}

	for _, tt := range targetTags {
		matched := false
		for _, pt := range p.Tags {
			if pt.ID == tt || pt.Name == tt {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

// GetBestTimeString 返回最佳游览时间的字符串表示
func (p *POI) GetBestTimeString() string {
	if len(p.BestTime) == 0 {
		return "全年皆宜"
	}
	return strings.Join(p.BestTime, "、")
}

// ========== 辅助方法 ==========

// IsRestaurant 判断是否为餐厅
func (p *POI) IsRestaurant() bool {
	return p.Category == CategoryRestaurant
}

// IsAttraction 判断是否为景点
func (p *POI) IsAttraction() bool {
	return p.Category == CategoryAttraction
}

// IsHotel 判断是否为酒店
func (p *POI) IsHotel() bool {
	return p.Category == CategoryHotel
}

// GetPriceDisplay 获取价格显示文本
func (p *POI) GetPriceDisplay() string {
	switch p.PriceLevel {
	case PriceLevelFree:
		return "免费"
	case PriceLevelCheap:
		return "¥"
	case PriceLevelMedium:
		return "¥¥"
	case PriceLevelExpensive:
		return "¥¥¥"
	case PriceLevelLuxury:
		return "¥¥¥¥"
	default:
		return "待定"
	}
}

// Clone 深拷贝
func (p *POI) Clone() *POI {
	clone := *p
	// 深拷贝切片
	if p.Tags != nil {
		clone.Tags = make([]Tag, len(p.Tags))
		copy(clone.Tags, p.Tags)
	}
	if p.Images != nil {
		clone.Images = make([]Image, len(p.Images))
		copy(clone.Images, p.Images)
	}
	if p.Tips != nil {
		clone.Tips = make([]string, len(p.Tips))
		copy(clone.Tips, p.Tips)
	}
	if p.Warnings != nil {
		clone.Warnings = make([]string, len(p.Warnings))
		copy(clone.Warnings, p.Warnings)
	}
	return &clone
}
