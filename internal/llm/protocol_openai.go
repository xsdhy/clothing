package llm

import (
	"clothing/internal/entity"
	"encoding/json"
	"strings"

	"bufio"
	"bytes"
	"context"

	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type orImageURL struct {
	URL string `json:"url"`
}
type orVideoURL struct {
	URL string `json:"url"`
}
type orImage struct {
	Type     string     `json:"type"` // "image_url"
	ImageURL orImageURL `json:"image_url"`
}
type orContentPart struct {
	Type string `json:"type"` // "text"
	Text string `json:"text,omitempty"`
}

type orDelta struct {
	ContentRaw json.RawMessage `json:"content"`
	Images     []orImage       `json:"images"`
}

func (d orDelta) Text() string {
	if len(d.ContentRaw) == 0 {
		return ""
	}

	var rawText string
	if err := json.Unmarshal(d.ContentRaw, &rawText); err == nil {
		return rawText
	}

	var parts []orContentPart
	if err := json.Unmarshal(d.ContentRaw, &parts); err == nil {
		var builder strings.Builder
		for _, part := range parts {
			if part.Text != "" {
				builder.WriteString(part.Text)
			}
		}
		return builder.String()
	}

	return ""
}

type orChoice struct {
	Delta              orDelta `json:"delta"`
	FinishReason       string  `json:"finish_reason"`
	NativeFinishReason string  `json:"native_finish_reason"`
	Index              int     `json:"index"`
}
type orStreamChunk struct {
	Choices []orChoice `json:"choices"`
}

type orMsgPart struct {
	Type     string      `json:"type"` // "text" | "image_url" | "video_url"
	Text     string      `json:"text,omitempty"`
	ImageURL *orImageURL `json:"image_url,omitempty"`
	VideoURL *orVideoURL `json:"video_url,omitempty"`
}
type orMessage struct {
	Role    string      `json:"role"` // "user"
	Content interface{} `json:"content"`
}

// 输入图片 data:URL 也行，http(s) 也行
func makeUserMessage(prompt string, refs, videos []string) orMessage {
	trimmedPrompt := strings.TrimSpace(prompt)
	parts := []orMsgPart{}
	if trimmedPrompt != "" {
		parts = append(parts, orMsgPart{Type: "text", Text: trimmedPrompt})
	}
	for _, r := range refs {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		parts = append(parts, orMsgPart{
			Type:     "image_url",
			ImageURL: &orImageURL{URL: r},
		})
	}

	for _, v := range videos {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		parts = append(parts, orMsgPart{
			Type:     "video_url",
			VideoURL: &orVideoURL{URL: v},
		})
	}
	return orMessage{Role: "user", Content: parts}
}

func GenerateContentByOpenaiProtocol(ctx context.Context, apiKey, baseURL, model, prompt string, refs, videos []string) (*entity.GenerateContentResponse, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("api key missing")
	}

	trimmedPrompt := strings.TrimSpace(prompt)

	logrus.WithFields(logrus.Fields{
		"model":                 model,
		"prompt_length":         len(trimmedPrompt),
		"reference_image_count": len(refs),
		"reference_video_count": len(videos),
		"reference_media_count": len(refs) + len(videos),
		"prompt_preview":        trimmedPrompt,
		"base_url":              truncateForLog(baseURL, 128),
	}).Info("GenerateContentOR called")

	modalities := []string{"text"}
	if len(refs) > 0 {
		modalities = append(modalities, "image")
	}
	if len(videos) > 0 {
		modalities = append(modalities, "video")
	}

	reqBody := map[string]any{
		"model":      model,
		"messages":   []orMessage{makeUserMessage(trimmedPrompt, refs, videos)},
		"modalities": modalities,
		"stream":     true,
	}

	bs, _ := json.Marshal(reqBody)
	req, _ := http.NewRequestWithContext(ctx, "POST", baseURL, bytes.NewReader(bs))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	httpCli := &http.Client{Timeout: 0}
	// SSE 不要超短超时
	resp, err := httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		logrus.WithFields(logrus.Fields{
			"apikey":  apiKey,
			"baseURL": baseURL,
			"status":  resp.StatusCode,
			"body":    buf.String(),
		}).Error("openai generate content failed")
		return nil, fmt.Errorf("openai http %d: %s", resp.StatusCode, buf.String())
	}
	// 处理 SSE 流式响应
	logrus.Info("openai stream response started")

	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	var imageDataURLs []string
	var assistantText string
	nativeFinishReasonText := ""
	seenImages := make(map[string]struct{})
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		logrus.WithFields(logrus.Fields{"data": line}).Info("stream chunk")
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk orStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		for i := 0; i < len(chunk.Choices); i++ {
			choice := chunk.Choices[i]

			// 累积文本和图片 URL
			delta := choice.Delta
			if text := delta.Text(); text != "" {
				assistantText += text
			}

			if choice.FinishReason != "" {
				nativeFinishReasonText += choice.NativeFinishReason
				logrus.WithFields(logrus.Fields{
					"finish_reason":        choice.FinishReason,
					"native_finish_reason": choice.NativeFinishReason,
				}).Info("stream chunk finish reason")
			}

			if len(delta.Images) > 0 {
				for _, img := range delta.Images {
					url := strings.TrimSpace(img.ImageURL.URL)
					if url == "" {
						continue
					}
					if _, exists := seenImages[url]; exists {
						continue
					}
					seenImages[url] = struct{}{}
					imageDataURLs = append(imageDataURLs, url)
				}
			}
		}

	}
	logrus.Info("openai stream response ended")
	if err := sc.Err(); err != nil {
		return &entity.GenerateContentResponse{
			TextContent: strings.TrimSpace(assistantText),
		}, err
	}

	assistantText = strings.TrimSpace(assistantText)
	if len(imageDataURLs) == 0 {
		if assistantText != "" {
			logrus.WithFields(logrus.Fields{
				"native_finish_reason": strings.TrimSpace(nativeFinishReasonText),
				"text_length":          len(assistantText),
			}).Warn("openai stream returned text only without images")
			return &entity.GenerateContentResponse{
				TextContent: assistantText,
			}, nil
		}
		if len(nativeFinishReasonText) > 0 {
			return nil, errors.New(nativeFinishReasonText)
		}
		return nil, errors.New("no image or text in streamed response")
	}
	return &entity.GenerateContentResponse{
		ImageAssets: imageDataURLs,
		TextContent: assistantText,
	}, nil
}
