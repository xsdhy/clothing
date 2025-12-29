package entity

// 重导出转换函数以保持向后兼容性。

import (
	"clothing/internal/entity/converter"
	"clothing/internal/entity/db"
	"clothing/internal/entity/dto"
	"strings"
)

// ProviderToAdminView 将 DbProvider 转换为 ProviderAdminView。
// 使用此函数而非方法以保持向后兼容性。
func ProviderToAdminView(p DbProvider, includeModels bool) ProviderAdminView {
	dbProvider := db.Provider(p)
	return converter.ProviderToAdminView(&dbProvider, includeModels)
}

// ModelToAdminView 将 DbModel 转换为 ProviderModelSummary。
// 使用此函数而非方法以保持向后兼容性。
func ModelToAdminView(m DbModel) ProviderModelSummary {
	dbModel := db.Model(m)
	return converter.ModelToSummary(&dbModel)
}

// toModalities 模态转换辅助函数（保留以兼容）
func toModalities(values StringArray, fallback StringArray) []Modality {
	var modalities []Modality
	source := values
	if len(source) == 0 {
		source = fallback
	}
	for _, item := range source {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			modalities = append(modalities, Modality(trimmed))
		}
	}
	return modalities
}

// UserToSummary 将 DbUser 转换为 UserSummary。
func UserToSummary(u *DbUser) UserSummary {
	dbUser := db.User(*u)
	return converter.UserToSummary(&dbUser)
}

// TagToDTO 将 DbTag 转换为 Tag DTO。
func TagToDTO(t *DbTag) dto.Tag {
	dbTag := db.Tag(*t)
	return converter.TagToDTO(&dbTag)
}
