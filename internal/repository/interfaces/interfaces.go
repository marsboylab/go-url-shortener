package interfaces

import (
	"context"
	"time"

	"go-url-shortener/internal/domain"
)

type URLRepository interface {
	Create(ctx context.Context, url *domain.URL) error
	GetByID(ctx context.Context, id string) (*domain.URL, error)
	Update(ctx context.Context, url *domain.URL) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, apiKey string, options domain.URLListOptions) ([]domain.URL, int64, error)
	ExistsByID(ctx context.Context, id string) (bool, error)
	IncrementClickCount(ctx context.Context, id string) error
	UpdateLastAccessed(ctx context.Context, id string) error
	GetExpiredURLs(ctx context.Context, limit int) ([]domain.URL, error)
	DeleteExpiredURLs(ctx context.Context, before time.Time) (int64, error)
}

type AnalyticsRepository interface {
	RecordClick(ctx context.Context, event *domain.ClickEvent) error
	GetURLAnalytics(ctx context.Context, urlID string, options domain.AnalyticsOptions) (*domain.URLAnalytics, error)
	GetClicksByDateRange(ctx context.Context, urlID string, startDate, endDate time.Time, granularity string) ([]domain.DailyClickStat, error)
	GetTopReferrers(ctx context.Context, urlID string, startDate, endDate time.Time, limit int) ([]domain.ReferrerStat, error)
	GetTopCountries(ctx context.Context, urlID string, startDate, endDate time.Time, limit int) ([]domain.CountryStat, error)
	GetTopBrowsers(ctx context.Context, urlID string, startDate, endDate time.Time, limit int) ([]domain.BrowserStat, error)
	GetTopDevices(ctx context.Context, urlID string, startDate, endDate time.Time, limit int) ([]domain.DeviceStat, error)
	GetRecentClicks(ctx context.Context, urlID string, limit int) ([]domain.ClickEvent, error)
	GetUniqueClickCount(ctx context.Context, urlID string, startDate, endDate time.Time) (int64, error)
	DeleteOldEvents(ctx context.Context, before time.Time) (int64, error)
}

type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetURL(ctx context.Context, url *domain.URL, expiration time.Duration) error
	GetURL(ctx context.Context, id string) (*domain.URL, error)
	DeleteURL(ctx context.Context, id string) error
	IncrementCounter(ctx context.Context, key string, expiration time.Duration) (int64, error)
	SetAnalytics(ctx context.Context, urlID string, analytics *domain.URLAnalytics, expiration time.Duration) error
	GetAnalytics(ctx context.Context, urlID string) (*domain.URLAnalytics, error)
	DeleteAnalytics(ctx context.Context, urlID string) error
}