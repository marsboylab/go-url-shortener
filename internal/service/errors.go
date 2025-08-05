package service

import "fmt"

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

type ServiceError struct {
	Code    ErrorCode              `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

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

func NewNotFoundError(resource string) *ServiceError {
	return &ServiceError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Details: map[string]interface{}{
			"resource": resource,
		},
	}
}

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

func NewInternalError(message string) *ServiceError {
	return &ServiceError{
		Code:    ErrCodeInternalError,
		Message: message,
	}
}

func NewUnauthorizedError(message string) *ServiceError {
	return &ServiceError{
		Code:    ErrCodeUnauthorized,
		Message: message,
	}
}

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

func NewExpiredError(resource string) *ServiceError {
	return &ServiceError{
		Code:    ErrCodeExpired,
		Message: fmt.Sprintf("%s has expired", resource),
		Details: map[string]interface{}{
			"resource": resource,
		},
	}
}