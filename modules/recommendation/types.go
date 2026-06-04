package recommendation

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

// Candidate is a profile candidate returned by any recommendation tier.
type Candidate struct {
	UserID    pgtype.UUID
	Latitude  float64
	Longitude float64
	// Source tracks which tier/circle produced this candidate.
	// Values: "same_club", "nearby_club", "country_wide", "proximity", "random"
	Source    string
	DistKm    float64
}

// UserClub holds minimal club info needed by the recommendation engine.
type UserClub struct {
	ID        pgtype.UUID
	Latitude  float64
	Longitude float64
}

// FilterParams mirrors user_preferences columns for SQL-level filtering.
type FilterParams struct {
	PreferredGender         *string
	AgeMin                  *int16
	AgeMax                  *int16
	HeightMin               *int16
	HeightMax               *int16
	PreferredGoal           *string
	PreferredProgram        *string
	PreferredCategories     []string
	PreferredCountry        *string
	PreferredCity           *string
	WantsPartnerToRelocate  *bool
	WantsPartnerToFinance   *string
}

// MyProfileData holds the current user's own profile attributes needed for
// mutual-match filtering (checking if MY profile passes a candidate's prefs).
type MyProfileData struct {
	Gender     string
	BirthDate  pgtype.Date
	HeightCm   pgtype.Int2
	Goal       string
	Program    string
	Categories []string
	City       string
	Country    string
}

// FeedParams is the input to every CandidateProvider.
type FeedParams struct {
	UserID     pgtype.UUID
	UserClubs  []UserClub
	Country    string
	Latitude   float64
	Longitude  float64
	Filters    FilterParams
	ExcludeIDs []pgtype.UUID
	Limit      int32
	// MyProfile holds the current user's profile data for mutual-match (Tier3).
	MyProfile  *MyProfileData
}

// CandidateProvider is the interface every recommendation tier implements.
type CandidateProvider interface {
	GetCandidates(ctx context.Context, params FeedParams) ([]Candidate, error)
}
