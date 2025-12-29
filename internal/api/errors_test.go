package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		status         int
		code           string
		message        string
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name:           "BadRequest",
			status:         http.StatusBadRequest,
			code:           ErrCodeInvalidRequest,
			message:        "无效的请求",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   ErrCodeInvalidRequest,
			expectedMsg:    "无效的请求",
		},
		{
			name:           "NotFound",
			status:         http.StatusNotFound,
			code:           ErrCodeProviderNotFound,
			message:        "服务商不存在",
			expectedStatus: http.StatusNotFound,
			expectedCode:   ErrCodeProviderNotFound,
			expectedMsg:    "服务商不存在",
		},
		{
			name:           "InternalError",
			status:         http.StatusInternalServerError,
			code:           ErrCodeInternalError,
			message:        "服务器内部错误",
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   ErrCodeInternalError,
			expectedMsg:    "服务器内部错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			ErrorResponse(c, tt.status, tt.code, tt.message)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response APIError
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if response.Code != tt.expectedCode {
				t.Errorf("expected code %s, got %s", tt.expectedCode, response.Code)
			}

			if response.Message != tt.expectedMsg {
				t.Errorf("expected message %s, got %s", tt.expectedMsg, response.Message)
			}
		})
	}
}

func TestErrorResponseWithDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	details := map[string]string{"field": "email"}
	ErrorResponseWithDetails(c, http.StatusBadRequest, ErrCodeMissingField, "缺少必填字段", details)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response APIError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Code != ErrCodeMissingField {
		t.Errorf("expected code %s, got %s", ErrCodeMissingField, response.Code)
	}

	if response.Details == nil {
		t.Error("expected details to be set")
	}
}

func TestShortcutFunctions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("BadRequest", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		BadRequest(c, ErrCodeInvalidRequest, "测试错误")

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		Unauthorized(c, "需要登录")

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("Forbidden", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		Forbidden(c, "没有权限")

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		NotFound(c, ErrCodeProviderNotFound, "资源不存在")

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("InternalError", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		InternalError(c, "服务器错误")

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("ServiceUnavailable", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		ServiceUnavailable(c, "服务不可用")

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
		}
	})

	t.Run("MissingField", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		MissingField(c, "email")

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		var response APIError
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Code != ErrCodeMissingField {
			t.Errorf("expected code %s, got %s", ErrCodeMissingField, response.Code)
		}
	})

	t.Run("InvalidPayload", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		InvalidPayload(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}
