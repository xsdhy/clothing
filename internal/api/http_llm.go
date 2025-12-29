package api

import (
	"clothing/internal/entity"
	"clothing/internal/llm"
	"clothing/internal/service"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ListProviders 列出可用的服务商
func (h *HTTPHandler) ListProviders(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusOK, gin.H{"providers": []entity.DbProvider{}})
		return
	}

	ctx := c.Request.Context()
	providers, err := h.repo.ListProviders(ctx, false)
	if err != nil {
		logrus.WithError(err).Error("failed to list providers from database")
		InternalError(c, "加载服务商列表失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// StreamGenerationEvents SSE 事件流
func (h *HTTPHandler) StreamGenerationEvents(c *gin.Context) {
	requestUser := CurrentUser(c)
	if requestUser == nil {
		Unauthorized(c, "需要登录")
		return
	}

	clientID := strings.TrimSpace(c.Query("client_id"))
	if clientID == "" {
		MissingField(c, "client_id")
		return
	}

	ctx := c.Request.Context()
	events := make(chan sseMessage, 8)
	h.registerSSEClient(clientID, events)
	defer h.unregisterSSEClient(clientID, events)

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}

	heartbeatTicker := time.NewTicker(10 * time.Second)
	defer heartbeatTicker.Stop()

	logrus.WithFields(logrus.Fields{
		"user_id":   requestUser.ID,
		"client_id": clientID,
	}).Info("generation sse connected")

	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			logrus.WithFields(logrus.Fields{
				"user_id":   requestUser.ID,
				"client_id": clientID,
			}).Info("generation sse disconnected")
			return false
		case <-heartbeatTicker.C:
			c.SSEvent("ping", gin.H{"ts": time.Now().UnixMilli()})
			return true
		case msg, ok := <-events:
			if !ok {
				return false
			}
			c.SSEvent(msg.event, msg.data)
			return true
		}
	})
}

// GenerateContent 提交内容生成请求
func (h *HTTPHandler) GenerateContent(c *gin.Context) {
	requestUser := CurrentUser(c)
	if requestUser == nil {
		Unauthorized(c, "需要登录")
		return
	}

	var request entity.GenerateContentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		InvalidPayload(c)
		return
	}

	request.Prompt = strings.TrimSpace(request.Prompt)
	if request.Prompt == "" {
		MissingField(c, "prompt")
		return
	}

	request.Options.Size = strings.TrimSpace(request.Options.Size)
	request.ClientID = strings.TrimSpace(request.ClientID)

	if h.repo == nil {
		InternalError(c, "服务商仓储未配置")
		return
	}

	providerID := strings.TrimSpace(request.ProviderID)
	if providerID == "" {
		MissingField(c, "provider")
		return
	}

	request.ModelID = strings.TrimSpace(request.ModelID)
	if request.ModelID == "" {
		MissingField(c, "model")
		return
	}

	ctx := c.Request.Context()

	// 加载并验证服务商
	dbProvider, err := h.repo.GetProvider(ctx, providerID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider": providerID,
		}).Error("failed to load provider from database")
		NotFound(c, ErrCodeProviderNotFound, "服务商不存在: "+request.ProviderID)
		return
	}
	if dbProvider == nil || !dbProvider.IsActive {
		ErrorResponse(c, http.StatusBadRequest, ErrCodeProviderDisabled, "服务商已禁用: "+request.ProviderID)
		return
	}

	// 加载并验证模型
	dbModel, err := h.repo.GetModel(ctx, providerID, request.ModelID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider": providerID,
			"model":    request.ModelID,
		}).Error("failed to load model from database")
		NotFound(c, ErrCodeModelNotFound, "模型不存在: "+request.ModelID)
		return
	}
	if dbModel == nil || !dbModel.IsActive {
		ErrorResponse(c, http.StatusBadRequest, ErrCodeModelDisabled, "模型已禁用: "+request.ModelID)
		return
	}

	// 初始化 LLM 服务
	llmService, err := llm.NewService(dbProvider)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider": providerID,
		}).Error("failed to initialise provider service")
		ErrorResponse(c, http.StatusBadRequest, ErrCodeProviderUnavailable, "服务商暂时不可用: "+request.ProviderID)
		return
	}

	// 验证标签
	tagIDs := deduplicatePositiveIDs(request.TagIDs)
	if len(tagIDs) > 0 {
		validateCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		tags, err := h.repo.FindTagsByIDs(validateCtx, tagIDs)
		if err != nil {
			logrus.WithError(err).Error("failed to validate tags")
			InternalError(c, "验证标签失败")
			return
		}
		if len(tags) != len(tagIDs) {
			BadRequest(c, ErrCodeInvalidTag, "部分标签不存在")
			return
		}
	}

	// 创建使用记录
	userID := requestUser.ID
	createCtx, cancelCreate := context.WithTimeout(ctx, 5*time.Second)
	defer cancelCreate()

	record := entity.DbUsageRecord{
		UserID:     userID,
		ProviderID: providerID,
		ModelID:    request.ModelID,
		Prompt:     request.Prompt,
		Size:       request.Options.Size,
	}

	if err := h.repo.CreateUsageRecord(createCtx, &record); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider": providerID,
			"model":    request.ModelID,
			"user_id":  userID,
		}).Error("failed to create usage record")
		InternalError(c, "创建使用记录失败")
		return
	}

	// 关联标签
	if len(tagIDs) > 0 {
		if err := h.repo.SetUsageRecordTags(createCtx, record.ID, tagIDs); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"record_id": record.ID,
			}).Warn("failed to set tags for usage record")
		}
	}

	logrus.WithFields(logrus.Fields{
		"record_id": record.ID,
		"provider":  providerID,
		"model":     request.ModelID,
		"user_id":   userID,
	}).Info("queued generation task")

	// 使用 Service 层异步处理生成任务
	h.generationService.GenerateContentAsync(service.GenerateContentRequest{
		Record:   record,
		Request:  request,
		Model:    *dbModel,
		Service:  llmService,
		ClientID: request.ClientID,
	})

	c.JSON(http.StatusAccepted, gin.H{
		"record_id": record.ID,
		"status":    "processing",
	})
}

// deduplicatePositiveIDs 去重正整数 ID 列表
func deduplicatePositiveIDs(values []uint) []uint {
	if len(values) == 0 {
		return []uint{}
	}
	result := make([]uint, 0, len(values))
	seen := make(map[uint]struct{}, len(values))
	for _, v := range values {
		if v == 0 {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	return result
}
