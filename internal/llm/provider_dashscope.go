package llm

import (
	"clothing/internal/config"
	"clothing/internal/entity"
	"clothing/internal/utils"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type Dashscope struct {
	client *openai.Client
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

	openAIConfig := openai.DefaultConfig(cfg.OpenRouterAPIKey)
	openAIConfig.BaseURL = "https://dashscope.aliyuncs.com/api/v1"

	d := &Dashscope{
		apiKey: cfg.DashscopeAPIKey,
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

func (d *Dashscope) GenerateImages(ctx context.Context, request entity.GenerateImageRequest) (string, string, error) {
	logger := providerLogger(ctx, d.ProviderID(), request.Model)
	logger.WithFields(logrus.Fields{
		"prompt_length":       len([]rune(request.Prompt)),
		"prompt_preview":      logSnippet(request.Prompt),
		"reference_image_cnt": len(request.Images),
	}).Info("llm_generate_images_start")

	if !d.SupportsModel(request.Model) {
		err := fmt.Errorf("dashscope model %q is not supported", request.Model)
		logger.WithError(err).Warn("llm_generate_images_invalid_model")
		return "", "", err
	}

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

	stream, err := d.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
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
