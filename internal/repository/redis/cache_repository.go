package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"go-url-shortener/internal/domain"
	"go-url-shortener/internal/repository/interfaces"
)

// cacheRepository는 Redis를 사용한 캐시 저장소 구현입니다
type cacheRepository struct {
	client *redis.Client
}

// NewCacheRepository는 새로운 캐시 저장소를 생성합니다
func NewCacheRepository(client *redis.Client) interfaces.CacheRepository {
	return &cacheRepository{client: client}
}

// Set은 키-값을 캐시에 저장합니다
func (r *cacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	
	err = r.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}
	
	return nil
}

// Get은 캐시에서 값을 조회합니다
func (r *cacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key '%s' not found in cache", key)
		}
		return fmt.Errorf("failed to get cache: %w", err)
	}
	
	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}
	
	return nil
}

// Delete는 캐시에서 키를 삭제합니다
func (r *cacheRepository) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}
	
	return nil
}

// Exists는 키가 존재하는지 확인합니다
func (r *cacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}
	
	return exists > 0, nil
}

// SetURL은 URL 객체를 캐시에 저장합니다
func (r *cacheRepository) SetURL(ctx context.Context, url *domain.URL, expiration time.Duration) error {
	key := r.urlCacheKey(url.ID)
	return r.Set(ctx, key, url, expiration)
}

// GetURL은 캐시에서 URL 객체를 조회합니다
func (r *cacheRepository) GetURL(ctx context.Context, id string) (*domain.URL, error) {
	key := r.urlCacheKey(id)
	var url domain.URL
	err := r.Get(ctx, key, &url)
	if err != nil {
		return nil, err
	}
	
	return &url, nil
}

// DeleteURL은 캐시에서 URL을 삭제합니다
func (r *cacheRepository) DeleteURL(ctx context.Context, id string) error {
	key := r.urlCacheKey(id)
	return r.Delete(ctx, key)
}

// IncrementCounter는 카운터를 증가시킵니다 (rate limiting 등에 사용)
func (r *cacheRepository) IncrementCounter(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	pipe := r.client.TxPipeline()
	
	// 카운터 증가
	incrCmd := pipe.Incr(ctx, key)
	
	// 만료 시간 설정 (첫 번째 증가인 경우)
	pipe.Expire(ctx, key, expiration)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment counter: %w", err)
	}
	
	return incrCmd.Val(), nil
}

// SetAnalytics는 분석 데이터를 캐시에 저장합니다
func (r *cacheRepository) SetAnalytics(ctx context.Context, urlID string, analytics *domain.URLAnalytics, expiration time.Duration) error {
	key := r.analyticsCacheKey(urlID)
	return r.Set(ctx, key, analytics, expiration)
}

// GetAnalytics는 캐시에서 분석 데이터를 조회합니다
func (r *cacheRepository) GetAnalytics(ctx context.Context, urlID string) (*domain.URLAnalytics, error) {
	key := r.analyticsCacheKey(urlID)
	var analytics domain.URLAnalytics
	err := r.Get(ctx, key, &analytics)
	if err != nil {
		return nil, err
	}
	
	return &analytics, nil
}

// DeleteAnalytics는 캐시에서 분석 데이터를 삭제합니다
func (r *cacheRepository) DeleteAnalytics(ctx context.Context, urlID string) error {
	key := r.analyticsCacheKey(urlID)
	return r.Delete(ctx, key)
}

// Helper methods for cache key generation

func (r *cacheRepository) urlCacheKey(id string) string {
	return fmt.Sprintf("url:%s", id)
}

func (r *cacheRepository) analyticsCacheKey(urlID string) string {
	return fmt.Sprintf("analytics:%s", urlID)
}

// Additional utility methods

// SetWithNX는 키가 존재하지 않을 때만 값을 설정합니다
func (r *cacheRepository) SetWithNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}
	
	success, err := r.client.SetNX(ctx, key, data, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set cache with NX: %w", err)
	}
	
	return success, nil
}

// GetTTL은 키의 남은 만료 시간을 조회합니다
func (r *cacheRepository) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}
	
	return ttl, nil
}

// FlushPattern은 패턴에 매칭되는 모든 키를 삭제합니다
func (r *cacheRepository) FlushPattern(ctx context.Context, pattern string) error {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys by pattern: %w", err)
	}
	
	if len(keys) > 0 {
		err = r.client.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}
	
	return nil
}