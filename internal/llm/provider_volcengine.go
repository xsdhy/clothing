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
	modelByID   map[string]entity.LlmModel
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
			ImageSizes:  []string{"1K", "2K", "4K"},
		},
	}

	v := &Volcengine{
		apiKey:      cfg.VolcengineAPIKey,
		models:      models,
		modelLookup: make(map[string]struct{}, len(models)),
		modelByID:   make(map[string]entity.LlmModel, len(models)),
	}
	for _, model := range models {
		v.modelLookup[model.ID] = struct{}{}
		v.modelByID[model.ID] = model
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
		if !ok || len(model.ImageSizes) == 0 {
			err := fmt.Errorf("volcengine model %q does not allow custom size", request.Model)
			logrus.WithError(err).Warn("llm_generate_images_invalid_size")
			return nil, "", err
		}
		valid := false
		for _, allowed := range model.ImageSizes {
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
