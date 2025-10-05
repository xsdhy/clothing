package entity

import "time"

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
	Time time.Time   `json:"time"`
}

type ResponseItems struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
	Meta *Meta       `json:"meta"`
	Time time.Time   `json:"time"`
}

type Meta struct {
	Page     int64 `json:"page"`
	PageSize int64 `json:"page_size"`
	Total    int64 `json:"total"`
}

type BaseParams struct {
	PageSize int64  `json:"page_size" form:"page_size" query:"page_size"`
	Page     int64  `json:"page" form:"page" query:"page"`
	SortBy   string `json:"sort_by" form:"sort_by" query:"sort_by"`
	SortDesc bool   `json:"sort_desc" form:"sort_desc" query:"sort_desc"`
}
