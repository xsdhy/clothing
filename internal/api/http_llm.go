package api

import (
	"clothing/internal/entity"
	"clothing/internal/llm"
	"clothing/internal/storage"
	"clothing/internal/utils"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *HTTPHandler) ListProviders(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusOK, gin.H{"providers": []entity.DbProvider{}})
		return
	}

	ctx := c.Request.Context()
	providers, err := h.repo.ListProviders(ctx, false)
	if err != nil {
		logrus.WithError(err).Error("failed to list providers from database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load provider catalogue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

func (h *HTTPHandler) StreamGenerationEvents(c *gin.Context) {
	requestUser := CurrentUser(c)
	if requestUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	clientID := strings.TrimSpace(c.Query("client_id"))
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
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

func (h *HTTPHandler) GenerateContent(c *gin.Context) {
	requestUser := CurrentUser(c)
	if requestUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var request entity.GenerateContentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	request.Prompt = strings.TrimSpace(request.Prompt)
	if request.Prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}

	request.Options.Size = strings.TrimSpace(request.Options.Size)
	request.ClientID = strings.TrimSpace(request.ClientID)

	if h.repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "provider repository not configured"})
		return
	}

	providerID := strings.TrimSpace(request.ProviderID)
	if providerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}

	request.ModelID = strings.TrimSpace(request.ModelID)
	if request.ModelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}

	ctx := c.Request.Context()

	dbProvider, err := h.repo.GetProvider(ctx, providerID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider": providerID,
		}).Error("failed to load provider from database")
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported provider: %s", request.ProviderID)})
		return
	}
	if dbProvider == nil || !dbProvider.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("provider %s is disabled", request.ProviderID)})
		return
	}

	dbModel, err := h.repo.GetModel(ctx, providerID, request.ModelID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider": providerID,
			"model":    request.ModelID,
		}).Error("failed to load model from database")
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported model: %s", request.ModelID)})
		return
	}
	if dbModel == nil || !dbModel.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported model: %s", request.ModelID)})
		return
	}

	service, err := llm.NewService(dbProvider)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider": providerID,
		}).Error("failed to initialise provider service")
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("provider %s is temporarily unavailable", request.ProviderID)})
		return
	}

	tagIDs := deduplicatePositiveIDs(request.TagIDs)
	if len(tagIDs) > 0 {
		validateCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		tags, err := h.repo.FindTagsByIDs(validateCtx, tagIDs)
		if err != nil {
			logrus.WithError(err).Error("failed to validate tags")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate tags"})
			return
		}
		if len(tags) != len(tagIDs) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "one or more tags do not exist"})
			return
		}
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create usage record"})
		return
	}

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

	go h.handleAsyncGeneration(record, request, *dbModel, service)

	c.JSON(http.StatusAccepted, gin.H{
		"record_id": record.ID,
		"status":    "processing",
	})
}

func (h *HTTPHandler) handleAsyncGeneration(record entity.DbUsageRecord, request entity.GenerateContentRequest, dbModel entity.DbModel, service llm.AIService) {
	if h.repo == nil {
		return
	}

	clientID := strings.TrimSpace(request.ClientID)
	genCtx, cancelGen := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancelGen()

	updates := make(map[string]interface{})
	completionError := ""
	var storageIssues []string

	if len(request.Inputs.Images) > 0 {
		inputPaths, err := h.saveMediaToStorage(genCtx, "inputs", request.Inputs.Images, request.ModelID)
		if len(inputPaths) > 0 {
			updates["input_images"] = entity.StringArray(inputPaths)
		}
		if err != nil {
			storageIssues = append(storageIssues, fmt.Sprintf("input images: %v", err))
			logrus.WithError(err).WithFields(logrus.Fields{
				"record_id": record.ID,
				"provider":  record.ProviderID,
				"model":     record.ModelID,
			}).Warn("failed to persist input images")
		}
	}

	outputs, text, err := service.GenerateContent(genCtx, request, dbModel)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"record_id": record.ID,
			"provider":  record.ProviderID,
			"model":     record.ModelID,
		}).Error("failed to generate content")

		errMsg := err.Error()
		if len(storageIssues) > 0 {
			errMsg = appendStorageNotes(errMsg, storageIssues)
		}

		updates["error_message"] = errMsg
		h.updateUsageRecord(record.ID, updates)
		h.notifyGenerationComplete(clientID, record.ID, "failure", errMsg)
		return
	}

	logrus.WithFields(logrus.Fields{
		"record_id": record.ID,
		"provider":  record.ProviderID,
		"model":     record.ModelID,
	}).Info("generated content")

	if text != "" {
		updates["output_text"] = text
	}

	if len(outputs) > 0 {
		outputPaths, err := h.saveMediaToStorage(genCtx, "outputs", outputs, request.ModelID)
		if len(outputPaths) > 0 {
			updates["output_images"] = entity.StringArray(outputPaths)
		}
		if err != nil {
			storageIssues = append(storageIssues, fmt.Sprintf("output assets: %v", err))
			logrus.WithError(err).WithFields(logrus.Fields{
				"record_id": record.ID,
				"provider":  record.ProviderID,
				"model":     record.ModelID,
			}).Warn("failed to persist output assets")
		}
	}

	if len(storageIssues) > 0 {
		existingError := ""
		if msg, ok := updates["error_message"].(string); ok {
			existingError = msg
		}
		combined := appendStorageNotes(existingError, storageIssues)
		updates["error_message"] = combined
		completionError = combined
	}

	h.updateUsageRecord(record.ID, updates)
	h.notifyGenerationComplete(clientID, record.ID, "success", completionError)
}

func (h *HTTPHandler) saveMediaToStorage(parentCtx context.Context, category string, payloads []string, model string) ([]string, error) {
	if h.storage == nil || len(payloads) == 0 {
		return nil, nil
	}

	if parentCtx == nil {
		parentCtx = context.Background()
	}

	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Minute)
	defer cancel()

	var (
		paths []string
		errs  []string
	)

	for idx, payload := range payloads {
		trimmed := strings.TrimSpace(payload)
		if trimmed == "" {
			continue
		}

		data, ext, err := h.resolveMediaPayload(ctx, trimmed)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%d: %v", idx, err))
			continue
		}

		saveOpts := storage.SaveOptions{Category: category, Extension: ext}
		switch strings.ToLower(strings.TrimSpace(category)) {
		case "inputs":
			saveOpts.SkipIfExists = true
			saveOpts.BaseName = computeInputBaseName(data)
		case "outputs":
			saveOpts.BaseName = buildOutputBaseName(model, idx)
		default:
			saveOpts.BaseName = ""
		}

		relPath, err := h.storage.Save(ctx, data, saveOpts)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%d: %v", idx, err))
			continue
		}
		paths = append(paths, relPath)
	}

	if len(errs) > 0 {
		return paths, fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	return paths, nil
}

func (h *HTTPHandler) resolveMediaPayload(ctx context.Context, payload string) ([]byte, string, error) {
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" {
		return nil, "", fmt.Errorf("empty payload")
	}

	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		reqCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, trimmed, nil)
		if err != nil {
			return nil, "", fmt.Errorf("create request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, "", fmt.Errorf("download image: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, "", fmt.Errorf("download image http %d", resp.StatusCode)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", fmt.Errorf("read image body: %w", err)
		}

		ext := utils.ExtensionFromMime(resp.Header.Get("Content-Type"))
		if ext == "" {
			ext = utils.ExtensionFromMime(http.DetectContentType(data))
		}
		if ext == "" {
			ext = "bin"
		}

		return data, ext, nil
	}

	data, ext, err := utils.DecodeMediaPayload(trimmed)
	if err == nil {
		return data, ext, nil
	}

	data, ext, err = utils.DecodeMediaPayload(utils.EnsureDataURL(trimmed))
	if err != nil {
		return nil, "", err
	}

	return data, ext, nil
}

func (h *HTTPHandler) updateUsageRecord(recordID uint, updates map[string]interface{}) {
	if h.repo == nil || recordID == 0 || len(updates) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.repo.UpdateUsageRecord(ctx, recordID, updates); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"record_id": recordID,
		}).Error("failed to update usage record")
	}
}

func (h *HTTPHandler) notifyGenerationComplete(clientID string, recordID uint, status string, errMsg string) {
	if strings.TrimSpace(clientID) == "" {
		return
	}
	payload := gin.H{
		"record_id": recordID,
		"status":    status,
	}
	if trimmed := strings.TrimSpace(errMsg); trimmed != "" {
		payload["error"] = trimmed
	}
	h.publishSSEMessage(clientID, sseMessage{
		event: "generation_completed",
		data:  payload,
	})
}

func appendStorageNotes(existing string, notes []string) string {
	if len(notes) == 0 {
		return existing
	}
	combined := strings.Join(notes, "; ")
	if strings.TrimSpace(existing) == "" {
		return combined
	}
	return existing + "; " + combined
}

func computeInputBaseName(data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:])
}

func buildOutputBaseName(model string, idx int) string {
	token := storage.SanitizeToken(model)
	if token == "" {
		token = "model"
	}
	if len(token) > 32 {
		token = token[:32]
	}
	suffix := time.Now().UTC().UnixNano()
	return fmt.Sprintf("%s_%d_%d", token, suffix, idx)
}

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
