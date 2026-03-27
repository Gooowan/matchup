package recommendation

import (
	"context"
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
	seen := make(map[[16]byte]bool)
	results := make([]Candidate, 0, params.Limit)

	for _, provider := range []CandidateProvider{r.Tier1, r.Tier2, r.Tier3} {
		if int32(len(results)) >= params.Limit {
			break
		}
		// Adjust limit to request only what's still needed
		remaining := params.Limit - int32(len(results))
		p := params
		p.Limit = remaining

		candidates, err := provider.GetCandidates(ctx, p)
		if err != nil {
			// Log but don't fail; try next tier
			continue
		}

		for _, c := range candidates {
			if int32(len(results)) >= params.Limit {
				break
			}
			if !seen[c.UserID.Bytes] {
				seen[c.UserID.Bytes] = true
				results = append(results, c)
			}
		}
	}

	return results, nil
}
