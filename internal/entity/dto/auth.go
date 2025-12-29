package dto

import "time"

// AuthStatusResponse indicates whether the system already has users.
type AuthStatusResponse struct {
	HasUser bool `json:"has_user"`
}

// AuthLoginRequest is the login request payload.
type AuthLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthRegisterRequest is the registration request payload.
type AuthRegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role,omitempty"`
}

// AuthResponse is returned after successful login/registration.
type AuthResponse struct {
	Token     string      `json:"token"`
	ExpiresAt time.Time   `json:"expires_at"`
	User      UserSummary `json:"user"`
}
