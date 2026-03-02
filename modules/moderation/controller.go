package moderation

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Gooowan/matchup/modules/users/auth"
	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
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

	if err := c.svc.BlockUser(ctx.Request.Context(), user.ID, targetID); err != nil {
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

	if err := c.svc.UnblockUser(ctx.Request.Context(), user.ID, targetID); err != nil {
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
