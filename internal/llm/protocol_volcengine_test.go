package llm

import (
	"context"
	"testing"

	volcModel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

func TestBuildVolcengineVideoPrompt(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		size     string
		duration int
		want     string
	}{
		{
			name:     "empty prompt",
			prompt:   "   ",
			size:     "720p",
			duration: 5,
			want:     "",
		},
		{
			name:     "append size and duration",
			prompt:   "hello world",
			size:     "720P",
			duration: 6,
			want:     "hello world --rs 720p --dur 6",
		},
		{
			name:     "skip existing size",
			prompt:   "scene --rs 480p",
			size:     "720p",
			duration: 0,
			want:     "scene --rs 480p",
		},
		{
			name:     "skip existing duration",
			prompt:   "scene --dur 4",
			size:     "",
			duration: 6,
			want:     "scene --dur 4",
		},
		{
			name:     "existing params in different case",
			prompt:   "scene --RS 480P --DUR 5",
			size:     "720p",
			duration: 6,
			want:     "scene --RS 480P --DUR 5",
		},
	}

	for _, tt := range tests {
		got := buildVolcengineVideoPrompt(tt.prompt, tt.size, tt.duration)
		if got != tt.want {
			t.Fatalf("%s: expected %q, got %q", tt.name, tt.want, got)
		}
	}
}

func TestBuildVolcengineVideoImages(t *testing.T) {
	ctx := context.Background()
	dataURL := "data:image/png;base64,AAA"

	t.Run("single image uses first frame", func(t *testing.T) {
		items, err := buildVolcengineVideoImages(ctx, "seedance-1-0-pro", []string{dataURL})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
		if items[0].Role == nil || *items[0].Role != "first_frame" {
			t.Fatalf("expected first_frame role, got %#v", items[0].Role)
		}
		if items[0].ImageURL == nil || items[0].ImageURL.URL != dataURL {
			t.Fatalf("expected image url %q, got %#v", dataURL, items[0].ImageURL)
		}
	})

	t.Run("two images uses first and last frame", func(t *testing.T) {
		items, err := buildVolcengineVideoImages(ctx, "seedance-1-0-pro", []string{dataURL, "data:image/png;base64,BBB"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(items))
		}
		if items[0].Role == nil || *items[0].Role != "first_frame" {
			t.Fatalf("expected first_frame role, got %#v", items[0].Role)
		}
		if items[1].Role == nil || *items[1].Role != "last_frame" {
			t.Fatalf("expected last_frame role, got %#v", items[1].Role)
		}
	})

	t.Run("three images defaults to first and last frame", func(t *testing.T) {
		items, err := buildVolcengineVideoImages(ctx, "seedance-1-0-pro", []string{
			dataURL,
			"data:image/png;base64,BBB",
			"data:image/png;base64,CCC",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(items))
		}
		if items[0].ImageURL == nil || items[0].ImageURL.URL != dataURL {
			t.Fatalf("expected first frame %q, got %#v", dataURL, items[0].ImageURL)
		}
		if items[1].ImageURL == nil || items[1].ImageURL.URL != "data:image/png;base64,CCC" {
			t.Fatalf("expected last frame, got %#v", items[1].ImageURL)
		}
	})

	t.Run("reference images for lite i2v model", func(t *testing.T) {
		items, err := buildVolcengineVideoImages(ctx, "doubao-seedance-1-0-lite-i2v", []string{
			dataURL,
			"data:image/png;base64,BBB",
			"data:image/png;base64,CCC",
			"data:image/png;base64,DDD",
			"data:image/png;base64,EEE",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 4 {
			t.Fatalf("expected 4 items, got %d", len(items))
		}
		for idx, item := range items {
			if item.Role == nil || *item.Role != "reference_image" {
				t.Fatalf("item %d expected reference_image role, got %#v", idx, item.Role)
			}
		}
	})
}

func TestCollectVolcengineVideoAssets(t *testing.T) {
	content := volcModel.Content{
		VideoURL:     " https://example.com/video.mp4 ",
		LastFrameURL: " https://example.com/last.png ",
	}
	assets := collectVolcengineVideoAssets(content)
	if len(assets) != 2 {
		t.Fatalf("expected 2 assets, got %d", len(assets))
	}
	if assets[0] != "https://example.com/video.mp4" {
		t.Fatalf("unexpected video url: %q", assets[0])
	}
	if assets[1] != "https://example.com/last.png" {
		t.Fatalf("unexpected last frame url: %q", assets[1])
	}
}
