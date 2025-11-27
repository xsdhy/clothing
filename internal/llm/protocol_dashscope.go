package llm

import (
	"bytes"
	"clothing/internal/entity"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"clothing/internal/utils"

	"github.com/sirupsen/logrus"
)

const defaultDashscopeGenerationURL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation"
const dashscopeImageToVideoURL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/video-generation/video-synthesis"
const dashscopeKeyframeToVideoURL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/image2video/video-synthesis"
const dashscopeTaskQueryURL = "https://dashscope.aliyuncs.com/api/v1/tasks/"

type dashscopeContent struct {
	Image    string `json:"image,omitempty"`
	Text     string `json:"text,omitempty"`
	Video    string `json:"video,omitempty"`
	VideoURL string `json:"video_url,omitempty"`
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

type dashscopeVideoRequest struct {
	Model      string                 `json:"model"`
	Input      dashscopeVideoInput    `json:"input"`
	Parameters dashscopeVideoParamers `json:"parameters,omitempty"`
}

type dashscopeVideoInput struct {
	Prompt        string `json:"prompt"`
	ImgURL        string `json:"img_url,omitempty"`
	FirstFrameURL string `json:"first_frame_url,omitempty"`
	LastFrameURL  string `json:"last_frame_url,omitempty"`
}

type dashscopeVideoParamers struct {
	Resolution   string `json:"resolution,omitempty"`
	PromptExtend *bool  `json:"prompt_extend,omitempty"`
	Duration     int    `json:"duration,omitempty"`
}

type dashscopeVideoResponse struct {
	Output     dashscopeVideoOutput `json:"output"`
	RequestID  string               `json:"request_id"`
	Code       string               `json:"code"`
	Message    string               `json:"message"`
	StatusCode int                  `json:"status_code"`
}

type dashscopeVideoOutput struct {
	VideoURL   string                `json:"video_url"`
	VideoURLs  []string              `json:"video_urls"`
	TaskID     string                `json:"task_id"`
	TaskStatus string                `json:"task_status"`
	Message    string                `json:"message"`
	Results    []dashscopeVideoAsset `json:"results"`
}

type dashscopeVideoAsset struct {
	URL      string `json:"url"`
	VideoURL string `json:"video_url"`
}

type dashscopeVideoConfig struct {
	Resolutions       []string
	DefaultResolution string
	Durations         []int
	DefaultDuration   int
}

func (o dashscopeVideoOutput) collectAssets() []string {
	seen := make(map[string]struct{})
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
	}

	add(o.VideoURL)
	for _, v := range o.VideoURLs {
		add(v)
	}
	for _, asset := range o.Results {
		add(asset.URL)
		add(asset.VideoURL)
	}

	out := make([]string, 0, len(seen))
	for v := range seen {
		out = append(out, v)
	}
	return out
}

func GenerateImagesByDashscopeProtocol(ctx context.Context, apiKey, endpoint string, model entity.LlmModel, prompt, size string, duration int, base64Images, videos []string) (assets []string, assistantText string, err error) {
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

	if isDashscopeVideoModel(model) {
		cfg := videoConfigFromModel(model)
		return generateDashscopeVideo(ctx, apiKey, targetEndpoint, model, cfg, trimmedPrompt, size, duration, base64Images)
	}

	logrus.WithFields(logrus.Fields{
		"model":           model.ID,
		"prompt_length":   len(trimmedPrompt),
		"image_count":     len(base64Images),
		"video_count":     len(videos),
		"reference_media": len(base64Images) + len(videos),
	}).Info("dashscope_generate_images_start")

	messageContents := make([]dashscopeContent, 0, len(base64Images)+len(videos)+1)
	for idx, img := range base64Images {
		img = strings.TrimSpace(img)
		if img == "" {
			continue
		}
		// Dashscope expects publicly reachable URLs or base64. Convert local/temporary URLs to base64 to avoid URL errors.
		if strings.HasPrefix(img, "http://") || strings.HasPrefix(img, "https://") {
			origURL := img
			b64, mimeType, err := downloadImageAsBase64(ctx, origURL)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"index": idx,
					"url":   truncateForLog(origURL, 120),
					"error": err,
				}).Warn("dashscope_failed_download_image_reference")
				continue
			}
			img = fmt.Sprintf("data:%s;base64,%s", mimeType, b64)
			logrus.WithFields(logrus.Fields{
				"index":    idx,
				"mime":     mimeType,
				"orig_url": truncateForLog(origURL, 120),
			}).Info("dashscope_inlined_image_reference")
		}
		messageContents = append(messageContents, dashscopeContent{Image: img})
	}

	for _, video := range videos {
		video = strings.TrimSpace(video)
		if video == "" {
			continue
		}
		content := dashscopeContent{}
		if strings.HasPrefix(video, "http://") || strings.HasPrefix(video, "https://") {
			content.VideoURL = video
		} else {
			content.Video = video
		}
		messageContents = append(messageContents, content)
	}

	messageContents = append(messageContents, dashscopeContent{Text: trimmedPrompt})

	watermark := false
	reqBody := dashscopeRequest{
		Model: model.ID,
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
			if text := strings.TrimSpace(content.Text); text != "" {
				if assistantText != "" {
					assistantText += "\n"
				}
				assistantText += text
			}
			if raw := strings.TrimSpace(content.Image); raw != "" {
				if _, exists := seen[raw]; !exists {
					seen[raw] = struct{}{}
					assets = append(assets, raw)
				}
			}
			if raw := strings.TrimSpace(content.Video); raw != "" {
				if _, exists := seen[raw]; !exists {
					seen[raw] = struct{}{}
					assets = append(assets, raw)
				}
			}
			if raw := strings.TrimSpace(content.VideoURL); raw != "" {
				if _, exists := seen[raw]; !exists {
					seen[raw] = struct{}{}
					assets = append(assets, raw)
				}
			}
		}
	}

	if len(assets) == 0 {
		return nil, assistantText, errors.New("dashscope no asset found in response")
	}

	return assets, assistantText, nil
}

func generateDashscopeVideo(ctx context.Context, apiKey, endpoint string, model entity.LlmModel, cfg dashscopeVideoConfig, prompt, size string, duration int, images []string) (assets []string, assistantText string, err error) {
	if len(images) == 0 {
		return nil, "", errors.New("dashscope video model requires at least one reference image")
	}

	modelLower := strings.ToLower(model.ID)
	useKeyframe := strings.Contains(modelLower, "kf2v") || len(images) > 1
	target := resolveDashscopeVideoEndpoint(endpoint, useKeyframe)

	logrus.WithFields(logrus.Fields{
		"model":           model.ID,
		"prompt_length":   len(prompt),
		"image_count":     len(images),
		"use_keyframe":    useKeyframe,
		"endpoint":        truncateForLog(target, 128),
		"resolution_hint": strings.TrimSpace(size),
		"duration_hint":   duration,
	}).Info("dashscope_generate_video_start")

	firstFrame, err := inlineDashscopeImage(ctx, images[0])
	if err != nil || firstFrame == "" {
		return nil, "", fmt.Errorf("prepare first frame: %w", err)
	}

	var lastFrame string
	if useKeyframe {
		last := images[len(images)-1]
		lastFrame, err = inlineDashscopeImage(ctx, last)
		if err != nil || lastFrame == "" {
			return nil, "", fmt.Errorf("prepare last frame: %w", err)
		}
	}

	promptExtend := true
	resolution := normalizeDashscopeResolution(cfg, size)
	durationValue := normalizeDashscopeDuration(cfg, duration)

	input := dashscopeVideoInput{Prompt: prompt, ImgURL: firstFrame}
	if useKeyframe {
		input = dashscopeVideoInput{
			Prompt:        prompt,
			FirstFrameURL: firstFrame,
			LastFrameURL:  lastFrame,
		}
	}

	reqBody := dashscopeVideoRequest{
		Model: model.ID,
		Input: input,
		Parameters: dashscopeVideoParamers{
			Resolution:   resolution,
			PromptExtend: &promptExtend,
			Duration:     durationValue,
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", fmt.Errorf("dashscope marshal video request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(payload))
	if err != nil {
		return nil, "", fmt.Errorf("dashscope create video request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-DashScope-Async", "enable")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("dashscope video request: %w", err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, "", fmt.Errorf("dashscope read video response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   buf.String(),
		}).Error("dashscope generate video http error")
		return nil, "", fmt.Errorf("dashscope http %d: %s", resp.StatusCode, buf.String())
	}

	var result dashscopeVideoResponse
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return nil, "", fmt.Errorf("dashscope unmarshal video response: %w", err)
	}

	if result.Code != "" && strings.ToLower(result.Code) != "success" {
		msg := strings.TrimSpace(result.Message)
		if msg == "" {
			msg = "dashscope video error"
		}
		return nil, "", fmt.Errorf("dashscope api error: %s", msg)
	}

	output := result.Output
	if status := strings.ToUpper(strings.TrimSpace(output.TaskStatus)); status != "" && status != "SUCCEEDED" {
		assets, err = waitForDashscopeVideo(ctx, apiKey, output.TaskID)
		if err != nil {
			return nil, "", fmt.Errorf("dashscope video task status %s: %w", status, err)
		}
		return assets, assistantText, nil
	}

	assets = output.collectAssets()
	if len(assets) == 0 {
		if strings.TrimSpace(output.TaskID) != "" {
			assets, err = waitForDashscopeVideo(ctx, apiKey, output.TaskID)
			if err != nil {
				return nil, "", err
			}
			return assets, assistantText, nil
		}
		return nil, "", errors.New("dashscope video response missing video url")
	}

	return assets, assistantText, nil
}

func resolveDashscopeVideoEndpoint(endpoint string, useKeyframe bool) string {
	base := strings.TrimSpace(endpoint)
	if strings.Contains(strings.ToLower(base), "video-generation") ||
		strings.Contains(strings.ToLower(base), "image2video") {
		return base
	}
	if useKeyframe {
		return dashscopeKeyframeToVideoURL
	}
	return dashscopeImageToVideoURL
}

func waitForDashscopeVideo(ctx context.Context, apiKey, taskID string) ([]string, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, errors.New("dashscope missing task id for async video")
	}

	maxWait := 5 * time.Minute
	pollInterval := 3 * time.Second
	deadline := time.Now().Add(maxWait)

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		assets, status, err := fetchDashscopeTask(ctx, apiKey, taskID)
		if err != nil {
			return nil, err
		}
		if len(assets) > 0 {
			return assets, nil
		}
		state := strings.ToUpper(strings.TrimSpace(status))
		if state == "FAILED" || state == "CANCELLED" {
			return nil, fmt.Errorf("dashscope task %s", state)
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("dashscope task timeout (last status: %s)", state)
		}
		logrus.WithFields(logrus.Fields{
			"task_id": taskID,
			"status":  state,
		}).Info("dashscope video task still running")
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

func fetchDashscopeTask(ctx context.Context, apiKey, taskID string) ([]string, string, error) {
	target := dashscopeTaskQueryURL + taskID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, "", fmt.Errorf("dashscope task request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("dashscope task query: %w", err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, "", fmt.Errorf("dashscope read task response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   buf.String(),
		}).Error("dashscope task http error")
		return nil, "", fmt.Errorf("dashscope task http %d: %s", resp.StatusCode, buf.String())
	}

	var result dashscopeVideoResponse
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return nil, "", fmt.Errorf("dashscope unmarshal task response: %w", err)
	}

	output := result.Output
	assets := output.collectAssets()
	return assets, output.TaskStatus, nil
}

func inlineDashscopeImage(ctx context.Context, payload string) (string, error) {
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" {
		return "", errors.New("empty image payload")
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		b64, mimeType, err := downloadImageAsBase64(ctx, trimmed)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("data:%s;base64,%s", fallbackMime(mimeType), b64), nil
	}
	return utils.EnsureDataURL(trimmed), nil
}

func isDashscopeVideoModel(model entity.LlmModel) bool {
	for _, modality := range model.Inputs.Modalities {
		if strings.EqualFold(string(modality), "video") {
			return true
		}
	}

	m := strings.ToLower(strings.TrimSpace(model.ID))
	if m == "" {
		return false
	}
	if strings.Contains(m, "i2v") || strings.Contains(m, "image2video") || strings.Contains(m, "kf2v") || strings.Contains(m, "video") {
		return true
	}
	return false
}

func videoConfigFromModel(model entity.LlmModel) dashscopeVideoConfig {
	cfg := dashscopeVideoConfig{
		Resolutions:       normaliseResolutions(model.Inputs.SupportedSizes),
		DefaultResolution: strings.ToUpper(strings.TrimSpace(model.Inputs.DefaultSize)),
		Durations:         normaliseDurations(model.Inputs.SupportedDurations),
		DefaultDuration:   model.Inputs.DefaultDuration,
	}

	if cfg.DefaultResolution == "" && len(cfg.Resolutions) > 0 {
		cfg.DefaultResolution = cfg.Resolutions[0]
	}
	if len(cfg.Resolutions) == 0 && cfg.DefaultResolution != "" {
		cfg.Resolutions = append(cfg.Resolutions, cfg.DefaultResolution)
	}

	if cfg.DefaultDuration <= 0 && len(cfg.Durations) > 0 {
		cfg.DefaultDuration = cfg.Durations[0]
	}
	if len(cfg.Durations) == 0 && cfg.DefaultDuration > 0 {
		cfg.Durations = append(cfg.Durations, cfg.DefaultDuration)
	}

	return cfg
}

func normaliseResolutions(values []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.ToUpper(strings.TrimSpace(value))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func normaliseDurations(values []int) []int {
	seen := make(map[int]struct{})
	out := make([]int, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func normalizeDashscopeResolution(cfg dashscopeVideoConfig, requested string) string {
	req := strings.ToUpper(strings.TrimSpace(requested))
	for _, r := range cfg.Resolutions {
		if req != "" && strings.EqualFold(req, r) {
			return r
		}
	}
	if len(cfg.Resolutions) > 0 {
		for _, r := range cfg.Resolutions {
			if cfg.DefaultResolution != "" && strings.EqualFold(cfg.DefaultResolution, r) {
				return r
			}
		}
		return cfg.Resolutions[0]
	}
	if cfg.DefaultResolution != "" {
		return strings.ToUpper(cfg.DefaultResolution)
	}
	return "720P"
}

func normalizeDashscopeDuration(cfg dashscopeVideoConfig, requested int) int {
	if requested > 0 {
		for _, v := range cfg.Durations {
			if v == requested {
				return v
			}
		}
	}
	if cfg.DefaultDuration > 0 {
		return cfg.DefaultDuration
	}
	if len(cfg.Durations) > 0 {
		return cfg.Durations[0]
	}
	return 5
}
