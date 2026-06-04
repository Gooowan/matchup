package push

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Gooowan/matchup/modules/core/types"
	corehttp "github.com/Gooowan/matchup/modules/core/http"
	"github.com/Gooowan/matchup/modules/users/auth"
)

type Controller struct {
	svc *Service
}

func NewController(svc *Service) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) RegisterRoutes(r *gin.RouterGroup, userAuth gin.HandlerFunc) {
	r.POST("/push-token", userAuth, c.registerToken)
}

func (c *Controller) registerToken(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "unauthorized"})
		return
	}

	var req struct {
		Token    string `json:"token" binding:"required"`
		Platform string `json:"platform"`
	}
	if !corehttp.BindJSON(ctx, &req) {
		return
	}
	if req.Platform == "" {
		req.Platform = "ios"
	}

	userIDStr := user.ID.String()
	if err := c.svc.RegisterToken(ctx.Request.Context(), userIDStr, req.Token, req.Platform); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "failed to save token"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}
