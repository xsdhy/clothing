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

func (o *AiHubMix) GenerateContent(ctx context.Context, request entity.GenerateContentRequest, dbModel entity.DbModel) ([]string, string, error) {
	logrus.WithFields(logrus.Fields{
		"prompt_preview":      request.Prompt,
		"reference_image_cnt": len(request.Inputs.Images),
		"reference_video_cnt": len(request.Inputs.Videos),
		"size":                strings.TrimSpace(request.Options.Size),
	}).Info("llm_generate_content_start")

	if strings.Contains(request.ModelID, "gemini") {
		return GenerateContentByGeminiProtocol(ctx, o.apiKey, o.geminiEndpoint, request.ModelID, request.Prompt, request.Inputs.Images)
	}

	return GenerateContentByOpenaiProtocol(ctx, o.apiKey, o.endpoint, request.ModelID, request.Prompt, request.Inputs.Images, request.Inputs.Videos)
}
