package utils

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// DecodeImagePayload decodes an inline base64 or data URL payload and returns
// the raw bytes together with a guessed file extension.
func DecodeImagePayload(payload string) ([]byte, string, error) {
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" {
		return nil, "", fmt.Errorf("empty image payload")
	}

	mimeType, base64Payload := SplitDataURL(trimmed)
	base64Payload = strings.TrimSpace(base64Payload)
	if base64Payload == "" {
		return nil, "", fmt.Errorf("empty base64 payload")
	}

	data, err := base64.StdEncoding.DecodeString(base64Payload)
	if err != nil {
		return nil, "", fmt.Errorf("decode base64: %w", err)
	}

	ext := ExtensionFromMime(mimeType)
	if ext == "" {
		ext = ExtensionFromMime(http.DetectContentType(data))
	}
	if ext == "" {
		ext = "png"
	}

	return data, ext, nil
}
