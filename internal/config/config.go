package config

import (
	"os"
	"strconv"
)

type Config struct {
	// server
	Environment string
	Port        string
	BaseURL     string
	APIKey      string

	// database
	DatabaseURL   string
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// url
	DefaultIDLength int
	MaxURLLength    int
	MaxDescLength   int

	// security
	RateLimitPerMinute int
	CacheExpiration    int // seconds
}

func Load() *Config {
	redisDB := 0
	if db := os.Getenv("REDIS_DB"); db != "" {
		if parsed, err := strconv.Atoi(db); err == nil {
			redisDB = parsed
		}
	}

	defaultIDLength := 6
	if length := os.Getenv("DEFAULT_ID_LENGTH"); length != "" {
		if parsed, err := strconv.Atoi(length); err == nil {
			defaultIDLength = parsed
		}
	}

	maxURLLength := 2048
	if length := os.Getenv("MAX_URL_LENGTH"); length != "" {
		if parsed, err := strconv.Atoi(length); err == nil {
			maxURLLength = parsed
		}
	}

	maxDescLength := 255
	if length := os.Getenv("MAX_DESC_LENGTH"); length != "" {
		if parsed, err := strconv.Atoi(length); err == nil {
			maxDescLength = parsed
		}
	}

	rateLimitPerMinute := 60
	if limit := os.Getenv("RATE_LIMIT_PER_MINUTE"); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			rateLimitPerMinute = parsed
		}
	}

	cacheExpiration := 300 // 5분
	if exp := os.Getenv("CACHE_EXPIRATION"); exp != "" {
		if parsed, err := strconv.Atoi(exp); err == nil {
			cacheExpiration = parsed
		}
	}

	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8080"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		APIKey:      getEnv("API_KEY", "sk_marsboy_dev_key"),

		DatabaseURL:   getEnv("DATABASE_URL", "postgres://user:password@localhost/urlshortener?sslmode=disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		DefaultIDLength: defaultIDLength,
		MaxURLLength:    maxURLLength,
		MaxDescLength:   maxDescLength,

		RateLimitPerMinute: rateLimitPerMinute,
		CacheExpiration:    cacheExpiration,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}