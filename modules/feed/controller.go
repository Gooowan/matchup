package feed

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Gooowan/matchup/modules/core/logging"
	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
	recgen "github.com/Gooowan/matchup/modules/recommendation/gen"
	"github.com/Gooowan/matchup/modules/users/auth"
)

type FeedController struct {
	svc *FeedService
}

func NewFeedController(svc *FeedService) *FeedController {
	return &FeedController{svc: svc}
}

func (c *FeedController) GetFeed(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	limit := int32(20)
	if l := ctx.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = int32(parsed)
		}
	}

	candidates, err := c.svc.GetFeed(ctx.Request.Context(), user.ID, limit)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	dtos := make([]recgen.FeedCandidateDTO, len(candidates))
	for i, c := range candidates {
		dtos[i] = c.ToFeedDTO()
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"candidates": dtos}})
}

func (c *FeedController) Swipe(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	var req struct {
		TargetUserID string `json:"target_user_id" binding:"required"`
		Action       string `json:"action" binding:"required,oneof=LIKE PASS"`
		Source       string `json:"source"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	targetID, err := utils.StringToUUID(req.TargetUserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid target_user_id"})
		return
	}

	result, err := c.svc.Swipe(ctx.Request.Context(), user.ID, targetID, req.Action, req.Source)
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to process swipe",
			"error", err, "action", req.Action, "target_user_id", req.TargetUserID)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to process swipe"})
		return
	}

	resp := gin.H{"is_mutual_match": result.IsMutualMatch}
	if result.ChatID != nil {
		resp["chat_id"] = utils.UUIDToString(*result.ChatID)
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: resp})
}

func (c *FeedController) Hide(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	var req struct {
		TargetUserID string `json:"target_user_id" binding:"required"`
		Reason       string `json:"reason"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	targetID, err := utils.StringToUUID(req.TargetUserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid target_user_id"})
		return
	}

	_, err = c.svc.Swipe(ctx.Request.Context(), user.ID, targetID, "PASS", "")
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to hide user", "error", err, "target_user_id", req.TargetUserID)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to hide user"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *FeedController) RegisterRoutes(rg *gin.RouterGroup, userAuth gin.HandlerFunc, swipeRL ...gin.HandlerFunc) {
	rg.Use(userAuth)
	rg.GET("/feed", c.GetFeed)
	swipeHandlers := []gin.HandlerFunc{c.Swipe}
	if len(swipeRL) > 0 {
		swipeHandlers = append([]gin.HandlerFunc{swipeRL[0]}, swipeHandlers...)
	}
	rg.POST("/swipe", swipeHandlers...)
	rg.POST("/hide", c.Hide)
}
