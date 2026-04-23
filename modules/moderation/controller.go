package moderation

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Gooowan/matchup/modules/core/logging"
	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
	gen "github.com/Gooowan/matchup/modules/moderation/gen"
	"github.com/Gooowan/matchup/modules/users/auth"
)

type ModerationController struct {
	svc *ModerationService
}

func NewModerationController(svc *ModerationService) *ModerationController {
	return &ModerationController{svc: svc}
}

func (c *ModerationController) BlockUser(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	targetID, err := utils.StringToUUID(ctx.Param("userId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid user ID"})
		return
	}

	if err := c.svc.Queries.CreateBlock(ctx.Request.Context(), gen.CreateBlockParams{
		BlockerID: user.ID,
		BlockedID: targetID,
	}); err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to block user", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to block user"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *ModerationController) UnblockUser(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	targetID, err := utils.StringToUUID(ctx.Param("userId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid user ID"})
		return
	}

	if err := c.svc.Queries.DeleteBlock(ctx.Request.Context(), gen.DeleteBlockParams{
		BlockerID: user.ID,
		BlockedID: targetID,
	}); err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to unblock user", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to unblock user"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *ModerationController) ReportUser(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	targetID, err := utils.StringToUUID(ctx.Param("userId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid user ID"})
		return
	}

	var req struct {
		Category string `json:"category" binding:"required"`
		Comment  string `json:"comment"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	if err := c.svc.ReportUser(ctx.Request.Context(), user.ID, targetID, req.Category, req.Comment); err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to report user", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to report user"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *ModerationController) RegisterRoutes(r *gin.Engine, userAuth gin.HandlerFunc) {
	users := r.Group("/users")
	users.Use(userAuth)
	users.POST("/:userId/block", c.BlockUser)
	users.DELETE("/:userId/block", c.UnblockUser)
	users.POST("/:userId/report", c.ReportUser)
}
