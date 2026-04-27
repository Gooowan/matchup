package subscriptions

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Gooowan/matchup/modules/core/logging"
	"github.com/Gooowan/matchup/modules/core/types"
	gen "github.com/Gooowan/matchup/modules/subscriptions/gen"
	"github.com/Gooowan/matchup/modules/core/utils"
)

// productPlanMap is loaded once from REVENUECAT_PLAN_MAP env var.
// Format: {"matchup_premium_monthly":"plan-uuid","matchup_premium_yearly":"plan-uuid"}
var productPlanMap map[string]string

func init() {
	raw := os.Getenv("REVENUECAT_PLAN_MAP")
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &productPlanMap)
	}
}

type revenuecatEvent struct {
	Type      string `json:"type"`
	AppUserID string `json:"app_user_id"`
	ProductID string `json:"product_id"`
	// Expiration in milliseconds epoch; present on INITIAL_PURCHASE and RENEWAL.
	ExpirationAtMs *int64 `json:"expiration_at_ms"`
}

type revenuecatPayload struct {
	Event revenuecatEvent `json:"event"`
}

// RegisterWebhook adds the RevenueCat webhook route (no auth — secret verified inline).
func (c *SubscriptionController) RegisterWebhook(r *gin.Engine) {
	r.POST("/webhooks/revenuecat", c.handleRevenuecatWebhook)
}

func (c *SubscriptionController) handleRevenuecatWebhook(ctx *gin.Context) {
	secret := os.Getenv("REVENUECAT_WEBHOOK_SECRET")
	if secret != "" {
		auth := ctx.GetHeader("Authorization")
		if !strings.EqualFold(auth, "Bearer "+secret) {
			ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "invalid webhook secret"})
			return
		}
	}

	var payload revenuecatPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "invalid payload"})
		return
	}

	log := logging.FromContext(ctx.Request.Context())
	ev := payload.Event
	log.Info("revenuecat webhook", "type", ev.Type, "app_user_id", ev.AppUserID, "product_id", ev.ProductID)

	switch ev.Type {
	case "INITIAL_PURCHASE", "RENEWAL", "UNCANCELLATION":
		userID, err := utils.StringToUUID(ev.AppUserID)
		if err != nil {
			log.Warn("revenuecat webhook: invalid app_user_id", "app_user_id", ev.AppUserID)
			ctx.JSON(http.StatusOK, types.Resp{Data: "ignored"})
			return
		}

		planID, err := c.resolvePlan(ctx, ev.ProductID)
		if err != nil {
			log.Error("revenuecat webhook: plan not found", "product_id", ev.ProductID, "error", err)
			ctx.JSON(http.StatusOK, types.Resp{Data: "plan not mapped"})
			return
		}

		now := time.Now()
		var expiredAt time.Time
		if ev.ExpirationAtMs != nil {
			expiredAt = time.UnixMilli(*ev.ExpirationAtMs)
		} else {
			expiredAt = now.Add(31 * 24 * time.Hour)
		}

		_, err = c.svc.Queries.CreateUserSubscription(ctx.Request.Context(), gen.CreateUserSubscriptionParams{
			UserID:         userID,
			SubscriptionID: planID,
			StartedAt:      pgtype.Timestamp{Time: now, Valid: true},
			ExpiredAt:      pgtype.Timestamp{Time: expiredAt, Valid: true},
		})
		if err != nil {
			log.Error("revenuecat webhook: failed to create subscription", "error", err)
			ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "failed to assign subscription"})
			return
		}

	case "CANCELLATION", "EXPIRATION", "SUBSCRIBER_ALIAS":
		if ev.Type != "CANCELLATION" && ev.Type != "EXPIRATION" {
			ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
			return
		}
		userID, err := utils.StringToUUID(ev.AppUserID)
		if err != nil {
			ctx.JSON(http.StatusOK, types.Resp{Data: "ignored"})
			return
		}
		activeSub, err := c.svc.Queries.GetActiveUserSubscription(ctx.Request.Context(), userID)
		if err == nil {
			status := "cancelled"
			if ev.Type == "EXPIRATION" {
				status = "finished"
			}
			_ = c.svc.Queries.UpdateUserSubscriptionStatus(ctx.Request.Context(), gen.UpdateUserSubscriptionStatusParams{
				ID:     activeSub.ID,
				Status: status,
			})
		}
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

// resolvePlan resolves a RevenueCat product_id to a local subscription plan UUID.
// Uses REVENUECAT_PLAN_MAP env var first, then falls back to name substring match.
func (c *SubscriptionController) resolvePlan(ctx *gin.Context, productID string) (pgtype.UUID, error) {
	if planIDStr, ok := productPlanMap[productID]; ok {
		return utils.StringToUUID(planIDStr)
	}

	// Fallback: match by plan name substring (e.g. "monthly" / "yearly")
	plans, err := c.svc.Queries.ListSubscriptions(ctx.Request.Context())
	if err != nil || len(plans) == 0 {
		return pgtype.UUID{}, err
	}
	lower := strings.ToLower(productID)
	for _, p := range plans {
		if strings.Contains(lower, strings.ToLower(p.Name)) ||
			strings.Contains(strings.ToLower(p.Name), lower) {
			return p.ID, nil
		}
	}
	// Last resort: first active plan
	return plans[0].ID, nil
}
