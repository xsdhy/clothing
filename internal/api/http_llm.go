package api

import (
	"clothing/internal/entity"
	"clothing/internal/storage"
	"clothing/internal/utils"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *HTTPHandler) ListProviders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"providers": h.providers})
}

func (h *HTTPHandler) GenerateImage(c *gin.Context) {
	var request entity.GenerateImageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	request.Prompt = strings.TrimSpace(request.Prompt)
	if request.Prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}

	request.Size = strings.TrimSpace(request.Size)

	providerID := strings.TrimSpace(request.Provider)
	provider, ok := h.providerMap[providerID]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported provider: %s", request.Provider)})
		return
	}

	request.Model = strings.TrimSpace(request.Model)
	if request.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}

	if !provider.SupportsModel(request.Model) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported model: %s", request.Model)})
		return
	}

	ctx := c.Request.Context()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}

	type sseMessage struct {
		event string
		data  interface{}
	}

	messages := make(chan sseMessage, 4)

	go func() {
		defer close(messages)
		messages <- sseMessage{event: "status", data: gin.H{"state": "processing"}}

		record := entity.DbUsageRecord{
			ProviderID: providerID,
			ModelID:    request.Model,
			Prompt:     request.Prompt,
			Size:       request.Size,
		}

		var storageIssues []string

		if len(request.Images) > 0 {
			inputPaths, err := h.saveImagesToStorage(providerID, "inputs", request.Images)
			if len(inputPaths) > 0 {
				record.InputImages = entity.StringArray(inputPaths)
			}
			if err != nil {
				storageIssues = append(storageIssues, fmt.Sprintf("input images: %v", err))
				logrus.WithError(err).WithFields(logrus.Fields{
					"provider": providerID,
					"model":    request.Model,
				}).Warn("failed to persist input images")
			}
		}

		images, text, err := provider.GenerateImages(ctx, request)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"provider": providerID,
				"model":    request.Model,
			}).Error("failed to generate image")
			record.ErrorMessage = err.Error()
			if len(storageIssues) > 0 {
				record.ErrorMessage = appendStorageNotes(record.ErrorMessage, storageIssues)
			}
			h.persistUsageRecord(&record)
			messages <- sseMessage{event: "error", data: gin.H{"error": err.Error()}}
			return
		}

		logrus.WithFields(logrus.Fields{
			"provider": providerID,
			"model":    request.Model,
		}).Info("generated image")

		record.OutputText = text

		if len(images) > 0 {
			outputPaths, err := h.saveImagesToStorage(providerID, "outputs", images)
			if len(outputPaths) > 0 {
				record.OutputImages = entity.StringArray(outputPaths)
			}
			if err != nil {
				storageIssues = append(storageIssues, fmt.Sprintf("output images: %v", err))
				logrus.WithError(err).WithFields(logrus.Fields{
					"provider": providerID,
					"model":    request.Model,
				}).Warn("failed to persist output images")
			}
		}

		if len(storageIssues) > 0 {
			record.ErrorMessage = appendStorageNotes(record.ErrorMessage, storageIssues)
		}

		h.persistUsageRecord(&record)

		response := entity.GenerateImageResponse{
			Text:   text,
			Images: images,
		}

		messages <- sseMessage{event: "result", data: response}
	}()

	heartbeatTicker := time.NewTicker(5 * time.Second)
	defer heartbeatTicker.Stop()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			logrus.WithFields(logrus.Fields{
				"provider": providerID,
				"model":    request.Model,
			}).Warn("generate image request cancelled by client")
			return false
		case <-heartbeatTicker.C:
			c.SSEvent("ping", gin.H{"ts": time.Now().UnixMilli()})
			return true
		case msg, ok := <-messages:
			if !ok {
				return false
			}
			logrus.WithFields(logrus.Fields{
				"event": msg.event,
				"data":  msg.data,
			}).Info("generate image request received")
			c.SSEvent(msg.event, msg.data)
			if msg.event == "result" || msg.event == "error" {
				return false
			}
			return true
		}
	})
}

func (h *HTTPHandler) ListUsageRecords(c *gin.Context) {
	if h.repo == nil {
		c.JSON(http.StatusOK, entity.UsageRecordListResponse{Records: []entity.UsageRecordItem{}, Meta: &entity.Meta{Page: 1, PageSize: 0, Total: 0}})
		return
	}

	var params entity.UsageRecordQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
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
		items = append(items, entity.UsageRecordItem{
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
		})
	}

	if meta == nil {
		meta = &entity.Meta{Page: int64(params.Page), PageSize: int64(params.PageSize), Total: int64(len(items))}
	}

	c.JSON(http.StatusOK, entity.UsageRecordListResponse{Records: items, Meta: meta})
}

func (h *HTTPHandler) saveImagesToStorage(providerID, category string, payloads []string) ([]string, error) {
	if h.storage == nil || len(payloads) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
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

		data, ext, err := h.resolveImagePayload(ctx, trimmed)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%d: %v", idx, err))
			continue
		}

		relPath, err := h.storage.Save(ctx, data, storage.SaveOptions{ProviderID: providerID, Category: category, Extension: ext})
		if err != nil {
			errs = append(errs, fmt.Sprintf("%d: %v", idx, err))
			continue
		}
		paths = append(paths, relPath)
	}

	if len(errs) > 0 {
		return paths, fmt.Errorf(strings.Join(errs, "; "))
	}

	return paths, nil
}

func (h *HTTPHandler) resolveImagePayload(ctx context.Context, payload string) ([]byte, string, error) {
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
			ext = "jpg"
		}

		return data, ext, nil
	}

	data, ext, err := utils.DecodeImagePayload(trimmed)
	if err == nil {
		return data, ext, nil
	}

	data, ext, err = utils.DecodeImagePayload(utils.EnsureDataURL(trimmed))
	if err != nil {
		return nil, "", err
	}

	return data, ext, nil
}

func (h *HTTPHandler) persistUsageRecord(record *entity.DbUsageRecord) {
	if h.repo == nil || record == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.repo.CreateUsageRecord(ctx, record); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider": record.ProviderID,
			"model":    record.ModelID,
		}).Error("failed to persist usage record")
	}
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

func (h *HTTPHandler) publicURL(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed
	}
	base := h.storagePublicBase
	if base == "" {
		base = "/files"
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(base, "/"), strings.TrimLeft(trimmed, "/"))
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
