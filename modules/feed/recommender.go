package feed

import (
	"context"
	"math/rand"
	"slices"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Gooowan/matchup/modules/recommendation"
	recgen "github.com/Gooowan/matchup/modules/recommendation/gen"
)

type FeedParams struct {
	UserID     pgtype.UUID
	Latitude   float64
	Longitude  float64
	Prefs      *recgen.UserPreference
	ExcludeIDs []pgtype.UUID
	Limit      int32
}

type RecommendationProvider interface {
	GetFeed(ctx context.Context, params FeedParams) ([]recgen.FindNearbyVisibleProfilesRow, error)
}

// NearestCandidatesProvider ranks by distance and filters by preferences
type NearestCandidatesProvider struct {
	RecommendationSvc *recommendation.RecommendationService
}

func NewNearestCandidatesProvider(recommendationSvc *recommendation.RecommendationService) *NearestCandidatesProvider {
	return &NearestCandidatesProvider{RecommendationSvc: recommendationSvc}
}

func (p *NearestCandidatesProvider) GetFeed(ctx context.Context, params FeedParams) ([]recgen.FindNearbyVisibleProfilesRow, error) {
	maxDist := 100.0 // default 100km
	if params.Prefs != nil && params.Prefs.MaxDistanceKm.Valid && params.Prefs.MaxDistanceKm.Float64 > 0 {
		maxDist = params.Prefs.MaxDistanceKm.Float64
	}

	// over-fetch to compensate for preference filtering
	fetchLimit := params.Limit * 3
	if fetchLimit < 30 {
		fetchLimit = 30
	}

	candidates, err := p.RecommendationSvc.FindNearbyVisibleProfiles(ctx, recgen.FindNearbyVisibleProfilesParams{
		Latitude:   params.Latitude,
		Longitude:  params.Longitude,
		UserID:     params.UserID,
		ExcludeIds: params.ExcludeIDs,
		LimitVal:   fetchLimit,
	})
	if err != nil {
		return nil, err
	}

	// filter by distance and preferences
	var filtered []recgen.FindNearbyVisibleProfilesRow
	for _, c := range candidates {
		if c.DistanceKm > maxDist {
			continue
		}
		if params.Prefs != nil && !matchesPreferences(c, params.Prefs) {
			continue
		}
		filtered = append(filtered, c)
		if int32(len(filtered)) >= params.Limit {
			break
		}
	}

	return filtered, nil
}

// RandomFallbackProvider returns random visible profiles (used if primary returns empty)
type RandomFallbackProvider struct {
	RecommendationSvc *recommendation.RecommendationService
}

func NewRandomFallbackProvider(recommendationSvc *recommendation.RecommendationService) *RandomFallbackProvider {
	return &RandomFallbackProvider{RecommendationSvc: recommendationSvc}
}

func (p *RandomFallbackProvider) GetFeed(ctx context.Context, params FeedParams) ([]recgen.FindNearbyVisibleProfilesRow, error) {
	// fetch more candidates and shuffle
	candidates, err := p.RecommendationSvc.FindNearbyVisibleProfiles(ctx, recgen.FindNearbyVisibleProfilesParams{
		Latitude:   params.Latitude,
		Longitude:  params.Longitude,
		UserID:     params.UserID,
		ExcludeIds: params.ExcludeIDs,
		LimitVal:   params.Limit * 5,
	})
	if err != nil {
		return nil, err
	}

	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	if int32(len(candidates)) > params.Limit {
		candidates = candidates[:params.Limit]
	}
	return candidates, nil
}

// FallbackProvider wraps a primary and fallback provider
type FallbackProvider struct {
	Primary  RecommendationProvider
	Fallback RecommendationProvider
}

func (p *FallbackProvider) GetFeed(ctx context.Context, params FeedParams) ([]recgen.FindNearbyVisibleProfilesRow, error) {
	result, err := p.Primary.GetFeed(ctx, params)
	if err == nil && len(result) > 0 {
		return result, nil
	}
	return p.Fallback.GetFeed(ctx, params)
}

// preference matching helpers

func matchesPreferences(c recgen.FindNearbyVisibleProfilesRow, prefs *recgen.UserPreference) bool {
	if prefs.PreferredRole.Valid && prefs.PreferredRole.String != "" && c.DanceRole.Valid {
		if !roleCompatible(prefs.PreferredRole.String, c.DanceRole.String) {
			return false
		}
	}

	if len(prefs.PreferredStyles) > 0 && len(c.DanceStyles) > 0 {
		if !stylesOverlap(prefs.PreferredStyles, c.DanceStyles) {
			return false
		}
	}

	if c.DanceLevel.Valid {
		if !levelInRange(c.DanceLevel.String, prefs.MinLevel, prefs.MaxLevel) {
			return false
		}
	}

	if c.HeightCm.Valid {
		if prefs.MinHeightCm.Valid && c.HeightCm.Int32 < prefs.MinHeightCm.Int32 {
			return false
		}
		if prefs.MaxHeightCm.Valid && c.HeightCm.Int32 > prefs.MaxHeightCm.Int32 {
			return false
		}
	}

	if c.BirthDate.Valid && (prefs.MinAge.Valid || prefs.MaxAge.Valid) {
		age := computeAge(c.BirthDate.Time)
		if prefs.MinAge.Valid && age < int(prefs.MinAge.Int32) {
			return false
		}
		if prefs.MaxAge.Valid && age > int(prefs.MaxAge.Int32) {
			return false
		}
	}

	if prefs.GenderPreference.Valid && prefs.GenderPreference.String != "" && c.Gender.Valid {
		if c.Gender.String != prefs.GenderPreference.String {
			return false
		}
	}

	return true
}

func roleCompatible(preferred, candidate string) bool {
	if preferred == "both" || candidate == "both" {
		return true
	}
	// leader wants follower and vice versa
	if preferred == "leader" && candidate == "follower" {
		return true
	}
	if preferred == "follower" && candidate == "leader" {
		return true
	}
	return false
}

func stylesOverlap(a, b []string) bool {
	for _, s := range a {
		if slices.Contains(b, s) {
			return true
		}
	}
	return false
}

var levelOrder = map[string]int{
	"beginner":     1,
	"intermediate": 2,
	"advanced":     3,
	"professional": 4,
}

func levelInRange(level string, min, max pgtype.Text) bool {
	ord, ok := levelOrder[level]
	if !ok {
		return true
	}
	if min.Valid && min.String != "" {
		if minOrd, ok := levelOrder[min.String]; ok && ord < minOrd {
			return false
		}
	}
	if max.Valid && max.String != "" {
		if maxOrd, ok := levelOrder[max.String]; ok && ord > maxOrd {
			return false
		}
	}
	return true
}

func computeAge(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()
	if now.YearDay() < birthDate.YearDay() {
		age--
	}
	return age
}
