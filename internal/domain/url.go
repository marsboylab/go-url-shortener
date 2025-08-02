package domain

import (
	"net/url"
	"strings"
	"time"
)

// URL은 단축 URL의 도메인 엔티티입니다
type URL struct {
	ID              string     `json:"id" db:"id" example:"my-project" format:"string" description:"단축 URL의 고유 식별자"`
	ShortURL        string     `json:"short_url" db:"-" example:"https://marsboy.dev/my-project" format:"uri" description:"완전한 단축 URL"`
	OriginalURL     string     `json:"original_url" db:"original_url" example:"https://github.com/username/awesome-project" format:"uri" description:"원본 URL"`
	QRCodeURL       string     `json:"qr_code_url" db:"-" example:"https://marsboy.dev/api/v1/urls/my-project/qr" format:"uri" description:"QR 코드 생성 URL"`
	Description     *string    `json:"description,omitempty" db:"description" example:"My awesome project repository" description:"URL에 대한 설명"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty" db:"expires_at" example:"2025-12-31T23:59:59Z" format:"date-time" description:"만료 일시"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at" example:"2025-08-02T10:30:00Z" format:"date-time" description:"생성 일시"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at" example:"2025-08-02T10:30:00Z" format:"date-time" description:"수정 일시"`
	ClickCount      int64      `json:"click_count" db:"click_count" example:"127" minimum:"0" description:"클릭 수"`
	IsActive        bool       `json:"is_active" db:"is_active" example:"true" description:"활성 상태"`
	LastAccessedAt  *time.Time `json:"last_accessed_at,omitempty" db:"last_accessed_at" example:"2025-08-02T15:45:30Z" format:"date-time" description:"마지막 접근 일시"`
	CreatedByAPIKey string     `json:"-" db:"created_by_api_key"`
}

// CreateURLRequest는 URL 생성 요청 구조체입니다
type CreateURLRequest struct {
	OriginalURL string     `json:"original_url" binding:"required,url,max=2048" example:"https://github.com/username/awesome-project/blob/main/README.md" format:"uri" description:"단축할 원본 URL (최대 2048자)"`
	CustomID    *string    `json:"custom_id,omitempty" binding:"omitempty,min=3,max=50" example:"my-project" minLength:"3" maxLength:"50" description:"커스텀 식별자 (3-50자, 영숫자와 하이픈만)"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" example:"2025-12-31T23:59:59Z" format:"date-time" description:"만료 일시 (ISO 8601 형식)"`
	Description *string    `json:"description,omitempty" binding:"omitempty,max=255" example:"My awesome project repository" maxLength:"255" description:"URL 설명 (최대 255자)"`
}

// UpdateURLRequest는 URL 업데이트 요청 구조체입니다
type UpdateURLRequest struct {
	OriginalURL *string    `json:"original_url,omitempty" binding:"omitempty,url,max=2048"`
	Description *string    `json:"description,omitempty" binding:"omitempty,max=255"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    *bool      `json:"is_active,omitempty"`
}

// URLListResponse는 URL 목록 응답 구조체입니다
type URLListResponse struct {
	URLs       []URL          `json:"urls" description:"URL 목록"`
	Pagination PaginationMeta `json:"pagination" description:"페이지네이션 정보"`
}

// PaginationMeta는 페이지네이션 메타데이터입니다
type PaginationMeta struct {
	CurrentPage int   `json:"current_page" example:"1" minimum:"1" description:"현재 페이지 번호"`
	PerPage     int   `json:"per_page" example:"20" minimum:"1" maximum:"100" description:"페이지당 항목 수"`
	TotalPages  int   `json:"total_pages" example:"5" minimum:"1" description:"전체 페이지 수"`
	TotalCount  int64 `json:"total_count" example:"95" minimum:"0" description:"전체 항목 수"`
	HasNext     bool  `json:"has_next" example:"true" description:"다음 페이지 존재 여부"`
	HasPrev     bool  `json:"has_prev" example:"false" description:"이전 페이지 존재 여부"`
}

// URLListOptions는 URL 목록 조회 옵션입니다
type URLListOptions struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	Limit    int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Sort     string `form:"sort" binding:"omitempty,oneof=created_at click_count last_accessed_at"`
	Order    string `form:"order" binding:"omitempty,oneof=asc desc"`
	IsActive *bool  `form:"is_active,omitempty"`
}

// NewURL은 새로운 URL 엔티티를 생성합니다
func NewURL(id, originalURL string, description *string, expiresAt *time.Time, apiKey string) *URL {
	now := time.Now()
	return &URL{
		ID:              id,
		OriginalURL:     originalURL,
		Description:     description,
		ExpiresAt:       expiresAt,
		CreatedAt:       now,
		UpdatedAt:       now,
		ClickCount:      0,
		IsActive:        true,
		CreatedByAPIKey: apiKey,
	}
}

// IsExpired는 URL이 만료되었는지 확인합니다
func (u *URL) IsExpired() bool {
	if u.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*u.ExpiresAt)
}

// IsAccessible은 URL에 접근 가능한지 확인합니다
func (u *URL) IsAccessible() bool {
	return u.IsActive && !u.IsExpired()
}

// IncrementClickCount는 클릭 수를 증가시킵니다
func (u *URL) IncrementClickCount() {
	u.ClickCount++
	now := time.Now()
	u.LastAccessedAt = &now
}

// BuildShortURL은 완전한 단축 URL을 구성합니다
func (u *URL) BuildShortURL(baseURL string) {
	u.ShortURL = strings.TrimRight(baseURL, "/") + "/" + u.ID
}

// BuildQRCodeURL은 QR 코드 URL을 구성합니다
func (u *URL) BuildQRCodeURL(baseURL string) {
	u.QRCodeURL = strings.TrimRight(baseURL, "/") + "/api/v1/urls/" + u.ID + "/qr"
}

// ValidateOriginalURL은 원본 URL의 유효성을 검사합니다
func ValidateOriginalURL(rawURL string) error {
	if rawURL == "" {
		return NewValidationError("original_url", "URL is required")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return NewValidationError("original_url", "Invalid URL format")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return NewValidationError("original_url", "URL must be http or https")
	}

	if parsed.Host == "" {
		return NewValidationError("original_url", "URL must have a valid host")
	}

	return nil
}

// ValidateCustomID는 커스텀 ID의 유효성을 검사합니다
func ValidateCustomID(customID string) error {
	if len(customID) < 3 || len(customID) > 50 {
		return NewValidationError("custom_id", "Custom ID must be between 3 and 50 characters")
	}

	// 영숫자와 하이픈만 허용
	for _, char := range customID {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '-') {
			return NewValidationError("custom_id", "Custom ID can only contain letters, numbers, and hyphens")
		}
	}

	// 예약된 키워드 확인
	reservedWords := []string{"api", "health", "admin", "www", "app", "dev", "stage", "prod"}
	lowerID := strings.ToLower(customID)
	for _, word := range reservedWords {
		if lowerID == word {
			return NewValidationError("custom_id", "Custom ID cannot use reserved word: "+word)
		}
	}

	return nil
}

// ValidationError는 도메인 레벨 유효성 검사 에러입니다
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError는 새로운 유효성 검사 에러를 생성합니다
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}