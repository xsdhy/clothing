package llm

import (
	"encoding/base64"
	"encoding/json"
	"mime"
	"strings"

	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"errors"
	"fmt"
	"net/http"

	"clothing/internal/utils"

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
	Delta              orDelta `json:"delta"`
	FinishReason       string  `json:"finish_reason"`
	NativeFinishReason string  `json:"native_finish_reason"`
	Index              int     `json:"index"`
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

func GenerateImagesByOpenaiProtocol(ctx context.Context, apiKey, baseURL, model, prompt string, refs []string) (imageDataURLs []string, assistantText string, err error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, "", errors.New("api key missing")
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

	httpCli := &http.Client{Timeout: 0}
	// SSE 不要超短超时
	resp, err := httpCli.Do(req)
	if err != nil {
		return nil, "", err
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
		return nil, "", fmt.Errorf("openrouter http %d: %s", resp.StatusCode, buf.String())
	}
	// 处理 SSE 流式响应
	logrus.Info("openrouter stream response started")

	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	nativeFinishReasonText := ""
	seenImages := make(map[string]struct{})
	for sc.Scan() {
		line := sc.Text()
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
			if delta.Content != "" {
				assistantText += delta.Content
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
					saveImageAsync(url)
				}
			}
		}

	}
	logrus.Info("openrouter stream response ended")
	if err := sc.Err(); err != nil {
		return nil, "", err
	}
	if len(imageDataURLs) == 0 {
		if len(nativeFinishReasonText) > 0 {
			return nil, "", errors.New(nativeFinishReasonText)
		}
		return nil, "", errors.New("no image in streamed response")
	}
	return imageDataURLs, strings.TrimSpace(assistantText), nil
}

const openaiImageSaveDir = "clothing-openai-images"

func saveImageAsync(imageURL string) {
	url := strings.TrimSpace(imageURL)
	if url == "" {
		return
	}

	go func(u string) {
		if err := persistImageLocally(u); err != nil {
			fields := logrus.Fields{"source": imageSourceForLog(u)}
			if !strings.HasPrefix(u, "data:") {
				fields["image_url"] = u
			}
			logrus.WithError(err).WithFields(fields).Warn("openai save streamed image failed")
		}
	}(url)
}

func persistImageLocally(imageURL string) error {
	var (
		data   []byte
		ext    string
		fields = logrus.Fields{"source": imageSourceForLog(imageURL)}
	)

	if strings.HasPrefix(imageURL, "data:") {
		mimeType, payload := utils.SplitDataURL(imageURL)
		if payload == "" {
			return errors.New("empty data url payload")
		}
		decoded, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return fmt.Errorf("decode data url: %w", err)
		}
		data = decoded
		ext = extensionFromMime(mimeType)
		if mimeType != "" {
			fields["mime"] = mimeType
		}
	} else {
		fields["image_url"] = imageURL
		reqCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, imageURL, nil)
		if err != nil {
			return fmt.Errorf("create image request: %w", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("download image: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("download image http %d", resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read image body: %w", err)
		}
		data = body
		ext = extensionFromMime(resp.Header.Get("Content-Type"))
		if ct := resp.Header.Get("Content-Type"); ct != "" {
			fields["mime"] = ct
		}
	}

	if len(data) == 0 {
		return errors.New("image payload empty")
	}

	if ext == "" {
		ext = extensionFromMime(http.DetectContentType(data))
	}
	if ext == "" {
		ext = "jpg"
	}

	dir := filepath.Join("./images/", openaiImageSaveDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create image dir: %w", err)
	}
	filename := fmt.Sprintf("openai_%d.%s", time.Now().UnixNano(), ext)
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write image file: %w", err)
	}

	fields["path"] = path
	fields["size"] = len(data)
	logrus.WithFields(fields).Info("openai streamed image saved")
	return nil
}

func extensionFromMime(mimeType string) string {
	if mimeType == "" {
		return ""
	}
	if parsed, _, err := mime.ParseMediaType(mimeType); err == nil {
		mimeType = parsed
	}

	switch strings.ToLower(mimeType) {
	case "image/png":
		return "png"
	case "image/jpeg", "image/jpg":
		return "jpg"
	case "image/webp":
		return "webp"
	case "image/gif":
		return "gif"
	case "image/bmp":
		return "bmp"
	case "image/svg+xml":
		return "svg"
	case "image/heic":
		return "heic"
	case "image/heif":
		return "heif"
	default:
		return ""
	}
}

func imageSourceForLog(imageURL string) string {
	if strings.HasPrefix(imageURL, "data:") {
		return "inline_data_url"
	}
	return "remote_image"
}
