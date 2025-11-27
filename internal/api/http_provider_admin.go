package api

import (
	"clothing/internal/entity"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var providerIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

func normaliseProviderID(raw string) (string, error) {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "" {
		return "", errors.New("provider id is required")
	}
	if !providerIDPattern.MatchString(trimmed) {
		return "", errors.New("provider id must contain only lowercase letters, numbers, hyphen, or underscore")
	}
	return trimmed, nil
}

func normaliseStringSlice(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func normaliseIntSlice(values []int) []int {
	out := make([]int, 0, len(values))
	seen := make(map[int]struct{})
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func (h *HTTPHandler) AdminListProviders(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provider repository not configured"})
		return
	}

	ctx := c.Request.Context()
	providers, err := h.repo.ListProviders(ctx, true)
	if err != nil {
		logrus.WithError(err).Error("failed to list providers")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list providers"})
		return
	}

	views := make([]entity.ProviderAdminView, 0, len(providers))
	for _, provider := range providers {
		views = append(views, provider.ToAdminView(true))
	}
	c.JSON(http.StatusOK, gin.H{"providers": views})
}

func (h *HTTPHandler) CreateProvider(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provider repository not configured"})
		return
	}

	var payload entity.CreateProviderRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	id, err := normaliseProviderID(payload.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(payload.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	driver := strings.TrimSpace(strings.ToLower(payload.Driver))
	if driver == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "driver is required"})
		return
	}

	isActive := true
	if payload.IsActive != nil {
		isActive = *payload.IsActive
	}

	provider := &entity.DbProvider{
		ID:          id,
		Name:        name,
		Driver:      driver,
		Description: strings.TrimSpace(payload.Description),
		APIKey:      strings.TrimSpace(payload.APIKey),
		BaseURL:     strings.TrimSpace(payload.BaseURL),
		IsActive:    isActive,
	}
	if len(payload.Config) > 0 {
		provider.Config = entity.JSONMap(payload.Config)
	}

	ctx := c.Request.Context()
	if err := h.repo.CreateProvider(ctx, provider); err != nil {
		logrus.WithError(err).WithField("provider_id", id).Error("failed to create provider")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"provider": provider.ToAdminView(false)})
}

func (h *HTTPHandler) GetProviderDetail(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provider repository not configured"})
		return
	}

	id, err := normaliseProviderID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	provider, err := h.repo.GetProvider(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
			return
		}
		logrus.WithError(err).WithField("provider_id", id).Error("failed to load provider")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"provider": provider.ToAdminView(true)})
}

func (h *HTTPHandler) UpdateProvider(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provider repository not configured"})
		return
	}

	id, err := normaliseProviderID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var payload entity.UpdateProviderRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	updates := make(map[string]interface{})

	if payload.Name != nil {
		trimmed := strings.TrimSpace(*payload.Name)
		if trimmed == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name cannot be empty"})
			return
		}
		updates["name"] = trimmed
	}
	if payload.Driver != nil {
		driver := strings.TrimSpace(strings.ToLower(*payload.Driver))
		if driver == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "driver cannot be empty"})
			return
		}
		updates["driver"] = driver
	}
	if payload.Description != nil {
		updates["description"] = strings.TrimSpace(*payload.Description)
	}
	if payload.APIKey != nil {
		updates["api_key"] = strings.TrimSpace(*payload.APIKey)
	}
	if payload.BaseURL != nil {
		updates["base_url"] = strings.TrimSpace(*payload.BaseURL)
	}
	if payload.Config != nil {
		updates["config"] = entity.JSONMap(payload.Config)
	}
	if payload.IsActive != nil {
		updates["is_active"] = *payload.IsActive
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "no changes supplied"})
		return
	}

	ctx := c.Request.Context()
	if err := h.repo.UpdateProvider(ctx, id, updates); err != nil {
		logrus.WithError(err).WithField("provider_id", id).Error("failed to update provider")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider, err := h.repo.GetProvider(ctx, id)
	if err != nil {
		logrus.WithError(err).WithField("provider_id", id).Error("failed to reload provider after update")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"provider": provider.ToAdminView(true)})
}

func (h *HTTPHandler) DeleteProvider(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provider repository not configured"})
		return
	}

	id, err := normaliseProviderID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.repo.DeleteProvider(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
			return
		}
		logrus.WithError(err).WithField("provider_id", id).Error("failed to delete provider")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete provider"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *HTTPHandler) ListProviderModels(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provider repository not configured"})
		return
	}

	id, err := normaliseProviderID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	provider, err := h.repo.GetProvider(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
			return
		}
		logrus.WithError(err).WithField("provider_id", id).Error("failed to load provider for model listing")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list models"})
		return
	}

	results := make([]entity.ProviderModelSummary, 0, len(provider.Models))
	for _, model := range provider.Models {
		results = append(results, model.ToAdminView())
	}
	c.JSON(http.StatusOK, gin.H{"models": results})
}

func (h *HTTPHandler) CreateProviderModel(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provider repository not configured"})
		return
	}

	id, err := normaliseProviderID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var payload entity.CreateModelRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	modelID := strings.TrimSpace(payload.ModelID)
	if modelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model_id is required"})
		return
	}

	name := strings.TrimSpace(payload.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	isActive := true
	if payload.IsActive != nil {
		isActive = *payload.IsActive
	}

	model := &entity.DbModel{
		ProviderID:         id,
		ModelID:            modelID,
		Name:               name,
		Description:        strings.TrimSpace(payload.Description),
		Price:              strings.TrimSpace(payload.Price),
		MaxImages:          0,
		Modalities:         entity.StringArray(normaliseStringSlice(payload.Modalities)),
		SupportedSizes:     entity.StringArray(normaliseStringSlice(payload.SupportedSizes)),
		SupportedDurations: entity.IntArray(normaliseIntSlice(payload.SupportedDurations)),
		DefaultSize:        strings.TrimSpace(payload.DefaultSize),
		DefaultDuration:    0,
		IsActive:           isActive,
	}
	if payload.MaxImages != nil {
		model.MaxImages = *payload.MaxImages
	}
	if payload.DefaultDuration != nil {
		model.DefaultDuration = *payload.DefaultDuration
	}
	if payload.Settings != nil {
		model.Settings = entity.JSONMap(payload.Settings)
	}

	ctx := c.Request.Context()
	if _, err := h.repo.GetProvider(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
			return
		}
		logrus.WithError(err).WithField("provider_id", id).Error("failed to load provider before creating model")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create model"})
		return
	}

	if err := h.repo.CreateModel(ctx, model); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider_id": id,
			"model_id":    modelID,
		}).Error("failed to create provider model")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"model": model.ToAdminView()})
}

func (h *HTTPHandler) UpdateProviderModel(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provider repository not configured"})
		return
	}

	id, err := normaliseProviderID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	modelID := strings.TrimSpace(c.Param("model_id"))
	if modelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model_id is required"})
		return
	}

	var payload entity.UpdateModelRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	updates := make(map[string]interface{})

	if payload.Name != nil {
		trimmed := strings.TrimSpace(*payload.Name)
		if trimmed == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name cannot be empty"})
			return
		}
		updates["name"] = trimmed
	}
	if payload.Description != nil {
		updates["description"] = strings.TrimSpace(*payload.Description)
	}
	if payload.Price != nil {
		updates["price"] = strings.TrimSpace(*payload.Price)
	}
	if payload.MaxImages != nil {
		updates["max_images"] = *payload.MaxImages
	}
	if payload.Modalities != nil {
		updates["modalities"] = entity.StringArray(normaliseStringSlice(*payload.Modalities))
	}
	if payload.SupportedSizes != nil {
		updates["supported_sizes"] = entity.StringArray(normaliseStringSlice(*payload.SupportedSizes))
	}
	if payload.SupportedDurations != nil {
		updates["supported_durations"] = entity.IntArray(normaliseIntSlice(*payload.SupportedDurations))
	}
	if payload.DefaultSize != nil {
		updates["default_size"] = strings.TrimSpace(*payload.DefaultSize)
	}
	if payload.DefaultDuration != nil {
		updates["default_duration"] = *payload.DefaultDuration
	}
	if payload.Settings != nil {
		updates["settings"] = entity.JSONMap(payload.Settings)
	}
	if payload.IsActive != nil {
		updates["is_active"] = *payload.IsActive
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "no changes supplied"})
		return
	}

	ctx := c.Request.Context()
	if err := h.repo.UpdateModel(ctx, id, modelID, updates); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
			return
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider_id": id,
			"model_id":    modelID,
		}).Error("failed to update provider model")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, model, err := h.repo.GetProviderWithModel(ctx, id, modelID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
			return
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider_id": id,
			"model_id":    modelID,
		}).Error("failed to reload model after update")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load model"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"model": model.ToAdminView()})
}

func (h *HTTPHandler) DeleteProviderModel(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provider repository not configured"})
		return
	}

	id, err := normaliseProviderID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	modelID := strings.TrimSpace(c.Param("model_id"))
	if modelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model_id is required"})
		return
	}

	ctx := c.Request.Context()
	if err := h.repo.DeleteModel(ctx, id, modelID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
			return
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider_id": id,
			"model_id":    modelID,
		}).Error("failed to delete provider model")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete model"})
		return
	}

	c.Status(http.StatusNoContent)
}
