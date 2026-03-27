package recommender

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"atlas-routex/internal/domain/entity"
	"atlas-routex/internal/domain/repository"
)

var (
	ErrInvalidRecommendInput = errors.New("recommender: invalid recommend input")
)

// RecommendMode 推荐取数策略（底层仍走仓储能力，便于与个性化策略并行演进）。
type RecommendMode string

const (
	// RecommendModeDefault 综合条件查询（FindByQuery）。
	RecommendModeDefault RecommendMode = ""
	// RecommendModePopular 热门榜。
	RecommendModePopular RecommendMode = "popular"
	// RecommendModeTopRated 高评分。
	RecommendModeTopRated RecommendMode = "top_rated"
	// RecommendModeSimilar 基于种子 POI 的相似扩展（SimilarPOIs + 同城同类回退）。
	RecommendModeSimilar RecommendMode = "similar"
)

// RecommendInput 推荐上下文：城市、筛选维度、模式与分页。
type RecommendInput struct {
	Mode    RecommendMode
	City    string
	Keyword string

	Category   *entity.Category
	Tags       []string
	TagsAll    []string
	MinRating  float64
	PriceLevel *entity.PriceLevel

	Lat      float64
	Lng      float64
	RadiusKM int

	// SeedPOIIDs 在 RecommendModeSimilar 下必填（至少一个）。
	SeedPOIIDs []string
	// ExcludePOIIDs 结果中剔除的 POI（例如已在行程中的点）。
	ExcludePOIIDs []string

	Limit     int
	SortBy    string
	SortOrder string
}

// Validate 入参校验。
func (in *RecommendInput) Validate() error {
	if in == nil {
		return ErrInvalidRecommendInput
	}
	if strings.TrimSpace(in.City) == "" {
		return ErrInvalidRecommendInput
	}
	if in.Limit < 0 {
		return ErrInvalidRecommendInput
	}
	if in.MinRating < 0 || in.MinRating > 5 {
		return ErrInvalidRecommendInput
	}
	if in.Mode == RecommendModeSimilar && len(in.SeedPOIIDs) == 0 {
		return ErrInvalidRecommendInput
	}
	if in.RadiusKM < 0 {
		return ErrInvalidRecommendInput
	}
	return nil
}

// RecommendUsecase 按上下文返回 POI 列表。
type RecommendUsecase struct {
	pois     repository.POIRepository
	strategy RecommendStrategy
}

// NewRecommendUsecase 构造用例（默认 RuleFilter + SortKey 排序）。
func NewRecommendUsecase(pois repository.POIRepository) *RecommendUsecase {
	return NewRecommendUsecaseWithStrategy(pois, nil)
}

// NewRecommendUsecaseWithStrategy 注入推荐策略；strategy 为 nil 时使用 NewDefaultRecommendStrategy。
func NewRecommendUsecaseWithStrategy(pois repository.POIRepository, strategy RecommendStrategy) *RecommendUsecase {
	s := strategy
	if s == nil {
		s = NewDefaultRecommendStrategy()
	}
	return &RecommendUsecase{pois: pois, strategy: s}
}

// Execute 校验 → 按模式取数 → 后处理（排除 ID、标签）→ 排序截断。
func (uc *RecommendUsecase) Execute(ctx context.Context, in *RecommendInput) ([]*entity.POI, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	if uc.pois == nil {
		return nil, fmt.Errorf("recommender: poi repository is nil")
	}
	strategy := uc.strategy
	if strategy == nil {
		strategy = NewDefaultRecommendStrategy()
	}

	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}

	var (
		list []*entity.POI
		err  error
	)
	switch in.Mode {
	case RecommendModePopular:
		list, err = uc.pois.GetPopularPOIs(ctx, in.City, in.Category, uc.oversampleLimit(in, limit))
	case RecommendModeTopRated:
		list, err = uc.pois.GetTopRatedPOIs(ctx, in.City, in.Category, uc.oversampleLimit(in, limit))
	case RecommendModeSimilar:
		list, err = uc.executeSimilar(ctx, in, limit)
	default:
		list, err = uc.executeQuery(ctx, in, limit)
	}
	if err != nil {
		return nil, err
	}

	list = strategy.Refine(list, in)
	return trimLimit(list, limit), nil
}

func (uc *RecommendUsecase) oversampleLimit(in *RecommendInput, limit int) int {
	if !needsOversample(in) {
		return limit
	}
	return capOversample(limit)
}

func (uc *RecommendUsecase) executeQuery(ctx context.Context, in *RecommendInput, limit int) ([]*entity.POI, error) {
	queryLimit := limit
	if needsOversample(in) {
		queryLimit = capOversample(limit)
	}

	qb := repository.NewPOIQueryBuilder().
		WithCity(in.City).
		WithPagination(1, queryLimit)

	if in.Category != nil && *in.Category != "" {
		qb = qb.WithCategory(*in.Category)
	}
	if in.Keyword != "" {
		qb = qb.WithKeyword(in.Keyword)
	}
	if in.MinRating > 0 {
		qb = qb.WithMinRating(in.MinRating)
	}
	if in.PriceLevel != nil {
		qb = qb.WithPriceLevel(*in.PriceLevel)
	}
	if len(in.Tags) > 0 {
		qb = qb.WithTags(in.Tags)
	}
	if in.RadiusKM > 0 {
		qb = qb.WithNearby(in.Lat, in.Lng, in.RadiusKM)
	}

	sortField, sortOrder := pickSortField(in), pickSortOrder(in)
	qb = qb.WithSort(sortField, sortOrder)

	q := qb.Build()
	if len(in.TagsAll) > 0 {
		q.TagsAll = append([]string(nil), in.TagsAll...)
	}

	return uc.pois.FindByQuery(ctx, q)
}

func (uc *RecommendUsecase) executeSimilar(ctx context.Context, in *RecommendInput, limit int) ([]*entity.POI, error) {
	seeds, err := uc.pois.FindByIDs(ctx, dedupeStrings(in.SeedPOIIDs))
	if err != nil {
		return nil, err
	}
	seedSet := make(map[string]struct{}, len(seeds))
	for _, s := range seeds {
		if s != nil && s.ID != "" {
			seedSet[s.ID] = struct{}{}
		}
	}

	var candidateIDs []string
	seen := make(map[string]struct{})
	for _, s := range seeds {
		if s == nil {
			continue
		}
		for _, id := range s.SimilarPOIs {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if _, isSeed := seedSet[id]; isSeed {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			candidateIDs = append(candidateIDs, id)
		}
	}

	const maxSimilarFetch = 200
	if len(candidateIDs) > maxSimilarFetch {
		candidateIDs = candidateIDs[:maxSimilarFetch]
	}

	var list []*entity.POI
	if len(candidateIDs) > 0 {
		list, err = uc.pois.FindByIDs(ctx, candidateIDs)
		if err != nil {
			return nil, err
		}
	}

	list = filterByCity(list, in.City)

	if len(list) < limit && len(seeds) > 0 && seeds[0] != nil {
		fallback, ferr := uc.fallbackSameCategory(ctx, in, seeds[0].Category, seedSet, uc.oversampleLimit(in, limit))
		if ferr != nil {
			return nil, ferr
		}
		list = mergeDedupe(list, fallback)
	}

	return list, nil
}

func (uc *RecommendUsecase) fallbackSameCategory(
	ctx context.Context,
	in *RecommendInput,
	cat entity.Category,
	seedSet map[string]struct{},
	fetchLimit int,
) ([]*entity.POI, error) {
	if !cat.IsValid() {
		return nil, nil
	}
	qb := repository.NewPOIQueryBuilder().
		WithCity(in.City).
		WithCategory(cat).
		WithPagination(1, fetchLimit).
		WithSort(pickSortField(in), pickSortOrder(in))
	q := qb.Build()
	if in.MinRating > 0 {
		q.MinRating = &in.MinRating
	}
	if len(in.Tags) > 0 {
		q.Tags = append([]string(nil), in.Tags...)
	}
	if len(in.TagsAll) > 0 {
		q.TagsAll = append([]string(nil), in.TagsAll...)
	}

	out, err := uc.pois.FindByQuery(ctx, q)
	if err != nil {
		return nil, err
	}
	filtered := out[:0]
	for _, p := range out {
		if p == nil {
			continue
		}
		if _, isSeed := seedSet[p.ID]; isSeed {
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered, nil
}

func needsOversample(in *RecommendInput) bool {
	return len(in.ExcludePOIIDs) > 0 ||
		len(in.Tags) > 0 ||
		len(in.TagsAll) > 0 ||
		in.Mode == RecommendModePopular ||
		in.Mode == RecommendModeTopRated
}

func capOversample(limit int) int {
	n := limit * 8
	if n < limit+40 {
		n = limit + 40
	}
	if n > 500 {
		n = 500
	}
	return n
}

func filterByCity(list []*entity.POI, city string) []*entity.POI {
	city = strings.TrimSpace(city)
	if city == "" {
		return list
	}
	out := list[:0]
	for _, p := range list {
		if p == nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(p.City), city) {
			out = append(out, p)
		}
	}
	return out
}

func mergeDedupe(a, b []*entity.POI) []*entity.POI {
	seen := make(map[string]struct{}, len(a)+len(b))
	out := make([]*entity.POI, 0, len(a)+len(b))
	for _, p := range a {
		if p == nil || p.ID == "" {
			continue
		}
		if _, ok := seen[p.ID]; ok {
			continue
		}
		seen[p.ID] = struct{}{}
		out = append(out, p)
	}
	for _, p := range b {
		if p == nil || p.ID == "" {
			continue
		}
		if _, ok := seen[p.ID]; ok {
			continue
		}
		seen[p.ID] = struct{}{}
		out = append(out, p)
	}
	return out
}

func trimLimit(list []*entity.POI, limit int) []*entity.POI {
	if limit <= 0 || len(list) <= limit {
		return list
	}
	return list[:limit]
}

func dedupeStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
