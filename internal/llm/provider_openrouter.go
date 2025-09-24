package llm

import (
	"clothing/internal/config"
	"clothing/internal/entity"
	"context"
	"errors"
	"strings"

	"github.com/sirupsen/logrus"
)

type OpenRouter struct {
	apiKey string

	models      []entity.LlmModel
	modelLookup map[string]struct{}
}

func NewOpenRouter(cfg config.Config) (*OpenRouter, error) {
	if strings.TrimSpace(cfg.OpenRouterAPIKey) == "" {
		return nil, errors.New("openrouter api key is not configured")
	}

	models := []entity.LlmModel{
		{
			ID:          "google/gemini-2.5-flash-image-preview",
			Name:        "Gemini 2.5 Flash Image Preview",
			Description: "Google 的高性能模型",
		},
	}

	o := &OpenRouter{
		models: models,
		apiKey: cfg.OpenRouterAPIKey,
		modelLookup: func() map[string]struct{} {
			lookup := make(map[string]struct{}, len(models))
			for _, model := range models {
				lookup[model.ID] = struct{}{}
			}
			return lookup
		}(),
	}

	return o, nil
}
func (o *OpenRouter) ProviderID() string {
	return "openrouter"
}

func (o *OpenRouter) Provider() entity.LlmProvider {
	return entity.LlmProvider{
		ID:     "openrouter",
		Name:   "OpenRouter",
		Models: o.Models(),
	}
}

func (o *OpenRouter) Models() []entity.LlmModel {
	return o.models
}

func (o *OpenRouter) SupportsModel(modelID string) bool {
	if o == nil || modelID == "" {
		return false
	}
	_, ok := o.modelLookup[modelID]
	return ok
}

func (o *OpenRouter) GenerateImages(ctx context.Context, request entity.GenerateImageRequest) (string, string, error) {
	logrus.WithFields(logrus.Fields{
		"prompt_preview":      request.Prompt,
		"reference_image_cnt": len(request.Images),
	}).Info("llm_generate_images_start")

	return GenerateImagesByOpenaiProtocol(ctx, o.apiKey, "https://openrouter.ai/api/v1/chat/completions", request.Model, request.Prompt, request.Images)
}
