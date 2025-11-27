package llm

import (
	"clothing/internal/entity"
	"context"
	"errors"
	"strings"

	"github.com/sirupsen/logrus"
)

type Dashscope struct {
	providerID   string
	providerName string

	apiKey   string
	endpoint string
}

func NewDashscope(provider *entity.DbProvider) (*Dashscope, error) {
	if provider == nil {
		return nil, errors.New("dashscope provider config is nil")
	}

	apiKey := strings.TrimSpace(provider.APIKey)
	if apiKey == "" {
		return nil, errors.New("dashscope api key is not configured")
	}

	name := strings.TrimSpace(provider.Name)
	if name == "" {
		name = provider.ID
	}

	return &Dashscope{
		providerID:   provider.ID,
		providerName: name,
		apiKey:       apiKey,
		endpoint:     strings.TrimSpace(provider.BaseURL),
	}, nil
}

func (d *Dashscope) GenerateContent(ctx context.Context, request entity.GenerateContentRequest, dbModel entity.DbModel) ([]string, string, error) {
	logrus.WithFields(logrus.Fields{
		"prompt_preview":      request.Prompt,
		"reference_image_cnt": len(request.Inputs.Images),
		"reference_video_cnt": len(request.Inputs.Videos),
		"size":                strings.TrimSpace(request.Options.Size),
		"duration":            request.Options.Duration,
	}).Info("llm_generate_content_start")

	if dbModel.IsVideoModel() {
		return GenerateDashscopeVideo(ctx, d.apiKey, d.endpoint, dbModel, request.Prompt, request.Options.Size, request.Options.Duration, request.Inputs.Images)
	}

	return GenerateImageByDashscopeProtocol(ctx, d.apiKey, d.endpoint, dbModel, request.Prompt, request.Options.Size, request.Options.Duration, request.Inputs.Images, request.Inputs.Videos)
}

func normalizeModelID(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}
