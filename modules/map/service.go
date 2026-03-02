package mapmod

import (
	"github.com/jackc/pgx/v5/pgxpool"

	gen "github.com/Gooowan/matchup/modules/map/gen"
	"github.com/Gooowan/matchup/modules/recommendation"
)

type MapService struct {
	Queries           *gen.Queries
	RecommendationSvc *recommendation.RecommendationService
}

func NewMapService(db *pgxpool.Pool, recommendationSvc *recommendation.RecommendationService) *MapService {
	return &MapService{Queries: gen.New(db), RecommendationSvc: recommendationSvc}
}
