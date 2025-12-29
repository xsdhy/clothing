package db

import (
	"clothing/internal/entity/common"
	"time"
)

// UsageRecord stores a generation usage record.
type UsageRecord struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserID uint  `gorm:"column:user_id;index" json:"user_id"`
	User   *User `gorm:"foreignKey:UserID" json:"-"`

	ProviderID string `gorm:"column:provider_id;type:varchar(255);index" json:"provider_id"`
	ModelID    string `gorm:"column:model_id;type:varchar(255);index" json:"model_id"`
	Prompt     string `gorm:"column:prompt;type:text" json:"prompt"`
	Size       string `gorm:"column:size;type:varchar(64)" json:"size"`

	InputImages  common.StringArray `gorm:"column:input_images;type:json" json:"input_images"`
	OutputImages common.StringArray `gorm:"column:output_images;type:json" json:"output_images"`

	OutputText   string `gorm:"column:output_text;type:text" json:"output_text"`
	ErrorMessage string `gorm:"column:error_message;type:text" json:"error_message"`

	ExternalTaskCode string `gorm:"column:external_task_code;type:varchar(255)" json:"external_task_code"` // 外部（第三方）任务code，或者任务ID
	RequestID        string `gorm:"column:request_id;type:varchar(255)" json:"request_id"`                 // 请求ID

	Tags []Tag `gorm:"many2many:usage_record_tags;foreignKey:ID;joinForeignKey:UsageRecordID;references:ID;joinReferences:TagID" json:"tags"`
}

// TableName 指定表名
func (UsageRecord) TableName() string {
	return "usage_records"
}
