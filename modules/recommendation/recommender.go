package recommendation

import (
	"context"
	"fmt"

	"github.com/Gooowan/matchup/modules/core/logging"
	"github.com/Gooowan/matchup/modules/core/metrics"
)

// Recommender orchestrates the 3-tier recommendation pipeline.
// It tries Tier 1 first. If the result is smaller than limit,
// it continues to Tier 2, then Tier 3, deduplicating along the way.
type Recommender struct {
	Tier1 CandidateProvider
	Tier2 CandidateProvider
	Tier3 CandidateProvider
}

func NewRecommender(t1, t2, t3 CandidateProvider) *Recommender {
	return &Recommender{Tier1: t1, Tier2: t2, Tier3: t3}
}

// GetCandidates runs the tier pipeline and returns up to params.Limit candidates.
func (r *Recommender) GetCandidates(ctx context.Context, params FeedParams) ([]Candidate, error) {
	logger := logging.FromContext(ctx)
	seen := make(map[[16]byte]bool)
	results := make([]Candidate, 0, params.Limit)

	tierNames := []string{"tier1", "tier2", "tier3"}
	providers := []CandidateProvider{r.Tier1, r.Tier2, r.Tier3}

	for i, provider := range providers {
		if int32(len(results)) >= params.Limit {
			break
		}
		tier := tierNames[i]
		remaining := params.Limit - int32(len(results))
		p := params
		p.Limit = remaining

		candidates, err := provider.GetCandidates(ctx, p)
		if err != nil {
			metrics.RecommendationTierErrors.WithLabelValues(tier).Inc()
			logger.Warn("recommendation tier error", "tier", tier, "error", fmt.Sprintf("%v", err))
			continue
		}
		if len(candidates) == 0 {
			metrics.RecommendationTierEmpty.WithLabelValues(tier).Inc()
			logger.Warn("recommendation tier returned empty", "tier", tier,
				"country", params.Country, "has_filters", params.Filters.PreferredGender != nil)
			continue
		}

		added := 0
		for _, c := range candidates {
			if int32(len(results)) >= params.Limit {
				break
			}
			if !seen[c.UserID.Bytes] {
				seen[c.UserID.Bytes] = true
				results = append(results, c)
				added++
			}
		}
		if added > 0 {
			metrics.RecommendationTierHits.WithLabelValues(tier).Add(float64(added))
		}
	}

	return results, nil
}
