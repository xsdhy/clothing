package db

import "time"

// Tag 表示用户定义的标签。
type Tag struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name        string `gorm:"size:64;uniqueIndex;not null" json:"name"`
	UsageCount  int64  `gorm:"-" json:"usage_count,omitempty"`
	Description string `gorm:"-" json:"description,omitempty"`
}

// TableName 指定表名
func (Tag) TableName() string {
	return "tags"
}

// UsageRecordTag 使用记录与标签的关联表。
type UsageRecordTag struct {
	UsageRecordID uint      `gorm:"primaryKey" json:"usage_record_id"`
	TagID         uint      `gorm:"primaryKey" json:"tag_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// TableName 指定表名
func (UsageRecordTag) TableName() string {
	return "usage_record_tags"
}
