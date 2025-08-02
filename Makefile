# Go URL Shortener Makefile

# 변수 설정
BINARY_NAME=urlshortener
MAIN_PATH=cmd/server/main.go
DOCKER_IMAGE=go-url-shortener
DOCKER_TAG=latest

# 기본 타겟
.DEFAULT_GOAL := help

# Go 관련 명령어
.PHONY: build
build: ## Go 바이너리 빌드
	@echo "Building $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

.PHONY: run
run: ## 로컬에서 서버 실행
	@echo "Running server locally..."
	go run $(MAIN_PATH)

.PHONY: test
test: ## 테스트 실행
	@echo "Running tests..."
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## 테스트 커버리지 실행
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: clean
clean: ## 빌드 파일 정리
	@echo "Cleaning up..."
	rm -rf bin/
	rm -f coverage.out coverage.html

.PHONY: deps
deps: ## 의존성 다운로드
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

.PHONY: lint
lint: ## 코드 린팅
	@echo "Running linter..."
	golangci-lint run

# Docker 관련 명령어
.PHONY: docker-build
docker-build: ## Docker 이미지 빌드
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-run
docker-run: ## Docker 컨테이너 실행
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-compose-up
docker-compose-up: ## Docker Compose로 전체 서비스 실행
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

.PHONY: docker-compose-dev
docker-compose-dev: ## Docker Compose 개발 환경 실행
	@echo "Starting development services with Docker Compose..."
	docker-compose -f docker-compose.dev.yml up -d

.PHONY: docker-compose-down
docker-compose-down: ## Docker Compose 서비스 중지
	@echo "Stopping services..."
	docker-compose down

.PHONY: docker-compose-logs
docker-compose-logs: ## Docker Compose 로그 확인
	docker-compose logs -f

# 데이터베이스 관련 명령어
.PHONY: db-migrate
db-migrate: ## 데이터베이스 마이그레이션 실행
	@echo "Running database migration..."
	psql $(DATABASE_URL) -f migrations/001_create_urls_table.sql

.PHONY: db-reset
db-reset: ## 데이터베이스 초기화
	@echo "Resetting database..."
	psql $(DATABASE_URL) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	$(MAKE) db-migrate

# 개발 환경 관련
.PHONY: dev-setup
dev-setup: ## 개발 환경 설정
	@echo "Setting up development environment..."
	$(MAKE) docker-compose-up
	@echo "Waiting for services to start..."
	sleep 10
	$(MAKE) db-migrate

.PHONY: dev-run
dev-run: ## 개발 모드로 실행 (hot reload)
	@echo "Starting development server..."
	air

# 유틸리티
.PHONY: fmt
fmt: ## 코드 포맷팅
	@echo "Formatting code..."
	go fmt ./...

.PHONY: vet
vet: ## 코드 검사
	@echo "Running go vet..."
	go vet ./...

.PHONY: mod-verify
mod-verify: ## 모듈 검증
	@echo "Verifying modules..."
	go mod verify

.PHONY: security-check
security-check: ## 보안 취약점 검사
	@echo "Running security check..."
	gosec ./...

# Swagger 관련 명령어
.PHONY: swagger-install
swagger-install: ## Swagger 도구 설치
	@echo "Installing Swagger..."
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: swagger-gen
swagger-gen: ## Swagger 문서 생성
	@echo "Generating Swagger documentation..."
	export PATH=$$PATH:~/go/bin && swag init -g cmd/server/main.go

.PHONY: swagger-serve
swagger-serve: swagger-gen build ## Swagger 문서와 함께 서버 실행
	@echo "Starting server with Swagger UI..."
	@echo "🔗 Swagger UI: http://localhost:8080/swagger/index.html"
	./bin/$(BINARY_NAME)

.PHONY: swagger-docker
swagger-docker: ## Docker에서 Swagger UI와 함께 서비스 실행
	@echo "Starting Docker services with Swagger UI..."
	@echo "🔗 Swagger UI: http://localhost:8080/swagger/index.html"
	$(MAKE) docker-compose-up

# 프로덕션 관련
.PHONY: build-prod
build-prod: ## 프로덕션용 빌드
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o bin/$(BINARY_NAME) $(MAIN_PATH)

.PHONY: docker-build-prod
docker-build-prod: ## 프로덕션용 Docker 이미지 빌드
	@echo "Building production Docker image..."
	docker build -t $(DOCKER_IMAGE):prod .

# 헬프
.PHONY: help
help: ## 사용 가능한 명령어 표시
	@echo "🔗 Go URL Shortener - 개발 도구"
	@echo ""
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "🔗 Quick Start:"
	@echo "  make docker-compose-up  # Docker로 전체 서비스 실행"
	@echo "  make swagger-serve       # Swagger UI와 함께 로컬 실행"
	@echo ""
	@echo "🔗 URLs:"
	@echo "  Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "  Health:     http://localhost:8080/health"
	@echo "  API:        http://localhost:8080/api/v1/"