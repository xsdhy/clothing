package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// StringArray stores a slice of strings in JSON format.
type StringArray []string

// Value implements the driver.Valuer interface.
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	raw, err := json.Marshal([]string(a))
	if err != nil {
		return nil, err
	}
	return string(raw), nil
}

// Scan implements the sql.Scanner interface.
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*a = []string{}
			return nil
		}
		return json.Unmarshal(v, (*[]string)(a))
	case string:
		if v == "" {
			*a = []string{}
			return nil
		}
		return json.Unmarshal([]byte(v), (*[]string)(a))
	default:
		return fmt.Errorf("unsupported type for StringArray: %T", value)
	}
}

// ToSlice returns a copy of the underlying slice.
func (a StringArray) ToSlice() []string {
	if len(a) == 0 {
		return []string{}
	}
	out := make([]string, len(a))
	copy(out, a)
	return out
}

// JSONMap stores a map as JSON text.
type JSONMap map[string]interface{}

// Value implements driver.Valuer for JSONMap.
func (m JSONMap) Value() (driver.Value, error) {
	if len(m) == 0 {
		return "{}", nil
	}
	raw, err := json.Marshal(map[string]interface{}(m))
	if err != nil {
		return nil, err
	}
	return string(raw), nil
}

// Scan implements sql.Scanner for JSONMap.
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*m = JSONMap{}
			return nil
		}
		return json.Unmarshal(v, (*map[string]interface{})(m))
	case string:
		if v == "" {
			*m = JSONMap{}
			return nil
		}
		return json.Unmarshal([]byte(v), (*map[string]interface{})(m))
	default:
		return fmt.Errorf("unsupported type for JSONMap: %T", value)
	}
}

// DbUsageRecord stores a generation usage record.
type DbUsageRecord struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserID uint    `gorm:"column:user_id;index" json:"user_id"`
	User   *DbUser `gorm:"foreignKey:UserID" json:"-"`

	ProviderID string `gorm:"column:provider_id;type:varchar(255);index" json:"provider_id"`
	ModelID    string `gorm:"column:model_id;type:varchar(255);index" json:"model_id"`
	Prompt     string `gorm:"column:prompt;type:text" json:"prompt"`
	Size       string `gorm:"column:size;type:varchar(64)" json:"size"`

	InputImages  StringArray `gorm:"column:input_images;type:json" json:"input_images"`
	OutputImages StringArray `gorm:"column:output_images;type:json" json:"output_images"`

	OutputText   string `gorm:"column:output_text;type:text" json:"output_text"`
	ErrorMessage string `gorm:"column:error_message;type:text" json:"error_message"`

	Tags []DbTag `gorm:"many2many:usage_record_tags;foreignKey:ID;joinForeignKey:UsageRecordID;references:ID;joinReferences:TagID" json:"tags"`
}

// TableName 指定表名
func (DbUsageRecord) TableName() string {
	return "usage_records"
}

// DbTag represents a user-defined tag.
type DbTag struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name        string `gorm:"size:64;uniqueIndex;not null" json:"name"`
	UsageCount  int64  `gorm:"-" json:"usage_count,omitempty"`
	Description string `gorm:"-" json:"description,omitempty"`
}

// TableName 指定表名
func (DbTag) TableName() string {
	return "tags"
}

// DbUsageRecordTag is the join table for usage records and tags.
type DbUsageRecordTag struct {
	UsageRecordID uint      `gorm:"primaryKey" json:"usage_record_id"`
	TagID         uint      `gorm:"primaryKey" json:"tag_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// TableName 指定表名
func (DbUsageRecordTag) TableName() string {
	return "usage_record_tags"
}
