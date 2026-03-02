package core

import (
	"github.com/jackc/pgx/v5/pgxpool"

	gen "github.com/Gooowan/matchup/modules/users/gen"
)

type UserService struct {
	Queries *gen.Queries
	DB      *pgxpool.Pool
}

func NewCoreService(db *pgxpool.Pool) *UserService {
	return &UserService{
		Queries: gen.New(db),
		DB:      db,
	}
}
