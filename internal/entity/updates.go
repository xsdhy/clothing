package entity

// UserUpdates 用户更新字段
type UserUpdates struct {
	DisplayName  *string
	Role         *string
	PasswordHash *string
	IsActive     *bool
}

// ToMap 转换为 GORM 更新 map（内部使用）
func (u UserUpdates) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})
	if u.DisplayName != nil {
		updates["display_name"] = *u.DisplayName
	}
	if u.Role != nil {
		updates["role"] = *u.Role
	}
	if u.PasswordHash != nil {
		updates["password_hash"] = *u.PasswordHash
	}
	if u.IsActive != nil {
		updates["is_active"] = *u.IsActive
	}
	return updates
}

// IsEmpty 检查是否没有任何更新字段
func (u UserUpdates) IsEmpty() bool {
	return len(u.ToMap()) == 0
}

// ProviderUpdates 提供商更新字段
type ProviderUpdates struct {
	Name        *string
	Driver      *string
	Description *string
	APIKey      *string
	BaseURL     *string
	Config      *JSONMap
	IsActive    *bool
}

// ToMap 转换为 GORM 更新 map（内部使用）
func (u ProviderUpdates) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})
	if u.Name != nil {
		updates["name"] = *u.Name
	}
	if u.Driver != nil {
		updates["driver"] = *u.Driver
	}
	if u.Description != nil {
		updates["description"] = *u.Description
	}
	if u.APIKey != nil {
		updates["api_key"] = *u.APIKey
	}
	if u.BaseURL != nil {
		updates["base_url"] = *u.BaseURL
	}
	if u.Config != nil {
		updates["config"] = *u.Config
	}
	if u.IsActive != nil {
		updates["is_active"] = *u.IsActive
	}
	return updates
}

// IsEmpty 检查是否没有任何更新字段
func (u ProviderUpdates) IsEmpty() bool {
	return len(u.ToMap()) == 0
}

// ModelUpdates 模型更新字段
type ModelUpdates struct {
	Name               *string
	Description        *string
	Price              *string
	MaxImages          *int
	InputModalities    *StringArray
	OutputModalities   *StringArray
	SupportedSizes     *StringArray
	SupportedDurations *IntArray
	DefaultSize        *string
	DefaultDuration    *int
	Settings           *JSONMap
	// 模型能力新字段
	GenerationMode    *string
	EndpointPath      *string
	SupportsStreaming *bool
	SupportsCancel    *bool
	IsActive          *bool
}

// ToMap 转换为 GORM 更新 map（内部使用）
func (u ModelUpdates) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})
	if u.Name != nil {
		updates["name"] = *u.Name
	}
	if u.Description != nil {
		updates["description"] = *u.Description
	}
	if u.Price != nil {
		updates["price"] = *u.Price
	}
	if u.MaxImages != nil {
		updates["max_images"] = *u.MaxImages
	}
	if u.InputModalities != nil {
		updates["input_modalities"] = *u.InputModalities
	}
	if u.OutputModalities != nil {
		updates["output_modalities"] = *u.OutputModalities
	}
	if u.SupportedSizes != nil {
		updates["supported_sizes"] = *u.SupportedSizes
	}
	if u.SupportedDurations != nil {
		updates["supported_durations"] = *u.SupportedDurations
	}
	if u.DefaultSize != nil {
		updates["default_size"] = *u.DefaultSize
	}
	if u.DefaultDuration != nil {
		updates["default_duration"] = *u.DefaultDuration
	}
	if u.Settings != nil {
		updates["settings"] = *u.Settings
	}
	if u.GenerationMode != nil {
		updates["generation_mode"] = *u.GenerationMode
	}
	if u.EndpointPath != nil {
		updates["endpoint_path"] = *u.EndpointPath
	}
	if u.SupportsStreaming != nil {
		updates["supports_streaming"] = *u.SupportsStreaming
	}
	if u.SupportsCancel != nil {
		updates["supports_cancel"] = *u.SupportsCancel
	}
	if u.IsActive != nil {
		updates["is_active"] = *u.IsActive
	}
	return updates
}

// IsEmpty 检查是否没有任何更新字段
func (u ModelUpdates) IsEmpty() bool {
	return len(u.ToMap()) == 0
}

// UsageRecordUpdates 使用记录更新字段
type UsageRecordUpdates struct {
	InputImages  *StringArray
	OutputImages *StringArray
	OutputText   *string
	ErrorMessage *string
	TaskID       *string
	RequestID    *string
}

// ToMap 转换为 GORM 更新 map（内部使用）
func (u UsageRecordUpdates) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})
	if u.InputImages != nil {
		updates["input_images"] = *u.InputImages
	}
	if u.OutputImages != nil {
		updates["output_images"] = *u.OutputImages
	}
	if u.OutputText != nil {
		updates["output_text"] = *u.OutputText
	}
	if u.ErrorMessage != nil {
		updates["error_message"] = *u.ErrorMessage
	}
	if u.TaskID != nil {
		updates["external_task_code"] = *u.TaskID
	}
	if u.RequestID != nil {
		updates["request_id"] = *u.RequestID
	}
	return updates
}

// IsEmpty 检查是否没有任何更新字段
func (u UsageRecordUpdates) IsEmpty() bool {
	return len(u.ToMap()) == 0
}

// TagUpdates 标签更新字段
type TagUpdates struct {
	Name *string
}

// ToMap 转换为 GORM 更新 map（内部使用）
func (u TagUpdates) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})
	if u.Name != nil {
		updates["name"] = *u.Name
	}
	return updates
}

// IsEmpty 检查是否没有任何更新字段
func (u TagUpdates) IsEmpty() bool {
	return len(u.ToMap()) == 0
}
