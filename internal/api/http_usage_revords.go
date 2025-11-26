package api

import (
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

func (h *HTTPHandler) ListUsageRecords(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusOK, entity.UsageRecordListResponse{Records: []entity.UsageRecordItem{}, Meta: &entity.Meta{Page: 1, PageSize: 0, Total: 0}})
		return
	}

	requestUser := CurrentUser(c)
	if requestUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var params entity.UsageRecordQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}
	params.TagIDs = parseUintListParam(
		append(c.QueryArray("tags"), c.QueryArray("tag_ids")...),
		c.Query("tags"),
		c.Query("tag_ids"),
	)
	if !params.HasOutputImages {
		params.HasOutputImages = parseBoolParam(
			c.Query("has_output_images"),
			c.Query("has_images"),
		)
	}

	params.Result = strings.ToLower(strings.TrimSpace(params.Result))
	if params.Result == "" {
		params.Result = "success"
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	if requestUser.IsAdmin() {
		params.IncludeAll = true
		if userFilter := strings.TrimSpace(c.Query("user_id")); userFilter != "" {
			if parsed, err := strconv.ParseUint(userFilter, 10, 64); err == nil && parsed > 0 {
				params.UserID = uint(parsed)
				params.IncludeAll = false
			}
		}
	} else {
		params.UserID = requestUser.ID
		params.IncludeAll = false
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	records, meta, err := h.repo.ListUsageRecords(ctx, &params)
	if err != nil {
		logrus.WithError(err).Error("failed to list usage records")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load usage records"})
		return
	}

	items := make([]entity.UsageRecordItem, 0, len(records))
	for _, record := range records {
		items = append(items, h.makeUsageRecordItem(record))
	}

	if meta == nil {
		meta = &entity.Meta{Page: int64(params.Page), PageSize: int64(params.PageSize), Total: int64(len(items))}
	}

	c.JSON(http.StatusOK, entity.UsageRecordListResponse{Records: items, Meta: meta})
}

func (h *HTTPHandler) makeUsageImages(paths []string) []entity.UsageImage {
	if len(paths) == 0 {
		return []entity.UsageImage{}
	}
	items := make([]entity.UsageImage, 0, len(paths))
	for _, p := range paths {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		items = append(items, entity.UsageImage{
			Path: trimmed,
			URL:  h.publicURL(trimmed),
		})
	}
	if len(items) == 0 {
		return []entity.UsageImage{}
	}
	return items
}

func (h *HTTPHandler) makeTags(tags []entity.DbTag) []entity.Tag {
	if len(tags) == 0 {
		return []entity.Tag{}
	}
	result := make([]entity.Tag, 0, len(tags))
	for _, tag := range tags {
		result = append(result, entity.Tag{
			ID:         tag.ID,
			Name:       tag.Name,
			UsageCount: tag.UsageCount,
			CreatedAt:  tag.CreatedAt,
			UpdatedAt:  tag.UpdatedAt,
		})
	}
	return result
}

func (h *HTTPHandler) makeUsageRecordItem(record entity.DbUsageRecord) entity.UsageRecordItem {
	return entity.UsageRecordItem{
		ID:           record.ID,
		ProviderID:   record.ProviderID,
		ModelID:      record.ModelID,
		Prompt:       record.Prompt,
		Size:         record.Size,
		OutputText:   record.OutputText,
		ErrorMessage: record.ErrorMessage,
		CreatedAt:    record.CreatedAt,
		InputImages:  h.makeUsageImages(record.InputImages.ToSlice()),
		OutputImages: h.makeUsageImages(record.OutputImages.ToSlice()),
		User:         makeUserSummary(record.User),
		Tags:         h.makeTags(record.Tags),
	}
}

func (h *HTTPHandler) GetUsageRecord(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "usage record repository not available"})
		return
	}

	idValue := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseUint(idValue, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid usage record id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	record, err := h.repo.GetUsageRecord(ctx, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "usage record not found"})
			return
		}
		logrus.WithError(err).WithField("id", id).Error("failed to load usage record")
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

	item := h.makeUsageRecordItem(*record)
	c.JSON(http.StatusOK, entity.UsageRecordDetailResponse{Record: item})
}

func (h *HTTPHandler) DeleteUsageRecord(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "usage record repository not available"})
		return
	}

	idValue := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseUint(idValue, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid usage record id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	requestUser := CurrentUser(c)
	if requestUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	record, err := h.repo.GetUsageRecord(ctx, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "usage record not found"})
			return
		}
		logrus.WithError(err).WithField("id", id).Error("failed to load usage record for deletion")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete usage record"})
		return
	}

	if !requestUser.IsAdmin() && record.UserID != requestUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.repo.DeleteUsageRecord(ctx, uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "usage record not found"})
			return
		}
		logrus.WithError(err).WithField("id", id).Error("failed to delete usage record")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete usage record"})
		return
	}

	c.Status(http.StatusNoContent)
}

func parseUintListParam(values []string, fallbacks ...string) []uint {
	items := make([]string, 0, len(values)+1)
	for _, val := range values {
		if trimmed := strings.TrimSpace(val); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	for _, fallback := range fallbacks {
		if trimmed := strings.TrimSpace(fallback); trimmed != "" {
			items = append(items, trimmed)
		}
	}

	result := make([]uint, 0, len(items))
	seen := make(map[uint]struct{}, len(items))
	for _, item := range items {
		parts := strings.Split(item, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed == "" {
				continue
			}
			if parsed, err := strconv.ParseUint(trimmed, 10, 64); err == nil && parsed > 0 {
				id := uint(parsed)
				if _, ok := seen[id]; !ok {
					seen[id] = struct{}{}
					result = append(result, id)
				}
			}
		}
	}
	return result
}

func parseBoolParam(values ...string) bool {
	for _, raw := range values {
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
		}
	}
	return false
}
