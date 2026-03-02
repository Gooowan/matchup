package gen

import (
	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

type ProfileDTO struct {
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
	dto := ProfileDTO{
		UserID:      utils.UUIDToString(p.UserID),
		DanceStyles: p.DanceStyles,
		DanceRole:   p.DanceRole.String,
		DanceLevel:  p.DanceLevel.String,
		HeightCm:    p.HeightCm.Int32,
		Bio:         p.Bio.String,
		Gender:      p.Gender.String,
		City:        p.City.String,
		Visible:     p.Visible,
		MediaUrls:   p.MediaUrls,
		CreatedAt:   p.CreatedAt.Time.UnixMilli(),
		UpdatedAt:   p.UpdatedAt.Time.UnixMilli(),
	}
	if p.BirthDate.Valid {
		dto.BirthDate = p.BirthDate.Time.Format("2006-01-02")
	}
	if p.Latitude.Valid {
		dto.Latitude = p.Latitude.Float64
	}
	if p.Longitude.Valid {
		dto.Longitude = p.Longitude.Float64
	}
	if dto.DanceStyles == nil {
		dto.DanceStyles = []string{}
	}
	if dto.MediaUrls == nil {
		dto.MediaUrls = []string{}
	}
	return dto
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
	dto := ProfilePreviewDTO{
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
	if dto.DanceStyles == nil {
		dto.DanceStyles = []string{}
	}
	if dto.MediaUrls == nil {
		dto.MediaUrls = []string{}
	}
	return dto
}

type PreferencesDTO struct {
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
	dto := PreferencesDTO{
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
	if dto.PreferredStyles == nil {
		dto.PreferredStyles = []string{}
	}
	return dto
}

type FeedCandidateDTO struct {
	UserID      string      `json:"user_id"`
	DanceStyles []string    `json:"dance_styles"`
	DanceRole   string      `json:"dance_role"`
	DanceLevel  string      `json:"dance_level"`
	HeightCm    int32       `json:"height_cm"`
	Bio         string      `json:"bio"`
	BirthDate   string      `json:"birth_date"`
	Gender      string      `json:"gender"`
	City        string      `json:"city"`
	MediaUrls   []string    `json:"media_urls"`
	ProfileData types.JSONB `json:"profile_data"`
	DistanceKm  float64     `json:"distance_km"`
}

func (r FindNearbyVisibleProfilesRow) ToFeedDTO() FeedCandidateDTO {
	dto := FeedCandidateDTO{
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
	if r.BirthDate.Valid {
		dto.BirthDate = r.BirthDate.Time.Format("2006-01-02")
	}
	if dto.DanceStyles == nil {
		dto.DanceStyles = []string{}
	}
	if dto.MediaUrls == nil {
		dto.MediaUrls = []string{}
	}
	return dto
}

type ChatDTO struct {
	ID          string      `json:"id"`
	OtherUserID string      `json:"other_user_id"`
	CreatedAt   int64       `json:"created_at"`
	LastMessage *MessageDTO `json:"last_message,omitempty"`
}

type MessageDTO struct {
	ID        string `json:"id"`
	ChatID    string `json:"chat_id"`
	SenderID  string `json:"sender_id"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

func (m Message) ToDTO() MessageDTO {
	return MessageDTO{
		ID:        utils.UUIDToString(m.ID),
		ChatID:    utils.UUIDToString(m.ChatID),
		SenderID:  utils.UUIDToString(m.SenderID),
		Type:      m.Type,
		Content:   m.Content,
		CreatedAt: m.CreatedAt.Time.UnixMilli(),
	}
}

type LocationDTO struct {
	UserID    string  `json:"user_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	UpdatedAt int64   `json:"updated_at"`
}

func (l UserLocation) ToDTO() LocationDTO {
	return LocationDTO{
		UserID:    utils.UUIDToString(l.UserID),
		Latitude:  l.Latitude,
		Longitude: l.Longitude,
		UpdatedAt: l.UpdatedAt.Time.UnixMilli(),
	}
}

type NearbyUserDTO struct {
	UserID     string  `json:"user_id"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	DistanceKm float64 `json:"distance_km"`
	UpdatedAt  int64   `json:"updated_at"`
}

// Helper to convert OtherUserID from interface{} to pgtype.UUID
func ParseOtherUserID(val interface{}) pgtype.UUID {
	if b, ok := val.([16]byte); ok {
		return pgtype.UUID{Bytes: b, Valid: true}
	}
	return pgtype.UUID{}
}
