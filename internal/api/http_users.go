package api

import (
	"clothing/internal/auth"
	"clothing/internal/entity"
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (h *HTTPHandler) ListUsers(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "user repository not available"})
		return
	}

	var query entity.UserQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	users, meta, err := h.repo.ListUsers(ctx, &query)
	if err != nil {
		logrus.WithError(err).Error("failed to list users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load users"})
		return
	}

	response := entity.UserListResponse{
		Users: make([]entity.UserSummary, 0, len(users)),
		Meta:  meta,
	}
	for idx := range users {
		response.Users = append(response.Users, makeUserSummary(&users[idx]))
	}

	c.JSON(http.StatusOK, response)
}

func (h *HTTPHandler) CreateUser(c *gin.Context) {
	requestUser := CurrentUser(c)
	if requestUser == nil || !requestUser.IsAdmin() {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
		return
	}

	var req entity.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user payload"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	role := sanitizeRole(req.Role)
	if role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}
	if role == entity.UserRoleSuperAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot create super admin"})
		return
	}
	if role == entity.UserRoleAdmin && !requestUser.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, gin.H{"error": "only super admin can create admin users"})
		return
	}

	password := strings.TrimSpace(req.Password)
	if password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		logrus.WithError(err).Error("failed to hash password for new user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	user := &entity.DbUser{
		Email:        email,
		DisplayName:  strings.TrimSpace(req.DisplayName),
		PasswordHash: hash,
		Role:         role,
		IsActive:     isActive,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.repo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email already registered"})
			return
		}
		logrus.WithError(err).Error("failed to create user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, makeUserSummary(user))
}

func (h *HTTPHandler) UpdateUser(c *gin.Context) {
	requestUser := CurrentUser(c)
	if requestUser == nil || !requestUser.IsAdmin() {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
		return
	}

	idValue := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseUint(idValue, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req entity.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user payload"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	dbUser, err := h.repo.GetUserByID(ctx, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		logrus.WithError(err).Error("failed to load user for update")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	if dbUser.Role == entity.UserRoleSuperAdmin && requestUser.ID != dbUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "super admin cannot be modified"})
		return
	}

	updates := make(map[string]interface{})

	if req.DisplayName != nil {
		updates["display_name"] = strings.TrimSpace(*req.DisplayName)
	}

	if req.Password != nil {
		password := strings.TrimSpace(*req.Password)
		if password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "password must not be empty"})
			return
		}
		hash, err := auth.HashPassword(password)
		if err != nil {
			logrus.WithError(err).Error("failed to hash password for update")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
			return
		}
		updates["password_hash"] = hash
	}

	if req.Role != nil {
		if !requestUser.IsSuperAdmin() {
			c.JSON(http.StatusForbidden, gin.H{"error": "only super admin can change roles"})
			return
		}
		targetRole := sanitizeRole(*req.Role)
		if targetRole == "" || targetRole == entity.UserRoleSuperAdmin {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
			return
		}
		updates["role"] = targetRole
	}

	if req.IsActive != nil {
		if dbUser.Role == entity.UserRoleSuperAdmin {
			c.JSON(http.StatusBadRequest, gin.H{"error": "super admin must remain active"})
			return
		}
		if dbUser.Role == entity.UserRoleAdmin && !requestUser.IsSuperAdmin() {
			c.JSON(http.StatusForbidden, gin.H{"error": "only super admin can change admin status"})
			return
		}
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, makeUserSummary(dbUser))
		return
	}

	if err := h.repo.UpdateUser(ctx, dbUser.ID, updates); err != nil {
		logrus.WithError(err).Error("failed to update user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	// Refresh
	updated, err := h.repo.GetUserByID(ctx, dbUser.ID)
	if err != nil {
		logrus.WithError(err).Error("failed to reload user after update")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load updated user"})
		return
	}

	c.JSON(http.StatusOK, makeUserSummary(updated))
}

func (h *HTTPHandler) DeleteUser(c *gin.Context) {
	requestUser := CurrentUser(c)
	if requestUser == nil || !requestUser.IsAdmin() {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
		return
	}

	idValue := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseUint(idValue, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if requestUser.ID == uint(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete current user"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	dbUser, err := h.repo.GetUserByID(ctx, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		logrus.WithError(err).Error("failed to load user for deletion")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	if dbUser.Role == entity.UserRoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "super admin cannot be deleted"})
		return
	}

	if dbUser.Role == entity.UserRoleAdmin && !requestUser.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, gin.H{"error": "only super admin can delete admin user"})
		return
	}

	if err := h.repo.DeleteUser(ctx, uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		logrus.WithError(err).Error("failed to delete user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.Status(http.StatusNoContent)
}

func sanitizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case entity.UserRoleAdmin:
		return entity.UserRoleAdmin
	case entity.UserRoleUser:
		return entity.UserRoleUser
	default:
		return ""
	}
}
