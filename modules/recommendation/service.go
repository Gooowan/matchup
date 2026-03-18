package recommendation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Gooowan/matchup/modules/core/types"
	gen "github.com/Gooowan/matchup/modules/recommendation/gen"
)

type RecommendationService struct {
	DB      *pgxpool.Pool
	Queries *gen.Queries
}

func NewRecommendationService(db *pgxpool.Pool) *RecommendationService {
	return &RecommendationService{DB: db, Queries: gen.New(db)}
}

// getProfileData extracts the data JSONB from a profile as a map
func getProfileData(data types.JSONB) types.JSONB {
	if data == nil {
		return types.JSONB{}
	}
	return data
}

func (s *RecommendationService) AddMediaURL(ctx context.Context, userID pgtype.UUID, url string) error {
	profile, err := s.Queries.GetProfileByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	data := getProfileData(profile.Data)

	var urls []string
	if raw, ok := data["media_urls"]; ok {
		b, _ := json.Marshal(raw)
		json.Unmarshal(b, &urls)
	}
	urls = append(urls, url)
	data["media_urls"] = urls

	return s.Queries.UpdateProfileData(ctx, gen.UpdateProfileDataParams{
		Data:   data,
		UserID: userID,
	})
}

func (s *RecommendationService) RemoveMediaURL(ctx context.Context, userID pgtype.UUID, url string) error {
	profile, err := s.Queries.GetProfileByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	data := getProfileData(profile.Data)

	var urls []string
	if raw, ok := data["media_urls"]; ok {
		b, _ := json.Marshal(raw)
		json.Unmarshal(b, &urls)
	}

	filtered := make([]string, 0, len(urls))
	for _, u := range urls {
		if u != url {
			filtered = append(filtered, u)
		}
	}
	data["media_urls"] = filtered

	return s.Queries.UpdateProfileData(ctx, gen.UpdateProfileDataParams{
		Data:   data,
		UserID: userID,
	})
}
