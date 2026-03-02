package profile

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	gen "github.com/Gooowan/matchup/modules/matchup/gen"
)

type ProfileService struct {
	DB      *pgxpool.Pool
	Queries *gen.Queries
}

func NewProfileService(db *pgxpool.Pool, queries *gen.Queries) *ProfileService {
	return &ProfileService{DB: db, Queries: queries}
}

func (s *ProfileService) CreateProfile(ctx context.Context, userID pgtype.UUID, params gen.CreateProfileParams) (gen.Profile, error) {
	params.UserID = userID
	return s.Queries.CreateProfile(ctx, params)
}

func (s *ProfileService) GetProfile(ctx context.Context, userID pgtype.UUID) (gen.Profile, error) {
	return s.Queries.GetProfileByUserID(ctx, userID)
}

func (s *ProfileService) UpdateProfile(ctx context.Context, userID pgtype.UUID, params gen.UpdateProfileParams) error {
	params.UserID = userID
	return s.Queries.UpdateProfile(ctx, params)
}

func (s *ProfileService) GetProfilePreview(ctx context.Context, userID pgtype.UUID) (gen.GetProfilePreviewRow, error) {
	return s.Queries.GetProfilePreview(ctx, userID)
}

func (s *ProfileService) GetPreferences(ctx context.Context, userID pgtype.UUID) (gen.UserPreference, error) {
	return s.Queries.GetPreferences(ctx, userID)
}

func (s *ProfileService) UpsertPreferences(ctx context.Context, userID pgtype.UUID, params gen.UpsertPreferencesParams) (gen.UserPreference, error) {
	params.UserID = userID
	return s.Queries.UpsertPreferences(ctx, params)
}

func (s *ProfileService) AddMediaURL(ctx context.Context, userID pgtype.UUID, url string) error {
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

func (s *ProfileService) RemoveMediaURL(ctx context.Context, userID pgtype.UUID, url string) error {
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
