package feed

import (
	"context"
	"math/rand"
	"slices"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Gooowan/matchup/modules/core/types"
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
	if params.Prefs != nil {
		if v, ok := jsonbFloat(params.Prefs.Data, "max_distance_km"); ok && v > 0 {
			maxDist = v
		}
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
		if params.Prefs != nil && !matchesPreferences(c, params.Prefs.Data) {
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

func matchesPreferences(c recgen.FindNearbyVisibleProfilesRow, prefs types.JSONB) bool {
	candidateData := getJSONB(c.Data)

	if prefRole, ok := jsonbString(prefs, "preferred_role"); ok && prefRole != "" {
		if candidateRole, ok := jsonbString(candidateData, "dance_role"); ok {
			if !roleCompatible(prefRole, candidateRole) {
				return false
			}
		}
	}

	if prefStyles := jsonbStringSlice(prefs, "preferred_styles"); len(prefStyles) > 0 {
		if len(c.DanceStyles) > 0 && !stylesOverlap(prefStyles, c.DanceStyles) {
			return false
		}
	}

	if candidateLevel, ok := jsonbString(candidateData, "dance_level"); ok {
		minLevel, _ := jsonbString(prefs, "min_level")
		maxLevel, _ := jsonbString(prefs, "max_level")
		if !levelInRange(candidateLevel, minLevel, maxLevel) {
			return false
		}
	}

	if heightCm, ok := jsonbFloat(candidateData, "height_cm"); ok {
		if minH, ok := jsonbFloat(prefs, "min_height_cm"); ok && heightCm < minH {
			return false
		}
		if maxH, ok := jsonbFloat(prefs, "max_height_cm"); ok && heightCm > maxH {
			return false
		}
	}

	if birthDateStr, ok := jsonbString(candidateData, "birth_date"); ok && birthDateStr != "" {
		if birthDate, err := time.Parse("2006-01-02", birthDateStr); err == nil {
			age := computeAge(birthDate)
			if minAge, ok := jsonbFloat(prefs, "min_age"); ok && age < int(minAge) {
				return false
			}
			if maxAge, ok := jsonbFloat(prefs, "max_age"); ok && age > int(maxAge) {
				return false
			}
		}
	}

	if genderPref, ok := jsonbString(prefs, "gender_preference"); ok && genderPref != "" {
		if candidateGender, ok := jsonbString(candidateData, "gender"); ok {
			if candidateGender != genderPref {
				return false
			}
		}
	}

	return true
}

func roleCompatible(preferred, candidate string) bool {
	if preferred == "both" || candidate == "both" {
		return true
	}
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

func levelInRange(level, min, max string) bool {
	ord, ok := levelOrder[level]
	if !ok {
		return true
	}
	if min != "" {
		if minOrd, ok := levelOrder[min]; ok && ord < minOrd {
			return false
		}
	}
	if max != "" {
		if maxOrd, ok := levelOrder[max]; ok && ord > maxOrd {
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

// JSONB helper functions

func getJSONB(v interface{}) types.JSONB {
	if m, ok := v.(types.JSONB); ok {
		return m
	}
	if m, ok := v.(map[string]any); ok {
		return types.JSONB(m)
	}
	return types.JSONB{}
}

func jsonbString(data types.JSONB, key string) (string, bool) {
	v, ok := data[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func jsonbFloat(data types.JSONB, key string) (float64, bool) {
	v, ok := data[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

func jsonbStringSlice(data types.JSONB, key string) []string {
	v, ok := data[key]
	if !ok {
		return nil
	}
	if arr, ok := v.([]any); ok {
		var result []string
		for _, item := range arr {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	if arr, ok := v.([]string); ok {
		return arr
	}
	return nil
}
