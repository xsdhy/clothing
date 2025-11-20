package llm

import (
	"clothing/internal/entity"
	"context"
	"errors"
	"strings"

	"github.com/sirupsen/logrus"
)

type OpenRouter struct {
	providerID   string
	providerName string

	apiKey   string
	endpoint string

	models      []entity.LlmModel
	modelLookup map[string]struct{}
}

func NewOpenRouter(provider *entity.DbProvider, models []entity.DbModel) (*OpenRouter, error) {
	if provider == nil {
		return nil, errors.New("openrouter provider config is nil")
	}

	apiKey := strings.TrimSpace(provider.APIKey)
	if apiKey == "" {
		return nil, errors.New("openrouter api key is not configured")
	}

	activeModels := make([]entity.LlmModel, 0, len(models))
	modelLookup := make(map[string]struct{}, len(models))
	for _, model := range models {
		if !model.IsActive {
			continue
		}
		llmModel := model.ToLlmModel()
		activeModels = append(activeModels, llmModel)
		modelLookup[llmModel.ID] = struct{}{}
	}

	if len(activeModels) == 0 {
		return nil, errors.New("openrouter has no active models configured")
	}

	endpoint := strings.TrimSpace(provider.BaseURL)
	if endpoint == "" {
		endpoint = "https://openrouter.ai/api/v1/chat/completions"
	}

	name := strings.TrimSpace(provider.Name)
	if name == "" {
		name = provider.ID
	}

	return &OpenRouter{
		providerID:   provider.ID,
		providerName: name,
		apiKey:       apiKey,
		endpoint:     endpoint,
		models:       activeModels,
		modelLookup:  modelLookup,
	}, nil
}

func (o *OpenRouter) ProviderID() string {
	return o.providerID
}

func (o *OpenRouter) Provider() entity.LlmProvider {
	return entity.LlmProvider{
		ID:     o.providerID,
		Name:   o.providerName,
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

func (o *OpenRouter) GenerateImages(ctx context.Context, request entity.GenerateImageRequest) ([]string, string, error) {
	logrus.WithFields(logrus.Fields{
		"prompt_preview":      request.Prompt,
		"reference_image_cnt": len(request.Images),
		"reference_video_cnt": len(request.Videos),
		"size":                strings.TrimSpace(request.Size),
	}).Info("llm_generate_images_start")

	return GenerateImagesByOpenaiProtocol(ctx, o.apiKey, o.endpoint, request.Model, request.Prompt, request.Images, request.Videos)
}
