package llm

import (
	"clothing/internal/entity"
	"context"
	"errors"
	"strings"

	"github.com/sirupsen/logrus"
)

type AiHubMix struct {
	providerID   string
	providerName string

	apiKey   string
	endpoint string
	// Optional custom Gemini base endpoint (e.g. https://aihubmix.com/gemini).
	geminiEndpoint string

	models      []entity.LlmModel
	modelLookup map[string]struct{}
}

func NewAiHubMix(provider *entity.DbProvider, models []entity.DbModel) (*AiHubMix, error) {
	if provider == nil {
		return nil, errors.New("aihubmix provider config is nil")
	}

	apiKey := strings.TrimSpace(provider.APIKey)
	if apiKey == "" {
		return nil, errors.New("aihubmix api key is not configured")
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
		return nil, errors.New("aihubmix has no active models configured")
	}

	baseURL := strings.TrimSpace(provider.BaseURL)
	endpoint := baseURL
	if strings.Contains(strings.ToLower(baseURL), "gemini") {
		// Prevent Gemini-specific base URL from polluting the OpenAI-compatible path.
		endpoint = ""
	}
	if endpoint == "" {
		endpoint = "https://aihubmix.com/v1/chat/completions"
	}

	// geminiEndpoint can come from config.gemini_base_url or fall back to a Gemini-looking base_url.
	geminiEndpoint := "https://aihubmix.com/gemini"

	name := strings.TrimSpace(provider.Name)
	if name == "" {
		name = provider.ID
	}

	return &AiHubMix{
		providerID:     provider.ID,
		providerName:   name,
		apiKey:         apiKey,
		endpoint:       endpoint,
		geminiEndpoint: geminiEndpoint,
		models:         activeModels,
		modelLookup:    modelLookup,
	}, nil
}

func (o *AiHubMix) ProviderID() string {
	return o.providerID
}

func (o *AiHubMix) Provider() entity.LlmProvider {
	return entity.LlmProvider{
		ID:     o.providerID,
		Name:   o.providerName,
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

	if strings.Contains(request.Model, "gemini") {
		return GenerateImagesByGeminiProtocol(ctx, o.apiKey, o.geminiEndpoint, request.Model, request.Prompt, request.Images)
	}

	return GenerateImagesByOpenaiProtocol(ctx, o.apiKey, o.endpoint, request.Model, request.Prompt, request.Images)
}
