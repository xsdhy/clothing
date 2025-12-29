package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"clothing/internal/entity"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	currentUserContextKey = "current-user"
)

// RequestUser 存储请求上下文中的认证用户信息
type RequestUser struct {
	ID          uint
	Email       string
	DisplayName string
	Role        string
}

// IsAdmin 判断用户是否具有管理员权限
func (u *RequestUser) IsAdmin() bool {
	if u == nil {
		return false
	}
	switch u.Role {
	case entity.UserRoleAdmin, entity.UserRoleSuperAdmin:
		return true
	default:
		return false
	}
}

// IsSuperAdmin 判断用户是否为超级管理员
func (u *RequestUser) IsSuperAdmin() bool {
	if u == nil {
		return false
	}
	return u.Role == entity.UserRoleSuperAdmin
}

// AuthMiddleware JWT 认证中间件
func (h *HTTPHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, APIError{
				Code:    ErrCodeUnauthorized,
				Message: "缺少授权头",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, APIError{
				Code:    ErrCodeUnauthorized,
				Message: "无效的授权头格式",
			})
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, APIError{
				Code:    ErrCodeUnauthorized,
				Message: "缺少 Bearer Token",
			})
			return
		}

		claims, err := h.authManager.ParseToken(tokenString)
		if err != nil {
			logrus.WithError(err).Warn("failed to parse jwt token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, APIError{
				Code:    ErrCodeSessionExpired,
				Message: "Token 无效或已过期",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		user, err := h.repo.GetUserByID(ctx, claims.UserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, APIError{
					Code:    ErrCodeUserNotFound,
					Message: "用户不存在",
				})
				return
			}
			logrus.WithError(err).WithField("user_id", claims.UserID).Error("failed to load user")
			c.AbortWithStatusJSON(http.StatusInternalServerError, APIError{
				Code:    ErrCodeInternalError,
				Message: "验证用户失败",
			})
			return
		}

		if !user.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, APIError{
				Code:    ErrCodeUserDisabled,
				Message: "账户已被禁用",
			})
			return
		}

		requestUser := &RequestUser{
			ID:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Role:        user.Role,
		}

		c.Set(currentUserContextKey, requestUser)
		c.Next()
	}
}

// RequireAdmin 管理员权限守卫中间件
func (h *HTTPHandler) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := CurrentUser(c)
		if user == nil || !user.IsAdmin() {
			c.AbortWithStatusJSON(http.StatusForbidden, APIError{
				Code:    ErrCodeForbidden,
				Message: "需要管理员权限",
			})
			return
		}
		c.Next()
	}
}

// CurrentUser 从上下文获取当前认证用户
func CurrentUser(c *gin.Context) *RequestUser {
	value, exists := c.Get(currentUserContextKey)
	if !exists {
		return nil
	}
	user, ok := value.(*RequestUser)
	if !ok {
		return nil
	}
	return user
}
