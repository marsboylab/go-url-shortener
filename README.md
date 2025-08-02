# 🔗 Go URL Shortener

개인 브랜딩을 위한 URL 단축 서비스 (marsboy.dev)

## 📋 프로젝트 개요

**목표**: 개인 도메인을 활용한 URL 단축 서비스 개발

### ✨ 주요 기능

- 🔗 짧고 깔끔한 링크로 포트폴리오/프로젝트 공유
- 📊 접근 통계 및 분석 기능
- 🎯 커스텀 ID 지원
- 📱 QR 코드 생성
- 🔒 API 키 기반 인증
- ⚡ Redis 캐싱으로 빠른 성능

### 💡 사용 예시

- `https://marsboy.dev/portfolio` → GitHub 포트폴리오
- `https://marsboy.dev/blog` → 개인 블로그
- `https://marsboy.dev/1` → 특정 프로젝트

## 🏗️ 시스템 아키텍처

### 핵심 동작 원리

1. **URL 단축**: 긴 URL → 고유 ID 생성 → 짧은 URL 반환
2. **리다이렉션**: 짧은 URL 접근 → 원본 URL 조회 → 301/302 리다이렉트
3. **통계 수집**: 클릭 수, 접근 시간, 리퍼러 등 분석

### 🛠️ 기술 스택

- **언어**: Go (Gin 프레임워크)
- **데이터베이스**: PostgreSQL (URL 매핑 저장)
- **캐시**: Redis (빠른 조회)
- **ID 생성**: Base62 인코딩 (0-9, a-z, A-Z)

## 📁 프로젝트 구조

```
go-url-shortener/
├── cmd/server/main.go              # 애플리케이션 진입점
├── internal/
│   ├── config/config.go            # 설정 관리
│   ├── domain/                     # 엔티티
│   │   ├── url.go
│   │   └── analytics.go
│   ├── repository/                 # 데이터 계층
│   │   ├── interfaces/
│   │   ├── postgres/url_repository.go
│   │   └── redis/cache_repository.go
│   ├── service/                    # 비즈니스 로직
│   │   ├── url_service.go
│   │   ├── analytics_service.go
│   │   └── id_generator.go
│   ├── handler/                    # HTTP 핸들러
│   │   └── url_handler.go
│   └── middleware/                 # 미들웨어
│       ├── rate_limit.go
│       ├── logging.go
│       ├── cors.go
│       └── auth.go
├── migrations/                     # DB 마이그레이션
├── docker/                         # 컨테이너 설정
└── Makefile                        # 빌드 스크립트
```

## 🚀 빠른 시작

### 1. 저장소 클론

```bash
git clone <repository-url>
cd go-url-shortener
```

### 2. 개발 환경 설정

```bash
# 환경 변수 복사
cp .env.example .env

# Docker Compose로 데이터베이스 및 Redis 실행
make docker-compose-up

# 데이터베이스 마이그레이션
make db-migrate
```

### 3. 서버 실행

```bash
# 로컬에서 실행
make run

# 또는 Docker로 실행
make docker-run
```

### 4. API 테스트

```bash
# 단축 URL 생성
curl -X POST http://localhost:8080/api/v1/urls \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk_marsboy_dev_key_1234567890" \
  -d '{
    "original_url": "https://github.com/username/awesome-project",
    "custom_id": "my-project",
    "description": "My awesome project"
  }'

# 리다이렉트 테스트
curl -L http://localhost:8080/my-project

# (선택사항) Swagger 문서 생성 후 UI 접속
make swagger-gen
open http://localhost:8080/swagger/index.html
```

## 📖 API 문서

### 🌐 Base URL

- **Development**: `http://localhost:8080`
- **Production**: `https://marsboy.dev`

### 🔑 인증

API 키를 헤더에 포함해서 요청:

```
X-API-Key: sk_marsboy_1234567890abcdef
```

### 📋 엔드포인트

#### 1. URL 단축 생성

```http
POST /api/v1/urls
Content-Type: application/json
X-API-Key: {your-api-key}

{
  "original_url": "https://example.com/very/long/url",
  "custom_id": "my-link",
  "expires_at": "2025-12-31T23:59:59Z",
  "description": "My awesome link"
}
```

#### 2. URL 정보 조회

```http
GET /api/v1/urls/{id}
X-API-Key: {your-api-key}
```

#### 3. URL 목록 조회

```http
GET /api/v1/urls?page=1&limit=20&sort=created_at&order=desc
X-API-Key: {your-api-key}
```

#### 4. 리다이렉션

```http
GET /{id}
```

#### 5. QR 코드 생성

```http
GET /api/v1/urls/{id}/qr?size=200
```

## 🔧 Base62 인코딩

### Base64 vs Base62

- **Base64**: A-Z, a-z, 0-9, +, / (64개 문자)
- **Base62**: A-Z, a-z, 0-9 (62개 문자, URL 안전)

### 변환 예시

- 숫자 1 → "1"
- 숫자 62 → "10"
- 숫자 123456 → "W7e"

**장점**: URL에 안전하고 짧은 ID 생성 가능

## 🧪 개발 도구

### Make 명령어

```bash
make help              # 사용 가능한 명령어 확인
make build             # 바이너리 빌드
make test              # 테스트 실행
make test-coverage     # 커버리지 테스트
make lint              # 코드 린팅
make fmt               # 코드 포맷팅
make dev-setup         # 개발 환경 자동 설정
```

### Docker 명령어

```bash
make docker-build           # Docker 이미지 빌드
make docker-compose-up      # 전체 서비스 실행
make docker-compose-down    # 서비스 중지
make docker-compose-logs    # 로그 확인
```

## 🚀 배포

### AWS 프로덕션 환경

```
Internet → Route53 → ALB → ECS → RDS + ElastiCache
```

### 구성 요소

- **Route53**: marsboy.dev 도메인 관리
- **ALB**: Application Load Balancer
- **ECS Fargate**: 컨테이너 실행
- **RDS PostgreSQL**: 메인 데이터베이스
- **ElastiCache Redis**: 캐싱 레이어

## 📊 모니터링 & 분석

### 기본 메트릭

- 클릭 수 및 고유 클릭 수
- 지리적 위치별 통계
- 브라우저/디바이스별 분석
- 리퍼러 추적

### 로그 관리

- 구조화된 JSON 로깅
- 요청/응답 추적
- 에러 모니터링
- 성능 메트릭

## 🔒 보안

### 보안 기능

- API 키 기반 인증
- Rate Limiting (분당 요청 제한)
- CORS 설정
- 입력 데이터 검증
- SQL Injection 방지

### 권장사항

- HTTPS 사용 필수
- 강력한 API 키 사용
- 정기적인 보안 업데이트
- 접근 로그 모니터링

## 🤝 기여하기

1. Fork the project
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 라이선스

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 연락처

- **Author**: marsboy
- **Website**: [marsboy.dev](https://marsboy.dev)
- **GitHub**: [@marsboy02](https://github.com/marsboy02)

---

⭐ 이 프로젝트가 도움이 되었다면 스타를 눌러주세요!
the project make url caching and short for sharing
