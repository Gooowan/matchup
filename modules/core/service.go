package core

import (
	"github.com/jackc/pgx/v5/pgxpool"

	gen "github.com/Gooowan/matchup/modules/core/gen"
)

type CoreService struct {
	Queries *gen.Queries
	DB      *pgxpool.Pool
}

func NewCoreService(db *pgxpool.Pool) *CoreService {
	return &CoreService{
		Queries: gen.New(db),
		DB:      db,
	}
}
