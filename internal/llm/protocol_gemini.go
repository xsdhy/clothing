package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"clothing/internal/utils"

	"github.com/sirupsen/logrus"
)

// Gemini uses a Google-style streaming endpoint instead of the OpenAI-compatible one.
// We keep the request/response contracts local so provider_aihubmix (and future Gemini
// providers) can call a single helper without duplicating glue code.
const geminiStreamEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse"

// Request payload pieces ----------------------------------------------------
type (
	geminiInlineData struct {
		MimeType string `json:"mimeType,omitempty"`
		Data     string `json:"data,omitempty"`
	}
	geminiFileData struct {
		FileURI  string `json:"fileUri,omitempty"`
		MimeType string `json:"mimeType,omitempty"`
	}
	geminiPart struct {
		Text       string            `json:"text,omitempty"`
		InlineData *geminiInlineData `json:"inlineData,omitempty"`
		FileData   *geminiFileData   `json:"fileData,omitempty"`
	}
	geminiContent struct {
		Role  string       `json:"role,omitempty"`
		Parts []geminiPart `json:"parts"`
	}
	geminiRequest struct {
		Contents []geminiContent `json:"contents"`
	}
)

// Response payload pieces ---------------------------------------------------
type (
	geminiCandidate struct {
		FinishReason string        `json:"finishReason,omitempty"`
		Content      geminiContent `json:"content"`
	}
	geminiError struct {
		Message string `json:"message"`
	}
	geminiStreamChunk struct {
		Candidates []geminiCandidate `json:"candidates"`
		Error      *geminiError      `json:"error,omitempty"`
	}
)

// GenerateContentByGeminiProtocol streams Gemini image generations via SSE.
// It behaves similarly to GenerateContentByOpenaiProtocol but understands Gemini
// payloads (candidates/parts with inlineData). More verbose logs are emitted to
// help diagnose integration issues, as Gemini responses can be picky about the
// shape of image payloads.
func GenerateContentByGeminiProtocol(ctx context.Context, apiKey, endpoint, model, prompt string, refs []string) (imageDataURLs []string, assistantText string, err error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, "", errors.New("api key missing")
	}
	if strings.TrimSpace(model) == "" {
		return nil, "", errors.New("model is required")
	}
	if strings.TrimSpace(prompt) == "" {
		return nil, "", errors.New("prompt is empty")
	}

	logrus.WithFields(logrus.Fields{
		"model":                 model,
		"endpoint":              truncateForLog(endpoint, 128),
		"prompt_preview":        truncateForLog(prompt, 64),
		"prompt_length":         len(prompt),
		"reference_image_count": len(refs),
	}).Info("gemini_generate_content_start")

	parts, errs := buildGeminiParts(ctx, prompt, refs)
	if len(parts) == 0 {
		return nil, "", errors.New("no valid prompt or image parts for gemini request")
	}
	if len(errs) > 0 {
		logrus.WithFields(logrus.Fields{
			"error_count": len(errs),
			"errors":      strings.Join(errs, "; "),
		}).Warn("gemini some references could not be parsed")
	}

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: parts,
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", fmt.Errorf("gemini marshal request: %w", err)
	}

	targetURL := resolveGeminiEndpoint(endpoint, model)
	logrus.WithField("target_url", truncateForLog(targetURL, 200)).Info("gemini resolved endpoint")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, "", fmt.Errorf("gemini create request: %w", err)
	}
	// Prefer header to avoid logging API key inside URLs; most gateways accept this.
	req.Header.Set("x-goog-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	httpCli := &http.Client{Timeout: 0} // disable client-level timeout for long-running streams
	resp, err := httpCli.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("gemini send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   buf.String(),
		}).Error("gemini generate images http error")
		return nil, "", fmt.Errorf("gemini http %d: %s", resp.StatusCode, buf.String())
	}

	// Stream parsing: Gemini with ?alt=sse emits `data: { ... }` lines similar to OpenAI.
	logrus.Info("gemini stream response started")

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	seenImages := make(map[string]struct{})
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		logrus.WithField("data", line).Info("gemini stream chunk received")

		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			logrus.Info("gemini stream reached end marker")
			break
		}

		var chunk geminiStreamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			logrus.WithError(err).Warn("gemini failed to unmarshal stream chunk")
			continue
		}
		if chunk.Error != nil && strings.TrimSpace(chunk.Error.Message) != "" {
			logrus.WithField("message", chunk.Error.Message).Error("gemini stream error chunk")
			assistantText = appendLine(assistantText, chunk.Error.Message)
			continue
		}
		if len(chunk.Candidates) == 0 {
			continue
		}

		for _, cand := range chunk.Candidates {
			if cand.FinishReason != "" {
				logrus.WithField("finish_reason", cand.FinishReason).Info("gemini finish signal")
			}
			for _, part := range cand.Content.Parts {
				if part.Text != "" {
					assistantText = appendLine(assistantText, part.Text)
				}
				// InlineData returns base64 image payload; wrap it into a data URL so downstream
				// consumers can persist or preview it without guessing MIME type.
				if part.InlineData != nil {
					dataURL := fmt.Sprintf("data:%s;base64,%s", fallbackMime(part.InlineData.MimeType), strings.TrimSpace(part.InlineData.Data))
					if dataURL != "" {
						if _, ok := seenImages[dataURL]; !ok {
							seenImages[dataURL] = struct{}{}
							imageDataURLs = append(imageDataURLs, dataURL)
							logrus.WithFields(logrus.Fields{
								"mime":        part.InlineData.MimeType,
								"image_len":   len(part.InlineData.Data),
								"image_count": len(imageDataURLs),
							}).Info("gemini collected inline image")
						}
					}
				}
				// Some gateways may stream out accessible file URIs instead of inline data.
				if part.FileData != nil {
					url := strings.TrimSpace(part.FileData.FileURI)
					if url == "" {
						continue
					}
					if _, ok := seenImages[url]; !ok {
						seenImages[url] = struct{}{}
						imageDataURLs = append(imageDataURLs, url)
						logrus.WithFields(logrus.Fields{
							"file_uri":    url,
							"mime":        part.FileData.MimeType,
							"image_count": len(imageDataURLs),
						}).Info("gemini collected file uri image")
					}
				}
			}
		}
	}

	logrus.Info("gemini stream response ended")

	if err := scanner.Err(); err != nil {
		return nil, assistantText, fmt.Errorf("gemini stream read error: %w", err)
	}
	if len(imageDataURLs) == 0 {
		if strings.TrimSpace(assistantText) == "" {
			return nil, "", errors.New("gemini response did not include image data")
		}
		return nil, assistantText, errors.New("gemini response did not include image data")
	}

	return imageDataURLs, strings.TrimSpace(assistantText), nil
}

// buildGeminiParts converts prompt and optional reference images into the Gemini
// Content/Part structure. Errors are collected (for logging) but skipped so one
// bad reference does not block the entire request.
func buildGeminiParts(ctx context.Context, prompt string, refs []string) ([]geminiPart, []string) {
	parts := []geminiPart{
		{Text: strings.TrimSpace(prompt)},
	}

	var errs []string
	for idx, ref := range refs {
		trimmed := strings.TrimSpace(ref)
		if trimmed == "" {
			continue
		}

		part, err := buildGeminiImagePart(ctx, trimmed)
		if err != nil {
			errs = append(errs, fmt.Sprintf("ref %d: %v", idx, err))
			logrus.WithFields(logrus.Fields{
				"index": idx,
				"error": err,
			}).Warn("gemini skip invalid reference image")
			continue
		}
		parts = append(parts, part)
	}

	return parts, errs
}

// buildGeminiImagePart tries to normalize a user-supplied image payload into the
// inlineData shape Gemini expects. We accept:
// - http(s) URLs: fetched and re-encoded as base64 for inlineData
// - data URLs or bare base64 strings: parsed and re-used
func buildGeminiImagePart(ctx context.Context, payload string) (geminiPart, error) {
	// Remote URL: download then encode.
	if strings.HasPrefix(payload, "http://") || strings.HasPrefix(payload, "https://") {
		b64, mimeType, err := downloadImageAsBase64(ctx, payload)
		if err != nil {
			return geminiPart{}, fmt.Errorf("download reference: %w", err)
		}
		return geminiPart{
			InlineData: &geminiInlineData{
				MimeType: mimeType,
				Data:     b64,
			},
		}, nil
	}

	// Inline payload: parse data URL or plain base64.
	mimeType, base64Payload := utils.SplitDataURL(utils.EnsureDataURL(payload))
	base64Payload = strings.TrimSpace(base64Payload)
	if base64Payload == "" {
		return geminiPart{}, errors.New("empty base64 payload")
	}

	raw, err := base64.StdEncoding.DecodeString(base64Payload)
	if err != nil {
		return geminiPart{}, fmt.Errorf("decode base64: %w", err)
	}

	// Re-encode to remove whitespace and keep a clean payload for Gemini.
	normalized := base64.StdEncoding.EncodeToString(raw)

	return geminiPart{
		InlineData: &geminiInlineData{
			MimeType: fallbackMime(mimeType),
			Data:     normalized,
		},
	}, nil
}

// downloadImageAsBase64 pulls the remote asset and encodes it for inlineData usage.
func downloadImageAsBase64(ctx context.Context, url string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", fmt.Errorf("create download request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("download image http %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("read image body: %w", err)
	}

	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}
	mimeType = fallbackMime(mimeType)

	logrus.WithFields(logrus.Fields{
		"mime":       mimeType,
		"size_bytes": len(data),
		"url":        truncateForLog(url, 128),
	}).Info("gemini downloaded reference image")

	return base64.StdEncoding.EncodeToString(data), mimeType, nil
}

// appendLine concatenates messages with newlines, avoiding empty prefixes.
func appendLine(current, next string) string {
	next = strings.TrimSpace(next)
	if next == "" {
		return current
	}
	if strings.TrimSpace(current) == "" {
		return next
	}
	return current + "\n" + next
}

// fallbackMime normalizes empty/unknown mime types to a sensible default.
func fallbackMime(mimeType string) string {
	v := strings.TrimSpace(mimeType)
	if v == "" {
		return "image/jpeg"
	}
	// Strip charset if present.
	if idx := strings.Index(v, ";"); idx > 0 {
		return strings.TrimSpace(v[:idx])
	}
	return v
}

// truncateForLog keeps logs short while still surfacing useful context.
func truncateForLog(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	return value[:max] + "..."
}

// resolveGeminiEndpoint builds the request URL from a provided endpoint template or base URL.
// - If endpoint contains "%s", it is treated as a fmt template and will be formatted with model.
// - If endpoint is a bare base URL, the default Gemini suffix is appended.
// - If empty, fall back to the public Gemini endpoint.
func resolveGeminiEndpoint(endpoint, model string) string {
	base := strings.TrimSpace(endpoint)
	if base == "" {
		return fmt.Sprintf(geminiStreamEndpoint, model)
	}

	// Template-style endpoint (allows full custom control).
	if strings.Contains(base, "%s") {
		return fmt.Sprintf(base, model)
	}

	// Base URL style (e.g. https://aihubmix.com/gemini).
	base = strings.TrimRight(base, "/")
	return fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?alt=sse", base, model)
}
