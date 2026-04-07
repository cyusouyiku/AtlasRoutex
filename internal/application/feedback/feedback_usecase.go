package feedback

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// 业务语义错误，供 HTTP/gRPC 映射。
var (
	ErrInvalidFeedback  = errors.New("feedback: invalid feedback")
	ErrEmptyFeedback    = errors.New("feedback: empty batch")
	ErrFeedbackTooLarge = errors.New("feedback: batch too large")
)

const (
	maxBatchItems    = 500
	maxCommentRunes  = 4000
	maxExtraEntries  = 32
	maxExtraKeyLen   = 64
	maxExtraValueLen = 512
)

// FeedbackKind 反馈类型（可随产品线扩展）。
type FeedbackKind string

const (
	FeedbackKindItineraryRating FeedbackKind = "itinerary_rating"
	FeedbackKindPOIRating       FeedbackKind = "poi_rating"
	FeedbackKindPOIReport       FeedbackKind = "poi_report"
	FeedbackKindSearchQuality   FeedbackKind = "search_quality"
	FeedbackKindGeneral         FeedbackKind = "general"
)

// IsValid 是否受支持的业务类型。
func (k FeedbackKind) IsValid() bool {
	switch k {
	case FeedbackKindItineraryRating, FeedbackKindPOIRating,
		FeedbackKindPOIReport, FeedbackKindSearchQuality, FeedbackKindGeneral:
		return true
	default:
		return false
	}
}

// FeedbackItem 一条已校验、可持久化的反馈（含聚合后的重复计数）。
type FeedbackItem struct {
	IdempotencyKey string
	UserID         string
	ItineraryID    string
	POIID          string
	Kind           FeedbackKind
	Rating         *float64
	Comment        string
	OccurredAt     time.Time
	RepeatCount    int
	Extra          map[string]string
}

// FeedbackBatch 交给 Sink 的一批反馈。
type FeedbackBatch struct {
	Items []FeedbackItem
}

// Validate 批次与条目的完整性（Sink 前应已通过 Collector，此处做最后一道守门）。
func (b *FeedbackBatch) Validate() error {
	if b == nil {
		return ErrInvalidFeedback
	}
	if len(b.Items) == 0 {
		return ErrEmptyFeedback
	}
	if len(b.Items) > maxBatchItems {
		return ErrFeedbackTooLarge
	}
	for i := range b.Items {
		if err := b.Items[i].validate(); err != nil {
			return err
		}
	}
	return nil
}

func (it *FeedbackItem) validate() error {
	if it == nil {
		return ErrInvalidFeedback
	}
	if !it.Kind.IsValid() {
		return ErrInvalidFeedback
	}
	if it.RepeatCount < 1 {
		return ErrInvalidFeedback
	}
	switch it.Kind {
	case FeedbackKindItineraryRating:
		if strings.TrimSpace(it.ItineraryID) == "" {
			return ErrInvalidFeedback
		}
	case FeedbackKindPOIRating, FeedbackKindPOIReport:
		if strings.TrimSpace(it.POIID) == "" {
			return ErrInvalidFeedback
		}
	}
	if it.Rating != nil {
		r := *it.Rating
		if r < 0 || r > 5 {
			return ErrInvalidFeedback
		}
	}
	if utf8.RuneCountInString(it.Comment) > maxCommentRunes {
		return ErrInvalidFeedback
	}
	if len(it.Extra) > maxExtraEntries {
		return ErrInvalidFeedback
	}
	for k, v := range it.Extra {
		if len(k) > maxExtraKeyLen || utf8.RuneCountInString(v) > maxExtraValueLen {
			return ErrInvalidFeedback
		}
	}
	return nil
}

// Clone 返回批次浅拷贝（供异步派发时避免与调用方共享底层 slice 被改写）。
func (b *FeedbackBatch) Clone() *FeedbackBatch {
	if b == nil {
		return nil
	}
	items := make([]FeedbackItem, len(b.Items))
	copy(items, b.Items)
	for i := range items {
		if items[i].Extra != nil {
			extra := make(map[string]string, len(items[i].Extra))
			for k, v := range items[i].Extra {
				extra[k] = v
			}
			items[i].Extra = extra
		}
	}
	return &FeedbackBatch{Items: items}
}

// FeedbackSink 持久化或下游投递端口（队列、仓库、分析管道等）。
type FeedbackSink interface {
	Persist(ctx context.Context, batch *FeedbackBatch) error
}

type noopSink struct{}

func (noopSink) Persist(ctx context.Context, batch *FeedbackBatch) error {
	_ = ctx
	_ = batch
	return nil
}

// FeedbackUsecase 反馈用例：校验批次并调用 Sink；可选异步占位。
type FeedbackUsecase struct {
	sink     FeedbackSink
	async    bool
	asyncTTL time.Duration
}

// FeedbackUsecaseOption 可选配置。
type FeedbackUsecaseOption func(*FeedbackUsecase)

// WithAsyncPersist 启用异步 Persist：尽快返回，在独立 goroutine 中用 Background+超时调用 Sink。
func WithAsyncPersist(timeout time.Duration) FeedbackUsecaseOption {
	return func(uc *FeedbackUsecase) {
		uc.async = true
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		uc.asyncTTL = timeout
	}
}

// NewFeedbackUsecase 构造用例；sink 为 nil 时使用 noop Sink（仅用于联调/占位）。
func NewFeedbackUsecase(sink FeedbackSink, opts ...FeedbackUsecaseOption) *FeedbackUsecase {
	s := sink
	if s == nil {
		s = noopSink{}
	}
	uc := &FeedbackUsecase{sink: s, asyncTTL: 30 * time.Second}
	for _, o := range opts {
		o(uc)
	}
	return uc
}

// Process 校验并交给 Sink；异步模式下拷贝批次后在后台执行 Persist。
func (uc *FeedbackUsecase) Process(ctx context.Context, batch *FeedbackBatch) error {
	if uc == nil {
		return fmt.Errorf("feedback: usecase is nil")
	}
	if err := batch.Validate(); err != nil {
		return err
	}
	if uc.async {
		payload := batch.Clone()
		ttl := uc.asyncTTL
		if ttl <= 0 {
			ttl = 30 * time.Second
		}
		go func() {
			bg, cancel := context.WithTimeout(context.Background(), ttl)
			defer cancel()
			_ = uc.sink.Persist(bg, payload)
		}()
		return nil
	}
	return uc.sink.Persist(ctx, batch)
}
