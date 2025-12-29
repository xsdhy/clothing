package llm

import (
	"clothing/internal/utils"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// MediaFormat specifies the desired output format for prepared media.
type MediaFormat string

const (
	// MediaFormatBase64 returns pure base64 encoded data without prefix.
	MediaFormatBase64 MediaFormat = "base64"
	// MediaFormatDataURL returns data URL format (data:mime;base64,...).
	MediaFormatDataURL MediaFormat = "data_url"
	// MediaFormatURL keeps the original URL if input is a URL, otherwise converts to data URL.
	MediaFormatURL MediaFormat = "url"
)

// PreparedMedia represents a processed media asset ready for use.
type PreparedMedia struct {
	// Base64 contains the pure base64 encoded data (without data: prefix).
	Base64 string
	// DataURL contains the full data URL (data:mime;base64,...).
	DataURL string
	// URL contains the original URL if the input was a URL.
	URL string
	// MimeType is the detected or provided MIME type.
	MimeType string
	// Size is the size in bytes of the decoded content.
	Size int64
}

// MediaService provides unified media processing capabilities.
type MediaService interface {
	// PrepareImage processes a single image input and returns it in the requested format.
	// Input can be: URL (http/https), data URL, or base64 string.
	PrepareImage(ctx context.Context, input string, format MediaFormat) (*PreparedMedia, error)

	// PrepareImages processes multiple image inputs in parallel.
	PrepareImages(ctx context.Context, inputs []string, format MediaFormat) ([]*PreparedMedia, error)

	// DownloadAsBase64 downloads a remote URL and returns base64 encoded content.
	DownloadAsBase64(ctx context.Context, url string) (base64Data, mimeType string, err error)
}

// defaultMediaService implements MediaService.
type defaultMediaService struct {
	httpClient *http.Client
}

// NewMediaService creates a new MediaService instance.
func NewMediaService() MediaService {
	return &defaultMediaService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PrepareImage processes a single image input.
func (s *defaultMediaService) PrepareImage(ctx context.Context, input string, format MediaFormat) (*PreparedMedia, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, errors.New("empty image input")
	}

	// Case 1: HTTP/HTTPS URL
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return s.prepareFromURL(ctx, trimmed, format)
	}

	// Case 2: Data URL or base64 string
	return s.prepareFromInline(trimmed, format)
}

// PrepareImages processes multiple image inputs.
func (s *defaultMediaService) PrepareImages(ctx context.Context, inputs []string, format MediaFormat) ([]*PreparedMedia, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	results := make([]*PreparedMedia, 0, len(inputs))
	var errs []string

	for idx, input := range inputs {
		prepared, err := s.PrepareImage(ctx, input, format)
		if err != nil {
			errs = append(errs, fmt.Sprintf("image %d: %v", idx, err))
			logrus.WithFields(logrus.Fields{
				"index": idx,
				"error": err,
			}).Warn("media_service: failed to prepare image")
			continue
		}
		results = append(results, prepared)
	}

	if len(results) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("all images failed: %s", strings.Join(errs, "; "))
	}

	return results, nil
}

// DownloadAsBase64 downloads a URL and returns base64 encoded content.
func (s *defaultMediaService) DownloadAsBase64(ctx context.Context, url string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", fmt.Errorf("create download request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("read image body: %w", err)
	}

	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}
	mimeType = normalizeMimeType(mimeType)

	logrus.WithFields(logrus.Fields{
		"mime":       mimeType,
		"size_bytes": len(data),
		"url":        truncateString(url, 128),
	}).Debug("media_service: downloaded image")

	return base64.StdEncoding.EncodeToString(data), mimeType, nil
}

// prepareFromURL downloads a URL and prepares the media.
func (s *defaultMediaService) prepareFromURL(ctx context.Context, url string, format MediaFormat) (*PreparedMedia, error) {
	// If format is URL, we can just return the URL directly
	if format == MediaFormatURL {
		return &PreparedMedia{
			URL:      url,
			MimeType: "",
		}, nil
	}

	// Download and encode
	b64, mimeType, err := s.DownloadAsBase64(ctx, url)
	if err != nil {
		return nil, err
	}

	raw, _ := base64.StdEncoding.DecodeString(b64)

	return &PreparedMedia{
		Base64:   b64,
		DataURL:  fmt.Sprintf("data:%s;base64,%s", mimeType, b64),
		URL:      url,
		MimeType: mimeType,
		Size:     int64(len(raw)),
	}, nil
}

// prepareFromInline processes inline data (data URL or base64 string).
func (s *defaultMediaService) prepareFromInline(input string, format MediaFormat) (*PreparedMedia, error) {
	// Normalize to data URL format
	dataURL := utils.EnsureDataURL(input)
	mimeType, b64Data := utils.SplitDataURL(dataURL)
	b64Data = strings.TrimSpace(b64Data)

	if b64Data == "" {
		return nil, errors.New("empty base64 payload")
	}

	// Validate base64
	raw, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}

	// Re-encode to normalize (remove whitespace etc.)
	normalized := base64.StdEncoding.EncodeToString(raw)

	if mimeType == "" {
		mimeType = http.DetectContentType(raw)
	}
	mimeType = normalizeMimeType(mimeType)

	return &PreparedMedia{
		Base64:   normalized,
		DataURL:  fmt.Sprintf("data:%s;base64,%s", mimeType, normalized),
		MimeType: mimeType,
		Size:     int64(len(raw)),
	}, nil
}

// normalizeMimeType cleans up MIME type strings.
func normalizeMimeType(mimeType string) string {
	v := strings.TrimSpace(mimeType)
	if v == "" {
		return "image/jpeg"
	}
	// Strip charset if present
	if idx := strings.Index(v, ";"); idx > 0 {
		return strings.TrimSpace(v[:idx])
	}
	return v
}

// truncateString truncates a string to max length for logging.
func truncateString(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// PickFirstImage selects the first valid image from a list.
// Returns the URL if it's a remote URL, or the base64 payload if inline.
// This is a utility function to maintain compatibility with existing code.
func PickFirstImage(images []string) (url string, base64Payload string) {
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

		// Assume it's a base64 string
		return "", trimmed
	}
	return "", ""
}
