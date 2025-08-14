package domain

import (
	"time"
)

type ClickEvent struct {
	ID          int64     `json:"id" db:"id"`
	URLId       string    `json:"url_id" db:"url_id"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	UserAgent   string    `json:"user_agent" db:"user_agent"`
	Referer     *string   `json:"referer,omitempty" db:"referer"`
	Country     *string   `json:"country,omitempty" db:"country"`
	City        *string   `json:"city,omitempty" db:"city"`
	Browser     *string   `json:"browser,omitempty" db:"browser"`
	OS          *string   `json:"os,omitempty" db:"os"`
	Device      *string   `json:"device,omitempty" db:"device"`
	ClickedAt   time.Time `json:"clicked_at" db:"clicked_at"`
	ProcessedAt time.Time `json:"processed_at" db:"processed_at"`
}

type URLAnalytics struct {
	URLID         string                   `json:"url_id"`
	TotalClicks   int64                    `json:"total_clicks"`
	UniqueClicks  int64                    `json:"unique_clicks"`
	ClicksByDate  []DailyClickStat         `json:"clicks_by_date"`
	TopReferrers  []ReferrerStat           `json:"top_referrers"`
	TopCountries  []CountryStat            `json:"top_countries"`
	TopBrowsers   []BrowserStat            `json:"top_browsers"`
	TopDevices    []DeviceStat             `json:"top_devices"`
	RecentClicks  []ClickEvent             `json:"recent_clicks"`
	GeneratedAt   time.Time                `json:"generated_at"`
}

type DailyClickStat struct {
	Date   string `json:"date" db:"date"`
	Clicks int64  `json:"clicks" db:"clicks"`
}

type ReferrerStat struct {
	Referer string `json:"referer" db:"referer"`
	Clicks  int64  `json:"clicks" db:"clicks"`
}

type CountryStat struct {
	Country string `json:"country" db:"country"`
	Clicks  int64  `json:"clicks" db:"clicks"`
}

type BrowserStat struct {
	Browser string `json:"browser" db:"browser"`
	Clicks  int64  `json:"clicks" db:"clicks"`
}

type DeviceStat struct {
	Device string `json:"device" db:"device"`
	Clicks int64  `json:"clicks" db:"clicks"`
}

type AnalyticsTimeRange struct {
	StartDate time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate   time.Time `form:"end_date" time_format:"2006-01-02"`
}

type AnalyticsOptions struct {
	TimeRange     AnalyticsTimeRange `form:",inline"`
	Granularity   string             `form:"granularity" binding:"omitempty,oneof=hour day week month"`
	IncludeEvents bool               `form:"include_events"`
	EventLimit    int                `form:"event_limit" binding:"omitempty,min=1,max=1000"`
}

func NewClickEvent(urlID, ipAddress, userAgent string, referer *string) *ClickEvent {
	now := time.Now()
	return &ClickEvent{
		URLId:       urlID,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Referer:     referer,
		ClickedAt:   now,
		ProcessedAt: now,
	}
}

func (c *ClickEvent) SetGeoLocation(country, city string) {
	if country != "" {
		c.Country = &country
	}
	if city != "" {
		c.City = &city
	}
}

func (c *ClickEvent) SetDeviceInfo(browser, os, device string) {
	if browser != "" {
		c.Browser = &browser
	}
	if os != "" {
		c.OS = &os
	}
	if device != "" {
		c.Device = &device
	}
}

func GetDefaultAnalyticsOptions() AnalyticsOptions {
	now := time.Now()
	return AnalyticsOptions{
		TimeRange: AnalyticsTimeRange{
			StartDate: now.AddDate(0, 0, -30), // 30일 전
			EndDate:   now,
		},
		Granularity:   "day",
		IncludeEvents: true,
		EventLimit:    100,
	}
}