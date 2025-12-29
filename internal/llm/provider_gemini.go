package llm

import (
	"clothing/internal/entity"
	"context"
	"errors"
	"net/http"
	"strings"
)

type GeminiService struct {
	providerID   string
	providerName string
	endpoint     string

	apiKey     string
	httpClient *http.Client
}

func NewGeminiService(provider *entity.DbProvider) (*GeminiService, error) {
	if provider == nil {
		return nil, errors.New("gemini provider config is nil")
	}

	apiKey := strings.TrimSpace(provider.APIKey)
	if apiKey == "" {
		return nil, errors.New("gemini api key is not configured")
	}

	name := strings.TrimSpace(provider.Name)
	if name == "" {
		name = provider.ID
	}

	return &GeminiService{
		providerID:   provider.ID,
		providerName: name,
		endpoint:     provider.BaseURL,
		apiKey:       apiKey,
		httpClient:   &http.Client{},
	}, nil
}

func (p *GeminiService) GenerateContent(ctx context.Context, request entity.GenerateContentRequest, dbModel entity.DbModel) (*entity.GenerateContentResponse, error) {
	return GenerateContentByGeminiProtocol(ctx, p.apiKey, p.endpoint, dbModel.ModelID, request.Prompt, request.GetImages())
}

// Capabilities returns the capabilities of the model.
func (p *GeminiService) Capabilities(model entity.DbModel) *ModelCapabilities {
	return &ModelCapabilities{
		InputModalities:    model.InputModalities,
		OutputModalities:   model.OutputModalities,
		MaxImages:          model.MaxImages,
		SupportedSizes:     model.SupportedSizes,
		SupportedDurations: model.SupportedDurations,
		SupportsStream:     true, // Gemini uses streaming by default
		SupportsCancel:     model.SupportsCancel,
		SupportsAsync:      model.IsVideoModel(),
	}
}

// Validate checks if the request is valid for the model.
func (p *GeminiService) Validate(request entity.GenerateContentRequest, model entity.DbModel) error {
	if strings.TrimSpace(request.Prompt) == "" {
		return errors.New("prompt is required")
	}
	return nil
}
