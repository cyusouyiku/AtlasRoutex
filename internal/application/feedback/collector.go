package feedback

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	rawMaxCommentRunes = maxCommentRunes
)

// RawFeedbackInput 自 API/消息层进入的原始反馈（collector 负责洗成 FeedbackItem）。
type RawFeedbackInput struct {
	IdempotencyKey string
	UserID         string
	ItineraryID    string
	POIID          string
	Kind           FeedbackKind
	Rating         *float64
	Comment        string
	OccurredAt     time.Time
	Extra          map[string]string
}

// Validate 收集前单条校验（与 FeedbackItem 规则对齐，RepeatCount 由聚合写入）。
func (r *RawFeedbackInput) Validate() error {
	if r == nil {
		return ErrInvalidFeedback
	}
	if !r.Kind.IsValid() {
		return ErrInvalidFeedback
	}
	switch r.Kind {
	case FeedbackKindItineraryRating:
		if strings.TrimSpace(r.ItineraryID) == "" {
			return ErrInvalidFeedback
		}
	case FeedbackKindPOIRating, FeedbackKindPOIReport:
		if strings.TrimSpace(r.POIID) == "" {
			return ErrInvalidFeedback
		}
	}
	if r.Rating != nil {
		v := *r.Rating
		if v < 0 || v > 5 {
			return ErrInvalidFeedback
		}
	}
	if utf8.RuneCountInString(r.Comment) > rawMaxCommentRunes {
		return ErrInvalidFeedback
	}
	if len(r.Extra) > maxExtraEntries {
		return ErrInvalidFeedback
	}
	for k, v := range r.Extra {
		if len(k) > maxExtraKeyLen || utf8.RuneCountInString(v) > maxExtraValueLen {
			return ErrInvalidFeedback
		}
	}
	return nil
}

// dedupeKey 批内/跨请求去重键：显式 IdempotencyKey 优先，否则语义哈希。
func dedupeKey(r *RawFeedbackInput) string {
	if r == nil {
		return ""
	}
	if k := strings.TrimSpace(r.IdempotencyKey); k != "" {
		return k
	}
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%s|%s|%s|%v|%s",
		r.UserID,
		r.ItineraryID,
		r.POIID,
		r.Kind,
		r.Comment,
		r.Rating,
		r.OccurredAt.UTC().Format(time.RFC3339Nano),
	)
	// 亚秒级抖动下仍可能多条被视为不同；调用方应优先传 IdempotencyKey。
	return hex.EncodeToString(h.Sum(nil))
}

type groupKey struct {
	UserID      string
	ItineraryID string
	POIID       string
	Kind        FeedbackKind
}

// FeedbackCollector 校验 → 去重 → 聚合 → 调用 FeedbackUsecase.Process。
type FeedbackCollector struct {
	uc  *FeedbackUsecase
	ttl time.Duration

	mu   sync.Mutex
	seen map[string]time.Time
}

// NewFeedbackCollector 构造收集器。ttl>0 时在内存中按 IdempotencyKey/dedupeKey 做过期去重；ttl==0 则仅批内去重。
func NewFeedbackCollector(uc *FeedbackUsecase, recentDedupeTTL time.Duration) *FeedbackCollector {
	c := &FeedbackCollector{uc: uc, ttl: recentDedupeTTL}
	if recentDedupeTTL > 0 {
		c.seen = make(map[string]time.Time)
	}
	return c
}

// Collect 处理多条原始反馈：校验、跨请求去重（若启用）、批内去重与聚合，再交给用例。
func (c *FeedbackCollector) Collect(ctx context.Context, raw []*RawFeedbackInput) error {
	if c == nil {
		return fmt.Errorf("feedback: collector is nil")
	}
	if c.uc == nil {
		return fmt.Errorf("feedback: usecase is nil")
	}
	if len(raw) == 0 {
		return ErrEmptyFeedback
	}

	active := make([]*RawFeedbackInput, 0, len(raw))
	for _, r := range raw {
		if r == nil {
			continue
		}
		if err := r.Validate(); err != nil {
			return err
		}
		key := dedupeKey(r)
		if key == "" {
			continue
		}
		if c.recentlySubmitted(key) {
			continue
		}
		active = append(active, r)
	}

	byKey := make(map[string]*RawFeedbackInput)
	order := make([]string, 0, len(active))
	for _, r := range active {
		k := dedupeKey(r)
		if _, ok := byKey[k]; ok {
			continue
		}
		byKey[k] = r
		order = append(order, k)
	}

	items := aggregateGroup(order, byKey)
	batch := &FeedbackBatch{Items: items}
	if err := c.uc.Process(ctx, batch); err != nil {
		return err
	}
	c.rememberDedupeKeys(order)
	return nil
}

// recentlySubmitted 跨请求 TTL 去重（成功 Persist 后才会写入 seen，同一次 Collect 内可安全合并重复键）。
func (c *FeedbackCollector) recentlySubmitted(key string) bool {
	if c.ttl <= 0 || c.seen == nil || key == "" {
		return false
	}
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, t := range c.seen {
		if now.Sub(t) > c.ttl {
			delete(c.seen, k)
		}
	}
	_, ok := c.seen[key]
	return ok
}

func (c *FeedbackCollector) rememberDedupeKeys(keys []string) {
	if c.ttl <= 0 || c.seen == nil {
		return
	}
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, k := range keys {
		if k == "" {
			continue
		}
		c.seen[k] = now
	}
}

func rawToItem(r *RawFeedbackInput, repeat int, rating *float64, comment string, at time.Time) FeedbackItem {
	extra := cloneExtra(r.Extra)
	key := strings.TrimSpace(r.IdempotencyKey)
	if key == "" {
		key = dedupeKey(r)
	}
	return FeedbackItem{
		IdempotencyKey: key,
		UserID:         strings.TrimSpace(r.UserID),
		ItineraryID:    strings.TrimSpace(r.ItineraryID),
		POIID:          strings.TrimSpace(r.POIID),
		Kind:           r.Kind,
		Rating:         rating,
		Comment:        comment,
		OccurredAt:     at,
		RepeatCount:    repeat,
		Extra:          extra,
	}
}

func cloneExtra(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// aggregateGroup 先按 Idempotency 去重后的列表，再按业务主键聚合评分与评论。
func aggregateGroup(order []string, byDedupe map[string]*RawFeedbackInput) []FeedbackItem {
	groups := make(map[groupKey]*aggSlot)
	keyOrder := make([]groupKey, 0)

	for _, dk := range order {
		r := byDedupe[dk]
		gk := groupKey{
			UserID:      strings.TrimSpace(r.UserID),
			ItineraryID: strings.TrimSpace(r.ItineraryID),
			POIID:       strings.TrimSpace(r.POIID),
			Kind:        r.Kind,
		}
		slot, ok := groups[gk]
		if !ok {
			slot = &aggSlot{}
			groups[gk] = slot
			keyOrder = append(keyOrder, gk)
		}
		slot.add(r)
	}

	out := make([]FeedbackItem, 0, len(keyOrder))
	for _, gk := range keyOrder {
		slot := groups[gk]
		rep := slot.representative()
		rating := slot.avgRating()
		comment := slot.joinComments()
		at := slot.latestAt()
		out = append(out, rawToItem(rep, slot.count, rating, comment, at))
	}
	return out
}

type aggSlot struct {
	count    int
	ratings  []float64
	comments []string
	latest   time.Time
	rep      *RawFeedbackInput
	firstAt  time.Time
}

func (s *aggSlot) add(r *RawFeedbackInput) {
	s.count++
	if s.rep == nil {
		s.rep = r
	}
	if r.Rating != nil {
		s.ratings = append(s.ratings, *r.Rating)
	}
	if c := strings.TrimSpace(r.Comment); c != "" {
		s.comments = append(s.comments, c)
	}
	ts := r.OccurredAt
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	if s.count == 1 {
		s.firstAt = ts
		s.latest = ts
	} else {
		if ts.After(s.latest) {
			s.latest = ts
		}
	}
}

func (s *aggSlot) representative() *RawFeedbackInput {
	if s.rep != nil {
		return s.rep
	}
	return &RawFeedbackInput{}
}

func (s *aggSlot) avgRating() *float64 {
	if len(s.ratings) == 0 {
		return nil
	}
	var sum float64
	for _, v := range s.ratings {
		sum += v
	}
	avg := sum / float64(len(s.ratings))
	return &avg
}

func (s *aggSlot) joinComments() string {
	if len(s.comments) == 0 {
		return ""
	}
	if len(s.comments) == 1 {
		return s.comments[0]
	}
	return strings.Join(s.comments, " | ")
}

func (s *aggSlot) latestAt() time.Time {
	if !s.latest.IsZero() {
		return s.latest
	}
	if !s.firstAt.IsZero() {
		return s.firstAt
	}
	return time.Now().UTC()
}
