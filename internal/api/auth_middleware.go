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

// RequestUser stores authenticated user details for request context.
type RequestUser struct {
	ID          uint
	Email       string
	DisplayName string
	Role        string
}

// IsAdmin returns true if the user has admin level privileges.
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

// IsSuperAdmin returns true if the user is a super admin.
func (u *RequestUser) IsSuperAdmin() bool {
	if u == nil {
		return false
	}
	return u.Role == entity.UserRoleSuperAdmin
}

// AuthMiddleware enforces JWT authentication.
func (h *HTTPHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		claims, err := h.authManager.ParseToken(tokenString)
		if err != nil {
			logrus.WithError(err).Warn("failed to parse jwt token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		user, err := h.repo.GetUserByID(ctx, claims.UserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
				return
			}
			logrus.WithError(err).WithField("user_id", claims.UserID).Error("failed to load user")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to verify user"})
			return
		}

		if !user.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user is disabled"})
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

// RequireAdmin ensures the current user has admin privileges.
func (h *HTTPHandler) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := CurrentUser(c)
		if user == nil || !user.IsAdmin() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
			return
		}
		c.Next()
	}
}

// CurrentUser fetches the authenticated user from context.
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
