package llm

import (
	"clothing/internal/entity"
	"context"
	"errors"
	"strings"
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

func (p *Dashscope) GenerateContent(ctx context.Context, request entity.GenerateContentRequest, dbModel entity.DbModel) (*entity.GenerateContentResponse, error) {
	if dbModel.IsVideoModel() {
		return GenerateDashscopeVideo(ctx, p.apiKey, p.endpoint, dbModel, request.Prompt, request.GetSize(), request.GetDuration(), request.GetImages())
	}
	return GenerateImageByDashscopeProtocol(ctx, p.apiKey, p.endpoint, dbModel, request.Prompt, request.GetSize(), request.GetDuration(), request.GetImages(), request.GetVideos())
}

// Capabilities returns the capabilities of the model.
func (p *Dashscope) Capabilities(model entity.DbModel) *ModelCapabilities {
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
func (p *Dashscope) Validate(request entity.GenerateContentRequest, model entity.DbModel) error {
	if strings.TrimSpace(request.Prompt) == "" {
		return errors.New("prompt is required")
	}
	return nil
}

func normalizeModelID(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}
