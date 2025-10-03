package utils

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const imageBaseDir = "./images"

// SaveImageAsync persists an image to disk in the background. It accepts both a URL and a base64 payload
// so that callers across providers can supply whichever format they receive.
func SaveImageAsync(provider, imageURL, base64Payload string) {
	trimmedURL := strings.TrimSpace(imageURL)
	trimmedBase64 := strings.TrimSpace(base64Payload)
	if trimmedURL == "" && trimmedBase64 == "" {
		return
	}

	go func(provider, imageURL, base64Payload string) {
		if err := persistImageLocally(provider, imageURL, base64Payload); err != nil {
			fields := logrus.Fields{
				"provider": sanitizeForPath(provider),
			}
			if imageURL != "" {
				fields["image_url"] = imageURL
			}
			if base64Payload != "" {
				fields["base64_len"] = len(base64Payload)
			}
			if source := imageSourceForLog(imageURL, base64Payload); source != "" {
				fields["source"] = source
			}
			logrus.WithError(err).WithFields(fields).Warn("save image failed")
		}
	}(provider, trimmedURL, trimmedBase64)
}

func persistImageLocally(provider, imageURL, base64Payload string) error {
	providerKey := sanitizeForPath(provider)
	if providerKey == "" {
		providerKey = "generic"
	}

	var (
		data     []byte
		ext      string
		mimeType string
		fields   = logrus.Fields{
			"provider": providerKey,
		}
	)

	if base64Payload != "" {
		decoded, err := base64.StdEncoding.DecodeString(base64Payload)
		if err != nil {
			return fmt.Errorf("decode base64: %w", err)
		}
		data = decoded
		fields["base64_len"] = len(base64Payload)
	} else if strings.HasPrefix(imageURL, "data:") {
		mimeType, base64Payload = SplitDataURL(imageURL)
		if base64Payload == "" {
			return errors.New("empty data url payload")
		}
		decoded, err := base64.StdEncoding.DecodeString(base64Payload)
		if err != nil {
			return fmt.Errorf("decode data url: %w", err)
		}
		data = decoded
	} else if imageURL != "" {
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
		mimeType = resp.Header.Get("Content-Type")
		if ct := strings.TrimSpace(mimeType); ct != "" {
			fields["mime"] = ct
		}
	} else {
		return errors.New("image payload empty")
	}

	if len(data) == 0 {
		return errors.New("image payload empty")
	}

	if mimeType == "" && base64Payload != "" {
		mimeType = http.DetectContentType(data)
	}

	ext = extensionFromMime(mimeType)
	if ext == "" {
		ext = extensionFromMime(http.DetectContentType(data))
	}
	if ext == "" {
		ext = "jpg"
	}

	dir := filepath.Join(imageBaseDir, fmt.Sprintf("clothing-%s-images", providerKey))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create image dir: %w", err)
	}

	filename := fmt.Sprintf("%s_%d.%s", providerKey, time.Now().UnixNano(), ext)
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write image file: %w", err)
	}

	fields["path"] = path
	fields["size"] = len(data)
	if imageURL != "" {
		fields["image_url"] = imageURL
	}
	logrus.WithFields(fields).Info("image saved")
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

func imageSourceForLog(imageURL, base64Payload string) string {
	if strings.TrimSpace(base64Payload) != "" {
		return "inline_base64"
	}
	if strings.HasPrefix(imageURL, "data:") {
		return "inline_data_url"
	}
	if strings.TrimSpace(imageURL) != "" {
		return "remote_image"
	}
	return ""
}

func sanitizeForPath(provider string) string {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		return ""
	}
	builder := strings.Builder{}
	builder.Grow(len(provider))
	for i := 0; i < len(provider); i++ {
		ch := provider[i]
		switch {
		case ch >= 'a' && ch <= 'z', ch >= '0' && ch <= '9':
			builder.WriteByte(ch)
		case ch >= 'A' && ch <= 'Z':
			builder.WriteByte(ch + ('a' - 'A'))
		default:
			builder.WriteByte('_')
		}
	}
	return builder.String()
}
