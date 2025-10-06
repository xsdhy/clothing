package api

import (
	"clothing/internal/entity"
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
			inputPaths, err := h.saveImagesToStorage("inputs", request.Images, request.Model)
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
			outputPaths, err := h.saveImagesToStorage("outputs", images, request.Model)
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

func (h *HTTPHandler) saveImagesToStorage(category string, payloads []string, model string) ([]string, error) {
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
