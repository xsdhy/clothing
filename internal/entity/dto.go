package entity

// 从 dto 包重导出类型，保持向后兼容性

import (
	"clothing/internal/entity/dto"
)

// 认证相关 DTO
type AuthStatusResponse = dto.AuthStatusResponse
type AuthLoginRequest = dto.AuthLoginRequest
type AuthRegisterRequest = dto.AuthRegisterRequest
type AuthResponse = dto.AuthResponse

// 用户相关 DTO
type UserSummary = dto.UserSummary
type UserQuery = dto.UserQuery
type UserCreateRequest = dto.UserCreateRequest
type UserUpdateRequest = dto.UserUpdateRequest
type UserListResponse = dto.UserListResponse

// 服务商相关 DTO
type CreateProviderRequest = dto.CreateProviderRequest
type UpdateProviderRequest = dto.UpdateProviderRequest
type CreateModelRequest = dto.CreateModelRequest
type UpdateModelRequest = dto.UpdateModelRequest
type ProviderAdminView = dto.ProviderAdminView
type ProviderModelSummary = dto.ProviderModelSummary

// 内容生成相关 DTO
type ContentInputs = dto.ContentInputs
type ContentOptions = dto.ContentOptions
type MediaInput = dto.MediaInput
type OutputConfig = dto.OutputConfig
type MediaOutput = dto.MediaOutput
type GenerateContentRequest = dto.GenerateContentRequest
type GenerateContentResponse = dto.GenerateContentResponse

// 使用记录相关 DTO
type UsageRecordQuery = dto.UsageRecordQuery
type UsageImage = dto.UsageImage
type UsageRecordItem = dto.UsageRecordItem
type UsageRecordListResponse = dto.UsageRecordListResponse
type UsageRecordDetailResponse = dto.UsageRecordDetailResponse

// 标签相关 DTO
type Tag = dto.Tag
type TagListResponse = dto.TagListResponse
type TagDetailResponse = dto.TagDetailResponse
