package core

import (
	"net/http"

	"github.com/gin-gonic/gin"

	gen "github.com/Gooowan/matchup/modules/core/gen"
	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
)

type AdminController struct {
	service *CoreService
}

func NewAdminController(service *CoreService) *AdminController {
	return &AdminController{
		service: service,
	}
}

func (c *AdminController) SearchUsers(ctx *gin.Context) {
	paginationParams := types.ParsePaginationParams(ctx)
	searchTerm := ctx.Query("q")

	users, err := c.service.Queries.AdminSearchUsers(ctx.Request.Context(), gen.AdminSearchUsersParams{
		SearchTerm: searchTerm,
		OffsetVal:  paginationParams.Offset(),
		LimitVal:   paginationParams.Limit(),
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to search users"})
		return
	}

	var totalCount int64 = 0
	if len(users) > 0 {
		totalCount = users[0].TotalCount
	}

	var responseUsers []*gen.AdminUserDTO
	for _, user := range users {
		responseUsers = append(responseUsers, user.ToAdminDTO())
	}

	response := types.NewPaginatedResp(paginationParams, responseUsers, totalCount)
	ctx.JSON(http.StatusOK, response)
}

func (c *AdminController) GetUser(ctx *gin.Context) {
	userIDStr := ctx.Param("id")

	userID, err := utils.StringToUUID(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid user ID format"})
		return
	}

	user, err := c.service.Queries.AdminGetUser(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "User not found"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: user.ToAdminDTO()})
}

func (c *AdminController) UpdateUser(ctx *gin.Context) {
	userIDStr := ctx.Param("id")

	userID, err := utils.StringToUUID(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid user ID format"})
		return
	}

	var req struct {
		ReplenishWallet *string `json:"replenish_wallet"`
		Comment         *string `json:"comment"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	updateParams := gen.AdminUpdateUserParams{
		UserID: userID,
	}

	if req.ReplenishWallet != nil || req.Comment != nil {
		metadataUpdate := types.JSONB{}

		if req.ReplenishWallet != nil {
			metadataUpdate["replenish_wallet"] = *req.ReplenishWallet
		}

		if req.Comment != nil {
			metadataUpdate["comment"] = *req.Comment
		}

		updateParams.Metadata = types.JSONB(metadataUpdate)
	}

	err = c.service.Queries.AdminUpdateUser(ctx.Request.Context(), updateParams)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to update user"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "User updated successfully"})
}

func (c *AdminController) GetStats(ctx *gin.Context) {
	totalUsers, err := c.service.Queries.AdminGetTotalUsers(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to get total users"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{
		Data: gin.H{
			"totalUsers": totalUsers,
		},
	})
}

func (c *AdminController) RegisterRoutes(rg *gin.RouterGroup, adminAuthMiddleware gin.HandlerFunc) {
	rg.Use(adminAuthMiddleware)
	rg.GET("/stats", c.GetStats)
	rg.GET("/users/search", c.SearchUsers)
	rg.GET("/users/:id", c.GetUser)
	rg.PUT("/users/:id", c.UpdateUser)
}
