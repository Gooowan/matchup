// Package tier2 implements one-way filter-match recommendation: candidates who
// pass MY filter preferences (age, height, goal, program, categories, city),
// ordered by club-to-club distance.
package tier2

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
	fp := toSQLCFilters(params.Filters)

	rows, err := p.queries.FindNearbyVisibleProfiles(ctx, gen.FindNearbyVisibleProfilesParams{
		Latitude:            params.Latitude,
		Longitude:           params.Longitude,
		UserID:              params.UserID,
		ExcludeIds:          params.ExcludeIDs,
		PreferredGender:     fp.preferredGender,
		AgeMin:              fp.ageMin,
		AgeMax:              fp.ageMax,
		HeightMin:           fp.heightMin,
		HeightMax:           fp.heightMax,
		PreferredGoal:       fp.preferredGoal,
		PreferredProgram:    fp.preferredProgram,
		PreferredCategories: fp.preferredCategories,
		PreferredCity:       fp.preferredCity,
		PreferredCountry:    fp.preferredCountry,
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
			Source:    "filter_match",
		})
	}
	return results, nil
}

// sqlcFilters holds pgtype-wrapped filter values for sqlc queries.
type sqlcFilters struct {
	preferredGender     pgtype.Text
	ageMin              pgtype.Int2
	ageMax              pgtype.Int2
	heightMin           pgtype.Int2
	heightMax           pgtype.Int2
	preferredGoal       pgtype.Text
	preferredProgram    pgtype.Text
	preferredCategories []string
	preferredCountry    pgtype.Text
	preferredCity       pgtype.Text
}

func toSQLCFilters(f rec.FilterParams) sqlcFilters {
	p := sqlcFilters{}
	if f.PreferredGender != nil {
		p.preferredGender = pgtype.Text{String: *f.PreferredGender, Valid: true}
	}
	if f.AgeMin != nil {
		p.ageMin = pgtype.Int2{Int16: *f.AgeMin, Valid: true}
	}
	if f.AgeMax != nil {
		p.ageMax = pgtype.Int2{Int16: *f.AgeMax, Valid: true}
	}
	if f.HeightMin != nil {
		p.heightMin = pgtype.Int2{Int16: *f.HeightMin, Valid: true}
	}
	if f.HeightMax != nil {
		p.heightMax = pgtype.Int2{Int16: *f.HeightMax, Valid: true}
	}
	if f.PreferredGoal != nil {
		p.preferredGoal = pgtype.Text{String: *f.PreferredGoal, Valid: true}
	}
	if f.PreferredProgram != nil {
		p.preferredProgram = pgtype.Text{String: *f.PreferredProgram, Valid: true}
	}
	if len(f.PreferredCategories) > 0 {
		p.preferredCategories = f.PreferredCategories
	}
	if f.PreferredCountry != nil {
		p.preferredCountry = pgtype.Text{String: *f.PreferredCountry, Valid: true}
	}
	if f.PreferredCity != nil {
		p.preferredCity = pgtype.Text{String: *f.PreferredCity, Valid: true}
	}
	return p
}
