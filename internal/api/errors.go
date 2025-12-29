package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 错误码定义
const (
	// 通用错误码 (1xxx)
	ErrCodeInvalidRequest   = "ERR_INVALID_REQUEST"
	ErrCodeUnauthorized     = "ERR_UNAUTHORIZED"
	ErrCodeForbidden        = "ERR_FORBIDDEN"
	ErrCodeNotFound         = "ERR_NOT_FOUND"
	ErrCodeInternalError    = "ERR_INTERNAL_ERROR"
	ErrCodeServiceUnavailable = "ERR_SERVICE_UNAVAILABLE"

	// 认证错误码 (2xxx)
	ErrCodeInvalidCredentials = "ERR_INVALID_CREDENTIALS"
	ErrCodeEmailExists        = "ERR_EMAIL_EXISTS"
	ErrCodeRegistrationClosed = "ERR_REGISTRATION_CLOSED"
	ErrCodeUserDisabled       = "ERR_USER_DISABLED"
	ErrCodeSessionExpired     = "ERR_SESSION_EXPIRED"

	// 资源错误码 (3xxx)
	ErrCodeProviderNotFound   = "ERR_PROVIDER_NOT_FOUND"
	ErrCodeProviderDisabled   = "ERR_PROVIDER_DISABLED"
	ErrCodeProviderUnavailable = "ERR_PROVIDER_UNAVAILABLE"
	ErrCodeModelNotFound      = "ERR_MODEL_NOT_FOUND"
	ErrCodeModelDisabled      = "ERR_MODEL_DISABLED"
	ErrCodeTagNotFound        = "ERR_TAG_NOT_FOUND"
	ErrCodeRecordNotFound     = "ERR_RECORD_NOT_FOUND"
	ErrCodeUserNotFound       = "ERR_USER_NOT_FOUND"

	// 业务逻辑错误码 (4xxx)
	ErrCodeMissingField       = "ERR_MISSING_FIELD"
	ErrCodeInvalidTag         = "ERR_INVALID_TAG"
	ErrCodeCannotDeleteSelf   = "ERR_CANNOT_DELETE_SELF"
	ErrCodeGenerationFailed   = "ERR_GENERATION_FAILED"
)

// APIError 统一的 API 错误响应结构
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// ErrorResponse 返回统一格式的错误响应
func ErrorResponse(c *gin.Context, status int, code string, message string) {
	c.JSON(status, APIError{
		Code:    code,
		Message: message,
	})
}

// ErrorResponseWithDetails 返回带详情的错误响应
func ErrorResponseWithDetails(c *gin.Context, status int, code string, message string, details any) {
	c.JSON(status, APIError{
		Code:    code,
		Message: message,
		Details: details,
	})
}

// 常用错误响应快捷函数

// BadRequest 400 错误请求
func BadRequest(c *gin.Context, code string, message string) {
	ErrorResponse(c, http.StatusBadRequest, code, message)
}

// Unauthorized 401 未授权
func Unauthorized(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

// Forbidden 403 禁止访问
func Forbidden(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, ErrCodeForbidden, message)
}

// NotFound 404 资源不存在
func NotFound(c *gin.Context, code string, message string) {
	ErrorResponse(c, http.StatusNotFound, code, message)
}

// InternalError 500 服务器内部错误
func InternalError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, message)
}

// ServiceUnavailable 503 服务不可用
func ServiceUnavailable(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusServiceUnavailable, ErrCodeServiceUnavailable, message)
}

// MissingField 缺少必填字段
func MissingField(c *gin.Context, field string) {
	ErrorResponseWithDetails(c, http.StatusBadRequest, ErrCodeMissingField, field+" is required", gin.H{"field": field})
}

// InvalidPayload 无效的请求体
func InvalidPayload(c *gin.Context) {
	ErrorResponse(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request payload")
}
