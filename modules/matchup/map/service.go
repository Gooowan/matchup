package mapmod

import (
	gen "github.com/Gooowan/matchup/modules/matchup/gen"
)

type MapService struct {
	Queries *gen.Queries
}

func NewMapService(queries *gen.Queries) *MapService {
	return &MapService{Queries: queries}
}
