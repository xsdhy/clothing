package dto

import (
	"clothing/internal/entity/common"
	"time"
)

// UserSummary is a lightweight user description returned to clients.
type UserSummary struct {
	ID          uint      `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserQuery supports listing users with pagination.
type UserQuery struct {
	common.BaseParams
	Role    string `json:"role" form:"role" query:"role"`
	Keyword string `json:"keyword" form:"keyword" query:"keyword"`
}

// UserCreateRequest is the payload for creating a user.
type UserCreateRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role" binding:"required"`
	IsActive    *bool  `json:"is_active"`
}

// UserUpdateRequest is the payload for updating a user.
type UserUpdateRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Role        *string `json:"role,omitempty"`
	Password    *string `json:"password,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// UserListResponse is the response for listing users.
type UserListResponse struct {
	Users []UserSummary `json:"users"`
	Meta  *common.Meta  `json:"meta"`
}
