package sql

import (
	"clothing/internal/entity"
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// CreateProvider inserts a new provider row.
func (r *GormRepository) CreateProvider(ctx context.Context, provider *entity.DbProvider) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repository not initialised")
	}
	if provider == nil {
		return fmt.Errorf("provider is nil")
	}
	provider.ID = strings.TrimSpace(provider.ID)
	provider.Driver = strings.TrimSpace(provider.Driver)
	provider.Name = strings.TrimSpace(provider.Name)
	if provider.ID == "" {
		return fmt.Errorf("provider id is required")
	}
	if provider.Driver == "" {
		return fmt.Errorf("provider driver is required")
	}
	if provider.Name == "" {
		return fmt.Errorf("provider name is required")
	}
	return r.db.WithContext(ctx).Create(provider).Error
}

// UpdateProvider updates provider fields using a map of updates.
func (r *GormRepository) UpdateProvider(ctx context.Context, id string, updates map[string]interface{}) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repository not initialised")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("provider id is required")
	}
	if len(updates) == 0 {
		return nil
	}

	delete(updates, "id")
	delete(updates, "ID")

	result := r.db.WithContext(ctx).
		Model(&entity.DbProvider{}).
		Where("id = ?", id).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteProvider deletes a provider by id.
func (r *GormRepository) DeleteProvider(ctx context.Context, id string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repository not initialised")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("provider id is required")
	}
	result := r.db.WithContext(ctx).Delete(&entity.DbProvider{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	// Cascade delete models
	if err := r.db.WithContext(ctx).Where("provider_id = ?", id).Delete(&entity.DbModel{}).Error; err != nil {
		return err
	}
	return nil
}

// ListProviders returns providers optionally filtering inactive ones.
func (r *GormRepository) ListProviders(ctx context.Context, includeInactive bool) ([]entity.DbProvider, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("repository not initialised")
	}

	query := r.db.WithContext(ctx).Model(&entity.DbProvider{})
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	query = query.Preload("Models", func(tx *gorm.DB) *gorm.DB {
		if !includeInactive {
			return tx.Where("is_active = ?", true)
		}
		return tx
	})

	var providers []entity.DbProvider
	if err := query.Order("id ASC").Find(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}

// GetProvider returns a single provider by ID (with models).
func (r *GormRepository) GetProvider(ctx context.Context, id string) (*entity.DbProvider, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("repository not initialised")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("provider id is required")
	}

	var provider entity.DbProvider
	if err := r.db.WithContext(ctx).Preload("Models").
		First(&provider, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

// GetProviderWithModel returns provider and a specific model.
func (r *GormRepository) GetProviderWithModel(ctx context.Context, providerID, modelID string, includeInactive bool) (*entity.DbProvider, *entity.DbModel, error) {
	if r == nil || r.db == nil {
		return nil, nil, fmt.Errorf("repository not initialised")
	}
	providerID = strings.TrimSpace(providerID)
	modelID = strings.TrimSpace(modelID)
	if providerID == "" || modelID == "" {
		return nil, nil, fmt.Errorf("provider and model id are required")
	}

	var provider entity.DbProvider
	providerQuery := r.db.WithContext(ctx).Where("id = ?", providerID)
	providerQuery = providerQuery.Preload("Models", func(tx *gorm.DB) *gorm.DB {
		if !includeInactive {
			return tx.Where("is_active = ?", true)
		}
		return tx
	})
	if !includeInactive {
		providerQuery = providerQuery.Where("is_active = ?", true)
	}
	if err := providerQuery.First(&provider).Error; err != nil {
		return nil, nil, err
	}

	var model entity.DbModel
	modelQuery := r.db.WithContext(ctx).
		Where("provider_id = ? AND model_id = ?", providerID, modelID)
	if !includeInactive {
		modelQuery = modelQuery.Where("is_active = ?", true)
	}
	if err := modelQuery.First(&model).Error; err != nil {
		return nil, nil, err
	}

	return &provider, &model, nil
}

// CreateModel inserts a new model row.
func (r *GormRepository) CreateModel(ctx context.Context, model *entity.DbModel) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repository not initialised")
	}
	if model == nil {
		return fmt.Errorf("model is nil")
	}
	model.ProviderID = strings.TrimSpace(model.ProviderID)
	model.ModelID = strings.TrimSpace(model.ModelID)
	model.Name = strings.TrimSpace(model.Name)
	if model.ProviderID == "" {
		return fmt.Errorf("provider id is required")
	}
	if model.ModelID == "" {
		return fmt.Errorf("model id is required")
	}
	if model.Name == "" {
		return fmt.Errorf("model name is required")
	}
	return r.db.WithContext(ctx).Create(model).Error
}

// UpdateModel updates model fields using a map.
func (r *GormRepository) UpdateModel(ctx context.Context, providerID, modelID string, updates map[string]interface{}) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repository not initialised")
	}
	providerID = strings.TrimSpace(providerID)
	modelID = strings.TrimSpace(modelID)
	if providerID == "" || modelID == "" {
		return fmt.Errorf("provider and model id are required")
	}
	if len(updates) == 0 {
		return nil
	}
	delete(updates, "provider_id")
	delete(updates, "ProviderID")
	delete(updates, "model_id")
	delete(updates, "ModelID")
	delete(updates, "id")
	delete(updates, "ID")

	result := r.db.WithContext(ctx).
		Model(&entity.DbModel{}).
		Where("provider_id = ? AND model_id = ?", providerID, modelID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteModel deletes a model row.
func (r *GormRepository) DeleteModel(ctx context.Context, providerID, modelID string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repository not initialised")
	}
	providerID = strings.TrimSpace(providerID)
	modelID = strings.TrimSpace(modelID)
	if providerID == "" || modelID == "" {
		return fmt.Errorf("provider and model id are required")
	}
	result := r.db.WithContext(ctx).
		Delete(&entity.DbModel{}, "provider_id = ? AND model_id = ?", providerID, modelID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListModels lists models for a provider.
func (r *GormRepository) ListModels(ctx context.Context, providerID string, includeInactive bool) ([]entity.DbModel, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("repository not initialised")
	}
	providerID = strings.TrimSpace(providerID)
	if providerID == "" {
		return nil, fmt.Errorf("provider id is required")
	}

	query := r.db.WithContext(ctx).
		Where("provider_id = ?", providerID)
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	var models []entity.DbModel
	if err := query.Order("model_id ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}
