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
		apiKey:       apiKey,
		httpClient:   &http.Client{},
	}, nil
}

func (g *GeminiService) GenerateContent(ctx context.Context, request entity.GenerateContentRequest, dbModel entity.DbModel) ([]string, string, error) {
	// TODO: Implement Gemini API integration.
	return nil, "", errors.New("gemini response did not include image data")
}
