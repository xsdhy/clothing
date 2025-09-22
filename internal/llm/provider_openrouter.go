package llm

import (
	"clothing/internal/config"
	"clothing/internal/entity"
	"clothing/internal/utils"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type OpenRouter struct {
	client *openai.Client

	models      []entity.LlmModel
	modelLookup map[string]struct{}
}

func NewOpenRouter(cfg config.Config) (*OpenRouter, error) {
	if strings.TrimSpace(cfg.OpenRouterAPIKey) == "" {
		return nil, errors.New("openrouter api key is not configured")
	}

	openAIConfig := openai.DefaultConfig(cfg.OpenRouterAPIKey)
	openAIConfig.BaseURL = "https://openrouter.ai/api/v1"

	models := []entity.LlmModel{
		{
			ID:          "google/gemini-2.5-flash-image-preview",
			Name:        "Gemini 2.5 Flash Image Preview",
			Description: "Google 的高性能模型",
		},
	}

	o := &OpenRouter{
		client: openai.NewClientWithConfig(openAIConfig),
		models: models,
		modelLookup: func() map[string]struct{} {
			lookup := make(map[string]struct{}, len(models))
			for _, model := range models {
				lookup[model.ID] = struct{}{}
			}
			return lookup
		}(),
	}

	return o, nil
}
func (o *OpenRouter) ProviderID() string {
	return "openrouter"
}

func (o *OpenRouter) Provider() entity.LlmProvider {
	return entity.LlmProvider{
		ID:     "openrouter",
		Name:   "OpenRouter",
		Models: o.Models(),
	}
}

func (o *OpenRouter) Models() []entity.LlmModel {
	return o.models
}

func (o *OpenRouter) SupportsModel(modelID string) bool {
	if o == nil || modelID == "" {
		return false
	}
	_, ok := o.modelLookup[modelID]
	return ok
}

func (o *OpenRouter) GenerateImages(ctx context.Context, request entity.GenerateImageRequest) (string, string, error) {
	logger := providerLogger(ctx, o.ProviderID(), request.Model)
	logger.WithFields(logrus.Fields{
		"prompt_length":       len([]rune(request.Prompt)),
		"prompt_preview":      logSnippet(request.Prompt),
		"reference_image_cnt": len(request.Images),
	}).Info("llm_generate_images_start")

	parts := []openai.ChatMessagePart{
		{
			Type: openai.ChatMessagePartTypeText,
			Text: request.Prompt,
		},
	}

	for idx, image := range request.Images {
		image = strings.TrimSpace(image)
		if image == "" {
			logger.WithField("image_index", idx).Warn("llm_generate_images_skip_empty_reference")
			continue
		}
		parts = append(parts, openai.ChatMessagePart{
			Type:     openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{URL: utils.EnsureDataURL(image)},
		})
	}

	stream, err := o.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: request.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:         openai.ChatMessageRoleUser,
				MultiContent: parts,
			},
		},
		MaxTokens: 1000,
	})
	if err != nil {
		logger.WithError(err).Error("llm_generate_images_stream_create_failed")
		return "", "", err
	}
	defer stream.Close()

	var (
		builder    strings.Builder
		chunkIndex int
	)

	for {
		resp, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			logger.WithField("chunks", chunkIndex).Info("llm_generate_images_stream_completed")
			break
		}
		if recvErr != nil {
			logger.WithFields(logrus.Fields{
				"chunks": chunkIndex,
			}).WithError(recvErr).Error("llm_generate_images_stream_recv_failed")
			return "", "", recvErr
		}

		chunkIndex++
		for choiceIdx, choice := range resp.Choices {
			fields := logrus.Fields{
				"chunk_index":         chunkIndex,
				"choice_index":        choiceIdx,
				"delta_content_len":   len(choice.Delta.Content),
				"delta_reasoning_len": len(choice.Delta.ReasoningContent),
				"finish_reason":       choice.FinishReason,
				"index":               choice.Index,
			}
			if choice.Delta.Content != "" {
				builder.WriteString(choice.Delta.Content)
				fields["delta_content_preview"] = logSnippet(choice.Delta.Content)
			}
			if choice.Delta.ReasoningContent != "" {
				builder.WriteString(choice.Delta.ReasoningContent)
				fields["delta_reasoning_preview"] = logSnippet(choice.Delta.ReasoningContent)
			}
			logger.WithFields(fields).Info("llm_generate_images_stream_chunk")
		}
	}

	content := builder.String()
	logger.WithFields(logrus.Fields{
		"combined_len":     len(content),
		"combined_preview": logSnippet(content),
	}).Info("llm_generate_images_stream_assembled")
	if strings.TrimSpace(content) == "" {
		logger.Warn("llm_generate_images_stream_empty_content")
		return "", "", errors.New("model did not return an image")
	}

	image, text := parseStreamedContent(content)
	if image != "" || text != "" {
		logger.WithFields(logrus.Fields{
			"has_image": image != "",
			"text_len":  len(text),
		}).Info("llm_generate_images_success")
		return image, text, nil
	}

	logger.Warn("llm_generate_images_no_parseable_content")
	return "", "", errors.New("model did not return an image")
}
