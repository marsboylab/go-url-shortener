package service

import "fmt"

// ErrorCode는 서비스 레이어 에러 코드를 정의합니다
type ErrorCode string

const (
	ErrCodeValidation     ErrorCode = "validation_failed"
	ErrCodeNotFound       ErrorCode = "not_found"
	ErrCodeConflict       ErrorCode = "conflict"
	ErrCodeInternalError  ErrorCode = "internal_error"
	ErrCodeUnauthorized   ErrorCode = "unauthorized"
	ErrCodeRateLimit      ErrorCode = "rate_limit_exceeded"
	ErrCodeExpired        ErrorCode = "expired"
)

// ServiceError는 서비스 레이어의 표준 에러 타입입니다
type ServiceError struct {
	Code    ErrorCode              `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewValidationError는 유효성 검사 에러를 생성합니다
func NewValidationError(field, message string, details map[string]interface{}) *ServiceError {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["field"] = field
	
	return &ServiceError{
		Code:    ErrCodeValidation,
		Message: message,
		Details: details,
	}
}

// NewNotFoundError는 리소스 없음 에러를 생성합니다
func NewNotFoundError(resource string) *ServiceError {
	return &ServiceError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Details: map[string]interface{}{
			"resource": resource,
		},
	}
}

// NewConflictError는 충돌 에러를 생성합니다
func NewConflictError(resource, identifier string) *ServiceError {
	return &ServiceError{
		Code:    ErrCodeConflict,
		Message: fmt.Sprintf("%s '%s' already exists", resource, identifier),
		Details: map[string]interface{}{
			"resource":   resource,
			"identifier": identifier,
		},
	}
}

// NewInternalError는 내부 서버 에러를 생성합니다
func NewInternalError(message string) *ServiceError {
	return &ServiceError{
		Code:    ErrCodeInternalError,
		Message: message,
	}
}

// NewUnauthorizedError는 인증 에러를 생성합니다
func NewUnauthorizedError(message string) *ServiceError {
	return &ServiceError{
		Code:    ErrCodeUnauthorized,
		Message: message,
	}
}

// NewRateLimitError는 요청 제한 에러를 생성합니다
func NewRateLimitError(limit int, window string) *ServiceError {
	return &ServiceError{
		Code:    ErrCodeRateLimit,
		Message: fmt.Sprintf("Rate limit exceeded: %d requests per %s", limit, window),
		Details: map[string]interface{}{
			"limit":  limit,
			"window": window,
		},
	}
}

// NewExpiredError는 만료 에러를 생성합니다
func NewExpiredError(resource string) *ServiceError {
	return &ServiceError{
		Code:    ErrCodeExpired,
		Message: fmt.Sprintf("%s has expired", resource),
		Details: map[string]interface{}{
			"resource": resource,
		},
	}
}