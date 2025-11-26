package sql

import (
	"clothing/internal/entity"
	"context"
	"fmt"

	"gorm.io/gorm"
)

// ListTags returns all tags with their usage counts.
func (r *GormRepository) ListTags(ctx context.Context) ([]entity.DbTag, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("repository not initialised")
	}

	var tags []entity.DbTag
	query := r.db.WithContext(ctx).
		Model(&entity.DbTag{}).
		Select("tags.*, COUNT(usage_record_tags.usage_record_id) as usage_count").
		Joins("LEFT JOIN usage_record_tags ON usage_record_tags.tag_id = tags.id").
		Group("tags.id").
		Order("tags.name ASC")

	if err := query.Find(&tags).Error; err != nil {
		return nil, err
	}

	return tags, nil
}

// CreateTag inserts a new tag.
func (r *GormRepository) CreateTag(ctx context.Context, tag *entity.DbTag) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repository not initialised")
	}
	if tag == nil {
		return fmt.Errorf("tag is nil")
	}
	return r.db.WithContext(ctx).Create(tag).Error
}

// UpdateTag updates tag fields.
func (r *GormRepository) UpdateTag(ctx context.Context, id uint, updates map[string]interface{}) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repository not initialised")
	}
	if id == 0 {
		return fmt.Errorf("invalid tag id")
	}
	if len(updates) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).Model(&entity.DbTag{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteTag removes a tag and its associations.
func (r *GormRepository) DeleteTag(ctx context.Context, id uint) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repository not initialised")
	}
	if id == 0 {
		return fmt.Errorf("invalid tag id")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tag_id = ?", id).Delete(&entity.DbUsageRecordTag{}).Error; err != nil {
			return err
		}

		result := tx.Delete(&entity.DbTag{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

// FindTagsByIDs fetches tags by ids.
func (r *GormRepository) FindTagsByIDs(ctx context.Context, ids []uint) ([]entity.DbTag, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("repository not initialised")
	}
	if len(ids) == 0 {
		return []entity.DbTag{}, nil
	}

	var tags []entity.DbTag
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

// SetUsageRecordTags replaces tags for a usage record.
func (r *GormRepository) SetUsageRecordTags(ctx context.Context, recordID uint, tagIDs []uint) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repository not initialised")
	}
	if recordID == 0 {
		return fmt.Errorf("invalid usage record id")
	}

	uniqueIDs := make([]uint, 0, len(tagIDs))
	seen := make(map[uint]struct{}, len(tagIDs))
	for _, id := range tagIDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var record entity.DbUsageRecord
		if err := tx.First(&record, recordID).Error; err != nil {
			return err
		}

		var tags []entity.DbTag
		if len(uniqueIDs) > 0 {
			if err := tx.Where("id IN ?", uniqueIDs).Find(&tags).Error; err != nil {
				return err
			}
			if len(tags) != len(uniqueIDs) {
				return fmt.Errorf("some tags do not exist")
			}
		}

		if err := tx.Model(&record).Association("Tags").Replace(tags); err != nil {
			return err
		}
		return nil
	})
}
