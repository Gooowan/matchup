// Package tier1 implements the club-proximity recommendation engine.
// It operates in three expanding circles:
//   Circle 1: Same club members (filtered by preferences)
//   Circle 2: Members of nearby clubs (ordered by club distance, interleaved via min-heap)
//   Circle 3: Country-wide fallback (random order, filtered by preferences)
package tier1

import (
	"context"
	"math"

	"github.com/jackc/pgx/v5/pgtype"

	rec "github.com/Gooowan/matchup/modules/recommendation"
	gen "github.com/Gooowan/matchup/modules/recommendation/gen"
)

// Provider implements rec.CandidateProvider using club-proximity logic.
type Provider struct {
	queries *gen.Queries
}

func NewProvider(queries *gen.Queries) *Provider {
	return &Provider{queries: queries}
}

// GetCandidates runs the 3-circle club-proximity pipeline.
func (p *Provider) GetCandidates(ctx context.Context, params rec.FeedParams) ([]rec.Candidate, error) {
	filterParams := buildFilterParams(params.Filters)
	excludeIDs := params.ExcludeIDs
	seen := make(map[[16]byte]bool)
	results := make([]rec.Candidate, 0, params.Limit)

	// --- Circle 1: Same club ---
	if len(params.UserClubs) > 0 {
		clubIDs := make([]pgtype.UUID, len(params.UserClubs))
		for i, c := range params.UserClubs {
			clubIDs[i] = c.ID
		}

		rows, err := p.queries.GetSameClubProfiles(ctx, gen.GetSameClubProfilesParams{
			ClubIds:               clubIDs,
			UserID:                params.UserID,
			ExcludeIds:            excludeIDs,
			PreferredGender:       filterParams.preferredGender,
			AgeMin:                filterParams.ageMin,
			AgeMax:                filterParams.ageMax,
			HeightMin:             filterParams.heightMin,
			HeightMax:             filterParams.heightMax,
			PreferredGoal:         filterParams.preferredGoal,
			PreferredProgram:      filterParams.preferredProgram,
			PreferredCategories:   filterParams.preferredCategories,
			PreferredCountry:      filterParams.preferredCountry,
			PreferredCity:         filterParams.preferredCity,
			WantsPartnerToFinance: filterParams.wantsPartnerToFinance,
			LimitVal:              params.Limit,
		})
		if err == nil {
			for _, row := range rows {
				if int32(len(results)) >= params.Limit {
					break
				}
				if !seen[row.UserID.Bytes] {
					seen[row.UserID.Bytes] = true
					results = append(results, rec.Candidate{
						UserID: row.UserID,
						Source: "same_club",
					})
				}
			}
		}
	}

	if int32(len(results)) >= params.Limit {
		return results, nil
	}

	// --- Circle 2: Nearby clubs ---
	if len(params.UserClubs) > 0 {
		userClubIDs := make([]pgtype.UUID, len(params.UserClubs))
		for i, c := range params.UserClubs {
			userClubIDs[i] = c.ID
		}

		// Use the centroid of user's clubs as the reference point
		refLat, refLng := clubCentroid(params.UserClubs)

		remaining := params.Limit - int32(len(results))
		rows, err := p.queries.GetNearbyClubProfiles(ctx, gen.GetNearbyClubProfilesParams{
			RefLatitude:           refLat,
			RefLongitude:          refLng,
			ExcludeClubIds:        userClubIDs,
			UserID:                params.UserID,
			ExcludeIds:            excludeIDs,
			PreferredGender:       filterParams.preferredGender,
			AgeMin:                filterParams.ageMin,
			AgeMax:                filterParams.ageMax,
			HeightMin:             filterParams.heightMin,
			HeightMax:             filterParams.heightMax,
			PreferredGoal:         filterParams.preferredGoal,
			PreferredProgram:      filterParams.preferredProgram,
			PreferredCategories:   filterParams.preferredCategories,
			PreferredCountry:      filterParams.preferredCountry,
			PreferredCity:         filterParams.preferredCity,
			WantsPartnerToFinance: filterParams.wantsPartnerToFinance,
			LimitVal:              remaining * 3, // over-fetch for dedup
		})
		if err == nil {
			// Merge interleaved by club distance
			entries := make([]candidateEntry, 0, len(rows))
			for _, row := range rows {
				if !seen[row.UserID.Bytes] {
					entries = append(entries, candidateEntry{
						userIDBytes: row.UserID.Bytes,
						distKm:      row.ClubDistKm,
						source:      "nearby_club",
					})
				}
			}
			merged := mergeByDistance([][]candidateEntry{entries}, int(remaining))
			for _, e := range merged {
				if int32(len(results)) >= params.Limit {
					break
				}
				if !seen[e.userIDBytes] {
					seen[e.userIDBytes] = true
					uid := pgtype.UUID{Bytes: e.userIDBytes, Valid: true}
					results = append(results, rec.Candidate{
						UserID: uid,
						Source: e.source,
						DistKm: e.distKm,
					})
				}
			}
		}
	}

	if int32(len(results)) >= params.Limit {
		return results, nil
	}

	// --- Circle 3: Country-wide ---
	if params.Country != "" {
		remaining := params.Limit - int32(len(results))
		rows, err := p.queries.GetCountryWideProfiles(ctx, gen.GetCountryWideProfilesParams{
			Country:               pgtype.Text{String: params.Country, Valid: true},
			UserID:                params.UserID,
			ExcludeIds:            excludeIDs,
			PreferredGender:       filterParams.preferredGender,
			AgeMin:                filterParams.ageMin,
			AgeMax:                filterParams.ageMax,
			HeightMin:             filterParams.heightMin,
			HeightMax:             filterParams.heightMax,
			PreferredGoal:         filterParams.preferredGoal,
			PreferredProgram:      filterParams.preferredProgram,
			PreferredCategories:   filterParams.preferredCategories,
			PreferredCity:         filterParams.preferredCity,
			WantsPartnerToFinance: filterParams.wantsPartnerToFinance,
			LimitVal:              remaining * 2,
		})
		if err == nil {
			for _, row := range rows {
				if int32(len(results)) >= params.Limit {
					break
				}
				if !seen[row.UserID.Bytes] {
					seen[row.UserID.Bytes] = true
					results = append(results, rec.Candidate{
						UserID: row.UserID,
						Source: "country_wide",
					})
				}
			}
		}
	}

	return results, nil
}

// clubCentroid computes the average lat/lng of a set of clubs.
func clubCentroid(clubs []rec.UserClub) (lat, lng float64) {
	if len(clubs) == 0 {
		return 0, 0
	}
	sumLat, sumLng := 0.0, 0.0
	for _, c := range clubs {
		sumLat += c.Latitude
		sumLng += c.Longitude
	}
	n := float64(len(clubs))
	return math.Round(sumLat/n*1e6) / 1e6, math.Round(sumLng/n*1e6) / 1e6
}

// sqlcFilterParams holds the generated pgtype values for sqlc calls.
type sqlcFilterParams struct {
	preferredGender        pgtype.Text
	ageMin                 pgtype.Int2
	ageMax                 pgtype.Int2
	heightMin              pgtype.Int2
	heightMax              pgtype.Int2
	preferredGoal          pgtype.Text
	preferredProgram       pgtype.Text
	preferredCategories    []string
	preferredCountry       pgtype.Text
	preferredCity          pgtype.Text
	wantsPartnerToFinance  pgtype.Text
}

func buildFilterParams(f rec.FilterParams) sqlcFilterParams {
	p := sqlcFilterParams{}

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
	if f.WantsPartnerToFinance != nil {
		p.wantsPartnerToFinance = pgtype.Text{String: *f.WantsPartnerToFinance, Valid: true}
	}

	return p
}
