package gen

import (
	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
)

type ProfileDTO struct {
	ID          string   `json:"id"`
	UserID      string   `json:"user_id"`
	DanceStyles []string `json:"dance_styles"`
	DanceRole   string   `json:"dance_role"`
	DanceLevel  string   `json:"dance_level"`
	HeightCm    int32    `json:"height_cm"`
	Bio         string   `json:"bio"`
	BirthDate   string   `json:"birth_date"`
	Gender      string   `json:"gender"`
	City        string   `json:"city"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	Visible     bool     `json:"visible"`
	MediaUrls   []string `json:"media_urls"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
}

func (p Profile) ToDTO() ProfileDTO {
	return ProfileDTO{
		ID:          utils.UUIDToString(p.ID),
		UserID:      utils.UUIDToString(p.UserID),
		DanceStyles: p.DanceStyles,
		DanceRole:   p.DanceRole.String,
		DanceLevel:  p.DanceLevel.String,
		HeightCm:    p.HeightCm.Int32,
		Bio:         p.Bio.String,
		BirthDate:   p.BirthDate.Time.Format("2006-01-02"),
		Gender:      p.Gender.String,
		City:        p.City.String,
		Latitude:    p.Latitude.Float64,
		Longitude:   p.Longitude.Float64,
		Visible:     p.Visible,
		MediaUrls:   p.MediaUrls,
		CreatedAt:   p.CreatedAt.Time.UnixMilli(),
		UpdatedAt:   p.UpdatedAt.Time.UnixMilli(),
	}
}

type PreferencesDTO struct {
	ID               string   `json:"id"`
	UserID           string   `json:"user_id"`
	PreferredStyles  []string `json:"preferred_styles"`
	PreferredRole    string   `json:"preferred_role"`
	MinLevel         string   `json:"min_level"`
	MaxLevel         string   `json:"max_level"`
	MinHeightCm      int32    `json:"min_height_cm"`
	MaxHeightCm      int32    `json:"max_height_cm"`
	MinAge           int32    `json:"min_age"`
	MaxAge           int32    `json:"max_age"`
	MaxDistanceKm    float64  `json:"max_distance_km"`
	GenderPreference string   `json:"gender_preference"`
}

func (p UserPreference) ToDTO() PreferencesDTO {
	return PreferencesDTO{
		ID:               utils.UUIDToString(p.ID),
		UserID:           utils.UUIDToString(p.UserID),
		PreferredStyles:  p.PreferredStyles,
		PreferredRole:    p.PreferredRole.String,
		MinLevel:         p.MinLevel.String,
		MaxLevel:         p.MaxLevel.String,
		MinHeightCm:      p.MinHeightCm.Int32,
		MaxHeightCm:      p.MaxHeightCm.Int32,
		MinAge:           p.MinAge.Int32,
		MaxAge:           p.MaxAge.Int32,
		MaxDistanceKm:    p.MaxDistanceKm.Float64,
		GenderPreference: p.GenderPreference.String,
	}
}

type ProfilePreviewDTO struct {
	UserID      string      `json:"user_id"`
	DanceStyles []string    `json:"dance_styles"`
	DanceRole   string      `json:"dance_role"`
	DanceLevel  string      `json:"dance_level"`
	HeightCm    int32       `json:"height_cm"`
	Bio         string      `json:"bio"`
	Gender      string      `json:"gender"`
	City        string      `json:"city"`
	MediaUrls   []string    `json:"media_urls"`
	ProfileData types.JSONB `json:"profile_data"`
}

func (p GetProfilePreviewRow) ToDTO() ProfilePreviewDTO {
	return ProfilePreviewDTO{
		UserID:      utils.UUIDToString(p.UserID),
		DanceStyles: p.DanceStyles,
		DanceRole:   p.DanceRole.String,
		DanceLevel:  p.DanceLevel.String,
		HeightCm:    p.HeightCm.Int32,
		Bio:         p.Bio.String,
		Gender:      p.Gender.String,
		City:        p.City.String,
		MediaUrls:   p.MediaUrls,
		ProfileData: p.ProfileData,
	}
}

type FeedCandidateDTO struct {
	UserID      string      `json:"user_id"`
	DanceStyles []string    `json:"dance_styles"`
	DanceRole   string      `json:"dance_role"`
	DanceLevel  string      `json:"dance_level"`
	HeightCm    int32       `json:"height_cm"`
	Bio         string      `json:"bio"`
	Gender      string      `json:"gender"`
	City        string      `json:"city"`
	MediaUrls   []string    `json:"media_urls"`
	ProfileData types.JSONB `json:"profile_data"`
	DistanceKm  float64     `json:"distance_km"`
}

func (r FindNearbyVisibleProfilesRow) ToFeedDTO() FeedCandidateDTO {
	return FeedCandidateDTO{
		UserID:      utils.UUIDToString(r.UserID),
		DanceStyles: r.DanceStyles,
		DanceRole:   r.DanceRole.String,
		DanceLevel:  r.DanceLevel.String,
		HeightCm:    r.HeightCm.Int32,
		Bio:         r.Bio.String,
		Gender:      r.Gender.String,
		City:        r.City.String,
		MediaUrls:   r.MediaUrls,
		ProfileData: r.ProfileData,
		DistanceKm:  r.DistanceKm,
	}
}
