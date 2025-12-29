package converter

import (
	"clothing/internal/entity/db"
	"clothing/internal/entity/dto"
	"strings"
)

// ProviderToAdminView converts db.Provider to dto.ProviderAdminView.
func ProviderToAdminView(p *db.Provider, includeModels bool) dto.ProviderAdminView {
	if p == nil {
		return dto.ProviderAdminView{}
	}

	hasAPIKey := strings.TrimSpace(p.APIKey) != ""
	view := dto.ProviderAdminView{
		ID:        p.ID,
		Name:      p.Name,
		Driver:    p.Driver,
		BaseURL:   p.BaseURL,
		Config:    p.Config,
		IsActive:  p.IsActive,
		HasAPIKey: hasAPIKey,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	if strings.TrimSpace(p.Description) != "" {
		view.Description = p.Description
	}

	if includeModels {
		view.Models = make([]dto.ProviderModelSummary, 0, len(p.Models))
		for _, model := range p.Models {
			view.Models = append(view.Models, ModelToSummary(&model))
		}
	}
	return view
}

// ProvidersToAdminViews converts a slice of db.Provider to dto.ProviderAdminView.
func ProvidersToAdminViews(providers []db.Provider, includeModels bool) []dto.ProviderAdminView {
	views := make([]dto.ProviderAdminView, len(providers))
	for i, p := range providers {
		views[i] = ProviderToAdminView(&p, includeModels)
	}
	return views
}

// ModelToSummary converts db.Model to dto.ProviderModelSummary.
func ModelToSummary(m *db.Model) dto.ProviderModelSummary {
	if m == nil {
		return dto.ProviderModelSummary{}
	}

	inputModalities := m.InputModalities.ToSlice()
	outputModalities := m.OutputModalities.ToSlice()

	// 如果新字段为空，则回退到旧的 Modalities 字段
	if len(inputModalities) == 0 && len(m.Modalities) > 0 {
		inputModalities = m.Modalities.ToSlice()
	}
	if len(outputModalities) == 0 && len(m.Modalities) > 0 {
		outputModalities = m.Modalities.ToSlice()
	}

	sizes := m.SupportedSizes.ToSlice()
	durations := db.MergeDurations(m.SupportedDurations, db.ParseDurations(m.Settings))
	defaultSize := db.ResolveDefaultSize(m.DefaultSize, sizes, m.Settings)
	defaultDuration := db.ResolveDefaultDuration(m.DefaultDuration, durations, m.Settings)

	return dto.ProviderModelSummary{
		ModelID:            m.ModelID,
		Name:               m.Name,
		Description:        m.Description,
		Price:              m.Price,
		MaxImages:          m.MaxImages,
		InputModalities:    inputModalities,
		OutputModalities:   outputModalities,
		SupportedSizes:     sizes,
		SupportedDurations: durations,
		DefaultSize:        defaultSize,
		DefaultDuration:    defaultDuration,
		Settings:           m.Settings,
		GenerationMode:     m.GenerationMode,
		EndpointPath:       m.EndpointPath,
		SupportsStreaming:  m.SupportsStreaming,
		SupportsCancel:     m.SupportsCancel,
		IsActive:           m.IsActive,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}

// ModelsToSummaries converts a slice of db.Model to dto.ProviderModelSummary.
func ModelsToSummaries(models []db.Model) []dto.ProviderModelSummary {
	summaries := make([]dto.ProviderModelSummary, len(models))
	for i, m := range models {
		summaries[i] = ModelToSummary(&m)
	}
	return summaries
}
