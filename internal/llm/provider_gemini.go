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

	models      []entity.LlmModel
	modelLookup map[string]struct{}
}

func NewGeminiService(provider *entity.DbProvider, models []entity.DbModel) (*GeminiService, error) {
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
		return nil, errors.New("gemini has no active models configured")
	}

	return &GeminiService{
		providerID:   provider.ID,
		providerName: name,
		apiKey:       apiKey,
		httpClient:   &http.Client{},
		models:       activeModels,
		modelLookup:  modelLookup,
	}, nil
}

func (g *GeminiService) ProviderID() string {
	return g.providerID
}

func (g *GeminiService) Provider() entity.LlmProvider {
	return entity.LlmProvider{
		ID:     g.providerID,
		Name:   g.providerName,
		Models: g.Models(),
	}
}

func (g *GeminiService) Models() []entity.LlmModel {
	return g.models
}

func (g *GeminiService) SupportsModel(modelID string) bool {
	if g == nil || modelID == "" {
		return false
	}
	_, ok := g.modelLookup[modelID]
	return ok
}

func (g *GeminiService) GenerateContent(ctx context.Context, request entity.GenerateContentRequest) ([]string, string, error) {
	// TODO: Implement Gemini API integration.
	return nil, "", errors.New("gemini response did not include image data")
}
