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
	modelByID   map[string]entity.LlmModel
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
	modelByID := make(map[string]entity.LlmModel, len(models))
	for _, model := range models {
		if !model.IsActive {
			continue
		}
		llmModel := model.ToLlmModel()
		activeModels = append(activeModels, llmModel)
		key := normalizeModelID(llmModel.ID)
		modelLookup[key] = struct{}{}
		modelByID[key] = llmModel
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
		modelByID:    modelByID,
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
	_, ok := d.modelLookup[normalizeModelID(modelID)]
	return ok
}

func (d *Dashscope) GenerateContent(ctx context.Context, request entity.GenerateContentRequest) ([]string, string, error) {

	logrus.WithFields(logrus.Fields{

		"prompt_preview":      request.Prompt,
		"reference_image_cnt": len(request.Inputs.Images),
		"reference_video_cnt": len(request.Inputs.Videos),
		"size":                strings.TrimSpace(request.Options.Size),
		"duration":            request.Options.Duration,
	}).Info("llm_generate_content_start")

	if !d.SupportsModel(request.ModelID) {
		err := fmt.Errorf("dashscope model %q is not supported", request.ModelID)
		logrus.WithError(err).Warn("llm_generate_content_invalid_model")
		return nil, "", err
	}

	modelConfig, ok := d.modelByID[normalizeModelID(request.ModelID)]
	if !ok {
		modelConfig = entity.LlmModel{ID: request.ModelID}
	}

	return GenerateContentByDashscopeProtocol(ctx, d.apiKey, d.endpoint, modelConfig, request.Prompt, request.Options.Size, request.Options.Duration, request.Inputs.Images, request.Inputs.Videos)
}

func normalizeModelID(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}
