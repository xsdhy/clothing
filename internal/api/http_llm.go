package api

import (
	"clothing/internal/entity"
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

		images, text, err := provider.GenerateImages(ctx, request)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"provider": providerID,
				"model":    request.Model,
			}).Error("failed to generate image")
			messages <- sseMessage{event: "error", data: gin.H{"error": err.Error()}}
			return
		}

		response := entity.GenerateImageResponse{
			Text: text,
		}
		if len(images) > 0 {
			response.Image = images[0]
			response.Images = images
		}

		messages <- sseMessage{event: "result", data: response}
	}()

	heartbeatTicker := time.NewTicker(15 * time.Second)
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
			c.SSEvent(msg.event, msg.data)
			if msg.event == "result" || msg.event == "error" {
				return false
			}
			return true
		}
	})
}
