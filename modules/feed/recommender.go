package feed

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Gooowan/matchup/modules/recommendation"
	recgen "github.com/Gooowan/matchup/modules/recommendation/gen"
)

// RecommendationProvider is the interface FeedService uses to get feed candidates.
// It returns rows from FindNearbyVisibleProfiles for backward compat with the controller.
type RecommendationProvider interface {
	GetFeed(ctx context.Context, params FeedParams) ([]recgen.FindNearbyVisibleProfilesRow, error)
}

// FeedParams carries everything needed to produce a feed for one user.
type FeedParams struct {
	UserID     pgtype.UUID
	Latitude   float64
	Longitude  float64
	Country    string
	UserClubs  []recommendation.UserClub
	Prefs      *recgen.UserPreference
	ExcludeIDs []pgtype.UUID
	Limit      int32
}

// TierRecommendationProvider delegates to the 3-tier recommendation engine
// and adapts the generic Candidate slice back to FindNearbyVisibleProfilesRow.
type TierRecommendationProvider struct {
	Recommender       *recommendation.Recommender
	RecommendationSvc *recommendation.RecommendationService
}

func NewTierRecommendationProvider(r *recommendation.Recommender, svc *recommendation.RecommendationService) *TierRecommendationProvider {
	return &TierRecommendationProvider{Recommender: r, RecommendationSvc: svc}
}

func (p *TierRecommendationProvider) GetFeed(ctx context.Context, params FeedParams) ([]recgen.FindNearbyVisibleProfilesRow, error) {
	recParams := recommendation.FeedParams{
		UserID:     params.UserID,
		Country:    params.Country,
		Latitude:   params.Latitude,
		Longitude:  params.Longitude,
		UserClubs:  params.UserClubs,
		ExcludeIDs: params.ExcludeIDs,
		Limit:      params.Limit,
	}

	// Map preference columns to FilterParams
	if params.Prefs != nil {
		f := &recParams.Filters
		if params.Prefs.PreferredGender.Valid {
			s := params.Prefs.PreferredGender.String
			f.PreferredGender = &s
		}
		if params.Prefs.AgeMin.Valid {
			v := params.Prefs.AgeMin.Int16
			f.AgeMin = &v
		}
		if params.Prefs.AgeMax.Valid {
			v := params.Prefs.AgeMax.Int16
			f.AgeMax = &v
		}
		if params.Prefs.HeightMin.Valid {
			v := params.Prefs.HeightMin.Int16
			f.HeightMin = &v
		}
		if params.Prefs.HeightMax.Valid {
			v := params.Prefs.HeightMax.Int16
			f.HeightMax = &v
		}
		if params.Prefs.PreferredGoal.Valid {
			s := params.Prefs.PreferredGoal.String
			f.PreferredGoal = &s
		}
		if params.Prefs.PreferredProgram.Valid {
			s := params.Prefs.PreferredProgram.String
			f.PreferredProgram = &s
		}
		if len(params.Prefs.PreferredCategories) > 0 {
			f.PreferredCategories = params.Prefs.PreferredCategories
		}
		if params.Prefs.PreferredCountry.Valid {
			s := params.Prefs.PreferredCountry.String
			f.PreferredCountry = &s
		}
		if params.Prefs.PreferredCity.Valid {
			s := params.Prefs.PreferredCity.String
			f.PreferredCity = &s
		}
		if params.Prefs.WantsPartnerToRelocate.Valid {
			v := params.Prefs.WantsPartnerToRelocate.Bool
			f.WantsPartnerToRelocate = &v
		}
		if params.Prefs.WantsPartnerToFinance.Valid {
			s := params.Prefs.WantsPartnerToFinance.String
			f.WantsPartnerToFinance = &s
		}
	}

	candidates, err := p.Recommender.GetCandidates(ctx, recParams)
	if err != nil {
		return nil, err
	}

	// Fetch full profile rows for each candidate to return enriched data
	rows := make([]recgen.FindNearbyVisibleProfilesRow, 0, len(candidates))
	for _, c := range candidates {
		profile, err := p.RecommendationSvc.Queries.GetProfileByUserID(ctx, c.UserID)
		if err != nil {
			continue
		}
		// GetProfileByUserID queries profiles table only; fetch profile_data from users via GetProfilePreview
		preview, _ := p.RecommendationSvc.Queries.GetProfilePreview(ctx, c.UserID)
		rows = append(rows, recgen.FindNearbyVisibleProfilesRow{
			ID:              profile.ID,
			UserID:          profile.UserID,
			DanceStyles:     profile.DanceStyles,
			Metadata:        profile.Metadata,
			Data:            profile.Data,
			Latitude:        profile.Latitude,
			Longitude:       profile.Longitude,
			Gender:          profile.Gender,
			BirthDate:       profile.BirthDate,
			HeightCm:        profile.HeightCm,
			Goal:            profile.Goal,
			Program:         profile.Program,
			Categories:      profile.Categories,
			Country:         profile.Country,
			City:            profile.City,
			ReadyToRelocate: profile.ReadyToRelocate,
			ReadyToFinance:  profile.ReadyToFinance,
			DistanceKm:      c.DistKm,
			ProfileData:     preview.ProfileData,
		})
	}
	return rows, nil
}
