package llm

import (
	"bytes"
	"clothing/internal/entity"
	"clothing/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	falDefaultAPIBaseURL         = "https://fal.run"
	falModeTextToImage   falMode = "text_to_image"
	falModeImageToImage  falMode = "image_to_image"

	falDefaultImageSize = "1024x1024"
	falPollInterval     = 2 * time.Second
	falMaxPollAttempts  = 60
)

type falMode string

type falModelConfig struct {
	endpoint string
	mode     falMode
}

type FalAI struct {
	providerID   string
	providerName string

	apiKey  string
	apiBase string

	httpClient *http.Client
}

func NewFalAI(provider *entity.DbProvider) (*FalAI, error) {
	if provider == nil {
		return nil, errors.New("fal.ai provider config is nil")
	}

	apiKey := strings.TrimSpace(provider.APIKey)
	if apiKey == "" {
		return nil, errors.New("fal.ai api key is not configured")
	}

	name := strings.TrimSpace(provider.Name)
	if name == "" {
		name = provider.ID
	}

	baseURL := strings.TrimSpace(provider.BaseURL)
	if baseURL == "" {
		baseURL = falDefaultAPIBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &FalAI{
		providerID:   provider.ID,
		providerName: name,
		apiKey:       apiKey,
		apiBase:      baseURL,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (f *FalAI) GenerateContent(ctx context.Context, request entity.GenerateContentRequest, dbModel entity.DbModel) (*entity.GenerateContentResponse, error) {
	if f == nil {
		return nil, errors.New("fal.ai provider not initialised")
	}

	// Use new EndpointPath field first, fallback to deriving from ModelID
	endpoint := strings.TrimSpace(dbModel.EndpointPath)
	if endpoint == "" {
		endpoint = "/" + strings.TrimLeft(dbModel.ModelID, "/")
	}
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}

	// Use new GenerationMode field first, fallback to inferring from ModelID
	mode := falMode(strings.TrimSpace(dbModel.GenerationMode))
	if mode == "" {
		if strings.Contains(strings.ToLower(dbModel.ModelID), "image-to-image") ||
			strings.Contains(strings.ToLower(dbModel.ModelID), "edit") {
			mode = falModeImageToImage
		} else {
			mode = falModeTextToImage
		}
	}

	logrus.WithFields(logrus.Fields{
		"model":               request.ModelID,
		"prompt_preview":      request.Prompt,
		"reference_image_cnt": len(request.GetImages()),
		"size":                strings.TrimSpace(request.GetSize()),
	}).Info("falai_generate_content_start")

	input, err := f.buildInputPayload(mode, request)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{"input": input}
	envelope, err := f.submitAndWait(ctx, endpoint, payload)
	if err != nil {
		return nil, err
	}

	requestID := envelope.RequestID
	if requestID == "" {
		requestID = "fal-" + utils.GenerateUUID()
	}

	if envelope.Status != "COMPLETED" && envelope.Status != "OK" {
		return &entity.GenerateContentResponse{RequestID: requestID}, fmt.Errorf("falai task failed status=%s error=%s", envelope.Status, envelope.Error)
	}

	var assets []string
	if envelope.Images != nil && len(envelope.Images) > 0 {
		for _, img := range envelope.Images {
			assets = append(assets, img.URL)
		}
	}

	return &entity.GenerateContentResponse{
		Outputs:   buildMediaOutputs(assets, "image"),
		RequestID: requestID,
	}, nil
}

func (f *FalAI) buildInputPayload(mode falMode, request entity.GenerateContentRequest) (map[string]any, error) {
	prompt := strings.TrimSpace(request.Prompt)
	if prompt == "" {
		return nil, errors.New("prompt is required")
	}

	input := map[string]any{"prompt": prompt}

	size := strings.TrimSpace(request.GetSize())
	if size == "" {
		size = falDefaultImageSize
	}

	switch mode {
	case falModeTextToImage:
		input["image_size"] = size
		input["num_images"] = 1
	case falModeImageToImage:
		imageURL, base64Payload := f.pickReferenceImage(request.GetImages())
		if imageURL == "" && base64Payload == "" {
			return nil, errors.New("image-to-image model requires at least one reference image")
		}
		if base64Payload != "" {
			input["image_base64"] = base64Payload
		} else {
			input["image_url"] = imageURL
		}
		input["image_size"] = size
		input["num_images"] = 1
	default:
		return nil, fmt.Errorf("unsupported fal.ai mode %q", mode)
	}

	return input, nil
}

func (f *FalAI) pickReferenceImage(images []string) (string, string) {
	for _, img := range images {
		trimmed := strings.TrimSpace(img)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
			return trimmed, ""
		}
		if strings.HasPrefix(trimmed, "data:") {
			_, payload := utils.SplitDataURL(trimmed)
			if payload != "" {
				return "", payload
			}
			continue
		}
		return "", trimmed
	}
	return "", ""
}

func (f *FalAI) submitAndWait(ctx context.Context, endpoint string, payload map[string]any) (*falGenerationEnvelope, error) {
	bs, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("fal.ai marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.apiBase+endpoint, bytes.NewReader(bs))
	if err != nil {
		return nil, fmt.Errorf("fal.ai create request: %w", err)
	}
	req.Header.Set("Authorization", "Key "+f.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fal.ai submit request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fal.ai read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("fal.ai http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var submission falSubmissionResponse
	if err := json.Unmarshal(body, &submission); err != nil {
		return nil, fmt.Errorf("fal.ai decode submission: %w", err)
	}

	envelope := submission.toEnvelope()
	if envelope.Error != nil {
		return nil, fmt.Errorf("fal.ai error: %s", envelope.Error.Message)
	}

	if strings.EqualFold(envelope.Status, "COMPLETED") && len(f.collectImagePayloads(envelope)) > 0 {
		return envelope, nil
	}

	responseURL := strings.TrimSpace(submission.ResponseURL)
	if responseURL == "" {
		responseURL = strings.TrimSpace(submission.StatusURL)
	}
	if responseURL == "" {
		return nil, errors.New("fal.ai response url missing")
	}

	return f.pollForCompletion(ctx, responseURL, submission.RequestID)
}

func (f *FalAI) pollForCompletion(ctx context.Context, responseURL, requestID string) (*falGenerationEnvelope, error) {
	attempts := 0
	ticker := time.NewTicker(falPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("fal.ai poll cancelled: %w", ctx.Err())
		case <-ticker.C:
			attempts++
			envelope, done, err := f.fetchResponse(ctx, responseURL)
			if err != nil {
				return nil, err
			}
			if !done {
				logrus.WithFields(logrus.Fields{
					"request_id": requestID,
					"status":     envelope.Status,
					"attempt":    attempts,
				}).Info("falai_poll_pending")
				if attempts >= falMaxPollAttempts {
					return nil, errors.New("fal.ai polling exceeded maximum attempts")
				}
				continue
			}
			if envelope.Error != nil {
				return nil, fmt.Errorf("fal.ai error: %s", envelope.Error.Message)
			}
			if len(f.collectImagePayloads(envelope)) == 0 {
				return envelope, errors.New("fal.ai completed without images")
			}
			return envelope, nil
		}
	}
}

func (f *FalAI) fetchResponse(ctx context.Context, url string) (*falGenerationEnvelope, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("fal.ai create poll request: %w", err)
	}
	req.Header.Set("Authorization", "Key "+f.apiKey)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("fal.ai poll request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("fal.ai poll read: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, false, fmt.Errorf("fal.ai poll http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var envelope falGenerationEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, false, fmt.Errorf("fal.ai poll decode: %w", err)
	}

	envelope.mergeInner()

	status := strings.ToUpper(strings.TrimSpace(envelope.Status))
	switch status {
	case "COMPLETED":
		return &envelope, true, nil
	case "FAILED", "CANCELLED", "ERROR":
		if envelope.Error != nil {
			return &envelope, true, fmt.Errorf("fal.ai error: %s", envelope.Error.Message)
		}
		return &envelope, true, fmt.Errorf("fal.ai job %s", strings.ToLower(status))
	default:
		return &envelope, false, nil
	}
}

func (f *FalAI) extractImagesAndText(envelope *falGenerationEnvelope) ([]string, string) {
	if envelope == nil {
		return nil, ""
	}

	payloads := f.collectImagePayloads(envelope)
	images := make([]string, 0, len(payloads))
	seen := make(map[string]struct{})

	for _, img := range payloads {
		url := strings.TrimSpace(img.firstURL())
		base64Payload := strings.TrimSpace(img.firstBase64())

		switch {
		case url != "":
			if _, exists := seen[url]; exists {
				continue
			}
			seen[url] = struct{}{}
			images = append(images, url)
		case base64Payload != "":
			dataURL := base64Payload
			if !strings.HasPrefix(base64Payload, "data:") {
				dataURL = utils.EnsureDataURL(base64Payload)
			}
			if _, exists := seen[dataURL]; exists {
				continue
			}
			seen[dataURL] = struct{}{}
			images = append(images, dataURL)
		}
	}

	text := strings.TrimSpace(envelope.Text)
	if text == "" {
		text = strings.TrimSpace(envelope.Message)
	}
	if text == "" {
		text = strings.TrimSpace(envelope.OutputText)
	}
	if text == "" && envelope.Response != nil {
		if t := strings.TrimSpace(envelope.Response.Text); t != "" {
			text = t
		} else if m := strings.TrimSpace(envelope.Response.Message); m != "" {
			text = m
		} else if ot := strings.TrimSpace(envelope.Response.OutputText); ot != "" {
			text = ot
		}
	}

	return images, text
}

func (f *FalAI) collectImagePayloads(envelope *falGenerationEnvelope) []falImagePayload {
	if envelope == nil {
		return nil
	}

	payloads := make([]falImagePayload, 0, len(envelope.Images)+len(envelope.Output)+len(envelope.Outputs)+len(envelope.Data)+len(envelope.Result)+len(envelope.Variants))
	payloads = append(payloads, envelope.Images...)
	payloads = append(payloads, envelope.Output...)
	payloads = append(payloads, envelope.Outputs...)
	payloads = append(payloads, envelope.Data...)
	payloads = append(payloads, envelope.Result...)
	payloads = append(payloads, envelope.Variants...)

	if envelope.Response != nil {
		payloads = append(payloads, envelope.Response.Images...)
		payloads = append(payloads, envelope.Response.Output...)
		payloads = append(payloads, envelope.Response.Outputs...)
		payloads = append(payloads, envelope.Response.Data...)
		payloads = append(payloads, envelope.Response.Result...)
		payloads = append(payloads, envelope.Response.Variants...)
	}

	return payloads
}

type falImagePayload struct {
	URL         string `json:"url"`
	ImageURL    string `json:"image_url"`
	ContentType string `json:"content_type"`
	Base64      string `json:"base64"`
	Data        string `json:"data"`
	B64JSON     string `json:"b64_json"`
}

func (p *falImagePayload) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if data[0] == '"' {
		var url string
		if err := json.Unmarshal(data, &url); err != nil {
			return err
		}
		p.URL = url
		return nil
	}

	type rawMap map[string]any
	var payload rawMap
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	if v, ok := payload["url"].(string); ok {
		p.URL = v
	}
	if v, ok := payload["image_url"].(string); ok {
		if p.URL == "" {
			p.URL = v
		}
		p.ImageURL = v
	}
	if v, ok := payload["uri"].(string); ok && p.URL == "" {
		p.URL = v
	}
	if v, ok := payload["href"].(string); ok && p.URL == "" {
		p.URL = v
	}
	if v, ok := payload["signed_url"].(string); ok && p.URL == "" {
		p.URL = v
	}
	if v, ok := payload["content_type"].(string); ok {
		p.ContentType = v
	}
	if v, ok := payload["base64"].(string); ok {
		p.Base64 = v
	}
	if v, ok := payload["image_base64"].(string); ok && p.Base64 == "" {
		p.Base64 = v
	}
	if v, ok := payload["base64_data"].(string); ok && p.Base64 == "" {
		p.Base64 = v
	}
	if v, ok := payload["data"].(string); ok {
		p.Data = v
	}
	if v, ok := payload["b64_json"].(string); ok {
		p.B64JSON = v
	}

	return nil
}

func (p falImagePayload) firstURL() string {
	if strings.TrimSpace(p.URL) != "" {
		return p.URL
	}
	return p.ImageURL
}

func (p falImagePayload) firstBase64() string {
	if strings.TrimSpace(p.Base64) != "" {
		return p.Base64
	}
	if strings.TrimSpace(p.Data) != "" {
		return p.Data
	}
	return p.B64JSON
}

type falGenerationEnvelope struct {
	RequestID  string            `json:"request_id"`
	Status     string            `json:"status"`
	Images     []falImagePayload `json:"images"`
	Output     []falImagePayload `json:"output"`
	Outputs    []falImagePayload `json:"outputs"`
	Data       []falImagePayload `json:"data"`
	Result     []falImagePayload `json:"result"`
	Variants   []falImagePayload `json:"variants"`
	Text       string            `json:"text"`
	Message    string            `json:"message"`
	OutputText string            `json:"output_text"`
	Error      *falAPIError      `json:"error"`
	Response   *falInnerResponse `json:"response"`
}

func (e *falGenerationEnvelope) mergeInner() {
	if e == nil || e.Response == nil {
		return
	}
	inner := e.Response
	if inner.Status != "" {
		e.Status = inner.Status
	}
	if e.Error == nil && inner.Error != nil {
		e.Error = inner.Error
	}
	if e.Text == "" {
		e.Text = inner.Text
	}
	if e.Message == "" {
		e.Message = inner.Message
	}
	if e.OutputText == "" {
		e.OutputText = inner.OutputText
	}
}

type falInnerResponse struct {
	Status     string            `json:"status"`
	Images     []falImagePayload `json:"images"`
	Output     []falImagePayload `json:"output"`
	Outputs    []falImagePayload `json:"outputs"`
	Data       []falImagePayload `json:"data"`
	Result     []falImagePayload `json:"result"`
	Variants   []falImagePayload `json:"variants"`
	Text       string            `json:"text"`
	Message    string            `json:"message"`
	OutputText string            `json:"output_text"`
	Error      *falAPIError      `json:"error"`
}

type falSubmissionResponse struct {
	RequestID   string            `json:"request_id"`
	Status      string            `json:"status"`
	StatusURL   string            `json:"status_url"`
	ResponseURL string            `json:"response_url"`
	Images      []falImagePayload `json:"images"`
	Output      []falImagePayload `json:"output"`
	Outputs     []falImagePayload `json:"outputs"`
	Data        []falImagePayload `json:"data"`
	Result      []falImagePayload `json:"result"`
	Variants    []falImagePayload `json:"variants"`
	Text        string            `json:"text"`
	Message     string            `json:"message"`
	OutputText  string            `json:"output_text"`
	Error       *falAPIError      `json:"error"`
	Response    *falInnerResponse `json:"response"`
}

func (s falSubmissionResponse) toEnvelope() *falGenerationEnvelope {
	envelope := &falGenerationEnvelope{
		RequestID:  s.RequestID,
		Status:     s.Status,
		Images:     append([]falImagePayload(nil), s.Images...),
		Output:     append([]falImagePayload(nil), s.Output...),
		Outputs:    append([]falImagePayload(nil), s.Outputs...),
		Data:       append([]falImagePayload(nil), s.Data...),
		Result:     append([]falImagePayload(nil), s.Result...),
		Variants:   append([]falImagePayload(nil), s.Variants...),
		Text:       s.Text,
		Message:    s.Message,
		OutputText: s.OutputText,
		Error:      s.Error,
		Response:   s.Response,
	}
	envelope.mergeInner()
	return envelope
}

type falAPIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Capabilities returns the capabilities of the model.
func (f *FalAI) Capabilities(model entity.DbModel) *ModelCapabilities {
	return &ModelCapabilities{
		InputModalities:    model.InputModalities,
		OutputModalities:   model.OutputModalities,
		MaxImages:          model.MaxImages,
		SupportedSizes:     model.SupportedSizes,
		SupportedDurations: model.SupportedDurations,
		SupportsStream:     model.SupportsStreaming,
		SupportsCancel:     model.SupportsCancel,
		SupportsAsync:      true, // FalAI always uses async polling
	}
}

// Validate checks if the request is valid for the model.
func (f *FalAI) Validate(request entity.GenerateContentRequest, model entity.DbModel) error {
	if strings.TrimSpace(request.Prompt) == "" {
		return errors.New("prompt is required")
	}

	// Check if image-to-image mode requires input images
	mode := strings.TrimSpace(model.GenerationMode)
	if mode == "" {
		// Infer from ModelID
		if strings.Contains(strings.ToLower(model.ModelID), "image-to-image") ||
			strings.Contains(strings.ToLower(model.ModelID), "edit") {
			mode = string(falModeImageToImage)
		}
	}

	if mode == string(falModeImageToImage) {
		images := request.GetImages()
		if len(images) == 0 {
			return errors.New("image-to-image model requires at least one reference image")
		}
	}

	return nil
}
