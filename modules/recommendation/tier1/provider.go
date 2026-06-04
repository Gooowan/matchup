// Package tier1 implements pure-distance recommendation: profiles ordered by
// club-to-club distance with only the gender preference applied as a hard filter.
// All other soft filters (age/height/goal/program/categories/city) are ignored so
// users always see someone even after mutual and one-way matches are exhausted.
package tier1

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	rec "github.com/Gooowan/matchup/modules/recommendation"
	gen "github.com/Gooowan/matchup/modules/recommendation/gen"
)

// Provider implements rec.CandidateProvider using pure-distance logic.
type Provider struct {
	queries *gen.Queries
}

func NewProvider(queries *gen.Queries) *Provider {
	return &Provider{queries: queries}
}

// GetCandidates returns candidates ordered by distance; only gender is filtered.
func (p *Provider) GetCandidates(ctx context.Context, params rec.FeedParams) ([]rec.Candidate, error) {
	// Apply only the gender filter — drop all other soft filters.
	var genderFilter pgtype.Text
	if params.Filters.PreferredGender != nil {
		genderFilter = pgtype.Text{String: *params.Filters.PreferredGender, Valid: true}
	}

	rows, err := p.queries.FindNearbyVisibleProfiles(ctx, gen.FindNearbyVisibleProfilesParams{
		Latitude:            params.Latitude,
		Longitude:           params.Longitude,
		UserID:              params.UserID,
		ExcludeIds:          params.ExcludeIDs,
		PreferredGender:     genderFilter,
		// All other filters intentionally left as zero-values (NULL in SQL → no filter).
		LimitVal:            params.Limit,
	})
	if err != nil || len(rows) == 0 {
		return nil, nil
	}

	results := make([]rec.Candidate, 0, len(rows))
	for _, row := range rows {
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
			DistKm:    row.DistanceKm,
			Source:    "proximity",
		})
	}
	return results, nil
}
