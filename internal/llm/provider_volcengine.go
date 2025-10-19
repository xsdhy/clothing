package llm

import (
	"clothing/internal/entity"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type Volcengine struct {
	providerID   string
	providerName string

	apiKey string

	models      []entity.LlmModel
	modelLookup map[string]struct{}
	modelByID   map[string]entity.LlmModel
}

func NewVolcengine(provider *entity.DbProvider, models []entity.DbModel) (*Volcengine, error) {
	if provider == nil {
		return nil, errors.New("volcengine provider config is nil")
	}

	apiKey := strings.TrimSpace(provider.APIKey)
	if apiKey == "" {
		return nil, errors.New("volcengine api key is not configured")
	}

	name := strings.TrimSpace(provider.Name)
	if name == "" {
		name = provider.ID
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
		modelLookup[llmModel.ID] = struct{}{}
		modelByID[llmModel.ID] = llmModel
	}

	if len(activeModels) == 0 {
		return nil, errors.New("volcengine has no active models configured")
	}

	return &Volcengine{
		providerID:   provider.ID,
		providerName: name,
		apiKey:       apiKey,
		models:       activeModels,
		modelLookup:  modelLookup,
		modelByID:    modelByID,
	}, nil
}

func (v *Volcengine) ProviderID() string {
	return v.providerID
}

func (v *Volcengine) Provider() entity.LlmProvider {
	return entity.LlmProvider{
		ID:     v.providerID,
		Name:   v.providerName,
		Models: v.Models(),
	}
}

func (v *Volcengine) Models() []entity.LlmModel {
	return v.models
}

func (v *Volcengine) SupportsModel(modelID string) bool {
	if v == nil || modelID == "" {
		return false
	}
	if len(v.modelLookup) == 0 {
		return true
	}
	_, ok := v.modelLookup[modelID]
	return ok
}

func (v *Volcengine) GenerateImages(ctx context.Context, request entity.GenerateImageRequest) ([]string, string, error) {
	logrus.WithFields(logrus.Fields{
		"prompt_preview":      request.Prompt,
		"reference_image_cnt": len(request.Images),
		"size":                strings.TrimSpace(request.Size),
	}).Info("llm_generate_images_start")

	if !v.SupportsModel(request.Model) {
		err := fmt.Errorf("volcengine model %q is not supported", request.Model)
		logrus.WithError(err).Warn("llm_generate_images_invalid_model")
		return nil, "", err
	}

	requestedSize := strings.TrimSpace(request.Size)
	model, ok := v.modelByID[request.Model]
	if requestedSize != "" {
		if !ok || len(model.Inputs.SupportedSizes) == 0 {
			err := fmt.Errorf("volcengine model %q does not allow custom size", request.Model)
			logrus.WithError(err).Warn("llm_generate_images_invalid_size")
			return nil, "", err
		}
		valid := false
		for _, allowed := range model.Inputs.SupportedSizes {
			if strings.EqualFold(allowed, requestedSize) {
				valid = true
				break
			}
		}
		if !valid {
			err := fmt.Errorf("volcengine model %q does not support size %q", request.Model, request.Size)
			logrus.WithError(err).Warn("llm_generate_images_invalid_size")
			return nil, "", err
		}
	}

	return GenerateImagesByVolcengineProtocol(ctx, v.apiKey, request.Model, request.Prompt, requestedSize, request.Images)
}
