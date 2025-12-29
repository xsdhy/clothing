package api

import (
	"clothing/internal/auth"
	"clothing/internal/entity"
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (h *HTTPHandler) Register(c *gin.Context) {
	if h.repo == nil {
		ServiceUnavailable(c, "用户服务不可用")
		return
	}

	var req entity.AuthRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		InvalidPayload(c)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	count, err := h.repo.CountUsers(ctx)
	if err != nil {
		logrus.WithError(err).Error("failed to count users during registration")
		InternalError(c, "注册处理失败")
		return
	}

	if count > 0 {
		ErrorResponse(c, http.StatusForbidden, ErrCodeRegistrationClosed, "注册功能已关闭")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)

	if email == "" {
		MissingField(c, "email")
		return
	}
	if password == "" {
		MissingField(c, "password")
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		logrus.WithError(err).Error("failed to hash password")
		InternalError(c, "注册用户失败")
		return
	}

	user := &entity.DbUser{
		Email:        email,
		PasswordHash: hash,
		DisplayName:  strings.TrimSpace(req.DisplayName),
		Role:         entity.UserRoleSuperAdmin,
		IsActive:     true,
	}

	if err := h.repo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			BadRequest(c, ErrCodeEmailExists, "邮箱已被注册")
			return
		}
		logrus.WithError(err).Error("failed to create initial user")
		InternalError(c, "注册用户失败")
		return
	}

	token, expiresAt, err := h.authManager.GenerateToken(user)
	if err != nil {
		logrus.WithError(err).Error("failed to create token for user")
		InternalError(c, "创建会话失败")
		return
	}

	c.JSON(http.StatusCreated, entity.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      makeUserSummary(user),
	})
}

func (h *HTTPHandler) Login(c *gin.Context) {
	if h.repo == nil {
		ServiceUnavailable(c, "用户服务不可用")
		return
	}

	var req entity.AuthLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		InvalidPayload(c)
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)
	if email == "" {
		MissingField(c, "email")
		return
	}
	if password == "" {
		MissingField(c, "password")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	user, err := h.repo.GetUserByEmail(ctx, email)
	if err != nil {
		logrus.WithError(err).WithField("email", email).Warn("login attempt failed")
		ErrorResponse(c, http.StatusUnauthorized, ErrCodeInvalidCredentials, "邮箱或密码错误")
		return
	}

	if !user.IsActive {
		ErrorResponse(c, http.StatusForbidden, ErrCodeUserDisabled, "账户已被禁用")
		return
	}

	if err := auth.VerifyPassword(user.PasswordHash, password); err != nil {
		logrus.WithError(err).WithField("email", email).Warn("password verification failed")
		ErrorResponse(c, http.StatusUnauthorized, ErrCodeInvalidCredentials, "邮箱或密码错误")
		return
	}

	token, expiresAt, err := h.authManager.GenerateToken(user)
	if err != nil {
		logrus.WithError(err).Error("failed to generate token")
		InternalError(c, "创建会话失败")
		return
	}

	c.JSON(http.StatusOK, entity.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      makeUserSummary(user),
	})
}

func (h *HTTPHandler) AuthStatus(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusOK, entity.AuthStatusResponse{HasUser: false})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	count, err := h.repo.CountUsers(ctx)
	if err != nil {
		logrus.WithError(err).Error("failed to count users for auth status")
		InternalError(c, "检查认证状态失败")
		return
	}
	c.JSON(http.StatusOK, entity.AuthStatusResponse{HasUser: count > 0})
}

func (h *HTTPHandler) Me(c *gin.Context) {
	user := CurrentUser(c)
	if user == nil {
		Unauthorized(c, "需要登录")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	dbUser, err := h.repo.GetUserByID(ctx, user.ID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", user.ID).Error("failed to load user profile")
		InternalError(c, "加载用户信息失败")
		return
	}

	c.JSON(http.StatusOK, makeUserSummary(dbUser))
}

func makeUserSummary(user *entity.DbUser) entity.UserSummary {
	if user == nil {
		return entity.UserSummary{}
	}
	return entity.UserSummary{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
