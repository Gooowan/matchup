// Package tier3 implements collaborative filtering by finding users with similar
// like patterns and recommending profiles that similar users liked but the current
// user has not yet seen.
package tier3

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	rec "github.com/Gooowan/matchup/modules/recommendation"
	gen "github.com/Gooowan/matchup/modules/recommendation/gen"
)

type Provider struct {
	queries *gen.Queries
}

func NewProvider(queries *gen.Queries) *Provider {
	return &Provider{queries: queries}
}

func (p *Provider) GetCandidates(ctx context.Context, params rec.FeedParams) ([]rec.Candidate, error) {
	// Find users who liked at least 2 of the same profiles as the current user.
	similarRows, err := p.queries.GetSimilarUsers(ctx, params.UserID)
	if err != nil || len(similarRows) == 0 {
		return nil, nil
	}

	similarIDs := make([]pgtype.UUID, 0, len(similarRows))
	for _, row := range similarRows {
		similarIDs = append(similarIDs, row.UserID)
	}

	// Get profiles liked by similar users that current user hasn't liked yet.
	likedIDs, err := p.queries.GetProfilesLikedBySimilarUsers(ctx, gen.GetProfilesLikedBySimilarUsersParams{
		SimilarUserIds: similarIDs,
		UserID:          params.UserID,
	})
	if err != nil || len(likedIDs) == 0 {
		return nil, nil
	}

	// Filter out already-excluded IDs.
	excludeSet := make(map[[16]byte]bool, len(params.ExcludeIDs)+1)
	excludeSet[params.UserID.Bytes] = true
	for _, id := range params.ExcludeIDs {
		excludeSet[id.Bytes] = true
	}

	candidateIDs := make([]pgtype.UUID, 0, len(likedIDs))
	for _, id := range likedIDs {
		if !excludeSet[id.Bytes] {
			candidateIDs = append(candidateIDs, id)
			excludeSet[id.Bytes] = true
		}
	}
	if len(candidateIDs) == 0 {
		return nil, nil
	}

	profiles, err := p.queries.GetProfilesByUserIDs(ctx, candidateIDs)
	if err != nil {
		return nil, nil
	}

	results := make([]rec.Candidate, 0, params.Limit)
	for _, row := range profiles {
		if int32(len(results)) >= params.Limit {
			break
		}
		lat, lng := 0.0, 0.0
		if row.Latitude.Valid {
			lat = row.Latitude.Float64
		}
		if row.Longitude.Valid {
			lng = row.Longitude.Float64
		}
		results = append(results, rec.Candidate{
			UserID:    row.UserID,
			Latitude:  lat,
			Longitude: lng,
			Source:    "collaborative",
		})
	}
	return results, nil
}
