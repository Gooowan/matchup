// Package tier2 implements the preference-model recommendation engine.
// It builds a preference profile from the user's like history and scores
// unseen candidates by feature similarity (categories, goal, program, height).
package tier2

import (
	"context"
	"encoding/json"
	"math"
	"sort"

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

// featureSnapshot mirrors the JSON stored in recommendation_likes_log.features.
type featureSnapshot struct {
	Categories []string `json:"categories"`
	Goal       string   `json:"goal"`
	Program    string   `json:"program"`
	HeightCm   *int16   `json:"height_cm,omitempty"`
}

type prefProfile struct {
	categoryFreq map[string]int
	goalFreq     map[string]int
	programFreq  map[string]int
	heightSum    float64
	heightCount  int
}

func buildPref(history []json.RawMessage) prefProfile {
	p := prefProfile{
		categoryFreq: make(map[string]int),
		goalFreq:     make(map[string]int),
		programFreq:  make(map[string]int),
	}
	for _, raw := range history {
		var snap featureSnapshot
		if err := json.Unmarshal(raw, &snap); err != nil {
			continue
		}
		for _, c := range snap.Categories {
			p.categoryFreq[c]++
		}
		if snap.Goal != "" {
			p.goalFreq[snap.Goal]++
		}
		if snap.Program != "" {
			p.programFreq[snap.Program]++
		}
		if snap.HeightCm != nil {
			p.heightSum += float64(*snap.HeightCm)
			p.heightCount++
		}
	}
	return p
}

func topN(freq map[string]int, n int) []string {
	type kv struct {
		key   string
		count int
	}
	pairs := make([]kv, 0, len(freq))
	for k, v := range freq {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].count > pairs[j].count })
	out := make([]string, 0, n)
	for i := 0; i < n && i < len(pairs); i++ {
		out = append(out, pairs[i].key)
	}
	return out
}

func mostCommon(freq map[string]int) string {
	best, bestCount := "", 0
	for k, v := range freq {
		if v > bestCount {
			best, bestCount = k, v
		}
	}
	return best
}

func score(row gen.GetCountryWideProfilesRow, topCats []string, topGoal, topProgram string, meanHeight float64) float64 {
	catSet := make(map[string]bool, len(row.Categories))
	for _, c := range row.Categories {
		catSet[c] = true
	}
	intersect := 0
	for _, c := range topCats {
		if catSet[c] {
			intersect++
		}
	}
	denom := len(topCats)
	if denom == 0 {
		denom = 1
	}
	catScore := float64(intersect) / float64(denom)

	goalScore := 0.0
	if topGoal != "" && row.Goal == topGoal {
		goalScore = 1.0
	}
	progScore := 0.0
	if topProgram != "" && row.Program == topProgram {
		progScore = 1.0
	}
	htScore := 0.0
	if meanHeight > 0 && row.HeightCm.Valid {
		htScore = math.Max(0, 1.0-math.Abs(float64(row.HeightCm.Int16)-meanHeight)/30.0)
	}
	return 0.40*catScore + 0.25*goalScore + 0.20*progScore + 0.15*htScore
}

func (p *Provider) GetCandidates(ctx context.Context, params rec.FeedParams) ([]rec.Candidate, error) {
	rawHistory, err := p.queries.GetLikeHistory(ctx, params.UserID)
	if err != nil || len(rawHistory) < 3 {
		return nil, nil
	}

	history := make([]json.RawMessage, 0, len(rawHistory))
	for _, j := range rawHistory {
		if raw, err := json.Marshal(j); err == nil {
			history = append(history, raw)
		}
	}

	pref := buildPref(history)
	topCats := topN(pref.categoryFreq, 3)
	topGoal := mostCommon(pref.goalFreq)
	topProgram := mostCommon(pref.programFreq)
	meanHeight := 0.0
	if pref.heightCount > 0 {
		meanHeight = pref.heightSum / float64(pref.heightCount)
	}

	if params.Country == "" {
		return nil, nil
	}

	excludeSet := make(map[[16]byte]bool, len(params.ExcludeIDs)+1)
	excludeSet[params.UserID.Bytes] = true
	for _, id := range params.ExcludeIDs {
		excludeSet[id.Bytes] = true
	}

	fetchLimit := params.Limit * 3
	if fetchLimit < 30 {
		fetchLimit = 30
	}

	countryRows, err := p.queries.GetCountryWideProfiles(ctx, gen.GetCountryWideProfilesParams{
		Country:             pgtype.Text{String: params.Country, Valid: true},
		UserID:              params.UserID,
		ExcludeIds:          params.ExcludeIDs,
		PreferredGender:     pgtype.Text{},
		AgeMin:              pgtype.Int2{},
		AgeMax:              pgtype.Int2{},
		HeightMin:           pgtype.Int2{},
		HeightMax:           pgtype.Int2{},
		PreferredGoal:       pgtype.Text{},
		PreferredProgram:    pgtype.Text{},
		PreferredCategories: nil,
		LimitVal:            fetchLimit,
	})
	if err != nil || len(countryRows) == 0 {
		return nil, nil
	}

	type sc struct {
		row   gen.GetCountryWideProfilesRow
		score float64
	}
	scored := make([]sc, 0, len(countryRows))
	for _, row := range countryRows {
		if excludeSet[row.UserID.Bytes] {
			continue
		}
		scored = append(scored, sc{row, score(row, topCats, topGoal, topProgram, meanHeight)})
	}
	sort.Slice(scored, func(i, j int) bool { return scored[i].score > scored[j].score })

	results := make([]rec.Candidate, 0, params.Limit)
	for _, s := range scored {
		if int32(len(results)) >= params.Limit {
			break
		}
		lat, lng := 0.0, 0.0
		if s.row.Latitude.Valid {
			lat = s.row.Latitude.Float64
		}
		if s.row.Longitude.Valid {
			lng = s.row.Longitude.Float64
		}
		results = append(results, rec.Candidate{
			UserID:    s.row.UserID,
			Latitude:  lat,
			Longitude: lng,
			Source:    "preference_model",
		})
	}
	return results, nil
}
