package llm

import (
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
type orImage struct {
	Type     string     `json:"type"` // "image_url"
	ImageURL orImageURL `json:"image_url"`
}

type orDelta struct {
	Content string    `json:"content"`
	Images  []orImage `json:"images"`
}
type orChoice struct {
	Delta        orDelta `json:"delta"`
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
}
type orStreamChunk struct {
	Choices []orChoice `json:"choices"`
}

type orMsgPart struct {
	Type     string      `json:"type"` // "text" | "image_url"
	Text     string      `json:"text,omitempty"`
	ImageURL *orImageURL `json:"image_url,omitempty"`
}
type orMessage struct {
	Role    string      `json:"role"` // "user"
	Content interface{} `json:"content"`
}

// 输入图片 data:URL 也行，http(s) 也行
func makeUserMessage(prompt string, refs []string) orMessage {
	parts := []orMsgPart{{Type: "text", Text: prompt}}
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
	return orMessage{Role: "user", Content: parts}
}

func GenerateImagesOR(ctx context.Context, apiKey, baseURL, model, prompt string, refs []string) (imageDataURL string, assistantText string, err error) {
	if strings.TrimSpace(apiKey) == "" {
		return "", "", errors.New("openrouter api key missing")
	}

	logrus.WithFields(logrus.Fields{
		"model":                 model,
		"prompt_length":         len(prompt),
		"reference_image_count": len(refs),
		"reference_cnt":         len(refs),
	}).Info("GenerateImagesOR called")

	reqBody := map[string]any{
		"model":      model,
		"messages":   []orMessage{makeUserMessage(prompt, refs)},
		"modalities": []string{"image", "text"},
		"stream":     true,
	}

	bs, _ := json.Marshal(reqBody)
	req, _ := http.NewRequestWithContext(ctx, "POST", baseURL, bytes.NewReader(bs))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	httpCli := &http.Client{Timeout: 0} // SSE 不要超短超时
	resp, err := httpCli.Do(req)
	if err != nil {
		return "", "", err
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
		}).Error("openrouter generate images failed")
		return "", "", fmt.Errorf("openrouter http %d: %s", resp.StatusCode, buf.String())
	}
	// 处理 SSE 流式响应
	logrus.Info("openrouter stream response started")

	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		logrus.WithFields(logrus.Fields{
			"data": data,
		}).Debug("stream chunk")

		var chunk orStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}

		delta := chunk.Choices[0].Delta
		if delta.Content != "" {
			assistantText += delta.Content
		}
		if len(delta.Images) > 0 && delta.Images[0].ImageURL.URL != "" && imageDataURL == "" {
			// 只取第一张（你也可以收集多张）
			imageDataURL = delta.Images[0].ImageURL.URL
		}
	}
	logrus.Info("openrouter stream response ended")
	if err := sc.Err(); err != nil {
		return "", "", err
	}
	if strings.TrimSpace(imageDataURL) == "" {
		return "", "", errors.New("no image in streamed response")
	}
	return imageDataURL, strings.TrimSpace(assistantText), nil
}
