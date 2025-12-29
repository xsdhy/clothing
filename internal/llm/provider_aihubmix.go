package llm

import (
	"clothing/internal/entity"
	"context"
	"errors"
	"strings"
)

type AiHubMix struct {
	providerID   string
	providerName string

	apiKey   string
	endpoint string
	// Optional custom Gemini base endpoint (e.g. https://aihubmix.com/gemini).
	geminiEndpoint string
}

func NewAiHubMix(provider *entity.DbProvider) (*AiHubMix, error) {
	if provider == nil {
		return nil, errors.New("aihubmix provider config is nil")
	}

	apiKey := strings.TrimSpace(provider.APIKey)
	if apiKey == "" {
		return nil, errors.New("aihubmix api key is not configured")
	}

	baseURL := strings.TrimSpace(provider.BaseURL)
	endpoint := strings.TrimRight(baseURL, "/")
	if strings.Contains(strings.ToLower(baseURL), "gemini") {
		// Prevent Gemini-specific base URL from polluting the OpenAI-compatible path.
		endpoint = ""
	}
	if endpoint == "" {
		endpoint = "https://aihubmix.com/v1/chat/completions"
	} else if !strings.Contains(strings.ToLower(endpoint), "chat/completions") {
		endpoint = endpoint + "/chat/completions"
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
	}, nil
}

func (p *AiHubMix) GenerateContent(ctx context.Context, request entity.GenerateContentRequest, dbModel entity.DbModel) (*entity.GenerateContentResponse, error) {
	if dbModel.IsVideoModel() {
		return nil, errors.New("aihubmix video not supported yet")
	}

	// AiHubMix uses Gemini-compatible protocol for image generation (as per current implementation assumption)
	// or OpenAI compatible?
	// Looking at previous implementation:
	// return GenerateContentByGeminiProtocol(ctx, p.apiKey, p.endpoint, dbModel.ModelID, request.Prompt, request.Inputs.Images)
	// Yes, it uses Gemini protocol.

	// The GenerateContentByGeminiProtocol function returns (*entity.GenerateContentResponse, error).
	return GenerateContentByGeminiProtocol(ctx, p.apiKey, p.endpoint, dbModel.ModelID, request.Prompt, request.GetImages())
}

// Capabilities returns the capabilities of the model.
func (p *AiHubMix) Capabilities(model entity.DbModel) *ModelCapabilities {
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
func (p *AiHubMix) Validate(request entity.GenerateContentRequest, model entity.DbModel) error {
	if strings.TrimSpace(request.Prompt) == "" {
		return errors.New("prompt is required")
	}
	if model.IsVideoModel() {
		return errors.New("video generation not supported")
	}
	return nil
}
