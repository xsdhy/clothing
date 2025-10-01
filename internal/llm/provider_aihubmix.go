package llm

import (
	"clothing/internal/config"
	"clothing/internal/entity"
	"context"
	"errors"
	"strings"

	"github.com/sirupsen/logrus"
)

type AiHubMix struct {
	apiKey string

	models      []entity.LlmModel
	modelLookup map[string]struct{}
}

func NewAiHubMix(cfg config.Config) (*AiHubMix, error) {
	if strings.TrimSpace(cfg.AiHubMixAPIKey) == "" {
		return nil, errors.New("AiHubMix api key is not configured")
	}

	models := []entity.LlmModel{
		{
			ID:          "qwen-image-edit",
			Name:        "qwen-image-edit",
			Price:       "<UNK>",
			Description: "",
		},
		{
			ID:          "gemini-2.5-flash-image-preview",
			Name:        "gemini-2.5-flash-image-preview",
			Price:       "0.03/IMG",
			Description: "",
		},
		{
			ID:          "imagen-4.0-fast-generate-001",
			Name:        "imagen-4.0-fast-generate-001",
			Price:       "0.03/IMG",
			Description: "",
		},
		{
			ID:          "gpt-4o-image",
			Name:        "gpt-4o-image",
			Price:       "0.005/IMG",
			Description: "",
		},
	}

	o := &AiHubMix{
		apiKey: cfg.AiHubMixAPIKey,
		models: models,
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

func (o *AiHubMix) ProviderID() string {
	return "AiHubMix"
}

func (o *AiHubMix) Provider() entity.LlmProvider {
	return entity.LlmProvider{
		ID:     "AiHubMix",
		Name:   "AiHubMix",
		Models: o.Models(),
	}
}

func (o *AiHubMix) Models() []entity.LlmModel {
	return o.models
}

func (o *AiHubMix) SupportsModel(modelID string) bool {
	if o == nil || modelID == "" {
		return false
	}
	_, ok := o.modelLookup[modelID]
	return ok
}

func (o *AiHubMix) GenerateImages(ctx context.Context, request entity.GenerateImageRequest) ([]string, string, error) {
	logrus.WithFields(logrus.Fields{
		"prompt_preview":      request.Prompt,
		"reference_image_cnt": len(request.Images),
		"size":                strings.TrimSpace(request.Size),
	}).Info("llm_generate_images_start")

	return GenerateImagesByOpenaiProtocol(ctx, o.apiKey, "https://aihubmix.com/v1/chat/completions", request.Model, request.Prompt, request.Images)
}
