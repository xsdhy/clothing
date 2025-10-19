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
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "user repository not available"})
		return
	}

	var req entity.AuthRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid registration payload"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	count, err := h.repo.CountUsers(ctx)
	if err != nil {
		logrus.WithError(err).Error("failed to count users during registration")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process registration"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "registration disabled"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)

	if email == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		logrus.WithError(err).Error("failed to hash password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "email already registered"})
			return
		}
		logrus.WithError(err).Error("failed to create initial user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	token, expiresAt, err := h.authManager.GenerateToken(user)
	if err != nil {
		logrus.WithError(err).Error("failed to create token for user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
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
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "user repository not available"})
		return
	}

	var req entity.AuthLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid login payload"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)
	if email == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	user, err := h.repo.GetUserByEmail(ctx, email)
	if err != nil {
		logrus.WithError(err).WithField("email", email).Warn("login attempt failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "user is disabled"})
		return
	}

	if err := auth.VerifyPassword(user.PasswordHash, password); err != nil {
		logrus.WithError(err).WithField("email", email).Warn("password verification failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	token, expiresAt, err := h.authManager.GenerateToken(user)
	if err != nil {
		logrus.WithError(err).Error("failed to generate token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check auth status"})
		return
	}
	c.JSON(http.StatusOK, entity.AuthStatusResponse{HasUser: count > 0})
}

func (h *HTTPHandler) Me(c *gin.Context) {
	user := CurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	dbUser, err := h.repo.GetUserByID(ctx, user.ID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", user.ID).Error("failed to load user profile")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
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
