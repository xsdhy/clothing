package dto

import "time"

// Tag is the DTO representation of a tag.
type Tag struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	UsageCount  int64     `json:"usage_count,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	Description string    `json:"description,omitempty"`
}

// TagListResponse is the response for listing tags.
type TagListResponse struct {
	Tags []Tag `json:"tags"`
}

// TagDetailResponse is the response for a single tag.
type TagDetailResponse struct {
	Tag Tag `json:"tag"`
}
