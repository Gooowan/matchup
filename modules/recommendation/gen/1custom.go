package gen

import (
	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
)

// ProfileDTO returns the full profile with JSONB data merged into a flat structure
type ProfileDTO struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id"`
	DanceStyles []string    `json:"dance_styles"`
	Latitude    float64     `json:"latitude"`
	Longitude   float64     `json:"longitude"`
	Visible     bool        `json:"visible"`
	Data        types.JSONB `json:"data"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
}

func (p Profile) ToDTO() ProfileDTO {
	return ProfileDTO{
		ID:          utils.UUIDToString(p.ID),
		UserID:      utils.UUIDToString(p.UserID),
		DanceStyles: p.DanceStyles,
		Latitude:    p.Latitude.Float64,
		Longitude:   p.Longitude.Float64,
		Visible:     p.Visible,
		Data:        p.Data,
		CreatedAt:   p.CreatedAt.Time.UnixMilli(),
		UpdatedAt:   p.UpdatedAt.Time.UnixMilli(),
	}
}

// PreferencesDTO returns preferences with JSONB data
type PreferencesDTO struct {
	ID     string      `json:"id"`
	UserID string      `json:"user_id"`
	Data   types.JSONB `json:"data"`
}

func (p UserPreference) ToDTO() PreferencesDTO {
	return PreferencesDTO{
		ID:     utils.UUIDToString(p.ID),
		UserID: utils.UUIDToString(p.UserID),
		Data:   p.Data,
	}
}

// ProfilePreviewDTO for viewing another user's profile
type ProfilePreviewDTO struct {
	UserID      string      `json:"user_id"`
	DanceStyles []string    `json:"dance_styles"`
	Data        types.JSONB `json:"data"`
	ProfileData types.JSONB `json:"profile_data"`
}

func (p GetProfilePreviewRow) ToDTO() ProfilePreviewDTO {
	return ProfilePreviewDTO{
		UserID:      utils.UUIDToString(p.UserID),
		DanceStyles: p.DanceStyles,
		Data:        p.Data,
		ProfileData: p.ProfileData,
	}
}

// FeedCandidateDTO for feed results
type FeedCandidateDTO struct {
	UserID      string      `json:"user_id"`
	DanceStyles []string    `json:"dance_styles"`
	Data        types.JSONB `json:"data"`
	ProfileData types.JSONB `json:"profile_data"`
	DistanceKm  float64     `json:"distance_km"`
}

func (r FindNearbyVisibleProfilesRow) ToFeedDTO() FeedCandidateDTO {
	return FeedCandidateDTO{
		UserID:      utils.UUIDToString(r.UserID),
		DanceStyles: r.DanceStyles,
		Data:        r.Data,
		ProfileData: r.ProfileData,
		DistanceKm:  r.DistanceKm,
	}
}
