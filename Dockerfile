# 멀티 스테이지 빌드를 사용한 최적화된 Dockerfile

# 빌드 스테이지
FROM golang:1.21-alpine AS builder

# 필요한 패키지 설치
RUN apk add --no-cache git

# 작업 디렉토리 설정
WORKDIR /app

# Go 모듈 파일 복사
COPY go.mod go.sum ./

# 의존성 다운로드
RUN go mod download

# 소스 코드 복사
COPY . .

# 바이너리 빌드 (정적 링크)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# 실행 스테이지
FROM alpine:latest

# 보안 업데이트 및 ca-certificates 설치
RUN apk --no-cache add ca-certificates tzdata curl

# 작업 디렉토리 설정
WORKDIR /root/

# 빌드된 바이너리 복사
COPY --from=builder /app/main .

# 권한 설정
RUN chmod +x ./main

# 포트 노출
EXPOSE 8080

# 헬스체크 추가
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# 실행
CMD ["./main"]