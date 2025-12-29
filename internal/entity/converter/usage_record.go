package converter

import (
	"clothing/internal/entity/db"
	"clothing/internal/entity/dto"
)

// TagToDTO converts db.Tag to dto.Tag.
func TagToDTO(t *db.Tag) dto.Tag {
	if t == nil {
		return dto.Tag{}
	}
	return dto.Tag{
		ID:          t.ID,
		Name:        t.Name,
		UsageCount:  t.UsageCount,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		Description: t.Description,
	}
}

// TagsToDTOs converts a slice of db.Tag to dto.Tag.
func TagsToDTOs(tags []db.Tag) []dto.Tag {
	dtos := make([]dto.Tag, len(tags))
	for i, t := range tags {
		dtos[i] = TagToDTO(&t)
	}
	return dtos
}

// UsageRecordToItem 将 db.UsageRecord 转换为 dto.UsageRecordItem。
// imageURLBuilder 用于将存储路径转换为公开 URL。
func UsageRecordToItem(r *db.UsageRecord, imageURLBuilder func(path string) dto.UsageImage) dto.UsageRecordItem {
	if r == nil {
		return dto.UsageRecordItem{}
	}

	var user dto.UserSummary
	if r.User != nil {
		user = UserToSummary(r.User)
	}

	inputImages := make([]dto.UsageImage, 0, len(r.InputImages))
	for _, path := range r.InputImages {
		inputImages = append(inputImages, imageURLBuilder(path))
	}

	outputImages := make([]dto.UsageImage, 0, len(r.OutputImages))
	for _, path := range r.OutputImages {
		outputImages = append(outputImages, imageURLBuilder(path))
	}

	tags := make([]dto.Tag, 0, len(r.Tags))
	for _, t := range r.Tags {
		tags = append(tags, TagToDTO(&t))
	}

	return dto.UsageRecordItem{
		ID:           r.ID,
		ProviderID:   r.ProviderID,
		ModelID:      r.ModelID,
		Prompt:       r.Prompt,
		Size:         r.Size,
		OutputText:   r.OutputText,
		ErrorMessage: r.ErrorMessage,
		CreatedAt:    r.CreatedAt,
		InputImages:  inputImages,
		OutputImages: outputImages,
		User:         user,
		Tags:         tags,
	}
}

// UsageRecordsToItems converts a slice of db.UsageRecord to dto.UsageRecordItem.
func UsageRecordsToItems(records []db.UsageRecord, imageURLBuilder func(path string) dto.UsageImage) []dto.UsageRecordItem {
	items := make([]dto.UsageRecordItem, len(records))
	for i, r := range records {
		items[i] = UsageRecordToItem(&r, imageURLBuilder)
	}
	return items
}
