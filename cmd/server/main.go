package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go-url-shortener/internal/config"
	"go-url-shortener/internal/handler"
	"go-url-shortener/internal/middleware"
	"go-url-shortener/internal/repository/postgres"
	redisRepo "go-url-shortener/internal/repository/redis"
	"go-url-shortener/internal/service"

	_ "go-url-shortener/docs" // Swagger 문서 임포트
)

// @title Go URL Shortener API
// @version 1.0
// @description 개인 브랜딩을 위한 URL 단축 서비스 API
// @termsOfService https://marsboy.dev/terms

// @contact.name marsboy
// @contact.url https://marsboy.dev
// @contact.email contact@marsboy.dev

// @license.name MIT License
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API Key 인증을 위해 X-API-Key 헤더에 API 키를 포함해주세요.

// @externalDocs.description Notion 프로젝트 문서
// @externalDocs.url https://www.notion.so/teamsparta/Go-URL-Shortener-Project-2432dc3ef51481998ac9d5b55bfd4ee3

func main() {
	// 환경 변수 로드
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// 설정 로드
	cfg := config.Load()

	// 데이터베이스 연결
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Redis 연결
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Repository 초기화
	urlRepo := postgres.NewURLRepository(db)
	cacheRepo := redisRepo.NewCacheRepository(rdb)

	// Service 초기화
	urlService := service.NewURLService(urlRepo, cacheRepo, cfg.BaseURL)

	// Handler 초기화
	urlHandler := handler.NewURLHandler(urlService)

	// Gin 라우터 설정
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit())

	// 헬스체크
	router.GET("/health", healthCheck)

	// API 라우트
	api := router.Group("/api/v1")
	{
		api.POST("/urls", middleware.APIKeyAuth(cfg.APIKey), urlHandler.CreateShortURL)
		api.GET("/urls/:id", middleware.APIKeyAuth(cfg.APIKey), urlHandler.GetURLInfo)
		api.GET("/urls", middleware.APIKeyAuth(cfg.APIKey), urlHandler.ListURLs)
		api.DELETE("/urls/:id", middleware.APIKeyAuth(cfg.APIKey), urlHandler.DeleteURL)
		api.GET("/urls/:id/qr", urlHandler.GetQRCode)
		api.GET("/urls/:id/analytics", middleware.APIKeyAuth(cfg.APIKey), urlHandler.GetAnalytics)
	}

	// Swagger UI 라우트
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 리다이렉트 라우트 (루트 레벨)
	router.GET("/:id", urlHandler.RedirectURL)

	// 서버 시작
	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("Base URL: %s", cfg.BaseURL)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// healthCheck 헬스체크 엔드포인트
// @Summary 서버 헬스체크
// @Description 서버가 정상적으로 동작하는지 확인합니다.
// @Tags Health
// @Accept */*
// @Produce json
// @Success 200 {object} domain.HealthResponse "서버 정상 상태"
// @Router /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}