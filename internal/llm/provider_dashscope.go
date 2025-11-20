package llm

import (
	"clothing/internal/entity"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type Dashscope struct {
	providerID   string
	providerName string

	apiKey   string
	endpoint string

	models      []entity.LlmModel
	modelLookup map[string]struct{}
}

func NewDashscope(provider *entity.DbProvider, models []entity.DbModel) (*Dashscope, error) {
	if provider == nil {
		return nil, errors.New("dashscope provider config is nil")
	}

	apiKey := strings.TrimSpace(provider.APIKey)
	if apiKey == "" {
		return nil, errors.New("dashscope api key is not configured")
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
		return nil, errors.New("dashscope has no active models configured")
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
		models:       activeModels,
		modelLookup:  modelLookup,
	}, nil
}

func (d *Dashscope) ProviderID() string {
	return d.providerID
}

func (d *Dashscope) Provider() entity.LlmProvider {
	return entity.LlmProvider{
		ID:     d.providerID,
		Name:   d.providerName,
		Models: d.Models(),
	}
}

func (d *Dashscope) Models() []entity.LlmModel {
	return d.models
}

func (d *Dashscope) SupportsModel(modelID string) bool {
	if d == nil || modelID == "" {
		return false
	}
	if len(d.modelLookup) == 0 {
		return true
	}
	_, ok := d.modelLookup[modelID]
	return ok
}

func (d *Dashscope) GenerateImages(ctx context.Context, request entity.GenerateImageRequest) ([]string, string, error) {

	logrus.WithFields(logrus.Fields{

		"prompt_preview":      request.Prompt,
		"reference_image_cnt": len(request.Images),
		"reference_video_cnt": len(request.Videos),
		"size":                strings.TrimSpace(request.Size),
		"duration":            request.Duration,
	}).Info("llm_generate_images_start")

	if !d.SupportsModel(request.Model) {
		err := fmt.Errorf("dashscope model %q is not supported", request.Model)
		logrus.WithError(err).Warn("llm_generate_images_invalid_model")
		return nil, "", err
	}

	return GenerateImagesByDashscopeProtocol(ctx, d.apiKey, d.endpoint, request.Model, request.Prompt, request.Size, request.Duration, request.Images, request.Videos)
}
