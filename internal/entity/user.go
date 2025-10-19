package entity

import "time"

const (
	UserRoleSuperAdmin = "super_admin"
	UserRoleAdmin      = "admin"
	UserRoleUser       = "user"
)

// DbUser represents a persisted user account.
type DbUser struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `gorm:"column:email;type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	DisplayName  string    `gorm:"column:display_name;type:varchar(255)" json:"display_name"`
	Role         string    `gorm:"column:role;type:varchar(50);index;not null" json:"role"`
	IsActive     bool      `gorm:"column:is_active;not null;default:true" json:"is_active"`
}

// TableName overrides default pluralised name.
func (DbUser) TableName() string {
	return "users"
}

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
	BaseParams
	Role    string `json:"role" form:"role" query:"role"`
	Keyword string `json:"keyword" form:"keyword" query:"keyword"`
}

// AuthStatusResponse indicates whether the system already has users.
type AuthStatusResponse struct {
	HasUser bool `json:"has_user"`
}

type AuthLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthRegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role,omitempty"`
}

type AuthResponse struct {
	Token     string      `json:"token"`
	ExpiresAt time.Time   `json:"expires_at"`
	User      UserSummary `json:"user"`
}

type UserCreateRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role" binding:"required"`
	IsActive    *bool  `json:"is_active"`
}

type UserUpdateRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Role        *string `json:"role,omitempty"`
	Password    *string `json:"password,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type UserListResponse struct {
	Users []UserSummary `json:"users"`
	Meta  *Meta         `json:"meta"`
}
