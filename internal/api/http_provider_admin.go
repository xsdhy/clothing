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
		return "", errors.New("服务商 ID 不能为空")
	}
	if !providerIDPattern.MatchString(trimmed) {
		return "", errors.New("服务商 ID 只能包含小写字母、数字、连字符或下划线")
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
		InternalError(c, "服务商仓储未配置")
		return
	}

	ctx := c.Request.Context()
	providers, err := h.repo.ListProviders(ctx, true)
	if err != nil {
		logrus.WithError(err).Error("failed to list providers")
		InternalError(c, "加载服务商列表失败")
		return
	}

	views := make([]entity.ProviderAdminView, 0, len(providers))
	for _, provider := range providers {
		views = append(views, entity.ProviderToAdminView(provider, true))
	}
	c.JSON(http.StatusOK, gin.H{"providers": views})
}

func (h *HTTPHandler) CreateProvider(c *gin.Context) {
	if h.repo == nil {
		InternalError(c, "服务商仓储未配置")
		return
	}

	var payload entity.CreateProviderRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		InvalidPayload(c)
		return
	}

	id, err := normaliseProviderID(payload.ID)
	if err != nil {
		BadRequest(c, ErrCodeInvalidRequest, err.Error())
		return
	}

	name := strings.TrimSpace(payload.Name)
	if name == "" {
		MissingField(c, "name")
		return
	}

	driver := strings.TrimSpace(strings.ToLower(payload.Driver))
	if driver == "" {
		MissingField(c, "driver")
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
		InternalError(c, "创建服务商失败: "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"provider": entity.ProviderToAdminView(*provider, false)})
}

func (h *HTTPHandler) GetProviderDetail(c *gin.Context) {
	if h.repo == nil {
		InternalError(c, "服务商仓储未配置")
		return
	}

	id, err := normaliseProviderID(c.Param("id"))
	if err != nil {
		BadRequest(c, ErrCodeInvalidRequest, err.Error())
		return
	}

	ctx := c.Request.Context()
	provider, err := h.repo.GetProvider(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NotFound(c, ErrCodeProviderNotFound, "服务商不存在")
			return
		}
		logrus.WithError(err).WithField("provider_id", id).Error("failed to load provider")
		InternalError(c, "加载服务商失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"provider": entity.ProviderToAdminView(*provider, true)})
}

func (h *HTTPHandler) UpdateProvider(c *gin.Context) {
	if h.repo == nil {
		InternalError(c, "服务商仓储未配置")
		return
	}

	id, err := normaliseProviderID(c.Param("id"))
	if err != nil {
		BadRequest(c, ErrCodeInvalidRequest, err.Error())
		return
	}

	var payload entity.UpdateProviderRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		InvalidPayload(c)
		return
	}

	var updates entity.ProviderUpdates

	if payload.Name != nil {
		trimmed := strings.TrimSpace(*payload.Name)
		if trimmed == "" {
			BadRequest(c, ErrCodeMissingField, "名称不能为空")
			return
		}
		updates.Name = &trimmed
	}
	if payload.Driver != nil {
		driver := strings.TrimSpace(strings.ToLower(*payload.Driver))
		if driver == "" {
			BadRequest(c, ErrCodeMissingField, "驱动不能为空")
			return
		}
		updates.Driver = &driver
	}
	if payload.Description != nil {
		description := strings.TrimSpace(*payload.Description)
		updates.Description = &description
	}
	if payload.APIKey != nil {
		apiKey := strings.TrimSpace(*payload.APIKey)
		updates.APIKey = &apiKey
	}
	if payload.BaseURL != nil {
		baseURL := strings.TrimSpace(*payload.BaseURL)
		updates.BaseURL = &baseURL
	}
	if payload.Config != nil {
		config := entity.JSONMap(payload.Config)
		updates.Config = &config
	}
	if payload.IsActive != nil {
		updates.IsActive = payload.IsActive
	}

	if updates.IsEmpty() {
		c.JSON(http.StatusOK, gin.H{"message": "无更新内容"})
		return
	}

	ctx := c.Request.Context()
	if err := h.repo.UpdateProvider(ctx, id, updates); err != nil {
		logrus.WithError(err).WithField("provider_id", id).Error("failed to update provider")
		InternalError(c, "更新服务商失败: "+err.Error())
		return
	}

	provider, err := h.repo.GetProvider(ctx, id)
	if err != nil {
		logrus.WithError(err).WithField("provider_id", id).Error("failed to reload provider after update")
		InternalError(c, "加载服务商失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"provider": entity.ProviderToAdminView(*provider, true)})
}

func (h *HTTPHandler) DeleteProvider(c *gin.Context) {
	if h.repo == nil {
		InternalError(c, "服务商仓储未配置")
		return
	}

	id, err := normaliseProviderID(c.Param("id"))
	if err != nil {
		BadRequest(c, ErrCodeInvalidRequest, err.Error())
		return
	}

	ctx := c.Request.Context()
	if err := h.repo.DeleteProvider(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NotFound(c, ErrCodeProviderNotFound, "服务商不存在")
			return
		}
		logrus.WithError(err).WithField("provider_id", id).Error("failed to delete provider")
		InternalError(c, "删除服务商失败")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *HTTPHandler) ListProviderModels(c *gin.Context) {
	if h.repo == nil {
		InternalError(c, "服务商仓储未配置")
		return
	}

	id, err := normaliseProviderID(c.Param("id"))
	if err != nil {
		BadRequest(c, ErrCodeInvalidRequest, err.Error())
		return
	}

	ctx := c.Request.Context()
	provider, err := h.repo.GetProvider(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NotFound(c, ErrCodeProviderNotFound, "服务商不存在")
			return
		}
		logrus.WithError(err).WithField("provider_id", id).Error("failed to load provider for model listing")
		InternalError(c, "加载模型列表失败")
		return
	}

	results := make([]entity.ProviderModelSummary, 0, len(provider.Models))
	for _, model := range provider.Models {
		results = append(results, entity.ModelToAdminView(model))
	}
	c.JSON(http.StatusOK, gin.H{"models": results})
}

func (h *HTTPHandler) CreateProviderModel(c *gin.Context) {
	if h.repo == nil {
		InternalError(c, "服务商仓储未配置")
		return
	}

	id, err := normaliseProviderID(c.Param("id"))
	if err != nil {
		BadRequest(c, ErrCodeInvalidRequest, err.Error())
		return
	}

	var payload entity.CreateModelRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		InvalidPayload(c)
		return
	}

	modelID := strings.TrimSpace(payload.ModelID)
	if modelID == "" {
		MissingField(c, "model_id")
		return
	}

	name := strings.TrimSpace(payload.Name)
	if name == "" {
		MissingField(c, "name")
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
		InputModalities:    entity.StringArray(normaliseStringSlice(payload.InputModalities)),
		OutputModalities:   entity.StringArray(normaliseStringSlice(payload.OutputModalities)),
		SupportedSizes:     entity.StringArray(normaliseStringSlice(payload.SupportedSizes)),
		SupportedDurations: entity.IntArray(normaliseIntSlice(payload.SupportedDurations)),
		DefaultSize:        strings.TrimSpace(payload.DefaultSize),
		DefaultDuration:    0,
		GenerationMode:     strings.TrimSpace(payload.GenerationMode),
		EndpointPath:       strings.TrimSpace(payload.EndpointPath),
		SupportsStreaming:  false,
		SupportsCancel:     false,
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
	if payload.SupportsStreaming != nil {
		model.SupportsStreaming = *payload.SupportsStreaming
	}
	if payload.SupportsCancel != nil {
		model.SupportsCancel = *payload.SupportsCancel
	}

	ctx := c.Request.Context()
	if _, err := h.repo.GetProvider(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NotFound(c, ErrCodeProviderNotFound, "服务商不存在")
			return
		}
		logrus.WithError(err).WithField("provider_id", id).Error("failed to load provider before creating model")
		InternalError(c, "创建模型失败")
		return
	}

	if err := h.repo.CreateModel(ctx, model); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider_id": id,
			"model_id":    modelID,
		}).Error("failed to create provider model")
		InternalError(c, "创建模型失败: "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"model": entity.ModelToAdminView(*model)})
}

func (h *HTTPHandler) UpdateProviderModel(c *gin.Context) {
	if h.repo == nil {
		InternalError(c, "服务商仓储未配置")
		return
	}

	id, err := normaliseProviderID(c.Param("id"))
	if err != nil {
		BadRequest(c, ErrCodeInvalidRequest, err.Error())
		return
	}

	modelID := strings.TrimSpace(c.Param("model_id"))
	if modelID == "" {
		MissingField(c, "model_id")
		return
	}

	var payload entity.UpdateModelRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		InvalidPayload(c)
		return
	}

	var updates entity.ModelUpdates

	if payload.Name != nil {
		trimmed := strings.TrimSpace(*payload.Name)
		if trimmed == "" {
			BadRequest(c, ErrCodeMissingField, "名称不能为空")
			return
		}
		updates.Name = &trimmed
	}
	if payload.Description != nil {
		description := strings.TrimSpace(*payload.Description)
		updates.Description = &description
	}
	if payload.Price != nil {
		price := strings.TrimSpace(*payload.Price)
		updates.Price = &price
	}
	if payload.MaxImages != nil {
		updates.MaxImages = payload.MaxImages
	}
	if payload.InputModalities != nil {
		inputModalities := entity.StringArray(normaliseStringSlice(*payload.InputModalities))
		updates.InputModalities = &inputModalities
	}
	if payload.OutputModalities != nil {
		outputModalities := entity.StringArray(normaliseStringSlice(*payload.OutputModalities))
		updates.OutputModalities = &outputModalities
	}
	if payload.SupportedSizes != nil {
		supportedSizes := entity.StringArray(normaliseStringSlice(*payload.SupportedSizes))
		updates.SupportedSizes = &supportedSizes
	}
	if payload.SupportedDurations != nil {
		supportedDurations := entity.IntArray(normaliseIntSlice(*payload.SupportedDurations))
		updates.SupportedDurations = &supportedDurations
	}
	if payload.DefaultSize != nil {
		defaultSize := strings.TrimSpace(*payload.DefaultSize)
		updates.DefaultSize = &defaultSize
	}
	if payload.DefaultDuration != nil {
		updates.DefaultDuration = payload.DefaultDuration
	}
	if payload.Settings != nil {
		settings := entity.JSONMap(payload.Settings)
		updates.Settings = &settings
	}
	if payload.GenerationMode != nil {
		generationMode := strings.TrimSpace(*payload.GenerationMode)
		updates.GenerationMode = &generationMode
	}
	if payload.EndpointPath != nil {
		endpointPath := strings.TrimSpace(*payload.EndpointPath)
		updates.EndpointPath = &endpointPath
	}
	if payload.SupportsStreaming != nil {
		updates.SupportsStreaming = payload.SupportsStreaming
	}
	if payload.SupportsCancel != nil {
		updates.SupportsCancel = payload.SupportsCancel
	}
	if payload.IsActive != nil {
		updates.IsActive = payload.IsActive
	}

	if updates.IsEmpty() {
		c.JSON(http.StatusOK, gin.H{"message": "无更新内容"})
		return
	}

	ctx := c.Request.Context()
	if err := h.repo.UpdateModel(ctx, id, modelID, updates); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NotFound(c, ErrCodeModelNotFound, "模型不存在")
			return
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider_id": id,
			"model_id":    modelID,
		}).Error("failed to update provider model")
		InternalError(c, "更新模型失败: "+err.Error())
		return
	}

	_, model, err := h.repo.GetProviderWithModel(ctx, id, modelID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NotFound(c, ErrCodeModelNotFound, "模型不存在")
			return
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider_id": id,
			"model_id":    modelID,
		}).Error("failed to reload model after update")
		InternalError(c, "加载模型失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"model": entity.ModelToAdminView(*model)})
}

func (h *HTTPHandler) DeleteProviderModel(c *gin.Context) {
	if h.repo == nil {
		InternalError(c, "服务商仓储未配置")
		return
	}

	id, err := normaliseProviderID(c.Param("id"))
	if err != nil {
		BadRequest(c, ErrCodeInvalidRequest, err.Error())
		return
	}

	modelID := strings.TrimSpace(c.Param("model_id"))
	if modelID == "" {
		MissingField(c, "model_id")
		return
	}

	ctx := c.Request.Context()
	if err := h.repo.DeleteModel(ctx, id, modelID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NotFound(c, ErrCodeModelNotFound, "模型不存在")
			return
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider_id": id,
			"model_id":    modelID,
		}).Error("failed to delete provider model")
		InternalError(c, "删除模型失败")
		return
	}

	c.Status(http.StatusNoContent)
}
