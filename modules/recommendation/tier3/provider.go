// Package tier3 is a stub for the collaborative-filtering recommendation engine.
// TODO: Find users with similar like patterns (cosine similarity on
//       recommendation_likes_log feature vectors) and recommend profiles
//       that similar users liked but the current user hasn't seen.
package tier3

import (
	"context"
	"log"

	rec "github.com/Gooowan/matchup/modules/recommendation"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) GetCandidates(ctx context.Context, params rec.FeedParams) ([]rec.Candidate, error) {
	log.Println("[tier3] collaborative filtering not yet implemented — returning empty")
	return nil, nil
}
