package api

import (
	"clothing/internal/entity"
	"fmt"
	"net/http"
	"strings"

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
	result, text, err := provider.GenerateImages(ctx, request)

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"provider": providerID,
			"model":    request.Model,
		}).Error("failed to generate image")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := entity.GenerateImageResponse{Image: result, Text: text}
	c.JSON(http.StatusOK, response)
}
