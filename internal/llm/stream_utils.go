package llm

import (
	"encoding/json"
	"strings"

	"clothing/internal/utils"
	"github.com/tidwall/gjson"
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

func responseEventType(current string, payload []byte) string {
	current = strings.TrimSpace(current)
	if current != "" {
		return current
	}
	if t := gjson.GetBytes(payload, "type"); t.Exists() {
		return strings.TrimSpace(t.String())
	}
	if e := gjson.GetBytes(payload, "event"); e.Exists() {
		return strings.TrimSpace(e.String())
	}
	return ""
}

func responseEventIndex(payload []byte) (int, bool) {
	paths := []string{
		"delta.index",
		"index",
		"output_index",
		"delta.output_index",
		"delta.content_index",
		"content_index",
		"response.index",
		"response.output_index",
		"partial_image_index",
	}
	for _, path := range paths {
		res := gjson.GetBytes(payload, path)
		if res.Exists() {
			return int(res.Int()), true
		}
	}
	return 0, false
}

func responseEventMIME(payload []byte) string {
	paths := []string{
		"delta.mime_type",
		"mime_type",
		"delta.content.#.mime_type",
		"delta.content.#.inline_data.mime_type",
		"delta.inline_data.mime_type",
		"delta.content.#.image_url.mime_type",
		"delta.image_url.mime_type",
		"inline_data.mime_type",
		"image_url.mime_type",
		"response.output.#.content.#.mime_type",
		"response.output.#.content.#.inline_data.mime_type",
	}
	for _, path := range paths {
		res := gjson.GetBytes(payload, path)
		if res.Exists() {
			if mime := strings.TrimSpace(res.String()); mime != "" {
				return mime
			}
		}
	}
	return ""
}

type responseImageCandidate struct {
	Index   int
	MIME    string
	Base64  string
	DataURL string
	URL     string
}

func (c responseImageCandidate) dataURL() string {
	if c.DataURL != "" {
		return c.DataURL
	}
	if c.Base64 != "" {
		mime := c.MIME
		if mime == "" {
			mime = "image/png"
		}
		return "data:" + mime + ";base64," + c.Base64
	}
	return ""
}

func normalizeBase64Chunk(chunk string) string {
	chunk = strings.TrimSpace(chunk)
	chunk = strings.ReplaceAll(chunk, "\n", "")
	chunk = strings.ReplaceAll(chunk, "\r", "")
	chunk = strings.ReplaceAll(chunk, " ", "")
	return chunk
}

func extractResponseImageCandidates(payload []byte, paths ...string) []responseImageCandidate {
	var candidates []responseImageCandidate
	for _, path := range paths {
		res := gjson.GetBytes(payload, path)
		if !res.Exists() {
			continue
		}
		candidates = append(candidates, collectImageCandidates(res)...)
	}
	return dedupeImageCandidates(candidates)
}

func collectImageCandidates(node gjson.Result) []responseImageCandidate {
	var candidates []responseImageCandidate
	if !node.Exists() {
		return candidates
	}
	if node.IsArray() {
		node.ForEach(func(_ gjson.Result, value gjson.Result) bool {
			candidates = append(candidates, collectImageCandidates(value)...)
			return true
		})
		return candidates
	}
	if !node.IsObject() {
		return candidates
	}

	if candidate, ok := imageCandidateFromObject(node); ok {
		candidates = append(candidates, candidate)
		return candidates
	}

	node.ForEach(func(key, value gjson.Result) bool {
		if key.String() == "type" || key.String() == "text" || key.String() == "value" || key.String() == "index" {
			return true
		}
		candidates = append(candidates, collectImageCandidates(value)...)
		return true
	})
	return candidates
}

func imageCandidateFromObject(obj gjson.Result) (responseImageCandidate, bool) {
	var candidate responseImageCandidate
	if !obj.IsObject() {
		return candidate, false
	}

	typ := strings.TrimSpace(obj.Get("type").String())
	hasImageData := obj.Get("image_base64").Exists() || obj.Get("b64_json").Exists() || obj.Get("data").Exists() || obj.Get("inline_data.data").Exists() || obj.Get("image_url.url").Exists() || obj.Get("partial_image_b64").Exists() || obj.Get("url").Exists()
	if !hasImageData && typ != "output_image" && typ != "image" && !strings.HasSuffix(typ, "_image") {
		return candidate, false
	}

	candidate.Index = int(obj.Get("index").Int())
	mime := responseEventMIMEFromObject(obj)
	if mime != "" {
		candidate.MIME = mime
	}

	fields := []string{
		"image_base64",
		"b64_json",
		"base64",
		"base64_data",
		"partial_image_b64",
		"data",
		"chunk",
		"inline_data.data",
		"image_url.url",
		"url",
	}
	for _, field := range fields {
		value := strings.TrimSpace(obj.Get(field).String())
		if value == "" {
			continue
		}
		if strings.HasPrefix(value, "data:image") {
			candidate.DataURL = value
			break
		}
		if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
			candidate.URL = value
			break
		}
		candidate.Base64 = normalizeBase64Chunk(value)
	}

	if candidate.MIME == "" && candidate.Base64 != "" {
		candidate.MIME = "image/png"
	}

	if candidate.DataURL == "" && candidate.Base64 == "" && candidate.URL == "" {
		return candidate, false
	}

	return candidate, true
}

func responseEventMIMEFromObject(obj gjson.Result) string {
	paths := []string{
		"mime_type",
		"inline_data.mime_type",
		"image_url.mime_type",
	}
	for _, path := range paths {
		value := strings.TrimSpace(obj.Get(path).String())
		if value != "" {
			return value
		}
	}
	return ""
}

func dedupeImageCandidates(candidates []responseImageCandidate) []responseImageCandidate {
	if len(candidates) <= 1 {
		return candidates
	}
	type key struct {
		Index int
		Data  string
	}
	seen := make(map[key]struct{}, len(candidates))
	result := make([]responseImageCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		k := key{Index: candidate.Index, Data: candidate.dataURL()}
		if candidate.URL != "" {
			k.Data = candidate.URL
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		result = append(result, candidate)
	}
	return result
}

func responseEventTextFragments(payload []byte, paths ...string) []string {
	var fragments []string
	for _, path := range paths {
		res := gjson.GetBytes(payload, path)
		if !res.Exists() {
			continue
		}
		fragments = append(fragments, collectTextFragments(res)...)
	}
	return dedupeStrings(fragments)
}

func collectTextFragments(node gjson.Result) []string {
	if !node.Exists() {
		return nil
	}
	if node.IsArray() {
		var fragments []string
		node.ForEach(func(_ gjson.Result, value gjson.Result) bool {
			fragments = append(fragments, collectTextFragments(value)...)
			return true
		})
		return fragments
	}
	if !node.IsObject() {
		value := strings.TrimSpace(node.String())
		if value == "" {
			return nil
		}
		return []string{value}
	}

	typ := strings.TrimSpace(node.Get("type").String())
	var fragments []string
	if isOutputTextType(typ) {
		if text := strings.TrimSpace(node.Get("text").String()); text != "" {
			fragments = append(fragments, text)
		}
		if text := strings.TrimSpace(node.Get("value").String()); text != "" {
			fragments = append(fragments, text)
		}
	}
	if delta := strings.TrimSpace(node.Get("delta").String()); delta != "" && (typ == "" || strings.Contains(typ, "text")) {
		fragments = append(fragments, delta)
	}

	node.ForEach(func(key, value gjson.Result) bool {
		if key.String() == "type" || key.String() == "text" || key.String() == "value" || key.String() == "index" {
			return true
		}
		fragments = append(fragments, collectTextFragments(value)...)
		return true
	})
	return fragments
}

func isOutputTextType(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return false
	}
	if strings.Contains(value, "output_text") {
		return true
	}
	switch value {
	case "text", "output.message", "message", "assistant_message", "output_message":
		return true
	default:
		return false
	}
}

func dedupeStrings(items []string) []string {
	if len(items) <= 1 {
		return items
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func responseErrorMessage(payload []byte) string {
	paths := []string{
		"error.message",
		"message",
		"error",
		"response.error.message",
		"response.message",
	}
	for _, path := range paths {
		if res := gjson.GetBytes(payload, path); res.Exists() {
			if msg := strings.TrimSpace(res.String()); msg != "" {
				return msg
			}
		}
	}
	return ""
}
