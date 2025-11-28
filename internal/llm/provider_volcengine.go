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
	requestedSize := request.Options.Size
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
			err := fmt.Errorf("volcengine model %q does not support size %q", request.ModelID, request.Options.Size)
			logrus.WithError(err).Warn("llm_generate_content_invalid_size")
			return nil, err
		}
	}

	return GenerateContentByVolcengineProtocol(ctx, p.apiKey, dbModel.ModelID, request.Prompt, requestedSize, request.Inputs.Images)
}
