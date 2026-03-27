// Package tier2 is a stub for the preference-model recommendation engine.
// TODO: Build a preference profile from like history,
//       compute a weighted average of features (height, age, goal, program),
//       and score unseen candidates by similarity.
package tier2

import (
	"context"
	"log"

	rec "github.com/Gooowan/matchup/modules/recommendation"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) GetCandidates(ctx context.Context, params rec.FeedParams) ([]rec.Candidate, error) {
	log.Println("[tier2] preference model not yet implemented — returning empty")
	return nil, nil
}
