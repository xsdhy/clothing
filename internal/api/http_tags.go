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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load tags"})
		return
	}

	c.JSON(http.StatusOK, entity.TagListResponse{Tags: h.makeTags(tags)})
}

func (h *HTTPHandler) CreateTag(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "tag repository not available"})
		return
	}

	var req tagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag payload"})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag name is required"})
		return
	}

	tag := &entity.DbTag{Name: name}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.repo.CreateTag(ctx, tag); err != nil {
		logrus.WithError(err).Error("failed to create tag")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create tag"})
		return
	}

	c.JSON(http.StatusCreated, entity.TagDetailResponse{Tag: h.makeTags([]entity.DbTag{*tag})[0]})
}

func (h *HTTPHandler) UpdateTag(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "tag repository not available"})
		return
	}

	rawID := strings.TrimSpace(c.Param("id"))
	tagID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || tagID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	var req tagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag payload"})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag name is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.repo.UpdateTag(ctx, uint(tagID), map[string]interface{}{"name": name}); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
			return
		}
		logrus.WithError(err).Error("failed to update tag")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tag"})
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
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "tag repository not available"})
		return
	}

	rawID := strings.TrimSpace(c.Param("id"))
	tagID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || tagID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.repo.DeleteTag(ctx, uint(tagID)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
			return
		}
		logrus.WithError(err).Error("failed to delete tag")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete tag"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *HTTPHandler) UpdateUsageRecordTags(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "tag repository not available"})
		return
	}

	rawID := strings.TrimSpace(c.Param("id"))
	recordID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || recordID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid usage record id"})
		return
	}

	var req updateUsageRecordTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag payload"})
		return
	}

	tagIDs := deduplicatePositiveIDs(req.TagIDs)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	record, err := h.repo.GetUsageRecord(ctx, uint(recordID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "usage record not found"})
			return
		}
		logrus.WithError(err).WithField("id", recordID).Error("failed to load usage record")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load usage record"})
		return
	}

	requestUser := CurrentUser(c)
	if requestUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	if !requestUser.IsAdmin() && record.UserID != requestUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.repo.SetUsageRecordTags(ctx, uint(recordID), tagIDs); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "usage record not found"})
			return
		}
		logrus.WithError(err).WithField("id", recordID).Error("failed to update usage record tags")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to update tags for usage record"})
		return
	}

	updatedRecord, err := h.repo.GetUsageRecord(ctx, uint(recordID))
	if err != nil {
		logrus.WithError(err).WithField("id", recordID).Error("failed to reload usage record after tag update")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load updated usage record"})
		return
	}

	c.JSON(http.StatusOK, entity.UsageRecordDetailResponse{Record: h.makeUsageRecordItem(*updatedRecord)})
}
