package recommendation

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	gen "github.com/Gooowan/matchup/modules/recommendation/gen"
)

type RecommendationService struct {
	DB      *pgxpool.Pool
	Queries *gen.Queries
}

func NewRecommendationService(db *pgxpool.Pool) *RecommendationService {
	return &RecommendationService{DB: db, Queries: gen.New(db)}
}

func (s *RecommendationService) CreateProfile(ctx context.Context, userID pgtype.UUID, params gen.CreateProfileParams) (gen.Profile, error) {
	params.UserID = userID
	return s.Queries.CreateProfile(ctx, params)
}

func (s *RecommendationService) GetProfile(ctx context.Context, userID pgtype.UUID) (gen.Profile, error) {
	return s.Queries.GetProfileByUserID(ctx, userID)
}

func (s *RecommendationService) UpdateProfile(ctx context.Context, userID pgtype.UUID, params gen.UpdateProfileParams) error {
	params.UserID = userID
	return s.Queries.UpdateProfile(ctx, params)
}

func (s *RecommendationService) GetProfilePreview(ctx context.Context, userID pgtype.UUID) (gen.GetProfilePreviewRow, error) {
	return s.Queries.GetProfilePreview(ctx, userID)
}

func (s *RecommendationService) GetPreferences(ctx context.Context, userID pgtype.UUID) (gen.UserPreference, error) {
	return s.Queries.GetPreferences(ctx, userID)
}

func (s *RecommendationService) UpsertPreferences(ctx context.Context, userID pgtype.UUID, params gen.UpsertPreferencesParams) (gen.UserPreference, error) {
	params.UserID = userID
	return s.Queries.UpsertPreferences(ctx, params)
}

func (s *RecommendationService) AddMediaURL(ctx context.Context, userID pgtype.UUID, url string) error {
	profile, err := s.Queries.GetProfileByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	urls := append(profile.MediaUrls, url)
	return s.Queries.UpdateProfileMediaURLs(ctx, gen.UpdateProfileMediaURLsParams{
		MediaUrls: urls,
		UserID:    userID,
	})
}

func (s *RecommendationService) RemoveMediaURL(ctx context.Context, userID pgtype.UUID, url string) error {
	profile, err := s.Queries.GetProfileByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	filtered := make([]string, 0, len(profile.MediaUrls))
	for _, u := range profile.MediaUrls {
		if u != url {
			filtered = append(filtered, u)
		}
	}
	return s.Queries.UpdateProfileMediaURLs(ctx, gen.UpdateProfileMediaURLsParams{
		MediaUrls: filtered,
		UserID:    userID,
	})
}

func (s *RecommendationService) FindNearbyVisibleProfiles(ctx context.Context, params gen.FindNearbyVisibleProfilesParams) ([]gen.FindNearbyVisibleProfilesRow, error) {
	return s.Queries.FindNearbyVisibleProfiles(ctx, params)
}
