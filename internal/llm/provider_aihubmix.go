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

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/sirupsen/logrus"
)

type AiHubMix struct {
	// 官方 openai-go 客户端（便于以后你用到其他 API）——本文件的图像流式走自定义 SSE，client 可留作他用
	client  openai.Client
	baseURL string
	apiKey  string

	models      []entity.LlmModel
	modelLookup map[string]struct{}
}

func NewAiHubMix(cfg config.Config) (*AiHubMix, error) {
	if strings.TrimSpace(cfg.AiHubMixAPIKey) == "" {
		return nil, errors.New("AiHubMix api key is not configured")
	}
	baseURL := "https://aihubmix.com/v1" // OpenAI 兼容网关
	client := openai.NewClient(
		option.WithAPIKey(cfg.AiHubMixAPIKey),
		option.WithBaseURL(baseURL),
	)

	models := []entity.LlmModel{
		{
			ID:          "qwen-image-edit",
			Name:        "qwen-image-edit",
			Price:       "<UNK>",
			Description: "",
		},
		{
			ID:          "imagen-4.0-fast-generate-001",
			Name:        "imagen-4.0-fast-generate-001",
			Price:       "0.03/IMG",
			Description: "",
		},
		{
			ID:          "gpt-4o-image",
			Name:        "gpt-4o-image",
			Price:       "0.005/IMG",
			Description: "",
		},
	}

	o := &AiHubMix{
		client:  client,
		baseURL: baseURL,
		apiKey:  cfg.AiHubMixAPIKey,
		models:  models,
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

func (o *AiHubMix) ProviderID() string {
	return "AiHubMix"
}

func (o *AiHubMix) Provider() entity.LlmProvider {
	return entity.LlmProvider{
		ID:     "AiHubMix",
		Name:   "AiHubMix",
		Models: o.Models(),
	}
}

func (o *AiHubMix) Models() []entity.LlmModel {
	return o.models
}

func (o *AiHubMix) SupportsModel(modelID string) bool {
	if o == nil || modelID == "" {
		return false
	}
	_, ok := o.modelLookup[modelID]
	return ok
}

/*
GenerateImages 通过 Responses API 的流式事件拿到图片与文本：
  - 文本：response.output_text.delta -> 累加到 strings.Builder
  - 图片：response.output_image.begin   -> 记录 mime、索引
    response.output_image.delta   -> 追加 base64 分片
    response.output_image.done    -> 标记该图完成
  - 结束：response.completed / EOF

返回值 (imageDataURL, text, error)：
- imageDataURL: 形如 "data:image/png;base64,...."
- text:         伴随生成的文字说明
*/
func (o *AiHubMix) GenerateImages(ctx context.Context, request entity.GenerateImageRequest) (string, string, error) {
	logger := providerLogger(ctx, o.ProviderID(), request.Model)
	logger.WithFields(logrus.Fields{
		"prompt_length":       len([]rune(request.Prompt)),
		"prompt_preview":      logSnippet(request.Prompt),
		"reference_image_cnt": len(request.Images),
	}).Info("llm_generate_images_start")

	// 构造 Responses API 请求体（启用内置的 image_generation 工具，非单独图片端点）
	inputContent := make([]map[string]any, 0, 1+len(request.Images))
	inputContent = append(inputContent, map[string]any{
		"type": "input_text",
		"text": request.Prompt,
	})
	for idx, img := range request.Images {
		img = strings.TrimSpace(img)
		if img == "" {
			logger.WithField("image_index", idx).Warn("llm_generate_images_skip_empty_reference")
			continue
		}
		inputContent = append(inputContent, map[string]any{
			"type":      "input_image",
			"image_url": utils.EnsureDataURL(img),
		})
	}

	reqBody := map[string]any{
		"model": request.Model,
		"input": []map[string]any{
			{
				"role":    "user",
				"content": inputContent,
			},
		},
		"tools":  []map[string]any{{"type": "image_generation"}},
		"stream": true, // 关键：启用流式
	}

	imageDataURL, text, err := o.streamResponsesImages(ctx, reqBody, logger)
	if err != nil {
		logger.WithError(err).Error("llm_generate_images_stream_failed")
		return "", "", err
	}

	logger.WithFields(logrus.Fields{
		"has_image": imageDataURL != "",
		"text_len":  len(text),
	}).Info("llm_generate_images_success")
	return imageDataURL, text, nil
}

func (o *AiHubMix) streamResponsesImages(ctx context.Context, body map[string]any, logger *logrus.Entry) (string, string, error) {
	reqBytes, _ := json.Marshal(body)
	url := strings.TrimRight(o.baseURL, "/") + "/responses"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBytes))
	if err != nil {
		return "", "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	// 不设超短超时，留足生成时间；你可根据业务把整体超时交给上层 ctx 控制
	httpClient := &http.Client{
		Timeout: 0,
	}

	logger.WithFields(logrus.Fields{
		"url":    url,
		"method": "POST",
	}).Info("responses_stream_begin")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", "", fmt.Errorf("responses stream bad status %d: %s", resp.StatusCode, string(b))
	}

	sc := bufio.NewScanner(resp.Body)
	// base64 片段可能很大，调大 buffer 和上限
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 10*1024*1024)

	var (
		curEvent    string
		dataBuf     []string
		textBuilder strings.Builder
		imgChunks   = make(map[int]*bytes.Buffer) // index -> b64 buffer
		imgMIMEs    = make(map[int]string)        // index -> mime
		imgURLs     = make(map[int]string)        // index -> url
		completed   bool
		chunkIndex  int
	)

	flushEvent := func() error {
		if len(dataBuf) == 0 {
			curEvent = ""
			return nil
		}
		payload := strings.Join(dataBuf, "\n")
		dataBuf = dataBuf[:0]

		payloadBytes := []byte(payload)
		chunkIndex++
		eventName := responseEventType(curEvent, payloadBytes)
		logger.WithFields(logrus.Fields{
			"event":       eventName,
			"chunk_index": chunkIndex,
			"len":         len(payloadBytes),
		}).Debug("responses_stream_event")

		handleTextFragments := func(paths ...string) {
			fragments := responseEventTextFragments(payloadBytes, paths...)
			for _, fragment := range fragments {
				if fragment == "" {
					continue
				}
				textBuilder.WriteString(fragment)
			}
		}

		handleImageDelta := func(defaultIndex int) error {
			index, ok := responseEventIndex(payloadBytes)
			if !ok {
				index = defaultIndex
			}
			if _, exists := imgChunks[index]; !exists {
				imgChunks[index] = &bytes.Buffer{}
			}
			if mime := responseEventMIME(payloadBytes); mime != "" {
				imgMIMEs[index] = mime
			} else if _, exists := imgMIMEs[index]; !exists {
				imgMIMEs[index] = "image/png"
			}

			candidates := extractResponseImageCandidates(payloadBytes, "delta", "content", "image", "choices")
			if len(candidates) == 0 {
				return nil
			}
			for _, candidate := range candidates {
				idx := index
				if candidate.Index != 0 {
					idx = candidate.Index
				}
				if candidate.MIME != "" {
					imgMIMEs[idx] = candidate.MIME
				}
				if candidate.DataURL != "" {
					return fmt.Errorf("__RETURN_IMMEDIATELY__|%s|%s", candidate.DataURL, textBuilder.String())
				}
				if candidate.URL != "" {
					imgURLs[idx] = candidate.URL
					continue
				}
				if candidate.Base64 != "" {
					imgChunks[idx].WriteString(candidate.Base64)
				}
			}
			return nil
		}

		if eventName == "" {
			handleTextFragments("choices.#.delta.content", "delta.content", "delta")
			return handleImageDelta(0)
		}

		switch eventName {
		case "response.output_text.delta", "response.reasoning_text.delta", "response.output_text.annotation.added":
			handleTextFragments("delta", "delta.content")
		case "response.output_image.begin", "response.image_generation_call.partial_image":
			index, ok := responseEventIndex(payloadBytes)
			if !ok {
				index = 0
			}
			if mime := responseEventMIME(payloadBytes); mime != "" {
				imgMIMEs[index] = mime
			} else if _, exists := imgMIMEs[index]; !exists {
				imgMIMEs[index] = "image/png"
			}
			if _, exists := imgChunks[index]; !exists {
				imgChunks[index] = &bytes.Buffer{}
			}
		case "response.output_image.delta", "response.image_generation_call.generating":
			if err := handleImageDelta(0); err != nil {
				return err
			}
		case "response.output_image.done", "response.image_generation_call.completed":
			// no-op
		case "response.completed":
			completed = true
			if textBuilder.Len() == 0 {
				handleTextFragments("response.output", "response.text", "response.output_items")
			}
			if len(imgChunks) == 0 && len(imgURLs) == 0 {
				candidates := extractResponseImageCandidates(payloadBytes, "response.output", "response.output_items")
				if len(candidates) > 0 {
					candidate := candidates[0]
					if url := candidate.dataURL(); url != "" {
						return fmt.Errorf("__RETURN_IMMEDIATELY__|%s|%s", url, textBuilder.String())
					}
					if candidate.URL != "" {
						return fmt.Errorf("__RETURN_IMMEDIATELY__|%s|%s", candidate.URL, textBuilder.String())
					}
				}
			}
		case "response.failed", "response.incomplete", "response.error", "error", "response.image_generation_call.failed":
			msg := responseErrorMessage(payloadBytes)
			if msg == "" {
				msg = payload
			}
			return errors.New(msg)
		default:
			handleTextFragments("delta.content", "delta", "content", "response.output")
			if err := handleImageDelta(0); err != nil {
				return err
			}
		}
		curEvent = ""
		return nil
	}

	for sc.Scan() {
		line := sc.Text()
		line = strings.TrimRight(line, "\r")
		if line == "" {
			// 一个事件块结束
			if err := flushEvent(); err != nil {
				// 特殊短路：用错误来带出“立即返回”的信号
				if strings.HasPrefix(err.Error(), "__RETURN_IMMEDIATELY__|") {
					parts := strings.SplitN(err.Error(), "|", 3)
					if len(parts) == 3 {
						return parts[1], parts[2], nil
					}
				}
				return "", "", err
			}
			continue
		}
		if strings.HasPrefix(line, "event:") {
			curEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			// 某些实现可能发 data: [DONE]
			if data == "[DONE]" {
				completed = true
				break
			}
			dataBuf = append(dataBuf, data)
			continue
		}
		// 其它前缀忽略（id:, retry:）
	}
	// 处理最后残留事件
	if err := flushEvent(); err != nil {
		if strings.HasPrefix(err.Error(), "__RETURN_IMMEDIATELY__|") {
			parts := strings.SplitN(err.Error(), "|", 3)
			if len(parts) == 3 {
				return parts[1], parts[2], nil
			}
		}
		return "", "", err
	}

	if err := sc.Err(); err != nil && !errors.Is(err, io.EOF) {
		return "", "", err
	}
	if !completed {
		logger.Warn("responses_stream_not_marked_completed")
	}

	// 选择第一张图输出（如需多图可扩展返回切片）
	if len(imgChunks) == 0 {
		if len(imgURLs) > 0 {
			minIdx := int(^uint(0) >> 1)
			for idx := range imgURLs {
				if idx < minIdx {
					minIdx = idx
				}
			}
			if url := strings.TrimSpace(imgURLs[minIdx]); url != "" {
				return url, textBuilder.String(), nil
			}
		}
		// 有些实现只在文本里内嵌 dataURL，这里兜底尝试
		text := textBuilder.String()
		if s := strings.TrimSpace(text); strings.Contains(s, "data:image") {
			// 粗略提取
			start := strings.Index(s, "data:image")
			end := strings.IndexAny(s[start:], " \"\n\t")
			if end == -1 {
				end = len(s)
			} else {
				end += start
			}
			return s[start:end], text, nil
		}
		return "", text, errors.New("model did not return an image in stream")
	}

	// 最小 index 作为第一张
	minIdx := int(^uint(0) >> 1)
	for i := range imgChunks {
		if i < minIdx {
			minIdx = i
		}
	}
	b64 := strings.TrimSpace(imgChunks[minIdx].String())
	if b64 == "" {
		if url := strings.TrimSpace(imgURLs[minIdx]); url != "" {
			return url, textBuilder.String(), nil
		}
		return "", textBuilder.String(), errors.New("model did not return an image in stream")
	}
	mime := imgMIMEs[minIdx]
	if mime == "" {
		mime = "image/png"
	}
	dataURL := "data:" + mime + ";base64," + b64

	return dataURL, textBuilder.String(), nil
}
