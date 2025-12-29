package api

import (
	"clothing/internal/entity"
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type tagRequest struct {
	Name string `json:"name" binding:"required"`
}

type updateUsageRecordTagsRequest struct {
	TagIDs []uint `json:"tag_ids"`
}

func (h *HTTPHandler) ListTags(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusOK, entity.TagListResponse{Tags: []entity.Tag{}})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	tags, err := h.repo.ListTags(ctx)
	if err != nil {
		logrus.WithError(err).Error("failed to list tags")
		InternalError(c, "加载标签列表失败")
		return
	}

	c.JSON(http.StatusOK, entity.TagListResponse{Tags: h.makeTags(tags)})
}

func (h *HTTPHandler) CreateTag(c *gin.Context) {
	if h.repo == nil {
		ServiceUnavailable(c, "标签服务不可用")
		return
	}

	var req tagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		InvalidPayload(c)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		MissingField(c, "name")
		return
	}

	tag := &entity.DbTag{Name: name}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.repo.CreateTag(ctx, tag); err != nil {
		logrus.WithError(err).Error("failed to create tag")
		InternalError(c, "创建标签失败")
		return
	}

	c.JSON(http.StatusCreated, entity.TagDetailResponse{Tag: h.makeTags([]entity.DbTag{*tag})[0]})
}

func (h *HTTPHandler) UpdateTag(c *gin.Context) {
	if h.repo == nil {
		ServiceUnavailable(c, "标签服务不可用")
		return
	}

	rawID := strings.TrimSpace(c.Param("id"))
	tagID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || tagID == 0 {
		BadRequest(c, ErrCodeInvalidRequest, "无效的标签 ID")
		return
	}

	var req tagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		InvalidPayload(c)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		MissingField(c, "name")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.repo.UpdateTag(ctx, uint(tagID), entity.TagUpdates{Name: &name}); err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, ErrCodeTagNotFound, "标签不存在")
			return
		}
		logrus.WithError(err).Error("failed to update tag")
		InternalError(c, "更新标签失败")
		return
	}

	updated, err := h.repo.FindTagsByIDs(ctx, []uint{uint(tagID)})
	if err != nil || len(updated) == 0 {
		c.JSON(http.StatusOK, entity.TagDetailResponse{Tag: entity.Tag{
			ID:   uint(tagID),
			Name: name,
		}})
		return
	}

	c.JSON(http.StatusOK, entity.TagDetailResponse{Tag: h.makeTags(updated)[0]})
}

func (h *HTTPHandler) DeleteTag(c *gin.Context) {
	if h.repo == nil {
		ServiceUnavailable(c, "标签服务不可用")
		return
	}

	rawID := strings.TrimSpace(c.Param("id"))
	tagID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || tagID == 0 {
		BadRequest(c, ErrCodeInvalidRequest, "无效的标签 ID")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.repo.DeleteTag(ctx, uint(tagID)); err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, ErrCodeTagNotFound, "标签不存在")
			return
		}
		logrus.WithError(err).Error("failed to delete tag")
		InternalError(c, "删除标签失败")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *HTTPHandler) UpdateUsageRecordTags(c *gin.Context) {
	if h.repo == nil {
		ServiceUnavailable(c, "标签服务不可用")
		return
	}

	rawID := strings.TrimSpace(c.Param("id"))
	recordID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || recordID == 0 {
		BadRequest(c, ErrCodeInvalidRequest, "无效的使用记录 ID")
		return
	}

	var req updateUsageRecordTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		InvalidPayload(c)
		return
	}

	tagIDs := deduplicatePositiveIDs(req.TagIDs)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	record, err := h.repo.GetUsageRecord(ctx, uint(recordID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, ErrCodeRecordNotFound, "使用记录不存在")
			return
		}
		logrus.WithError(err).WithField("id", recordID).Error("failed to load usage record")
		InternalError(c, "加载使用记录失败")
		return
	}

	requestUser := CurrentUser(c)
	if requestUser == nil {
		Unauthorized(c, "需要登录")
		return
	}
	if !requestUser.IsAdmin() && record.UserID != requestUser.ID {
		Forbidden(c, "无权访问此记录")
		return
	}

	if err := h.repo.SetUsageRecordTags(ctx, uint(recordID), tagIDs); err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, ErrCodeRecordNotFound, "使用记录不存在")
			return
		}
		logrus.WithError(err).WithField("id", recordID).Error("failed to update usage record tags")
		InternalError(c, "更新使用记录标签失败")
		return
	}

	updatedRecord, err := h.repo.GetUsageRecord(ctx, uint(recordID))
	if err != nil {
		logrus.WithError(err).WithField("id", recordID).Error("failed to reload usage record after tag update")
		InternalError(c, "加载更新后的使用记录失败")
		return
	}

	c.JSON(http.StatusOK, entity.UsageRecordDetailResponse{Record: h.makeUsageRecordItem(*updatedRecord)})
}
