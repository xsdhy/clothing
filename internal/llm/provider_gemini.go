package llm

import (
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
)

type GeminiService struct {
	httpClient   *http.Client
	GeminiAPIKey string

	modelLookup map[string]struct{}
}

func NewGeminiService(cfg config.Config) (*GeminiService, error) {
	httpClient := &http.Client{}
	g := &GeminiService{
		httpClient:   httpClient,
		GeminiAPIKey: cfg.GeminiAPIKey,
	}
	models := g.Models()
	for _, model := range models {
		g.modelLookup[model.ID] = struct{}{}
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
	return []entity.LlmModel{
		{
			ID:          "gemini-2.5-flash-image-preview",
			Name:        "Gemini 2.5 Flash Image Preview",
			Description: "最新的实验性模型，支持图像生成",
		},
	}
}
func (g *GeminiService) SupportsModel(modelID string) bool {
	if g == nil || modelID == "" {
		return false
	}
	_, ok := g.modelLookup[modelID]
	return ok
}

func (g *GeminiService) GenerateImages(ctx context.Context, request entity.GenerateImageRequest) (string, string, error) {

	parts := []geminiContentPart{{Text: request.Prompt}}
	for _, image := range request.Images {
		image = strings.TrimSpace(image)
		if image == "" {
			continue
		}
		mime, data := utils.SplitDataURL(image)
		if data == "" {
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
		return "", "", err
	}

	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", request.Model, g.GeminiAPIKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		var apiErr geminiErrorResponse
		if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Error.Message != "" {
			return "", "", errors.New(apiErr.Error.Message)
		}
		return "", "", fmt.Errorf("gemini request failed with status %d", resp.StatusCode)
	}

	var apiResponse geminiResponse
	if err := json.Unmarshal(respBody, &apiResponse); err != nil {
		return "", "", err
	}

	for _, candidate := range apiResponse.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil && part.InlineData.Data != "" {
				mimeType := part.InlineData.MimeType
				if mimeType == "" {
					mimeType = "image/png"
				}
				return fmt.Sprintf("data:%s;base64,%s", mimeType, part.InlineData.Data), "", nil
			}
			if part.Text != "" {
				return "", part.Text, nil
			}
		}
	}

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
