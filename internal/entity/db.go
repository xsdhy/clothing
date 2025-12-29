package entity

// Re-export common types from the common package for backward compatibility.

import (
	"clothing/internal/entity/common"
)

// Type aliases for common types
type StringArray = common.StringArray
type IntArray = common.IntArray
type JSONMap = common.JSONMap
type Response = common.Response
type ResponseItems = common.ResponseItems
type Meta = common.Meta
type BaseParams = common.BaseParams
type Modality = common.Modality

// Constants
const (
	ModText  = common.ModText
	ModImage = common.ModImage
	ModVideo = common.ModVideo
)
