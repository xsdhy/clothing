package llm

import (
	"clothing/internal/config"
	"clothing/internal/entity"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type Volcengine struct {
	apiKey string

	models      []entity.LlmModel
	modelLookup map[string]struct{}
}

func NewVolcengine(cfg config.Config) (*Volcengine, error) {
	if strings.TrimSpace(cfg.VolcengineAPIKey) == "" {
		return nil, errors.New("volcengine api key is not configured")
	}

	models := []entity.LlmModel{
		{
			ID:          "doubao-seedream-4-0-250828",
			Name:        "Doubao Seedream 4.0",
			Description: "火山引擎图像生成模型",
		},
	}

	v := &Volcengine{
		apiKey: cfg.VolcengineAPIKey,
		models: models,
		modelLookup: func() map[string]struct{} {
			lookup := make(map[string]struct{}, len(models))
			for _, model := range models {
				lookup[model.ID] = struct{}{}
			}
			return lookup
		}(),
	}

	return v, nil
}

func (v *Volcengine) ProviderID() string {
	return "volcengine"
}

func (v *Volcengine) Provider() entity.LlmProvider {
	return entity.LlmProvider{
		ID:     v.ProviderID(),
		Name:   "Volcengine",
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

func (v *Volcengine) GenerateImages(ctx context.Context, request entity.GenerateImageRequest) (string, string, error) {
	logrus.WithFields(logrus.Fields{
		"prompt_preview":      request.Prompt,
		"reference_image_cnt": len(request.Images),
	}).Info("llm_generate_images_start")

	if !v.SupportsModel(request.Model) {
		err := fmt.Errorf("volcengine model %q is not supported", request.Model)
		logrus.WithError(err).Warn("llm_generate_images_invalid_model")
		return "", "", err
	}

	return GenerateImagesByVolcengineProtocol(ctx, v.apiKey, request.Model, request.Prompt, request.Images)
}
