package llm

import (
	"encoding/json"
	"strings"

	"clothing/internal/utils"
)

type streamContentPart struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL *struct {
		URL string `json:"url"`
	} `json:"image_url,omitempty"`
	ImageBase64 string `json:"image_base64,omitempty"`
	B64JSON     string `json:"b64_json,omitempty"`
	Data        string `json:"data,omitempty"`
	MimeType    string `json:"mime_type,omitempty"`
}

func parseStreamedContent(raw string) (string, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}

	if img, text, ok := parseStreamJSON(raw); ok {
		return img, text
	}

	if data := extractFirstDataURL(raw); data != "" {
		return data, strings.TrimSpace(strings.Replace(raw, data, "", 1))
	}

	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw, ""
	}

	if idx := strings.Index(raw, "https://"); idx != -1 {
		end := findURLEnd(raw[idx:])
		url := raw[idx : idx+end]
		return url, strings.TrimSpace(strings.Replace(raw, url, "", 1))
	}

	return "", raw
}

func parseStreamJSON(raw string) (string, string, bool) {
	var single streamContentPart
	if err := json.Unmarshal([]byte(raw), &single); err == nil && single.Type != "" {
		if img := extractImageFromPart(single); img != "" {
			return img, "", true
		}
		if text := strings.TrimSpace(single.Text); text != "" {
			return "", text, true
		}
	}

	var parts []streamContentPart
	if err := json.Unmarshal([]byte(raw), &parts); err == nil && len(parts) > 0 {
		for _, part := range parts {
			if img := extractImageFromPart(part); img != "" {
				return img, "", true
			}
		}

		var textParts []string
		for _, part := range parts {
			if text := strings.TrimSpace(part.Text); text != "" {
				textParts = append(textParts, text)
			}
		}
		if len(textParts) > 0 {
			return "", strings.Join(textParts, "\n"), true
		}

		return "", "", true
	}

	return "", "", false
}

func extractImageFromPart(part streamContentPart) string {
	if part.ImageURL != nil && strings.TrimSpace(part.ImageURL.URL) != "" {
		return strings.TrimSpace(part.ImageURL.URL)
	}

	if data := strings.TrimSpace(part.ImageBase64); data != "" {
		return utils.EnsureDataURL(data)
	}

	if data := strings.TrimSpace(part.B64JSON); data != "" {
		return utils.EnsureDataURL(data)
	}

	if data := strings.TrimSpace(part.Data); data != "" {
		if part.MimeType != "" {
			return "data:" + part.MimeType + ";base64," + data
		}
		return utils.EnsureDataURL(data)
	}

	return ""
}

func extractFirstDataURL(raw string) string {
	if strings.HasPrefix(raw, "data:") {
		return raw
	}

	idx := strings.Index(raw, "data:")
	if idx == -1 {
		return ""
	}

	rest := raw[idx:]
	end := findDataURLEnd(rest)
	if end == -1 {
		return strings.TrimSpace(rest)
	}

	return strings.TrimSpace(rest[:end])
}

func findDataURLEnd(raw string) int {
	for i, r := range raw {
		if r == '\n' || r == '\r' || r == '\t' || r == ' ' {
			return i
		}
	}
	return -1
}

func findURLEnd(raw string) int {
	for i, r := range raw {
		if r == '\n' || r == '\r' || r == '\t' || r == ' ' || r == '"' || r == '\'' {
			return i
		}
	}
	return len(raw)
}
