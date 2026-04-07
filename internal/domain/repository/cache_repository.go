package repository

import (
	"context"
	"time"

	"atlas-routex/internal/domain/entity"
)

// CacheRepository 定义缓存层需要暴露的能力。
type CacheRepository interface {
	GetPOI(ctx context.Context, poiID string) (*entity.POI, error)
	SetPOI(ctx context.Context, poi *entity.POI, ttl time.Duration) error
	DeletePOI(ctx context.Context, poiID string) error
	BatchGetPOIs(ctx context.Context, poiIDs []string) (map[string]*entity.POI, error)
	BatchSetPOIs(ctx context.Context, pois map[string]*entity.POI, ttl time.Duration) error
	GetNearbyPOIs(ctx context.Context, lat, lng float64, radius int, category string) ([]string, error)
	SetNearbyPOIs(ctx context.Context, lat, lng float64, radius int, category string, poiIDs []string, ttl time.Duration) error
	GetItinerary(ctx context.Context, itineraryID string) (*entity.Itinerary, error)
	SetItinerary(ctx context.Context, itinerary *entity.Itinerary, ttl time.Duration) error
	DeleteItinerary(ctx context.Context, itineraryID string) error
	GetPlanningResult(ctx context.Context, userID string, constraints map[string]interface{}) ([]*entity.Itinerary, error)
	SetPlanningResult(ctx context.Context, userID string, constraints map[string]interface{}, itineraries []*entity.Itinerary, ttl time.Duration) error
	SetProgress(ctx context.Context, taskID string, progress int, status string) error
	GetProgress(ctx context.Context, taskID string) (int, string, error)
	SaveDraft(ctx context.Context, userID string, draft *entity.Itinerary, ttl time.Duration) error
	GetDraft(ctx context.Context, userID string) (*entity.Itinerary, error)
	DeleteDraft(ctx context.Context, userID string) error
	AcquireLock(ctx context.Context, key string, ttl time.Duration) (string, error)
	ReleaseLock(ctx context.Context, key, lockValue string) error
	IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error)
	GetRateLimit(ctx context.Context, key string) (int64, error)
	BatchDelete(ctx context.Context, pattern string) error
}
