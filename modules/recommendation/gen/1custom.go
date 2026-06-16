package gen

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
)

// normalizeMinioURL rewrites legacy http://localhost:9000 or http://minio:9000
// URLs to the current MINIO_PUBLIC_ENDPOINT, matching the pattern used for avatars.
func normalizeMinioURL(url string) string {
	if url == "" {
		return url
	}
	pub := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	if pub == "" || strings.HasPrefix(url, pub) {
		return url
	}
	for _, legacy := range []string{"http://localhost:9000", "http://minio:9000"} {
		if strings.HasPrefix(url, legacy) {
			return pub + url[len(legacy):]
		}
	}
	return url
}

// normalizeMetadataURLs rewrites legacy MinIO URLs in-place inside a metadata
// JSONB map. Touches "media_urls" ([]string) and "avatar" (string).
func normalizeMetadataURLs(meta types.JSONB) types.JSONB {
	if meta == nil {
		return meta
	}
	changed := false

	if raw, ok := meta["media_urls"]; ok {
		b, _ := json.Marshal(raw)
		var urls []string
		if json.Unmarshal(b, &urls) == nil {
			normalized := make([]string, len(urls))
			for i, u := range urls {
				normalized[i] = normalizeMinioURL(u)
			}
			// Only copy the map if something actually changed.
			for i, u := range urls {
				if normalized[i] != u {
					changed = true
					break
				}
			}
			if changed {
				out := make(types.JSONB, len(meta))
				for k, v := range meta {
					out[k] = v
				}
				out["media_urls"] = normalized
				return out
			}
		}
	}

	if avatar, ok := meta["avatar"].(string); ok && avatar != "" {
		if n := normalizeMinioURL(avatar); n != avatar {
			if !changed {
				out := make(types.JSONB, len(meta))
				for k, v := range meta {
					out[k] = v
				}
				out["avatar"] = n
				return out
			}
		}
	}

	return meta
}

// ProfileToFeatures converts a Profile into a JSONB feature snapshot for
// recommendation_likes_log. Only stores filterable fields needed by Tier 2/3.
func ProfileToFeatures(p Profile) types.JSONB {
	f := types.JSONB{
		"categories": p.Categories,
		"goal":       p.Goal,
		"program":    p.Program,
		"gender":     p.Gender,
	}
	if p.BirthDate.Valid {
		f["birth_date"] = p.BirthDate.Time.Format("2006-01-02")
	}
	if p.HeightCm.Valid {
		f["height_cm"] = p.HeightCm.Int16
	}
	if p.Country.Valid {
		f["country"] = p.Country.String
	}
	if p.City.Valid {
		f["city"] = p.City.String
	}
	return f
}

// ProfileDTO returns the full profile with all filter columns as flat fields.
type ProfileDTO struct {
	ID              string      `json:"id"`
	UserID          string      `json:"user_id"`
	AccountType     string      `json:"account_type"`
	DanceStyles     []string    `json:"dance_styles"`
	Latitude        float64     `json:"latitude"`
	Longitude       float64     `json:"longitude"`
	Visible         bool        `json:"visible"`
	Gender          string      `json:"gender"`
	BirthDate       string      `json:"birth_date,omitempty"`
	HeightCm        *int16      `json:"height_cm,omitempty"`
	Goal            string      `json:"goal"`
	Program         string      `json:"program"`
	Categories      []string    `json:"categories"`
	Country         string      `json:"country,omitempty"`
	City            string      `json:"city,omitempty"`
	PrimaryClubID   *string     `json:"primary_club_id,omitempty"`
	ReadyToRelocate bool        `json:"ready_to_relocate"`
	ReadyToFinance  string      `json:"ready_to_finance,omitempty"`
	Metadata        types.JSONB `json:"metadata"`
	CreatedAt       int64       `json:"created_at"`
	UpdatedAt       int64       `json:"updated_at"`
}

func (p Profile) ToDTO() ProfileDTO {
	dto := ProfileDTO{
		ID:          utils.UUIDToString(p.ID),
		UserID:      utils.UUIDToString(p.UserID),
		AccountType: p.AccountType,
		DanceStyles: p.DanceStyles,
		Latitude:    p.Latitude.Float64,
		Longitude:   p.Longitude.Float64,
		Visible:     p.Visible,
		Gender:      p.Gender,
		Goal:        p.Goal,
		Program:     p.Program,
		Categories:  p.Categories,
		Metadata:    normalizeMetadataURLs(p.Metadata),
		CreatedAt:   p.CreatedAt.Time.UnixMilli(),
		UpdatedAt:   p.UpdatedAt.Time.UnixMilli(),
	}
	if p.BirthDate.Valid {
		t := p.BirthDate.Time
		dto.BirthDate = t.Format("2006-01-02")
	}
	if p.HeightCm.Valid {
		v := p.HeightCm.Int16
		dto.HeightCm = &v
	}
	if p.Country.Valid {
		dto.Country = p.Country.String
	}
	if p.City.Valid {
		dto.City = p.City.String
	}
	if p.ReadyToRelocate.Valid {
		dto.ReadyToRelocate = p.ReadyToRelocate.Bool
	}
	if p.ReadyToFinance.Valid {
		dto.ReadyToFinance = p.ReadyToFinance.String
	}
	if p.PrimaryClubID.Valid {
		s := utils.UUIDToString(p.PrimaryClubID)
		dto.PrimaryClubID = &s
	}
	return dto
}

// PreferencesDTO returns preferences as flat typed fields.
type PreferencesDTO struct {
	ID                     string      `json:"id"`
	UserID                 string      `json:"user_id"`
	PreferredGender        string      `json:"preferred_gender,omitempty"`
	AgeMin                 *int16      `json:"age_min,omitempty"`
	AgeMax                 *int16      `json:"age_max,omitempty"`
	HeightMin              *int16      `json:"height_min,omitempty"`
	HeightMax              *int16      `json:"height_max,omitempty"`
	PreferredGoal          string      `json:"preferred_goal,omitempty"`
	PreferredProgram       string      `json:"preferred_program,omitempty"`
	PreferredCategories    []string    `json:"preferred_categories"`
	PreferredCountry       string      `json:"preferred_country,omitempty"`
	PreferredCity          string      `json:"preferred_city,omitempty"`
	WantsPartnerToRelocate *bool       `json:"wants_partner_to_relocate,omitempty"`
	WantsPartnerToFinance  string      `json:"wants_partner_to_finance,omitempty"`
	Metadata               types.JSONB `json:"metadata"`
}

func (p UserPreference) ToDTO() PreferencesDTO {
	dto := PreferencesDTO{
		ID:                  utils.UUIDToString(p.ID),
		UserID:              utils.UUIDToString(p.UserID),
		PreferredCategories: p.PreferredCategories,
		Metadata:            p.Metadata,
	}
	if p.PreferredGender.Valid {
		dto.PreferredGender = p.PreferredGender.String
	}
	if p.AgeMin.Valid {
		v := p.AgeMin.Int16
		dto.AgeMin = &v
	}
	if p.AgeMax.Valid {
		v := p.AgeMax.Int16
		dto.AgeMax = &v
	}
	if p.HeightMin.Valid {
		v := p.HeightMin.Int16
		dto.HeightMin = &v
	}
	if p.HeightMax.Valid {
		v := p.HeightMax.Int16
		dto.HeightMax = &v
	}
	if p.PreferredGoal.Valid {
		dto.PreferredGoal = p.PreferredGoal.String
	}
	if p.PreferredProgram.Valid {
		dto.PreferredProgram = p.PreferredProgram.String
	}
	if p.PreferredCountry.Valid {
		dto.PreferredCountry = p.PreferredCountry.String
	}
	if p.PreferredCity.Valid {
		dto.PreferredCity = p.PreferredCity.String
	}
	if p.WantsPartnerToRelocate.Valid {
		v := p.WantsPartnerToRelocate.Bool
		dto.WantsPartnerToRelocate = &v
	}
	if p.WantsPartnerToFinance.Valid {
		dto.WantsPartnerToFinance = p.WantsPartnerToFinance.String
	}
	return dto
}

// ProfilePreviewDTO for viewing another user's profile.
type ProfilePreviewDTO struct {
	UserID         string      `json:"user_id"`
	DanceStyles    []string    `json:"dance_styles"`
	Visible        bool        `json:"visible"`
	Gender         string      `json:"gender"`
	BirthDate      string      `json:"birth_date,omitempty"`
	HeightCm       *int16      `json:"height_cm,omitempty"`
	Goal           string      `json:"goal"`
	Program        string      `json:"program"`
	Categories     []string    `json:"categories"`
	Country        string      `json:"country,omitempty"`
	City           string      `json:"city,omitempty"`
	PrimaryClubID  *string     `json:"primary_club_id,omitempty"`
	ClubName       *string     `json:"club_name,omitempty"`
	Metadata       types.JSONB `json:"metadata"`
	ProfileData    types.JSONB `json:"profile_data"`
}

func (p GetProfilePreviewRow) ToDTO() ProfilePreviewDTO {
	dto := ProfilePreviewDTO{
		UserID:      utils.UUIDToString(p.UserID),
		DanceStyles: p.DanceStyles,
		Visible:     p.Visible,
		Gender:      p.Gender,
		Goal:        p.Goal,
		Program:     p.Program,
		Categories:  p.Categories,
		Metadata:    normalizeMetadataURLs(p.Metadata),
		ProfileData: p.ProfileData,
	}
	if p.BirthDate.Valid {
		dto.BirthDate = p.BirthDate.Time.Format("2006-01-02")
	}
	if p.HeightCm.Valid {
		v := p.HeightCm.Int16
		dto.HeightCm = &v
	}
	if p.Country.Valid {
		dto.Country = p.Country.String
	}
	if p.City.Valid {
		dto.City = p.City.String
	}
	if p.PrimaryClubID.Valid {
		s := utils.UUIDToString(p.PrimaryClubID)
		dto.PrimaryClubID = &s
	}
	if p.ClubName.Valid {
		dto.ClubName = &p.ClubName.String
	}
	return dto
}

// TrainerCardDTO is the public representation of a trainer profile for the trainers feed tab.
// Only ships what cards render — no full JSONB blobs.
type TrainerCardDTO struct {
	UserID     string   `json:"user_id"`
	FirstName  string   `json:"first_name"`
	LastName   string   `json:"last_name,omitempty"`
	Avatar     string   `json:"avatar,omitempty"`
	Gender     string   `json:"gender"`
	City       string   `json:"city,omitempty"`
	Bio        string   `json:"bio,omitempty"`
	Categories []string `json:"categories"`
}

func (r ListTrainersRow) ToTrainerCardDTO() TrainerCardDTO {
	pd := r.ProfileData
	if pd == nil {
		pd = types.JSONB{}
	}
	meta := r.Metadata
	if meta == nil {
		meta = types.JSONB{}
	}
	firstName, _ := pd["first_name"].(string)
	lastName, _ := pd["last_name"].(string)
	avatar, _ := meta["avatar"].(string)
	// avatar may also be stored in profile_data on some accounts
	if avatar == "" {
		avatar, _ = pd["avatar"].(string)
	}
	city, _ := meta["city"].(string)
	bio, _ := meta["bio"].(string)
	return TrainerCardDTO{
		UserID:     utils.UUIDToString(r.UserID),
		FirstName:  firstName,
		LastName:   lastName,
		Avatar:     avatar,
		Gender:     r.Gender,
		City:       city,
		Bio:        bio,
		Categories: r.Categories,
	}
}

// FeedCandidateDTO for feed results from FindNearbyVisibleProfiles.
type FeedCandidateDTO struct {
	UserID      string      `json:"user_id"`
	DanceStyles []string    `json:"dance_styles"`
	Gender      string      `json:"gender"`
	BirthDate   string      `json:"birth_date,omitempty"`
	HeightCm    *int16      `json:"height_cm,omitempty"`
	Goal        string      `json:"goal"`
	Program     string      `json:"program"`
	Categories  []string    `json:"categories"`
	Country     string      `json:"country,omitempty"`
	City        string      `json:"city,omitempty"`
	Metadata    types.JSONB `json:"metadata"`
	ProfileData types.JSONB `json:"profile_data"`
	DistanceKm  float64     `json:"distance_km"`
	Source      string      `json:"source,omitempty"`
}

func (r FindNearbyVisibleProfilesRow) ToFeedDTO() FeedCandidateDTO {
	dto := FeedCandidateDTO{
		UserID:      utils.UUIDToString(r.UserID),
		DanceStyles: r.DanceStyles,
		Gender:      r.Gender,
		Goal:        r.Goal,
		Program:     r.Program,
		Categories:  r.Categories,
		Metadata:    normalizeMetadataURLs(r.Metadata),
		ProfileData: r.ProfileData,
		DistanceKm:  r.DistanceKm,
	}
	if r.BirthDate.Valid {
		dto.BirthDate = r.BirthDate.Time.Format("2006-01-02")
	}
	if r.HeightCm.Valid {
		v := r.HeightCm.Int16
		dto.HeightCm = &v
	}
	if r.Country.Valid {
		dto.Country = r.Country.String
	}
	if r.City.Valid {
		dto.City = r.City.String
	}
	return dto
}
