package llm

import (
	"clothing/internal/entity"
	"context"
)

type AIService interface {
	ProviderID() string
	Provider() entity.LlmProvider
	Models() []entity.LlmModel
	SupportsModel(modelID string) bool

	GenerateContent(ctx context.Context, request entity.GenerateContentRequest) ([]string, string, error)
}
