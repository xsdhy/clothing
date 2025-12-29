package db

import "time"

const (
	UserRoleSuperAdmin = "super_admin"
	UserRoleAdmin      = "admin"
	UserRoleUser       = "user"
)

// User 表示持久化的用户账户。
type User struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `gorm:"column:email;type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	DisplayName  string    `gorm:"column:display_name;type:varchar(255)" json:"display_name"`
	Role         string    `gorm:"column:role;type:varchar(50);index;not null" json:"role"`
	IsActive     bool      `gorm:"column:is_active;not null;default:true" json:"is_active"`
}

// TableName 指定表名。
func (User) TableName() string {
	return "users"
}
