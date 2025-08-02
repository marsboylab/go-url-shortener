package service

import (
	"context"
	"log"
	"strings"
	"time"

	"go-url-shortener/internal/domain"
	"go-url-shortener/internal/repository/interfaces"
)

// URLService는 URL 관련 비즈니스 로직을 처리하는 서비스입니다
type URLService struct {
	urlRepo     interfaces.URLRepository
	cacheRepo   interfaces.CacheRepository
	idGenerator *IDGenerator
	baseURL     string
}

// NewURLService는 새로운 URL 서비스를 생성합니다
func NewURLService(urlRepo interfaces.URLRepository, cacheRepo interfaces.CacheRepository, baseURL string) *URLService {
	return &URLService{
		urlRepo:     urlRepo,
		cacheRepo:   cacheRepo,
		idGenerator: NewIDGenerator(6),
		baseURL:     baseURL,
	}
}

// CreateShortURL은 새로운 단축 URL을 생성합니다
func (s *URLService) CreateShortURL(ctx context.Context, req domain.CreateURLRequest, apiKey string) (*domain.URL, error) {
	// 원본 URL 유효성 검사
	if err := domain.ValidateOriginalURL(req.OriginalURL); err != nil {
		return nil, NewValidationError("original_url", err.Error(), nil)
	}

	// 커스텀 ID 처리
	var id string

	if req.CustomID != nil && *req.CustomID != "" {
		customID := strings.TrimSpace(*req.CustomID)
		
		// 커스텀 ID 유효성 검사
		if err := domain.ValidateCustomID(customID); err != nil {
			return nil, NewValidationError("custom_id", err.Error(), nil)
		}
		
		// 커스텀 ID 중복 확인
		exists, err := s.urlRepo.ExistsByID(ctx, customID)
		if err != nil {
			return nil, NewInternalError("Failed to check custom ID availability")
		}
		if exists {
			return nil, NewConflictError("Custom ID", customID)
		}
		
		id = customID
	} else {
		// 랜덤 ID 생성 (중복 방지)
		for attempts := 0; attempts < 10; attempts++ {
			generatedID, err := s.idGenerator.Generate()
			if err != nil {
				return nil, NewInternalError("Failed to generate ID")
			}
			
			exists, err := s.urlRepo.ExistsByID(ctx, generatedID)
			if err != nil {
				return nil, NewInternalError("Failed to check ID availability")
			}
			
			if !exists {
				id = generatedID
				break
			}
		}
		
		if id == "" {
			return nil, NewInternalError("Failed to generate unique ID after multiple attempts")
		}
	}

	// URL 엔티티 생성
	url := domain.NewURL(id, req.OriginalURL, req.Description, req.ExpiresAt, apiKey)
	
	// URL 빌드
	url.BuildShortURL(s.baseURL)
	url.BuildQRCodeURL(s.baseURL)

	// 데이터베이스에 저장
	if err := s.urlRepo.Create(ctx, url); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, NewConflictError("URL ID", id)
		}
		log.Printf("Failed to create URL in database: %v", err)
		return nil, NewInternalError("Failed to save URL")
	}

	// 캐시에 저장
	if err := s.cacheRepo.SetURL(ctx, url, 5*time.Minute); err != nil {
		log.Printf("Failed to cache URL: %v", err)
		// 캐시 실패는 치명적이지 않으므로 계속 진행
	}

	return url, nil
}

// GetURL은 단축 URL 정보를 조회합니다
func (s *URLService) GetURL(ctx context.Context, id string) (*domain.URL, error) {
	// 캐시에서 먼저 조회
	url, err := s.cacheRepo.GetURL(ctx, id)
	if err == nil {
		url.BuildShortURL(s.baseURL)
		url.BuildQRCodeURL(s.baseURL)
		return url, nil
	}

	// 캐시에 없으면 데이터베이스에서 조회
	url, err = s.urlRepo.GetByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, NewNotFoundError("Short URL")
		}
		log.Printf("Failed to get URL from database: %v", err)
		return nil, NewInternalError("Failed to retrieve URL")
	}

	// URL 접근 가능성 확인
	if !url.IsAccessible() {
		if url.IsExpired() {
			return nil, NewExpiredError("Short URL")
		}
		return nil, NewNotFoundError("Short URL")
	}

	// URL 빌드
	url.BuildShortURL(s.baseURL)
	url.BuildQRCodeURL(s.baseURL)

	// 캐시에 저장
	if err := s.cacheRepo.SetURL(ctx, url, 5*time.Minute); err != nil {
		log.Printf("Failed to cache URL: %v", err)
	}

	return url, nil
}

// GetURLForRedirect는 리다이렉션을 위한 URL 조회입니다 (클릭 수 증가 포함)
func (s *URLService) GetURLForRedirect(ctx context.Context, id string) (*domain.URL, error) {
	url, err := s.GetURL(ctx, id)
	if err != nil {
		return nil, err
	}

	// 클릭 수 증가 (비동기적으로 처리)
	go func() {
		bgCtx := context.Background()
		if err := s.urlRepo.IncrementClickCount(bgCtx, id); err != nil {
			log.Printf("Failed to increment click count for URL %s: %v", id, err)
		}
		
		// 캐시 무효화
		if err := s.cacheRepo.DeleteURL(bgCtx, id); err != nil {
			log.Printf("Failed to invalidate cache for URL %s: %v", id, err)
		}
	}()

	return url, nil
}

// ListURLs는 URL 목록을 조회합니다
func (s *URLService) ListURLs(ctx context.Context, apiKey string, options domain.URLListOptions) (*domain.URLListResponse, error) {
	// 기본값 설정
	if options.Page <= 0 {
		options.Page = 1
	}
	if options.Limit <= 0 {
		options.Limit = 20
	}
	if options.Limit > 100 {
		options.Limit = 100
	}

	urls, totalCount, err := s.urlRepo.List(ctx, apiKey, options)
	if err != nil {
		log.Printf("Failed to list URLs: %v", err)
		return nil, NewInternalError("Failed to retrieve URL list")
	}

	// URL 빌드
	for i := range urls {
		urls[i].BuildShortURL(s.baseURL)
		urls[i].BuildQRCodeURL(s.baseURL)
	}

	// 페이지네이션 메타데이터 계산
	totalPages := int((totalCount + int64(options.Limit) - 1) / int64(options.Limit))
	if totalPages == 0 {
		totalPages = 1
	}

	pagination := domain.PaginationMeta{
		CurrentPage: options.Page,
		PerPage:     options.Limit,
		TotalPages:  totalPages,
		TotalCount:  totalCount,
		HasNext:     options.Page < totalPages,
		HasPrev:     options.Page > 1,
	}

	return &domain.URLListResponse{
		URLs:       urls,
		Pagination: pagination,
	}, nil
}

// UpdateURL은 URL을 업데이트합니다
func (s *URLService) UpdateURL(ctx context.Context, id string, req domain.UpdateURLRequest, apiKey string) (*domain.URL, error) {
	// 기존 URL 조회
	url, err := s.urlRepo.GetByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, NewNotFoundError("Short URL")
		}
		return nil, NewInternalError("Failed to retrieve URL")
	}

	// 권한 확인
	if url.CreatedByAPIKey != apiKey {
		return nil, NewUnauthorizedError("You don't have permission to update this URL")
	}

	// 업데이트 필드 적용
	if req.OriginalURL != nil {
		if err := domain.ValidateOriginalURL(*req.OriginalURL); err != nil {
			return nil, NewValidationError("original_url", err.Error(), nil)
		}
		url.OriginalURL = *req.OriginalURL
	}

	if req.Description != nil {
		url.Description = req.Description
	}

	if req.ExpiresAt != nil {
		url.ExpiresAt = req.ExpiresAt
	}

	if req.IsActive != nil {
		url.IsActive = *req.IsActive
	}

	url.UpdatedAt = time.Now()

	// 데이터베이스 업데이트
	if err := s.urlRepo.Update(ctx, url); err != nil {
		log.Printf("Failed to update URL: %v", err)
		return nil, NewInternalError("Failed to update URL")
	}

	// 캐시 무효화
	if err := s.cacheRepo.DeleteURL(ctx, id); err != nil {
		log.Printf("Failed to invalidate cache for URL %s: %v", id, err)
	}

	// URL 빌드
	url.BuildShortURL(s.baseURL)
	url.BuildQRCodeURL(s.baseURL)

	return url, nil
}

// DeleteURL은 URL을 삭제합니다 (soft delete)
func (s *URLService) DeleteURL(ctx context.Context, id string, apiKey string) error {
	// 기존 URL 조회
	url, err := s.urlRepo.GetByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return NewNotFoundError("Short URL")
		}
		return NewInternalError("Failed to retrieve URL")
	}

	// 권한 확인
	if url.CreatedByAPIKey != apiKey {
		return NewUnauthorizedError("You don't have permission to delete this URL")
	}

	// 삭제 처리
	if err := s.urlRepo.Delete(ctx, id); err != nil {
		log.Printf("Failed to delete URL: %v", err)
		return NewInternalError("Failed to delete URL")
	}

	// 캐시 무효화
	if err := s.cacheRepo.DeleteURL(ctx, id); err != nil {
		log.Printf("Failed to invalidate cache for URL %s: %v", id, err)
	}

	return nil
}

// GetURLStats는 URL의 기본 통계를 조회합니다
func (s *URLService) GetURLStats(ctx context.Context, id string, apiKey string) (*domain.URL, error) {
	url, err := s.urlRepo.GetByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, NewNotFoundError("Short URL")
		}
		return nil, NewInternalError("Failed to retrieve URL")
	}

	// 권한 확인
	if url.CreatedByAPIKey != apiKey {
		return nil, NewUnauthorizedError("You don't have permission to view this URL's stats")
	}

	// URL 빌드
	url.BuildShortURL(s.baseURL)
	url.BuildQRCodeURL(s.baseURL)

	return url, nil
}

// CleanupExpiredURLs는 만료된 URL들을 정리합니다
func (s *URLService) CleanupExpiredURLs(ctx context.Context) (int64, error) {
	deleted, err := s.urlRepo.DeleteExpiredURLs(ctx, time.Now())
	if err != nil {
		log.Printf("Failed to cleanup expired URLs: %v", err)
		return 0, NewInternalError("Failed to cleanup expired URLs")
	}

	log.Printf("Cleaned up %d expired URLs", deleted)
	return deleted, nil
}