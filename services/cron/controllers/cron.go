package controllers

import (
	"context"
	"log"
	"time"

	"github.com/Gooowan/matchup/modules/subscriptions"
)

type CronController struct {
	subscriptions *subscriptions.SubscriptionService
}

func NewCronController(subscriptionSvc *subscriptions.SubscriptionService) *CronController {
	return &CronController{
		subscriptions: subscriptionSvc,
	}
}

// ExpireSubscriptions transitions active subscriptions past their expired_at to 'finished'.
func (c *CronController) ExpireSubscriptions() {
	ctx := context.Background()
	start := time.Now()

	count, err := c.subscriptions.Queries.FinishExpiredSubscriptions(ctx)
	if err != nil {
		log.Printf("[CRON] Failed to expire subscriptions: %v", err)
		return
	}

	if count > 0 {
		log.Printf("[CRON] Expired %d subscriptions in %s", count, time.Since(start))
	}
}

// NotifyExpiringSoon logs subscriptions expiring within 1 day.
// Placeholder for future notification integration (email, push, etc.).
func (c *CronController) NotifyExpiringSoon() {
	ctx := context.Background()

	expiring, err := c.subscriptions.Queries.ListSubscriptionsExpiring1Day(ctx)
	if err != nil {
		log.Printf("[CRON] Failed to fetch expiring subscriptions: %v", err)
		return
	}

	if len(expiring) > 0 {
		log.Printf("[CRON] %d subscriptions expiring within 1 day", len(expiring))
		// TODO: Send notifications to users
	}
}
