package llm

import (
	"bufio"
	"bytes"
	"clothing/internal/config"
	"clothing/internal/entity"
	"clothing/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

type GeminiService struct {
	httpClient   *http.Client
	GeminiAPIKey string

	modelLookup map[string]struct{}
	models      []entity.LlmModel
}

func NewGeminiService(cfg config.Config) (*GeminiService, error) {
	if strings.TrimSpace(cfg.GeminiAPIKey) == "" {
		return nil, errors.New("gemini api key is not configured")
	}

	httpClient := &http.Client{}

	models := []entity.LlmModel{
		{
			ID:          "gemini-2.5-flash-image-preview",
			Name:        "Gemini 2.5 Flash Image Preview",
			Description: "最新的实验性模型，支持图像生成",
		},
	}

	g := &GeminiService{
		httpClient:   httpClient,
		GeminiAPIKey: cfg.GeminiAPIKey,
		models:       models,
		modelLookup: func() map[string]struct{} {
			lookup := make(map[string]struct{}, len(models))
			for _, model := range models {
				lookup[model.ID] = struct{}{}
			}
			return lookup
		}(),
	}

	return g, nil
}

func (g *GeminiService) ProviderID() string {
	return "gemini"
}

func (g *GeminiService) Provider() entity.LlmProvider {
	return entity.LlmProvider{
		ID:     "google",
		Name:   "Google Gemini",
		Models: g.Models(),
	}
}
func (g *GeminiService) Models() []entity.LlmModel {
	return g.models
}
func (g *GeminiService) SupportsModel(modelID string) bool {
	if g == nil || modelID == "" {
		return false
	}
	_, ok := g.modelLookup[modelID]
	return ok
}

func (g *GeminiService) GenerateImages(ctx context.Context, request entity.GenerateImageRequest) (string, string, error) {
	logger := providerLogger(ctx, g.ProviderID(), request.Model)
	logger.WithFields(logrus.Fields{
		"prompt_length":       len([]rune(request.Prompt)),
		"prompt_preview":      logSnippet(request.Prompt),
		"reference_image_cnt": len(request.Images),
	}).Info("llm_generate_images_start")

	parts := []geminiContentPart{{Text: request.Prompt}}
	for idx, image := range request.Images {
		image = strings.TrimSpace(image)
		if image == "" {
			logger.WithField("image_index", idx).Warn("llm_generate_images_skip_empty_reference")
			continue
		}
		mime, data := utils.SplitDataURL(image)
		if data == "" {
			logger.WithField("image_index", idx).Warn("llm_generate_images_skip_invalid_reference")
			continue
		}
		parts = append(parts, geminiContentPart{
			InlineData: &geminiInlineData{
				MimeType: mime,
				Data:     data,
			},
		})
	}

	payload := geminiRequest{
		Contents: []geminiContent{
			{Role: "user", Parts: parts},
		},
		GenerationConfig: &geminiConfig{
			MaxOutputTokens: 2048,
			Temperature:     0.8,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		logger.WithError(err).Error("llm_generate_images_payload_marshal_failed")
		return "", "", err
	}

	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", request.Model, g.GeminiAPIKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		logger.WithError(err).Error("llm_generate_images_request_build_failed")
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")

	logger.WithFields(logrus.Fields{
		"endpoint":      endpoint,
		"payload_bytes": len(body),
	}).Info("llm_generate_images_request_send")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		logger.WithError(err).Error("llm_generate_images_request_failed")
		return "", "", err
	}
	defer resp.Body.Close()

	logger.WithField("status", resp.StatusCode).Info("llm_generate_images_response_status")
	if resp.StatusCode >= http.StatusBadRequest {
		respBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			logger.WithFields(logrus.Fields{
				"status": resp.StatusCode,
			}).WithError(readErr).Error("llm_generate_images_response_read_failed")
			return "", "", fmt.Errorf("gemini request failed with status %d", resp.StatusCode)
		}
		logger.WithFields(logrus.Fields{
			"status":       resp.StatusCode,
			"body_preview": logSnippet(string(respBody)),
		}).Warn("llm_generate_images_response_error")
		var apiErr geminiErrorResponse
		if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Error.Message != "" {
			return "", "", errors.New(apiErr.Error.Message)
		}
		return "", "", fmt.Errorf("gemini request failed with status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 1024), 16*1024*1024)

	var (
		imageData     string
		imageMimeType string
		textBuilder   strings.Builder
		rawBuffer     strings.Builder
		chunkIndex    int
	)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		rawBuffer.WriteString(line)
		rawBuffer.WriteByte('\n')
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data:") {
			logger.WithFields(logrus.Fields{
				"chunk_index":  chunkIndex,
				"line_preview": logSnippet(line),
			}).Info("llm_generate_images_stream_event")
			continue
		}

		payloadLine := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payloadLine == "" {
			logger.WithField("chunk_index", chunkIndex).Warn("llm_generate_images_stream_empty_payload")
			continue
		}
		if payloadLine == "[DONE]" {
			logger.WithField("chunks", chunkIndex).Info("llm_generate_images_stream_completed")
			break
		}

		chunkIndex++
		logger.WithFields(logrus.Fields{
			"chunk_index":     chunkIndex,
			"payload_len":     len(payloadLine),
			"payload_preview": logSnippet(payloadLine),
		}).Info("llm_generate_images_stream_chunk_raw")

		var chunk geminiResponse
		if err := json.Unmarshal([]byte(payloadLine), &chunk); err != nil {
			logger.WithField("chunk_index", chunkIndex).WithError(err).Error("llm_generate_images_stream_chunk_unmarshal_failed")
			return "", "", err
		}

		var (
			chunkTextBuilder strings.Builder
			chunkImageBytes  int
			finishReasons    []string
		)

		for _, candidate := range chunk.Candidates {
			if candidate.FinishReason != "" {
				finishReasons = append(finishReasons, candidate.FinishReason)
			}
			for _, part := range candidate.Content.Parts {
				if part.InlineData != nil && part.InlineData.Data != "" {
					if part.InlineData.MimeType != "" {
						imageMimeType = part.InlineData.MimeType
					} else if imageMimeType == "" {
						imageMimeType = "image/png"
					}
					imageData += part.InlineData.Data
					chunkImageBytes += len(part.InlineData.Data)
				}
				if text := strings.TrimSpace(part.Text); text != "" {
					if textBuilder.Len() > 0 {
						textBuilder.WriteString("\n")
					}
					textBuilder.WriteString(text)

					if chunkTextBuilder.Len() > 0 {
						chunkTextBuilder.WriteString("\n")
					}
					chunkTextBuilder.WriteString(text)
				}
			}
		}

		chunkText := chunkTextBuilder.String()
		fields := logrus.Fields{
			"chunk_index":       chunkIndex,
			"chunk_image_bytes": chunkImageBytes,
			"chunk_text_len":    len([]rune(chunkText)),
		}
		if chunkText != "" {
			fields["chunk_text_preview"] = logSnippet(chunkText)
		}
		if len(finishReasons) > 0 {
			fields["finish_reasons"] = strings.Join(finishReasons, ",")
		}
		logger.WithFields(fields).Info("llm_generate_images_stream_chunk_processed")
	}

	if err := scanner.Err(); err != nil {
		logger.WithError(err).Error("llm_generate_images_stream_error")
		return "", "", err
	}

	textResult := strings.TrimSpace(textBuilder.String())
	logger.WithFields(logrus.Fields{
		"chunks":            chunkIndex,
		"image_bytes_total": len(imageData),
		"text_length":       len([]rune(textResult)),
		"text_preview":      logSnippet(textResult),
		"image_mime_type":   imageMimeType,
	}).Info("llm_generate_images_stream_summary")

	if imageData != "" {
		if imageMimeType == "" {
			imageMimeType = "image/png"
		}
		logger.WithFields(logrus.Fields{
			"has_text": textResult != "",
			"result":   "image",
		}).Info("llm_generate_images_success")
		return fmt.Sprintf("data:%s;base64,%s", imageMimeType, imageData), textResult, nil
	}
	if textResult != "" {
		logger.WithFields(logrus.Fields{
			"has_image": false,
			"result":    "text",
		}).Info("llm_generate_images_success")
		return "", textResult, nil
	}

	respBody := strings.TrimSpace(rawBuffer.String())
	if respBody != "" {
		logger.Info("llm_generate_images_stream_fallback")
		var apiResponse geminiResponse
		if err := json.Unmarshal([]byte(respBody), &apiResponse); err == nil {
			for _, candidate := range apiResponse.Candidates {
				for _, part := range candidate.Content.Parts {
					if part.InlineData != nil && part.InlineData.Data != "" {
						mimeType := part.InlineData.MimeType
						if mimeType == "" {
							mimeType = "image/png"
						}
						logger.WithField("fallback", true).Info("llm_generate_images_success")
						return fmt.Sprintf("data:%s;base64,%s", mimeType, part.InlineData.Data), "", nil
					}
					if strings.TrimSpace(part.Text) != "" {
						result := strings.TrimSpace(part.Text)
						logger.WithFields(logrus.Fields{
							"fallback":     true,
							"result":       "text",
							"text_preview": logSnippet(result),
						}).Info("llm_generate_images_success")
						return "", result, nil
					}
				}
			}
		}
	}

	logger.Warn("llm_generate_images_no_parseable_content")
	return "", "", errors.New("gemini response did not include image data")
}

type geminiContentPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inlineData,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiContent struct {
	Role  string              `json:"role"`
	Parts []geminiContentPart `json:"parts"`
}

type geminiConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	Temperature     float32 `json:"temperature,omitempty"`
}

type geminiRequest struct {
	Contents         []geminiContent `json:"contents"`
	GenerationConfig *geminiConfig   `json:"generationConfig,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiContentPart `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason,omitempty"`
	} `json:"candidates"`
}

type geminiErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
