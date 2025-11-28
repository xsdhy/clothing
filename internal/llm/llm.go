package llm

import (
	"clothing/internal/entity"
	"context"
)

type AIService interface {
	GenerateContent(ctx context.Context, request entity.GenerateContentRequest, dbModel entity.DbModel) (*entity.GenerateContentResponse, error)
}
