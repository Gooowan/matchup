package subscriptions

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
	gen "github.com/Gooowan/matchup/modules/subscriptions/gen"
	"github.com/Gooowan/matchup/modules/users/auth"
)

type SubscriptionController struct {
	svc *SubscriptionService
}

func NewSubscriptionController(svc *SubscriptionService) *SubscriptionController {
	return &SubscriptionController{svc: svc}
}

func (c *SubscriptionController) RegisterRoutes(r *gin.RouterGroup, adminAuth, userAuth gin.HandlerFunc) {
	admin := r.Group("/admin")
	admin.Use(adminAuth)
	admin.POST("/plans", c.CreatePlan)
	admin.GET("/plans", c.ListAllPlans)
	admin.GET("/plans/:id", c.GetPlan)
	admin.PUT("/plans/:id", c.UpdatePlan)
	admin.DELETE("/plans/:id", c.DeactivatePlan)
	admin.POST("/assign", c.AssignSubscription)

	user := r.Group("")
	user.Use(userAuth)
	user.GET("/plans", c.ListActivePlans)
	user.GET("/my", c.GetMySubscriptions)
	user.GET("/my/active", c.GetMyActiveSubscription)
}

// Admin handlers

func (c *SubscriptionController) CreatePlan(ctx *gin.Context) {
	var req struct {
		Name         string `json:"name" binding:"required"`
		Description  string `json:"description"`
		DurationDays int32  `json:"duration_days" binding:"required,min=1"`
		PriceCents   int64  `json:"price_cents" binding:"min=0"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	plan, err := c.svc.Queries.CreateSubscription(ctx.Request.Context(), gen.CreateSubscriptionParams{
		Name:         req.Name,
		Description:  pgtype.Text{String: req.Description, Valid: req.Description != ""},
		DurationDays: req.DurationDays,
		PriceCents:   req.PriceCents,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to create subscription plan"})
		return
	}

	ctx.JSON(http.StatusCreated, types.Resp{Data: plan})
}

func (c *SubscriptionController) ListAllPlans(ctx *gin.Context) {
	plans, err := c.svc.Queries.ListAllSubscriptions(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to list plans"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: plans})
}

func (c *SubscriptionController) GetPlan(ctx *gin.Context) {
	id, err := utils.StringToUUID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid plan ID"})
		return
	}

	plan, err := c.svc.Queries.GetSubscription(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Plan not found"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: plan})
}

func (c *SubscriptionController) UpdatePlan(ctx *gin.Context) {
	id, err := utils.StringToUUID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid plan ID"})
		return
	}

	var req struct {
		Name         string `json:"name" binding:"required"`
		Description  string `json:"description"`
		DurationDays int32  `json:"duration_days" binding:"required,min=1"`
		PriceCents   int64  `json:"price_cents" binding:"min=0"`
		IsActive     bool   `json:"is_active"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	if err := c.svc.Queries.UpdateSubscription(ctx.Request.Context(), gen.UpdateSubscriptionParams{
		ID:           id,
		Name:         req.Name,
		Description:  pgtype.Text{String: req.Description, Valid: req.Description != ""},
		DurationDays: req.DurationDays,
		PriceCents:   req.PriceCents,
		IsActive:     req.IsActive,
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to update plan"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *SubscriptionController) DeactivatePlan(ctx *gin.Context) {
	id, err := utils.StringToUUID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid plan ID"})
		return
	}

	if err := c.svc.Queries.DeactivateSubscription(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to deactivate plan"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *SubscriptionController) AssignSubscription(ctx *gin.Context) {
	var req struct {
		UserID         string `json:"user_id" binding:"required"`
		SubscriptionID string `json:"subscription_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	userID, err := utils.StringToUUID(req.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid user ID"})
		return
	}

	subID, err := utils.StringToUUID(req.SubscriptionID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid subscription ID"})
		return
	}

	userSub, err := c.svc.AssignSubscription(ctx.Request.Context(), userID, subID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to assign subscription"})
		return
	}

	ctx.JSON(http.StatusCreated, types.Resp{Data: userSub})
}

// User handlers

func (c *SubscriptionController) ListActivePlans(ctx *gin.Context) {
	plans, err := c.svc.Queries.ListSubscriptions(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to list plans"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: plans})
}

func (c *SubscriptionController) GetMySubscriptions(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	subs, err := c.svc.Queries.ListUserSubscriptions(ctx.Request.Context(), user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to get subscriptions"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: subs})
}

func (c *SubscriptionController) GetMyActiveSubscription(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	sub, err := c.svc.Queries.GetActiveUserSubscription(ctx.Request.Context(), user.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "No active subscription"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: sub})
}
