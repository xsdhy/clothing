package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// demo表
type DbDemo struct {
	ID        uint  `gorm:"primarykey" json:"id"`
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`

	UUID string `gorm:"column:uuid;type:varchar(255);uniqueIndex" json:"uuid"`

	Phone             string `gorm:"column:phone;type:varchar(255);index" json:"phone"`
	Nickname          string `gorm:"column:nickname;type:varchar(255)" json:"nickname"`
	LastHeartbeatTime int64  `gorm:"column:last_heartbeat_time;type:int(11);default:0" json:"last_heartbeat_time,omitempty"` //上一次心跳时间，时间戳
}

// TableName 指定表名
func (DbDemo) TableName() string {
	return "demo"
}

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

// DbUsageRecord stores a generation usage record.
type DbUsageRecord struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ProviderID string `gorm:"column:provider_id;type:varchar(255);index" json:"provider_id"`
	ModelID    string `gorm:"column:model_id;type:varchar(255);index" json:"model_id"`
	Prompt     string `gorm:"column:prompt;type:text" json:"prompt"`
	Size       string `gorm:"column:size;type:varchar(64)" json:"size"`

	InputImages  StringArray `gorm:"column:input_images;type:json" json:"input_images"`
	OutputImages StringArray `gorm:"column:output_images;type:json" json:"output_images"`

	OutputText   string `gorm:"column:output_text;type:text" json:"output_text"`
	ErrorMessage string `gorm:"column:error_message;type:text" json:"error_message"`
}

// TableName 指定表名
func (DbUsageRecord) TableName() string {
	return "usage_records"
}
