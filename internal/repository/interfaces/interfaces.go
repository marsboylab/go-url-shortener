package interfaces

import (
	"context"
	"time"

	"go-url-shortener/internal/domain"
)

// URLRepository는 URL 데이터 접근을 위한 인터페이스입니다
type URLRepository interface {
	// Create는 새로운 URL을 생성합니다
	Create(ctx context.Context, url *domain.URL) error
	
	// GetByID는 ID로 URL을 조회합니다
	GetByID(ctx context.Context, id string) (*domain.URL, error)
	
	// Update는 URL을 업데이트합니다
	Update(ctx context.Context, url *domain.URL) error
	
	// Delete는 URL을 삭제합니다 (soft delete)
	Delete(ctx context.Context, id string) error
	
	// List는 URL 목록을 조회합니다
	List(ctx context.Context, apiKey string, options domain.URLListOptions) ([]domain.URL, int64, error)
	
	// ExistsByID는 ID가 이미 존재하는지 확인합니다
	ExistsByID(ctx context.Context, id string) (bool, error)
	
	// IncrementClickCount는 클릭 수를 증가시킵니다
	IncrementClickCount(ctx context.Context, id string) error
	
	// UpdateLastAccessed는 마지막 접근 시간을 업데이트합니다
	UpdateLastAccessed(ctx context.Context, id string) error
	
	// GetExpiredURLs는 만료된 URL 목록을 조회합니다
	GetExpiredURLs(ctx context.Context, limit int) ([]domain.URL, error)
	
	// DeleteExpiredURLs는 만료된 URL들을 삭제합니다
	DeleteExpiredURLs(ctx context.Context, before time.Time) (int64, error)
}

// AnalyticsRepository는 분석 데이터 접근을 위한 인터페이스입니다
type AnalyticsRepository interface {
	// RecordClick은 클릭 이벤트를 기록합니다
	RecordClick(ctx context.Context, event *domain.ClickEvent) error
	
	// GetURLAnalytics는 URL의 분석 데이터를 조회합니다
	GetURLAnalytics(ctx context.Context, urlID string, options domain.AnalyticsOptions) (*domain.URLAnalytics, error)
	
	// GetClicksByDateRange는 기간별 클릭 통계를 조회합니다
	GetClicksByDateRange(ctx context.Context, urlID string, startDate, endDate time.Time, granularity string) ([]domain.DailyClickStat, error)
	
	// GetTopReferrers는 상위 리퍼러를 조회합니다
	GetTopReferrers(ctx context.Context, urlID string, startDate, endDate time.Time, limit int) ([]domain.ReferrerStat, error)
	
	// GetTopCountries는 상위 국가를 조회합니다
	GetTopCountries(ctx context.Context, urlID string, startDate, endDate time.Time, limit int) ([]domain.CountryStat, error)
	
	// GetTopBrowsers는 상위 브라우저를 조회합니다
	GetTopBrowsers(ctx context.Context, urlID string, startDate, endDate time.Time, limit int) ([]domain.BrowserStat, error)
	
	// GetTopDevices는 상위 디바이스를 조회합니다
	GetTopDevices(ctx context.Context, urlID string, startDate, endDate time.Time, limit int) ([]domain.DeviceStat, error)
	
	// GetRecentClicks는 최근 클릭 이벤트를 조회합니다
	GetRecentClicks(ctx context.Context, urlID string, limit int) ([]domain.ClickEvent, error)
	
	// GetUniqueClickCount는 고유 클릭 수를 조회합니다
	GetUniqueClickCount(ctx context.Context, urlID string, startDate, endDate time.Time) (int64, error)
	
	// DeleteOldEvents는 오래된 이벤트를 삭제합니다
	DeleteOldEvents(ctx context.Context, before time.Time) (int64, error)
}

// CacheRepository는 캐시 데이터 접근을 위한 인터페이스입니다
type CacheRepository interface {
	// Set은 키-값을 캐시에 저장합니다
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	
	// Get은 캐시에서 값을 조회합니다
	Get(ctx context.Context, key string, dest interface{}) error
	
	// Delete는 캐시에서 키를 삭제합니다
	Delete(ctx context.Context, key string) error
	
	// Exists는 키가 존재하는지 확인합니다
	Exists(ctx context.Context, key string) (bool, error)
	
	// SetURL은 URL 객체를 캐시에 저장합니다
	SetURL(ctx context.Context, url *domain.URL, expiration time.Duration) error
	
	// GetURL은 캐시에서 URL 객체를 조회합니다
	GetURL(ctx context.Context, id string) (*domain.URL, error)
	
	// DeleteURL은 캐시에서 URL을 삭제합니다
	DeleteURL(ctx context.Context, id string) error
	
	// IncrementCounter는 카운터를 증가시킵니다 (rate limiting 등에 사용)
	IncrementCounter(ctx context.Context, key string, expiration time.Duration) (int64, error)
	
	// SetAnalytics는 분석 데이터를 캐시에 저장합니다
	SetAnalytics(ctx context.Context, urlID string, analytics *domain.URLAnalytics, expiration time.Duration) error
	
	// GetAnalytics는 캐시에서 분석 데이터를 조회합니다
	GetAnalytics(ctx context.Context, urlID string) (*domain.URLAnalytics, error)
	
	// DeleteAnalytics는 캐시에서 분석 데이터를 삭제합니다
	DeleteAnalytics(ctx context.Context, urlID string) error
}