package converter

import (
	"clothing/internal/entity/db"
	"clothing/internal/entity/dto"
)

// UserToSummary converts a db.User to dto.UserSummary.
func UserToSummary(u *db.User) dto.UserSummary {
	if u == nil {
		return dto.UserSummary{}
	}
	return dto.UserSummary{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
		IsActive:    u.IsActive,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// UsersToSummaries converts a slice of db.User to dto.UserSummary.
func UsersToSummaries(users []db.User) []dto.UserSummary {
	summaries := make([]dto.UserSummary, len(users))
	for i, u := range users {
		summaries[i] = UserToSummary(&u)
	}
	return summaries
}
