package controllers

import (
	"context"
	"log/slog"
	"time"

	"github.com/Gooowan/matchup/modules/core/metrics"
	"github.com/Gooowan/matchup/modules/subscriptions"
)

type CronController struct {
	subscriptions *subscriptions.SubscriptionService
	log           *slog.Logger
}

func NewCronController(subscriptionSvc *subscriptions.SubscriptionService, logger *slog.Logger) *CronController {
	return &CronController{
		subscriptions: subscriptionSvc,
		log:           logger,
	}
}

// ExpireSubscriptions transitions active subscriptions past their expired_at to 'finished'.
func (c *CronController) ExpireSubscriptions() {
	ctx := context.Background()
	start := time.Now()

	count, err := c.subscriptions.Queries.FinishExpiredSubscriptions(ctx)
	if err != nil {
		c.log.Error("failed to expire subscriptions", "error", err)
		metrics.CronJobFailureTotal.WithLabelValues("expire_subscriptions").Inc()
		return
	}

	metrics.CronJobSuccessTotal.WithLabelValues("expire_subscriptions").Inc()
	if count > 0 {
		c.log.Info("subscriptions expired", "count", count, "duration_ms", time.Since(start).Milliseconds())
	}
}

// NotifyExpiringSoon logs subscriptions expiring within 1 day.
// Placeholder for future notification integration (email, push, etc.).
func (c *CronController) NotifyExpiringSoon() {
	ctx := context.Background()

	expiring, err := c.subscriptions.Queries.ListSubscriptionsExpiring1Day(ctx)
	if err != nil {
		c.log.Error("failed to fetch expiring subscriptions", "error", err)
		metrics.CronJobFailureTotal.WithLabelValues("notify_expiring_soon").Inc()
		return
	}

	metrics.CronJobSuccessTotal.WithLabelValues("notify_expiring_soon").Inc()
	if len(expiring) > 0 {
		c.log.Info("subscriptions expiring within 1 day", "count", len(expiring))
		// TODO: Send push notifications / emails to users
	}
}
