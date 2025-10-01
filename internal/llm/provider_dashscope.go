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

type Dashscope struct {
	apiKey string

	models      []entity.LlmModel
	modelLookup map[string]struct{}
}

func NewDashscope(cfg config.Config) (*Dashscope, error) {
	if strings.TrimSpace(cfg.DashscopeAPIKey) == "" {
		return nil, errors.New("dashscope api key is not configured")
	}

	models := []entity.LlmModel{
		{
			ID:          "qwen-image-edit",
			Name:        "qwen-image-edit",
			Description: "qwen-image-edit",
		},
	}

	d := &Dashscope{
		apiKey: cfg.DashscopeAPIKey,

		models: models,
		modelLookup: func() map[string]struct{} {
			lookup := make(map[string]struct{}, len(models))
			for _, model := range models {
				lookup[model.ID] = struct{}{}
			}
			return lookup
		}(),
	}

	return d, nil
}

func (d *Dashscope) ProviderID() string {
	return "dashscope"
}

func (d *Dashscope) Provider() entity.LlmProvider {
	return entity.LlmProvider{
		ID:     d.ProviderID(),
		Name:   "Dashscope",
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
		"size":                strings.TrimSpace(request.Size),
	}).Info("llm_generate_images_start")

	if !d.SupportsModel(request.Model) {
		err := fmt.Errorf("dashscope model %q is not supported", request.Model)
		logrus.WithError(err).Warn("llm_generate_images_invalid_model")
		return nil, "", err
	}

	return GenerateImagesByDashscopeProtocol(ctx, d.apiKey, request.Model, request.Prompt, request.Images)
}
