package mapmod

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
	gen "github.com/Gooowan/matchup/modules/map/gen"
	"github.com/Gooowan/matchup/modules/users/auth"
)

type MapController struct {
	svc *MapService
}

func NewMapController(svc *MapService) *MapController {
	return &MapController{svc: svc}
}

func (c *MapController) UpdateLocation(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	var req struct {
		Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
		Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	loc, err := c.svc.Queries.UpsertUserLocation(ctx.Request.Context(), gen.UpsertUserLocationParams{
		UserID:    user.ID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to update location"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: loc.ToDTO()})
}

func (c *MapController) GetLocation(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	loc, err := c.svc.Queries.GetUserLocation(ctx.Request.Context(), user.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Location not found"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: loc.ToDTO()})
}

func (c *MapController) DeleteLocation(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	if err := c.svc.Queries.DeleteUserLocation(ctx.Request.Context(), user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to delete location"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *MapController) FindNearbyByCount(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	var req struct {
		Latitude   float64 `json:"latitude" binding:"required,min=-90,max=90"`
		Longitude  float64 `json:"longitude" binding:"required,min=-180,max=180"`
		MaxResults int32   `json:"max_results" binding:"required,min=1,max=100"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	rows, err := c.svc.Queries.FindNearbyUsersByCount(ctx.Request.Context(), gen.FindNearbyUsersByCountParams{
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		UserID:     user.ID,
		MaxResults: req.MaxResults,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to find nearby users"})
		return
	}

	dtos := make([]gen.NearbyUserDTO, len(rows))
	for i, r := range rows {
		dtos[i] = gen.NearbyUserDTO{
			UserID:     utils.UUIDToString(r.UserID),
			Latitude:   r.Latitude,
			Longitude:  r.Longitude,
			DistanceKm: r.DistanceKm,
			UpdatedAt:  r.UpdatedAt.Time.UnixMilli(),
		}
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: dtos})
}

func (c *MapController) FindNearbyByRadius(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	var req struct {
		Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
		Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
		RadiusKm  float64 `json:"radius_km" binding:"required,min=0.1,max=500"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	rows, err := c.svc.Queries.FindNearbyUsersWithinRadius(ctx.Request.Context(), gen.FindNearbyUsersWithinRadiusParams{
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		UserID:    user.ID,
		RadiusKm:  req.RadiusKm,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to find nearby users"})
		return
	}

	dtos := make([]gen.NearbyUserDTO, len(rows))
	for i, r := range rows {
		dtos[i] = gen.NearbyUserDTO{
			UserID:     utils.UUIDToString(r.UserID),
			Latitude:   r.Latitude,
			Longitude:  r.Longitude,
			DistanceKm: r.DistanceKm,
			UpdatedAt:  r.UpdatedAt.Time.UnixMilli(),
		}
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: dtos})
}

func (c *MapController) RegisterRoutes(rg *gin.RouterGroup, userAuth gin.HandlerFunc) {
	rg.Use(userAuth)

	rg.POST("/location", c.UpdateLocation)
	rg.GET("/location", c.GetLocation)
	rg.DELETE("/location", c.DeleteLocation)

	rg.POST("/nearby/count", c.FindNearbyByCount)
	rg.POST("/nearby/radius", c.FindNearbyByRadius)
}
