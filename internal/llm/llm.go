package llm

import (
	"clothing/internal/entity"
	"context"
	"errors"
	"strings"
)

// ModelCapabilities describes the capabilities of a model.
type ModelCapabilities struct {
	InputModalities    []string
	OutputModalities   []string
	SupportsAsync      bool
	SupportsStream     bool
	SupportsCancel     bool
	MaxImages          int
	SupportedSizes     []string
	SupportedDurations []int
}

// AIService defines the interface for AI content generation services.
type AIService interface {
	// GenerateContent performs synchronous content generation.
	GenerateContent(ctx context.Context, request entity.GenerateContentRequest, dbModel entity.DbModel) (*entity.GenerateContentResponse, error)

	// Capabilities returns the capabilities of a model.
	Capabilities(model entity.DbModel) *ModelCapabilities

	// Validate checks if a request is valid for a given model.
	Validate(request entity.GenerateContentRequest, model entity.DbModel) error
}

// BaseProvider provides default implementations for common AIService methods.
// Embed this in your provider implementations to avoid boilerplate.
type BaseProvider struct {
	ProviderID   string
	ProviderName string
	APIKey       string
}

// Capabilities returns the capabilities of a model based on its configuration.
func (b *BaseProvider) Capabilities(model entity.DbModel) *ModelCapabilities {
	return &ModelCapabilities{
		InputModalities:    model.InputModalities,
		OutputModalities:   model.OutputModalities,
		MaxImages:          model.MaxImages,
		SupportedSizes:     model.SupportedSizes,
		SupportedDurations: model.SupportedDurations,
		SupportsStream:     model.SupportsStreaming,
		SupportsCancel:     model.SupportsCancel,
		SupportsAsync:      model.IsVideoModel(), // Video models typically use async generation
	}
}

// Validate performs basic validation on the request.
func (b *BaseProvider) Validate(request entity.GenerateContentRequest, model entity.DbModel) error {
	if strings.TrimSpace(request.Prompt) == "" {
		return errors.New("prompt is required")
	}

	// Check if the request has required images for image-to-image models
	if model.GenerationMode == "image_to_image" || model.GenerationMode == "image_to_video" {
		images := request.GetImages()
		if len(images) == 0 {
			return errors.New("at least one input image is required for this model")
		}
	}

	return nil
}
