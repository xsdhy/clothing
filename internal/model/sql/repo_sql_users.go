package sql

import (
	"clothing/internal/entity"
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// CreateUser persists a new user record.
func (r *GormRepository) CreateUser(ctx context.Context, user *entity.DbUser) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repository not initialised")
	}
	if user == nil {
		return fmt.Errorf("user is nil")
	}
	return r.db.WithContext(ctx).Create(user).Error
}

// UpdateUser updates an existing user entry.
func (r *GormRepository) UpdateUser(ctx context.Context, id uint, updates map[string]interface{}) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repository not initialised")
	}
	if id == 0 {
		return fmt.Errorf("invalid user")
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&entity.DbUser{}).Where("id = ?", id).Updates(updates).Error
}

// GetUserByEmail loads a user by email.
func (r *GormRepository) GetUserByEmail(ctx context.Context, email string) (*entity.DbUser, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("repository not initialised")
	}
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return nil, fmt.Errorf("email is empty")
	}

	var user entity.DbUser
	if err := r.db.WithContext(ctx).Where("LOWER(email) = ?", strings.ToLower(trimmed)).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID loads a user by ID.
func (r *GormRepository) GetUserByID(ctx context.Context, id uint) (*entity.DbUser, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("repository not initialised")
	}
	if id == 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	var user entity.DbUser
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ListUsers returns paginated users.
func (r *GormRepository) ListUsers(ctx context.Context, params *entity.UserQuery) ([]entity.DbUser, *entity.Meta, error) {
	if r == nil || r.db == nil {
		return nil, nil, fmt.Errorf("repository not initialised")
	}

	query := r.db.WithContext(ctx).Model(&entity.DbUser{})
	if params != nil {
		if trimmed := strings.TrimSpace(params.Role); trimmed != "" {
			query = query.Where("role = ?", trimmed)
		}
		if keyword := strings.TrimSpace(params.Keyword); keyword != "" {
			kw := "%" + strings.ToLower(keyword) + "%"
			query = query.Where("LOWER(email) LIKE ? OR LOWER(display_name) LIKE ?", kw, kw)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	page := 1
	pageSize := 20
	if params != nil {
		if params.Page > 0 {
			page = int(params.Page)
		}
		if params.PageSize > 0 {
			pageSize = int(params.PageSize)
		}
	}

	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	var users []entity.DbUser
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, nil, err
	}

	meta := r.calculatePagination(total, page, pageSize)
	return users, meta, nil
}

// DeleteUser removes a user by ID.
func (r *GormRepository) DeleteUser(ctx context.Context, id uint) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repository not initialised")
	}
	if id == 0 {
		return fmt.Errorf("invalid user id")
	}
	result := r.db.WithContext(ctx).Delete(&entity.DbUser{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CountUsers returns total user count.
func (r *GormRepository) CountUsers(ctx context.Context) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("repository not initialised")
	}
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.DbUser{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
