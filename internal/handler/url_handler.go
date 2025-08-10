package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-url-shortener/internal/domain"
	"go-url-shortener/internal/middleware"
	"go-url-shortener/internal/service"
)

type URLHandler struct {
	urlService *service.URLService
}

func NewURLHandler(urlService *service.URLService) *URLHandler {
	return &URLHandler{
		urlService: urlService,
	}
}

// @Summary 단축 URL 생성
// @Description 긴 URL을 짧은 URL로 단축합니다. 커스텀 ID, 만료시간, 설명을 선택적으로 설정할 수 있습니다.
// @Tags URLs
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body domain.CreateURLRequest true "URL 생성 요청"
// @Success 201 {object} domain.URL "생성된 단축 URL 정보"
// @Failure 400 {object} domain.ErrorResponse "잘못된 요청"
// @Failure 401 {object} domain.ErrorResponse "인증 실패"
// @Failure 409 {object} domain.ErrorResponse "커스텀 ID 중복"
// @Failure 500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router /api/v1/urls [post]
func (h *URLHandler) CreateShortURL(c *gin.Context) {
	var req domain.CreateURLRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "Invalid request body",
			"details": map[string]interface{}{
				"validation_error": err.Error(),
			},
		})
		return
	}
	
	apiKey := middleware.GetAPIKeyFromContext(c)
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "API key is required",
		})
		return
	}
	
	url, err := h.urlService.CreateShortURL(c.Request.Context(), req, apiKey)
	if err != nil {
		h.handleError(c, err)
		return
	}
	
	c.JSON(http.StatusCreated, url)
}

// @Summary 단축 URL 정보 조회
// @Description 단축 URL의 상세 정보를 조회합니다. 클릭 수, 생성일, 만료일 등의 정보를 확인할 수 있습니다.
// @Tags URLs
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "단축 URL ID" example:"my-project"
// @Success 200 {object} domain.URL "단축 URL 정보"
// @Failure 400 {object} domain.ErrorResponse "잘못된 요청"
// @Failure 401 {object} domain.ErrorResponse "인증 실패"
// @Failure 404 {object} domain.ErrorResponse "URL을 찾을 수 없음"
// @Failure 500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router /api/v1/urls/{id} [get]
func (h *URLHandler) GetURLInfo(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "URL ID is required",
		})
		return
	}
	
	apiKey := middleware.GetAPIKeyFromContext(c)
	
	url, err := h.urlService.GetURLStats(c.Request.Context(), id, apiKey)
	if err != nil {
		h.handleError(c, err)
		return
	}
	
	c.JSON(http.StatusOK, url)
}

// @Summary URL 목록 조회
// @Description 내가 생성한 단축 URL 목록을 페이지네이션과 함께 조회합니다. 정렬 및 필터링이 가능합니다.
// @Tags URLs
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "페이지 번호" default(1) minimum(1)
// @Param limit query int false "페이지당 항목 수" default(20) minimum(1) maximum(100)
// @Param sort query string false "정렬 기준" Enums(created_at,click_count,last_accessed_at) default(created_at)
// @Param order query string false "정렬 순서" Enums(asc,desc) default(desc)
// @Param is_active query bool false "활성 상태 필터"
// @Success 200 {object} domain.URLListResponse "URL 목록과 페이지네이션 정보"
// @Failure 400 {object} domain.ErrorResponse "잘못된 요청"
// @Failure 401 {object} domain.ErrorResponse "인증 실패"
// @Failure 500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router /api/v1/urls [get]
func (h *URLHandler) ListURLs(c *gin.Context) {
	var options domain.URLListOptions
	
	if err := c.ShouldBindQuery(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "Invalid query parameters",
			"details": map[string]interface{}{
				"validation_error": err.Error(),
			},
		})
		return
	}
	
	apiKey := middleware.GetAPIKeyFromContext(c)
	
	response, err := h.urlService.ListURLs(c.Request.Context(), apiKey, options)
	if err != nil {
		h.handleError(c, err)
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// PUT /api/v1/urls/:id
func (h *URLHandler) UpdateURL(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "URL ID is required",
		})
		return
	}
	
	var req domain.UpdateURLRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "Invalid request body",
			"details": map[string]interface{}{
				"validation_error": err.Error(),
			},
		})
		return
	}
	
	apiKey := middleware.GetAPIKeyFromContext(c)
	
	url, err := h.urlService.UpdateURL(c.Request.Context(), id, req, apiKey)
	if err != nil {
		h.handleError(c, err)
		return
	}
	
	c.JSON(http.StatusOK, url)
}

// DELETE /api/v1/urls/:id
func (h *URLHandler) DeleteURL(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "URL ID is required",
		})
		return
	}
	
	apiKey := middleware.GetAPIKeyFromContext(c)
	
	err := h.urlService.DeleteURL(c.Request.Context(), id, apiKey)
	if err != nil {
		h.handleError(c, err)
		return
	}
	
	c.JSON(http.StatusNoContent, nil)
}

// @Summary URL 리다이렉션
// @Description 단축 URL에 접근하면 원본 URL로 리다이렉트합니다. 클릭 수가 자동으로 증가합니다.
// @Tags Redirect
// @Accept */*
// @Produce html
// @Param id path string true "단축 URL ID" example:"my-project"
// @Success 301 "원본 URL로 영구 리다이렉트"
// @Failure 404 {object} domain.ErrorResponse "URL을 찾을 수 없음"
// @Failure 410 {object} domain.ErrorResponse "만료된 URL"
// @Failure 500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router /{id} [get]
func (h *URLHandler) RedirectURL(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "url_not_found",
			"message": "Short URL not found",
		})
		return
	}
	
	url, err := h.urlService.GetURLForRedirect(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	
	// 301 영구 리다이렉트 (SEO에 좋음) 또는 302 임시 리다이렉트
	// 여기서는 301 사용
	c.Header("Cache-Control", "public, max-age=300") // 5분 캐시
	c.Redirect(http.StatusMovedPermanently, url.OriginalURL)
}

// @Summary QR 코드 생성
// @Description 단축 URL의 QR 코드를 생성합니다. 크기를 조정할 수 있습니다.
// @Tags QR Code
// @Accept */*
// @Produce image/png
// @Param id path string true "단축 URL ID" example:"my-project"
// @Param size query int false "QR 코드 크기" default(200) minimum(50) maximum(1000)
// @Success 301 "QR 코드 이미지로 리다이렉트"
// @Failure 400 {object} domain.ErrorResponse "잘못된 요청"
// @Failure 404 {object} domain.ErrorResponse "URL을 찾을 수 없음"
// @Failure 500 {object} domain.ErrorResponse "서버 내부 오류"
// @Router /api/v1/urls/{id}/qr [get]
func (h *URLHandler) GetQRCode(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "URL ID is required",
		})
		return
	}
	
	// QR 코드 크기 파라미터
	size := c.DefaultQuery("size", "200")
	sizeInt, err := strconv.Atoi(size)
	if err != nil || sizeInt < 50 || sizeInt > 1000 {
		sizeInt = 200 // 기본 크기
	}
	
	url, err := h.urlService.GetURL(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	
	// QR 코드 생성
	// TODO: 실제 구현에서는 qr 라이브러리 사용
	// 여기서는 외부 서비스로 리다이렉트
	qrURL := "https://api.qrserver.com/v1/create-qr-code/?size=" + 
			 strconv.Itoa(sizeInt) + "x" + strconv.Itoa(sizeInt) + 
			 "&data=" + url.ShortURL
	
	c.Redirect(http.StatusMovedPermanently, qrURL)
}

// GET /api/v1/urls/:id/analytics
func (h *URLHandler) GetAnalytics(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "URL ID is required",
		})
		return
	}
	
	apiKey := middleware.GetAPIKeyFromContext(c)
	
	// URL 존재 및 권한 확인
	_, err := h.urlService.GetURLStats(c.Request.Context(), id, apiKey)
	if err != nil {
		h.handleError(c, err)
		return
	}
	
	// 기본 분석 옵션으로 응답
	// TODO: 실제 분석 서비스 구현 필요
	analytics := gin.H{
		"url_id":       id,
		"total_clicks": 0,
		"unique_clicks": 0,
		"message":      "Analytics service will be implemented in future version",
	}
	
	c.JSON(http.StatusOK, analytics)
}

func (h *URLHandler) handleError(c *gin.Context, err error) {
	if serviceErr, ok := err.(*service.ServiceError); ok {
		statusCode := h.getHTTPStatusFromErrorCode(serviceErr.Code)
		c.JSON(statusCode, serviceErr)
		return
	}
	
	// 알 수 없는 에러
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   "internal_error",
		"message": "An unexpected error occurred",
	})
}

func (h *URLHandler) getHTTPStatusFromErrorCode(code service.ErrorCode) int {
	switch code {
	case service.ErrCodeValidation:
		return http.StatusBadRequest
	case service.ErrCodeNotFound:
		return http.StatusNotFound
	case service.ErrCodeConflict:
		return http.StatusConflict
	case service.ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case service.ErrCodeRateLimit:
		return http.StatusTooManyRequests
	case service.ErrCodeExpired:
		return http.StatusGone
	case service.ErrCodeInternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}