// Package apispec 是面向资源的接口契约：统一响应、带原因码的错误、分页、排序、
// 输入校验、方法策略、幂等与资源标识。新模块可以选用（见 docs/API_SPEC.md）；
// system 模块保持原有的 {code,data,msg} 约定，两者并存。
package apispec

import (
	"errors"
	"fmt"
	"net/http"
)

// Code 是标准错误码。模块细分原因用 Error.Reason 表达，不扩展数字码。
type Code struct {
	Value     int32
	Name      string
	Status    int
	Retryable bool
}

var (
	CodeInternalServiceError  = Code{100001, "InternalServiceError", http.StatusInternalServerError, true}
	CodeMissingParameter      = Code{100002, "MissingParameter", http.StatusBadRequest, false}
	CodeInvalidParameter      = Code{100003, "InvalidParameter", http.StatusBadRequest, false}
	CodeUnauthorizedAccess    = Code{100004, "UnauthorizedAccess", http.StatusUnauthorized, false}
	CodeResourceAlreadyExists = Code{100005, "ResourceAlreadyExists", http.StatusConflict, false}
	CodeResourceNotFound      = Code{100006, "ResourceNotFound", http.StatusNotFound, false}
	CodeBadRequest            = Code{100007, "BadRequest", http.StatusBadRequest, false}
	CodeNotImplemented        = Code{100008, "NotImplemented", http.StatusNotImplemented, false}
	CodePermissionDenied      = Code{100009, "PermissionDenied", http.StatusForbidden, false}
	CodeFailedPrecondition    = Code{100010, "FailedPrecondition", http.StatusBadRequest, false}
	CodeConflict              = Code{100011, "Conflict", http.StatusConflict, false}
	CodePayloadTooLarge       = Code{100012, "PayloadTooLarge", http.StatusRequestEntityTooLarge, false}
	CodeUnprocessableEntity   = Code{100013, "UnprocessableEntity", http.StatusUnprocessableEntity, false}
	CodeRateLimited           = Code{100014, "RateLimited", http.StatusTooManyRequests, true}
	CodeQuotaExceeded         = Code{100015, "QuotaExceeded", http.StatusTooManyRequests, true}
	CodeServiceUnavailable    = Code{100016, "ServiceUnavailable", http.StatusServiceUnavailable, true}
	CodeDeadlineExceeded      = Code{100017, "DeadlineExceeded", http.StatusGatewayTimeout, true}
)

// FieldViolation 指出一个字段的问题；不回显被拒绝的原始值。
type FieldViolation struct {
	Field          string `json:"field"`
	Reason         string `json:"reason"`
	Issue          string `json:"issue"`
	ExpectedFormat string `json:"expected_format,omitempty"`
	Suggestion     string `json:"suggestion,omitempty"`
}

// Error 是 Service 层返回的 typed domain error，由 WriteError 统一转换为响应。
type Error struct {
	Code              Code
	Reason            string // 稳定点分原因码，如 task.already_claimed
	Message           string // 面向人的默认文案，不含内部细节
	Retryable         bool
	RetryAfterSeconds int32
	FieldViolations   []FieldViolation
	Metadata          map[string]string
	cause             error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s (%s): %s: %v", e.Code.Name, e.Reason, e.Message, e.cause)
	}
	return fmt.Sprintf("%s (%s): %s", e.Code.Name, e.Reason, e.Message)
}

func (e *Error) Unwrap() error { return e.cause }

// NewError 按错误码默认的可重试性创建错误。
func NewError(code Code, reason, message string) *Error {
	return &Error{Code: code, Reason: reason, Message: message, Retryable: code.Retryable}
}

// WithCause 记录内部原因，只进日志，不进响应。
func (e *Error) WithCause(err error) *Error { e.cause = err; return e }

func (e *Error) WithMetadata(key, value string) *Error {
	if e.Metadata == nil {
		e.Metadata = map[string]string{}
	}
	e.Metadata[key] = value
	return e
}

// WithRetryAfter 设置最早重试时间并标记可重试。
func (e *Error) WithRetryAfter(seconds int32) *Error {
	e.RetryAfterSeconds = seconds
	e.Retryable = true
	return e
}

func InvalidParameter(reason, message string, violations ...FieldViolation) *Error {
	e := NewError(CodeInvalidParameter, reason, message)
	e.FieldViolations = violations
	return e
}

func NotFound(reason, message string) *Error {
	return NewError(CodeResourceNotFound, reason, message)
}

func FailedPrecondition(reason, message string) *Error {
	return NewError(CodeFailedPrecondition, reason, message)
}

func AlreadyExists(reason, message string) *Error {
	return NewError(CodeResourceAlreadyExists, reason, message)
}

func Unauthorized(reason, message string) *Error {
	return NewError(CodeUnauthorizedAccess, reason, message)
}

func PermissionDenied(reason, message string) *Error {
	return NewError(CodePermissionDenied, reason, message)
}

func RateLimited(retryAfterSeconds int32) *Error {
	return NewError(CodeRateLimited, "rate.limited", "too many requests").WithRetryAfter(retryAfterSeconds)
}

// QuotaExceeded 只有能给出等待时长时才可重试。
func QuotaExceeded(reason, message string, retryAfterSeconds int32) *Error {
	e := NewError(CodeQuotaExceeded, reason, message)
	e.Retryable = false
	if retryAfterSeconds > 0 {
		e.WithRetryAfter(retryAfterSeconds)
	}
	return e
}

func Internal(err error) *Error {
	return NewError(CodeInternalServiceError, "internal.unexpected", "internal service error").WithCause(err)
}

func Unavailable(reason string, err error) *Error {
	return NewError(CodeServiceUnavailable, reason, "service temporarily unavailable").WithCause(err)
}

// AsError 把任意错误转换为 *Error；未知错误视为内部错误，不泄露细节。
func AsError(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return Internal(err)
}
