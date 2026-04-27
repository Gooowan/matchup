package moderation

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Gooowan/matchup/modules/core/logging"
	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
	gen "github.com/Gooowan/matchup/modules/moderation/gen"
	core "github.com/Gooowan/matchup/modules/users"
	"github.com/Gooowan/matchup/modules/users/auth"
	usergen "github.com/Gooowan/matchup/modules/users/gen"
)

type ModerationController struct {
	svc     *ModerationService
	userSvc *core.UserService
}

func NewModerationController(svc *ModerationService, userSvc *core.UserService) *ModerationController {
	return &ModerationController{svc: svc, userSvc: userSvc}
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

func (c *ModerationController) AdminListReports(ctx *gin.Context) {
	reports, err := c.svc.ListAllReports(ctx.Request.Context(), 100)
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to list reports", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "failed to list reports"})
		return
	}
	if reports == nil {
		reports = []ReportRow{}
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: reports})
}

func (c *ModerationController) AdminBanUser(ctx *gin.Context) {
	targetID, err := utils.StringToUUID(ctx.Param("userId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "invalid user ID"})
		return
	}

	if err := c.userSvc.Queries.UpdateUserRole(ctx.Request.Context(), usergen.UpdateUserRoleParams{
		Role:   "banned",
		UserID: targetID,
	}); err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to ban user", "error", err, "user_id", ctx.Param("userId"))
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "failed to ban user"})
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
