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
		ServiceUnavailable(c, "用户服务不可用")
		return
	}

	var query entity.UserQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		InvalidPayload(c)
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
		InternalError(c, "加载用户列表失败")
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
		Forbidden(c, "需要管理员权限")
		return
	}

	var req entity.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		InvalidPayload(c)
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		MissingField(c, "email")
		return
	}

	role := sanitizeRole(req.Role)
	if role == "" {
		BadRequest(c, ErrCodeInvalidRequest, "无效的角色")
		return
	}
	if role == entity.UserRoleSuperAdmin {
		BadRequest(c, ErrCodeInvalidRequest, "无法创建超级管理员")
		return
	}
	if role == entity.UserRoleAdmin && !requestUser.IsSuperAdmin() {
		Forbidden(c, "只有超级管理员才能创建管理员用户")
		return
	}

	password := strings.TrimSpace(req.Password)
	if password == "" {
		MissingField(c, "password")
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		logrus.WithError(err).Error("failed to hash password for new user")
		InternalError(c, "创建用户失败")
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
			BadRequest(c, ErrCodeEmailExists, "邮箱已被注册")
			return
		}
		logrus.WithError(err).Error("failed to create user")
		InternalError(c, "创建用户失败")
		return
	}

	c.JSON(http.StatusCreated, makeUserSummary(user))
}

func (h *HTTPHandler) UpdateUser(c *gin.Context) {
	requestUser := CurrentUser(c)
	if requestUser == nil || !requestUser.IsAdmin() {
		Forbidden(c, "需要管理员权限")
		return
	}

	idValue := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseUint(idValue, 10, 64)
	if err != nil || id == 0 {
		BadRequest(c, ErrCodeInvalidRequest, "无效的用户 ID")
		return
	}

	var req entity.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		InvalidPayload(c)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	dbUser, err := h.repo.GetUserByID(ctx, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NotFound(c, ErrCodeUserNotFound, "用户不存在")
			return
		}
		logrus.WithError(err).Error("failed to load user for update")
		InternalError(c, "更新用户失败")
		return
	}

	if dbUser.Role == entity.UserRoleSuperAdmin && requestUser.ID != dbUser.ID {
		Forbidden(c, "超级管理员不可被修改")
		return
	}

	var updates entity.UserUpdates

	if req.DisplayName != nil {
		displayName := strings.TrimSpace(*req.DisplayName)
		updates.DisplayName = &displayName
	}

	if req.Password != nil {
		password := strings.TrimSpace(*req.Password)
		if password == "" {
			BadRequest(c, ErrCodeMissingField, "密码不能为空")
			return
		}
		hash, err := auth.HashPassword(password)
		if err != nil {
			logrus.WithError(err).Error("failed to hash password for update")
			InternalError(c, "更新用户失败")
			return
		}
		updates.PasswordHash = &hash
	}

	if req.Role != nil {
		if !requestUser.IsSuperAdmin() {
			Forbidden(c, "只有超级管理员才能修改角色")
			return
		}
		targetRole := sanitizeRole(*req.Role)
		if targetRole == "" || targetRole == entity.UserRoleSuperAdmin {
			BadRequest(c, ErrCodeInvalidRequest, "无效的角色")
			return
		}
		updates.Role = &targetRole
	}

	if req.IsActive != nil {
		if dbUser.Role == entity.UserRoleSuperAdmin {
			BadRequest(c, ErrCodeInvalidRequest, "超级管理员必须保持激活状态")
			return
		}
		if dbUser.Role == entity.UserRoleAdmin && !requestUser.IsSuperAdmin() {
			Forbidden(c, "只有超级管理员才能修改管理员状态")
			return
		}
		updates.IsActive = req.IsActive
	}

	if updates.IsEmpty() {
		c.JSON(http.StatusOK, makeUserSummary(dbUser))
		return
	}

	if err := h.repo.UpdateUser(ctx, dbUser.ID, updates); err != nil {
		logrus.WithError(err).Error("failed to update user")
		InternalError(c, "更新用户失败")
		return
	}

	// 刷新用户信息
	updated, err := h.repo.GetUserByID(ctx, dbUser.ID)
	if err != nil {
		logrus.WithError(err).Error("failed to reload user after update")
		InternalError(c, "加载更新后的用户失败")
		return
	}

	c.JSON(http.StatusOK, makeUserSummary(updated))
}

func (h *HTTPHandler) DeleteUser(c *gin.Context) {
	requestUser := CurrentUser(c)
	if requestUser == nil || !requestUser.IsAdmin() {
		Forbidden(c, "需要管理员权限")
		return
	}

	idValue := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseUint(idValue, 10, 64)
	if err != nil || id == 0 {
		BadRequest(c, ErrCodeInvalidRequest, "无效的用户 ID")
		return
	}

	if requestUser.ID == uint(id) {
		BadRequest(c, ErrCodeCannotDeleteSelf, "不能删除当前登录用户")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	dbUser, err := h.repo.GetUserByID(ctx, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NotFound(c, ErrCodeUserNotFound, "用户不存在")
			return
		}
		logrus.WithError(err).Error("failed to load user for deletion")
		InternalError(c, "删除用户失败")
		return
	}

	if dbUser.Role == entity.UserRoleSuperAdmin {
		Forbidden(c, "超级管理员不可被删除")
		return
	}

	if dbUser.Role == entity.UserRoleAdmin && !requestUser.IsSuperAdmin() {
		Forbidden(c, "只有超级管理员才能删除管理员用户")
		return
	}

	if err := h.repo.DeleteUser(ctx, uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NotFound(c, ErrCodeUserNotFound, "用户不存在")
			return
		}
		logrus.WithError(err).Error("failed to delete user")
		InternalError(c, "删除用户失败")
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
