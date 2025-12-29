package common

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// StringArray 以 JSON 格式存储字符串切片。
type StringArray []string

// Value 实现 driver.Valuer 接口。
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

// Scan 实现 sql.Scanner 接口。
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

// ToSlice 返回底层切片的副本。
func (a StringArray) ToSlice() []string {
	if len(a) == 0 {
		return []string{}
	}
	out := make([]string, len(a))
	copy(out, a)
	return out
}

// Contains 检查数组是否包含给定的字符串。
func (a StringArray) Contains(s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}

// IntArray 以 JSON 格式存储整数切片。
type IntArray []int

// Value 实现 driver.Valuer 接口。
func (a IntArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	raw, err := json.Marshal([]int(a))
	if err != nil {
		return nil, err
	}
	return string(raw), nil
}

// Scan 实现 sql.Scanner 接口。
func (a *IntArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*a = []int{}
			return nil
		}
		return json.Unmarshal(v, (*[]int)(a))
	case string:
		if v == "" {
			*a = []int{}
			return nil
		}
		return json.Unmarshal([]byte(v), (*[]int)(a))
	default:
		return fmt.Errorf("unsupported type for IntArray: %T", value)
	}
}

// ToSlice 返回底层切片的副本。
func (a IntArray) ToSlice() []int {
	if len(a) == 0 {
		return []int{}
	}
	out := make([]int, len(a))
	copy(out, a)
	return out
}

// JSONMap 以 JSON 文本格式存储 map。
type JSONMap map[string]interface{}

// Value 实现 driver.Valuer 接口。
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

// Scan 实现 sql.Scanner 接口。
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

// Response 是标准 API 响应结构。
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
	Time time.Time   `json:"time"`
}

// ResponseItems 是带分页的标准 API 响应。
type ResponseItems struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
	Meta *Meta       `json:"meta"`
	Time time.Time   `json:"time"`
}

// Meta 包含分页元数据。
type Meta struct {
	Page     int64 `json:"page"`
	PageSize int64 `json:"page_size"`
	Total    int64 `json:"total"`
}

// BaseParams 包含通用的分页和排序参数。
type BaseParams struct {
	PageSize int64  `json:"page_size" form:"page_size" query:"page_size"`
	Page     int64  `json:"page" form:"page" query:"page"`
	SortBy   string `json:"sort_by" form:"sort_by" query:"sort_by"`
	SortDesc bool   `json:"sort_desc" form:"sort_desc" query:"sort_desc"`
}

// Modality 表示内容模态类型。
type Modality string

const (
	ModText  Modality = "text"
	ModImage Modality = "image"
	ModVideo Modality = "video"
)
