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
}

func NewVolcengine(provider *entity.DbProvider) (*Volcengine, error) {
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

	return &Volcengine{
		providerID:   provider.ID,
		providerName: name,
		apiKey:       apiKey,
	}, nil
}

func (p *Volcengine) GenerateContent(ctx context.Context, request entity.GenerateContentRequest, dbModel entity.DbModel) (*entity.GenerateContentResponse, error) {
	requestedSize := request.GetSize()
	if requestedSize != "" {
		if len(dbModel.SupportedSizes) == 0 {
			err := fmt.Errorf("volcengine model %q does not allow custom size", request.ModelID)
			logrus.WithError(err).Warn("llm_generate_content_invalid_size")
			return nil, err
		}
		valid := false
		for _, allowed := range dbModel.SupportedSizes {
			if strings.EqualFold(allowed, requestedSize) {
				valid = true
				break
			}
		}
		if !valid {
			err := fmt.Errorf("volcengine model %q does not support size %q", request.ModelID, requestedSize)
			logrus.WithError(err).Warn("llm_generate_content_invalid_size")
			return nil, err
		}
	}

	if dbModel.IsVideoModel() {
		return GenerateVolcengineVideo(ctx, p.apiKey, dbModel, request.Prompt, request.GetSize(), request.GetDuration(), request.GetImages())
	}

	return GenerateContentByVolcengineProtocol(ctx, p.apiKey, dbModel.ModelID, request.Prompt, requestedSize, request.GetImages())
}

// Capabilities returns the capabilities of the model.
func (p *Volcengine) Capabilities(model entity.DbModel) *ModelCapabilities {
	return &ModelCapabilities{
		InputModalities:    model.InputModalities,
		OutputModalities:   model.OutputModalities,
		MaxImages:          model.MaxImages,
		SupportedSizes:     model.SupportedSizes,
		SupportedDurations: model.SupportedDurations,
		SupportsStream:     true, // Volcengine uses streaming
		SupportsCancel:     model.SupportsCancel,
		SupportsAsync:      model.IsVideoModel(),
	}
}

// Validate checks if the request is valid for the model.
func (p *Volcengine) Validate(request entity.GenerateContentRequest, model entity.DbModel) error {
	if strings.TrimSpace(request.Prompt) == "" {
		return errors.New("prompt is required")
	}

	// Validate size if provided
	requestedSize := request.GetSize()
	if requestedSize != "" && len(model.SupportedSizes) > 0 {
		valid := false
		for _, allowed := range model.SupportedSizes {
			if strings.EqualFold(allowed, requestedSize) {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("size %q not supported, available: %v", requestedSize, model.SupportedSizes)
		}
	}

	return nil
}
