//这个文件是对redis的热点缓存,一共设计四个不同模块，分别是POI相关缓存，行程相关缓存，会话与状态，还有分布式协调

//三种缓存策略设计：过期策略，更新策略和失效策略
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"atlasroutex/internal/domain/entity"
	"atlasroutex/internal/domain/repository"
	"atlasroutex/internal/domain/valueobject"
)

// CacheRepository Redis缓存仓储实现
type CacheRepository struct {
	client *Client
}

// NewCacheRepository 创建缓存仓储实例
func NewCacheRepository(client *Client) repository.CacheRepository {
	return &CacheRepository{
		client: client,
	}
}

// ==================== POI相关缓存 ====================

// GetPOI 获取POI缓存
func (r *CacheRepository) GetPOI(ctx context.Context, poiID string) (*entity.POI, error) {
	key := r.buildPOIKey(poiID)
	
	data, err := r.client.GetClient().Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // 缓存未命中
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get POI from cache")
	}
	
	var poi entity.POI
	if err := json.Unmarshal([]byte(data), &poi); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal POI")
	}
	
	return &poi, nil
}

// SetPOI 设置POI缓存
func (r *CacheRepository) SetPOI(ctx context.Context, poi *entity.POI, ttl time.Duration) error {
	key := r.buildPOIKey(poi.ID)
	
	data, err := json.Marshal(poi)
	if err != nil {
		return errors.Wrap(err, "failed to marshal POI")
	}
	
	if err := r.client.GetClient().Set(ctx, key, data, ttl).Err(); err != nil {
		return errors.Wrap(err, "failed to set POI cache")
	}
	
	return nil
}

// DeletePOI 删除POI缓存
func (r *CacheRepository) DeletePOI(ctx context.Context, poiID string) error {
	key := r.buildPOIKey(poiID)
	return r.client.GetClient().Del(ctx, key).Err()
}

// BatchGetPOIs 批量获取POI缓存
func (r *CacheRepository) BatchGetPOIs(ctx context.Context, poiIDs []string) (map[string]*entity.POI, error) {
	if len(poiIDs) == 0 {
		return make(map[string]*entity.POI), nil
	}
	
	// 构建keys
	keys := make([]string, len(poiIDs))
	for i, id := range poiIDs {
		keys[i] = r.buildPOIKey(id)
	}
	
	// 批量获取
	values, err := r.client.GetClient().MGet(ctx, keys...).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to batch get POIs")
	}
	
	result := make(map[string]*entity.POI)
	for i, val := range values {
		if val == nil {
			continue
		}
		
		var poi entity.POI
		if err := json.Unmarshal([]byte(val.(string)), &poi); err != nil {
			r.client.logger.Warn("failed to unmarshal POI in batch",
				zap.String("poi_id", poiIDs[i]),
				zap.Error(err))
			continue
		}
		result[poiIDs[i]] = &poi
	}
	
	return result, nil
}

// BatchSetPOIs 批量设置POI缓存
func (r *CacheRepository) BatchSetPOIs(ctx context.Context, pois map[string]*entity.POI, ttl time.Duration) error {
	if len(pois) == 0 {
		return nil
	}
	
	// 使用Pipeline批量操作
	pipe := r.client.GetClient().Pipeline()
	for id, poi := range pois {
		key := r.buildPOIKey(id)
		data, err := json.Marshal(poi)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal POI %s", id)
		}
		pipe.Set(ctx, key, data, ttl)
	}
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to batch set POIs")
	}
	
	return nil
}

// ==================== 周边POI缓存 ====================

// GetNearbyPOIs 获取周边POI列表缓存
func (r *CacheRepository) GetNearbyPOIs(ctx context.Context, lat, lng float64, radius int, category string) ([]string, error) {
	key := r.buildNearbyKey(lat, lng, radius, category)
	
	// 使用ZSET存储，按距离排序
	results, err := r.client.GetClient().ZRevRange(ctx, key, 0, -1).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get nearby POIs")
	}
	
	return results, nil
}

// SetNearbyPOIs 设置周边POI列表缓存
func (r *CacheRepository) SetNearbyPOIs(ctx context.Context, lat, lng float64, radius int, category string, poiIDs []string, ttl time.Duration) error {
	key := r.buildNearbyKey(lat, lng, radius, category)
	
	pipe := r.client.GetClient().Pipeline()
	
	// 先删除旧数据
	pipe.Del(ctx, key)
	
	// 添加新数据，使用距离作为score
	for i, poiID := range poiIDs {
		// 使用索引作为距离分数（实际应该使用真实距离）
		pipe.ZAdd(ctx, key, &redis.Z{
			Score:  float64(i),
			Member: poiID,
		})
	}
	
	// 设置过期时间
	pipe.Expire(ctx, key, ttl)
	
	_, err := pipe.Exec(ctx)
	return errors.Wrap(err, "failed to set nearby POIs")
}

// ==================== 行程相关缓存 ====================

// GetItinerary 获取行程缓存
func (r *CacheRepository) GetItinerary(ctx context.Context, itineraryID string) (*entity.Itinerary, error) {
	key := r.buildItineraryKey(itineraryID)
	
	data, err := r.client.GetClient().Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get itinerary from cache")
	}
	
	var itinerary entity.Itinerary
	if err := json.Unmarshal([]byte(data), &itinerary); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal itinerary")
	}
	
	return &itinerary, nil
}

// SetItinerary 设置行程缓存
func (r *CacheRepository) SetItinerary(ctx context.Context, itinerary *entity.Itinerary, ttl time.Duration) error {
	key := r.buildItineraryKey(itinerary.ID)
	
	data, err := json.Marshal(itinerary)
	if err != nil {
		return errors.Wrap(err, "failed to marshal itinerary")
	}
	
	return r.client.GetClient().Set(ctx, key, data, ttl).Err()
}

// DeleteItinerary 删除行程缓存
func (r *CacheRepository) DeleteItinerary(ctx context.Context, itineraryID string) error {
	key := r.buildItineraryKey(itineraryID)
	return r.client.GetClient().Del(ctx, key).Err()
}

// ==================== 规划结果缓存 ====================

// GetPlanningResult 获取规划结果缓存
func (r *CacheRepository) GetPlanningResult(ctx context.Context, userID string, constraints map[string]interface{}) ([]*entity.Itinerary, error) {
	key := r.buildPlanningKey(userID, constraints)
	
	data, err := r.client.GetClient().Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get planning result")
	}
	
	var itineraries []*entity.Itinerary
	if err := json.Unmarshal([]byte(data), &itineraries); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal itineraries")
	}
	
	return itineraries, nil
}

// SetPlanningResult 设置规划结果缓存
func (r *CacheRepository) SetPlanningResult(ctx context.Context, userID string, constraints map[string]interface{}, itineraries []*entity.Itinerary, ttl time.Duration) error {
	key := r.buildPlanningKey(userID, constraints)
	
	data, err := json.Marshal(itineraries)
	if err != nil {
		return errors.Wrap(err, "failed to marshal itineraries")
	}
	
	return r.client.GetClient().Set(ctx, key, data, ttl).Err()
}

// ==================== 进度状态缓存 ====================

// SetProgress 设置任务进度
func (r *CacheRepository) SetProgress(ctx context.Context, taskID string, progress int, status string) error {
	key := r.buildProgressKey(taskID)
	
	pipe := r.client.GetClient().Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"progress": progress,
		"status":   status,
		"updated":  time.Now().Unix(),
	})
	pipe.Expire(ctx, key, 1*time.Hour)
	
	_, err := pipe.Exec(ctx)
	return errors.Wrap(err, "failed to set progress")
}

// GetProgress 获取任务进度
func (r *CacheRepository) GetProgress(ctx context.Context, taskID string) (int, string, error) {
	key := r.buildProgressKey(taskID)
	
	data, err := r.client.GetClient().HGetAll(ctx, key).Result()
	if err != nil {
		return 0, "", errors.Wrap(err, "failed to get progress")
	}
	
	if len(data) == 0 {
		return 0, "", nil
	}
	
	progress, _ := strconv.Atoi(data["progress"])
	status := data["status"]
	
	return progress, status, nil
}

// ==================== 临时草稿缓存 ====================

// SaveDraft 保存行程草稿
func (r *CacheRepository) SaveDraft(ctx context.Context, userID string, draft *entity.Itinerary, ttl time.Duration) error {
	key := r.buildDraftKey(userID)
	
	data, err := json.Marshal(draft)
	if err != nil {
		return errors.Wrap(err, "failed to marshal draft")
	}
	
	return r.client.GetClient().Set(ctx, key, data, ttl).Err()
}

// GetDraft 获取行程草稿
func (r *CacheRepository) GetDraft(ctx context.Context, userID string) (*entity.Itinerary, error) {
	key := r.buildDraftKey(userID)
	
	data, err := r.client.GetClient().Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get draft")
	}
	
	var itinerary entity.Itinerary
	if err := json.Unmarshal([]byte(data), &itinerary); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal draft")
	}
	
	return &itinerary, nil
}

// DeleteDraft 删除行程草稿
func (r *CacheRepository) DeleteDraft(ctx context.Context, userID string) error {
	key := r.buildDraftKey(userID)
	return r.client.GetClient().Del(ctx, key).Err()
}

// ==================== 分布式锁 ====================

// AcquireLock 获取分布式锁
func (r *CacheRepository) AcquireLock(ctx context.Context, key string, ttl time.Duration) (string, error) {
	lockKey := r.buildLockKey(key)
	lockValue := uuid.New().String()
	
	// 使用SET NX EX命令
	ok, err := r.client.GetClient().SetNX(ctx, lockKey, lockValue, ttl).Result()
	if err != nil {
		return "", errors.Wrap(err, "failed to acquire lock")
	}
	
	if !ok {
		return "", nil // 获取锁失败
	}
	
	return lockValue, nil
}

// ReleaseLock 释放分布式锁
func (r *CacheRepository) ReleaseLock(ctx context.Context, key, lockValue string) error {
	lockKey := r.buildLockKey(key)
	
	// 使用Lua脚本确保只有持有锁的客户端才能释放
	luaScript := redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`)
	
	result, err := luaScript.Run(ctx, r.client.GetClient(), []string{lockKey}, lockValue).Int()
	if err != nil {
		return errors.Wrap(err, "failed to release lock")
	}
	
	if result == 0 {
		return errors.New("lock not held by this client")
	}
	
	return nil
}

// ==================== 限流计数器 ====================

// IncrementRateLimit 增加限流计数
func (r *CacheRepository) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	rateKey := r.buildRateLimitKey(key)
	
	pipe := r.client.GetClient().Pipeline()
	
	// 增加计数
	incr := pipe.Incr(ctx, rateKey)
	// 设置过期时间
	pipe.Expire(ctx, rateKey, window)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to increment rate limit")
	}
	
	return incr.Val(), nil
}

// GetRateLimit 获取当前限流计数
func (r *CacheRepository) GetRateLimit(ctx context.Context, key string) (int64, error) {
	rateKey := r.buildRateLimitKey(key)
	
	count, err := r.client.GetClient().Get(ctx, rateKey).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, errors.Wrap(err, "failed to get rate limit")
	}
	
	return count, nil
}

// ==================== 批量操作 ====================

// BatchDelete 批量删除缓存
func (r *CacheRepository) BatchDelete(ctx context.Context, pattern string) error {
	// 使用SCAN命令查找匹配的key
	var cursor uint64
	var keys []string
	
	for {
		var scanKeys []string
		var err error
		
		scanKeys, cursor, err = r.client.GetClient().Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return errors.Wrap(err, "failed to scan keys")
		}
		
		keys = append(keys, scanKeys...)
		
		if cursor == 0 {
			break
		}
	}
	
	if len(keys) > 0 {
		if err := r.client.GetClient().Del(ctx, keys...).Err(); err != nil {
			return errors.Wrap(err, "failed to delete keys")
		}
	}
	
	return nil
}

// ==================== 辅助方法 ====================

// buildPOIKey 构建POI缓存key
func (r *CacheRepository) buildPOIKey(poiID string) string {
	return fmt.Sprintf("atlas:poi:%s", poiID)
}

// buildNearbyKey 构建周边POI缓存key
func (r *CacheRepository) buildNearbyKey(lat, lng float64, radius int, category string) string {
	// 使用Geohash精度控制，简化key
	latKey := fmt.Sprintf("%.2f", lat)
	lngKey := fmt.Sprintf("%.2f", lng)
	return fmt.Sprintf("atlas:nearby:%s:%s:%d:%s", latKey, lngKey, radius, category)
}

// buildItineraryKey 构建行程缓存key
func (r *CacheRepository) buildItineraryKey(itineraryID string) string {
	return fmt.Sprintf("atlas:itinerary:%s", itineraryID)
}

// buildPlanningKey 构建规划结果缓存key
func (r *CacheRepository) buildPlanningKey(userID string, constraints map[string]interface{}) string {
	// 简化：将约束条件序列化后hash
	// 实际应该使用更优雅的方式
	return fmt.Sprintf("atlas:planning:%s:%v", userID, constraints)
}

// buildProgressKey 构建进度缓存key
func (r *CacheRepository) buildProgressKey(taskID string) string {
	return fmt.Sprintf("atlas:progress:%s", taskID)
}

// buildDraftKey 构建草稿缓存key
func (r *CacheRepository) buildDraftKey(userID string) string {
	return fmt.Sprintf("atlas:draft:%s", userID)
}

// buildLockKey 构建锁key
func (r *CacheRepository) buildLockKey(key string) string {
	return fmt.Sprintf("atlas:lock:%s", key)
}

// buildRateLimitKey 构建限流key
func (r *CacheRepository) buildRateLimitKey(key string) string {
	return fmt.Sprintf("atlas:ratelimit:%s", key)
}
