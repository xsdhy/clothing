package llm

import (
	"clothing/internal/entity"
	"context"
	"errors"
	"strings"
)

type OpenRouter struct {
	providerID   string
	providerName string

	apiKey   string
	endpoint string
}

func NewOpenRouter(provider *entity.DbProvider) (*OpenRouter, error) {
	if provider == nil {
		return nil, errors.New("openrouter provider config is nil")
	}

	apiKey := strings.TrimSpace(provider.APIKey)
	if apiKey == "" {
		return nil, errors.New("openrouter api key is not configured")
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
	}, nil
}

func (o *OpenRouter) GenerateContent(ctx context.Context, request entity.GenerateContentRequest, dbModel entity.DbModel) (*entity.GenerateContentResponse, error) {
	return GenerateContentByOpenaiProtocol(ctx, o.apiKey, o.endpoint, dbModel.ModelID, request.Prompt, request.GetImages(), request.GetVideos())
}

// Capabilities returns the capabilities of the model.
func (o *OpenRouter) Capabilities(model entity.DbModel) *ModelCapabilities {
	return &ModelCapabilities{
		InputModalities:    model.InputModalities,
		OutputModalities:   model.OutputModalities,
		MaxImages:          model.MaxImages,
		SupportedSizes:     model.SupportedSizes,
		SupportedDurations: model.SupportedDurations,
		SupportsStream:     model.SupportsStreaming,
		SupportsCancel:     model.SupportsCancel,
		SupportsAsync:      model.IsVideoModel(),
	}
}

// Validate checks if the request is valid for the model.
func (o *OpenRouter) Validate(request entity.GenerateContentRequest, model entity.DbModel) error {
	if strings.TrimSpace(request.Prompt) == "" {
		return errors.New("prompt is required")
	}
	return nil
}
