package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

const defaultDashscopeGenerationURL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation"

type dashscopeContent struct {
	Image string `json:"image,omitempty"`
	Text  string `json:"text,omitempty"`
}

type dashscopeMessage struct {
	Role    string             `json:"role"`
	Content []dashscopeContent `json:"content"`
}

type dashscopeRequest struct {
	Model      string              `json:"model"`
	Input      dashscopeInput      `json:"input"`
	Parameters dashscopeParameters `json:"parameters,omitempty"`
}

type dashscopeInput struct {
	Messages []dashscopeMessage `json:"messages"`
}

type dashscopeParameters struct {
	NegativePrompt string `json:"negative_prompt,omitempty"`
	Watermark      *bool  `json:"watermark,omitempty"`
}

type dashscopeResponse struct {
	Output     dashscopeOutput `json:"output"`
	RequestID  string          `json:"request_id"`
	Code       string          `json:"code"`
	Message    string          `json:"message"`
	StatusCode int             `json:"status_code"`
}

type dashscopeOutput struct {
	Choices []dashscopeChoice `json:"choices"`
}

type dashscopeChoice struct {
	FinishReason string             `json:"finish_reason"`
	Message      dashscopeChoiceMsg `json:"message"`
}

type dashscopeChoiceMsg struct {
	Role    string             `json:"role"`
	Content []dashscopeContent `json:"content"`
}

func GenerateImagesByDashscopeProtocol(ctx context.Context, apiKey, endpoint, model, prompt string, base64Images []string) (imageDataURLs []string, assistantText string, err error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, "", errors.New("api key missing")
	}

	trimmedPrompt := strings.TrimSpace(prompt)
	if trimmedPrompt == "" {
		return nil, "", errors.New("prompt is empty")
	}

	targetEndpoint := strings.TrimSpace(endpoint)
	if targetEndpoint == "" {
		targetEndpoint = defaultDashscopeGenerationURL
	}

	logrus.WithFields(logrus.Fields{
		"model":         model,
		"prompt_length": len(trimmedPrompt),
		"image_count":   len(base64Images),
	}).Info("dashscope_generate_images_start")

	messageContents := make([]dashscopeContent, 0, len(base64Images)+1)
	for _, img := range base64Images {
		img = strings.TrimSpace(img)
		if img == "" {
			continue
		}
		messageContents = append(messageContents, dashscopeContent{Image: img})
	}

	messageContents = append(messageContents, dashscopeContent{Text: trimmedPrompt})

	watermark := false
	reqBody := dashscopeRequest{
		Model: model,
		Input: dashscopeInput{
			Messages: []dashscopeMessage{
				{
					Role:    "user",
					Content: messageContents,
				},
			},
		},
		Parameters: dashscopeParameters{
			Watermark:      &watermark,
			NegativePrompt: "",
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", fmt.Errorf("dashscope marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetEndpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, "", fmt.Errorf("dashscope create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("dashscope request: %w", err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, "", fmt.Errorf("dashscope read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   buf.String(),
		}).Error("dashscope generate images http error")
		return nil, "", fmt.Errorf("dashscope http %d: %s", resp.StatusCode, buf.String())
	}

	var result dashscopeResponse
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return nil, "", fmt.Errorf("dashscope unmarshal response: %w", err)
	}

	if result.Code != "" && strings.ToLower(result.Code) != "success" {
		msg := result.Message
		if msg == "" {
			msg = "dashscope error"
		}
		return nil, "", fmt.Errorf("dashscope api error: %s", msg)
	}

	if len(result.Output.Choices) == 0 {
		return nil, "", errors.New("dashscope no choices in response")
	}
	seen := make(map[string]struct{})

	for _, choice := range result.Output.Choices {
		for _, content := range choice.Message.Content {
			if raw := strings.TrimSpace(content.Image); raw != "" {
				if _, exists := seen[raw]; !exists {
					seen[raw] = struct{}{}
					imageDataURLs = append(imageDataURLs, raw)
				}
			}
			if text := strings.TrimSpace(content.Text); text != "" {
				if assistantText != "" {
					assistantText += "\n"
				}
				assistantText += text
			}
		}
	}

	if len(imageDataURLs) == 0 {
		return nil, assistantText, errors.New("dashscope no image found in response")
	}

	return imageDataURLs, assistantText, nil
}
