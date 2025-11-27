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

func (o *OpenRouter) GenerateContent(ctx context.Context, request entity.GenerateContentRequest, dbModel entity.DbModel) ([]string, string, error) {
	logrus.WithFields(logrus.Fields{
		"prompt_preview":      request.Prompt,
		"reference_image_cnt": len(request.Inputs.Images),
		"reference_video_cnt": len(request.Inputs.Videos),
		"size":                strings.TrimSpace(request.Options.Size),
	}).Info("llm_generate_content_start")

	return GenerateContentByOpenaiProtocol(ctx, o.apiKey, o.endpoint, request.ModelID, request.Prompt, request.Inputs.Images, request.Inputs.Videos)
}
