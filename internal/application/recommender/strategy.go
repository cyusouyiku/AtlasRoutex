package recommender

import (
	"sort"
	"strings"

	"atlas-routex/internal/domain/entity"
)

// -----------------------------------------------------------------------------
// 端口：无状态或可注入，便于单测与后续替换为学习式排序。
// -----------------------------------------------------------------------------

// RecommendStrategy 对仓储侧拉回的一批候选做规则精炼与排序。
type RecommendStrategy interface {
	Refine(candidates []*entity.POI, in *RecommendInput) []*entity.POI
}

// POIFilter 规则过滤（返回新切片，不修改入参切片头以外的共享状态）。
type POIFilter interface {
	Filter(candidates []*entity.POI, in *RecommendInput) []*entity.POI
}

// POIScorer 为 POI 计算排序分，分数越大越优先（与 SortOrder 配合）。
type POIScorer interface {
	Score(p *entity.POI, in *RecommendInput) float64
}

// -----------------------------------------------------------------------------
// PipelineStrategy：过滤器链 + 单一打分器。
// -----------------------------------------------------------------------------

// PipelineStrategy 组合多段过滤与最终打分排序。
type PipelineStrategy struct {
	Filters []POIFilter
	Scorer  POIScorer
}

// NewPipelineStrategy 构造流水线；filters 可为 nil（跳过过滤）。
func NewPipelineStrategy(filters []POIFilter, scorer POIScorer) *PipelineStrategy {
	return &PipelineStrategy{Filters: filters, Scorer: scorer}
}

// NewDefaultRecommendStrategy 默认：排除/标签 + 与查询一致的 SortKey 打分。
func NewDefaultRecommendStrategy() RecommendStrategy {
	return NewPipelineStrategy([]POIFilter{
		RuleFilter{},
	}, SortKeyScorer{})
}

// NewWeightedRecommendStrategy 加权融合（评分、热度、标签重合、锚点距离），仍先走 RuleFilter。
func NewWeightedRecommendStrategy(w WeightedScoreWeights) RecommendStrategy {
	return NewPipelineStrategy([]POIFilter{
		RuleFilter{},
	}, WeightedScorer{Weights: w})
}

// Refine 依次过滤，再按 Scorer 与 in.SortOrder 排序。
func (p *PipelineStrategy) Refine(candidates []*entity.POI, in *RecommendInput) []*entity.POI {
	list := candidates
	for _, f := range p.Filters {
		if f == nil {
			continue
		}
		list = f.Filter(list, in)
	}
	if p.Scorer == nil || len(list) < 2 {
		return list
	}
	sortByScorer(list, in, p.Scorer)
	return list
}

// -----------------------------------------------------------------------------
// RuleFilter：排除 ID、标签、最低分（补 popular/top_rated 等路径）。
// -----------------------------------------------------------------------------

// RuleFilter 无状态规则过滤器。
type RuleFilter struct{}

// Filter 应用 ExcludePOIIDs、Tags、TagsAll、MinRating。
func (RuleFilter) Filter(list []*entity.POI, in *RecommendInput) []*entity.POI {
	if len(list) == 0 {
		return list
	}
	exclude := make(map[string]struct{}, len(in.ExcludePOIIDs))
	for _, id := range in.ExcludePOIIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			exclude[id] = struct{}{}
		}
	}

	out := make([]*entity.POI, 0, len(list))
	for _, p := range list {
		if p == nil {
			continue
		}
		if _, skip := exclude[p.ID]; skip {
			continue
		}
		if in.MinRating > 0 && p.Rating < in.MinRating {
			continue
		}
		if len(in.Tags) > 0 && !p.MatchTags(in.Tags) {
			continue
		}
		if len(in.TagsAll) > 0 && !p.MatchAllTags(in.TagsAll) {
			continue
		}
		out = append(out, p)
	}
	return out
}

// -----------------------------------------------------------------------------
// SortKeyScorer：与仓储 Sort 字段语义对齐的单维打分。
// -----------------------------------------------------------------------------

// SortKeyScorer 根据 in.SortBy / in.Mode 选择排序键（默认 rating desc）。
type SortKeyScorer struct{}

func (SortKeyScorer) Score(p *entity.POI, in *RecommendInput) float64 {
	return scoreForSortKey(p, pickSortField(in))
}

// -----------------------------------------------------------------------------
// WeightedScorer：多信号线性融合（便于后续接入用户画像权重）。
// -----------------------------------------------------------------------------

// WeightedScoreWeights 各维度权重，未设置的项视为 0。
type WeightedScoreWeights struct {
	Rating     float64
	Popularity float64
	Proximity  float64
	TagOverlap float64
}

// WeightedScorer 使用 Weight 对若干信号加权求和。
type WeightedScorer struct {
	Weights WeightedScoreWeights
}

func (w WeightedScorer) Score(p *entity.POI, in *RecommendInput) float64 {
	if p == nil {
		return 0
	}
	var s float64
	g := w.Weights
	if g.Rating != 0 {
		s += g.Rating * (p.Rating / 5.0)
	}
	if g.Popularity != 0 {
		s += g.Popularity * (p.Popularity / 100.0)
	}
	if g.Proximity != 0 && anchorValid(in) && p.Location.IsValid() {
		d := p.Location.DistanceTo(entity.Location{Lat: in.Lat, Lng: in.Lng})
		s += g.Proximity * (1.0 / (1.0 + d))
	}
	if g.TagOverlap != 0 && len(in.Tags) > 0 {
		s += g.TagOverlap * tagOverlapFraction(p, in.Tags)
	}
	return s
}

func anchorValid(in *RecommendInput) bool {
	return in != nil && entity.Location{Lat: in.Lat, Lng: in.Lng}.IsValid() &&
		!(in.Lat == 0 && in.Lng == 0)
}

func tagOverlapFraction(p *entity.POI, want []string) float64 {
	if p == nil || len(want) == 0 {
		return 0
	}
	match := 0
	for _, tt := range want {
		tt = strings.TrimSpace(tt)
		if tt == "" {
			continue
		}
		for _, pt := range p.Tags {
			if pt.ID == tt || pt.Name == tt {
				match++
				break
			}
		}
	}
	return float64(match) / float64(len(want))
}

// -----------------------------------------------------------------------------
// 与 usecase 共用：查询键、排序方向、分数字段。
// -----------------------------------------------------------------------------

func pickSortField(in *RecommendInput) string {
	if in == nil {
		return "rating"
	}
	if strings.TrimSpace(in.SortBy) != "" {
		return strings.TrimSpace(in.SortBy)
	}
	switch in.Mode {
	case RecommendModePopular:
		return "popularity"
	case RecommendModeTopRated:
		return "rating"
	default:
		return "rating"
	}
}

func pickSortOrder(in *RecommendInput) string {
	if in == nil {
		return "desc"
	}
	o := strings.ToLower(strings.TrimSpace(in.SortOrder))
	if o == "asc" {
		return "asc"
	}
	return "desc"
}

func scoreForSortKey(p *entity.POI, field string) float64 {
	if p == nil {
		return 0
	}
	field = strings.ToLower(strings.TrimSpace(field))
	switch field {
	case "popularity":
		return p.Popularity
	case "price":
		return p.TicketPrice + p.AvgPrice
	case "rank":
		if p.Rank <= 0 {
			return 1e9
		}
		return float64(p.Rank)
	default:
		return p.Rating
	}
}

func sortByScorer(list []*entity.POI, in *RecommendInput, scorer POIScorer) {
	asc := pickSortOrder(in) == "asc"
	sort.SliceStable(list, func(i, j int) bool {
		ai, aj := scorer.Score(list[i], in), scorer.Score(list[j], in)
		if ai == aj {
			idi, idj := "", ""
			if list[i] != nil {
				idi = list[i].ID
			}
			if list[j] != nil {
				idj = list[j].ID
			}
			return idi < idj
		}
		if asc {
			return ai < aj
		}
		return ai > aj
	})
}
