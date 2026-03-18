package subscriptions

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	gen "github.com/Gooowan/matchup/modules/subscriptions/gen"
)

type SubscriptionService struct {
	DB      *pgxpool.Pool
	Queries *gen.Queries
}

func NewSubscriptionService(db *pgxpool.Pool) *SubscriptionService {
	return &SubscriptionService{
		DB:      db,
		Queries: gen.New(db),
	}
}

// AssignSubscription fetches the plan to compute expired_at, then creates the user subscription.
func (s *SubscriptionService) AssignSubscription(ctx context.Context, userID, subscriptionID pgtype.UUID) (gen.UserSubscription, error) {
	plan, err := s.Queries.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return gen.UserSubscription{}, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	now := time.Now()
	expiredAt := now.Add(time.Duration(plan.DurationDays) * 24 * time.Hour)

	return s.Queries.CreateUserSubscription(ctx, gen.CreateUserSubscriptionParams{
		UserID:         userID,
		SubscriptionID: subscriptionID,
		StartedAt:      pgtype.Timestamp{Time: now, Valid: true},
		ExpiredAt:      pgtype.Timestamp{Time: expiredAt, Valid: true},
	})
}
