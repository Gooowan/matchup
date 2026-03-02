package recommendation

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
	gen "github.com/Gooowan/matchup/modules/recommendation/gen"
	"github.com/Gooowan/matchup/modules/users/auth"
)

type RecommendationController struct {
	svc *RecommendationService
}

func NewRecommendationController(svc *RecommendationService) *RecommendationController {
	return &RecommendationController{svc: svc}
}

func (c *RecommendationController) GetProfile(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	profile, err := c.svc.GetProfile(ctx.Request.Context(), user.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Profile not found"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: profile.ToDTO()})
}

func (c *RecommendationController) CreateOrUpdateProfile(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	var req struct {
		DanceStyles []string `json:"dance_styles"`
		Latitude    float64  `json:"latitude"`
		Longitude   float64  `json:"longitude"`
		Visible     *bool    `json:"visible"`
		// Fields stored in JSONB data
		DanceRole  string `json:"dance_role"`
		DanceLevel string `json:"dance_level"`
		HeightCm   int32  `json:"height_cm"`
		Bio        string `json:"bio"`
		BirthDate  string `json:"birth_date"`
		Gender     string `json:"gender"`
		City       string `json:"city"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	visible := true
	if req.Visible != nil {
		visible = *req.Visible
	}

	// Build JSONB data from request fields
	data := types.JSONB{}
	if req.DanceRole != "" {
		data["dance_role"] = req.DanceRole
	}
	if req.DanceLevel != "" {
		data["dance_level"] = req.DanceLevel
	}
	if req.HeightCm > 0 {
		data["height_cm"] = req.HeightCm
	}
	if req.Bio != "" {
		data["bio"] = req.Bio
	}
	if req.BirthDate != "" {
		data["birth_date"] = req.BirthDate
	}
	if req.Gender != "" {
		data["gender"] = req.Gender
	}
	if req.City != "" {
		data["city"] = req.City
	}

	// Try update first, create if not exists
	existing, err := c.svc.GetProfile(ctx.Request.Context(), user.ID)
	if err != nil {
		// Create
		profile, err := c.svc.CreateProfile(ctx.Request.Context(), user.ID, gen.CreateProfileParams{
			DanceStyles: req.DanceStyles,
			Latitude:    pgtype.Float8{Float64: req.Latitude, Valid: req.Latitude != 0},
			Longitude:   pgtype.Float8{Float64: req.Longitude, Valid: req.Longitude != 0},
			Visible:     visible,
			Data:        data,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to create profile"})
			return
		}
		ctx.JSON(http.StatusCreated, types.Resp{Data: profile.ToDTO()})
		return
	}

	// Update — merge JSONB data with existing
	existingData := getProfileData(existing.Data)
	for k, v := range data {
		existingData[k] = v
	}

	styles := req.DanceStyles
	if styles == nil {
		styles = existing.DanceStyles
	}

	lat := pgtype.Float8{Float64: req.Latitude, Valid: req.Latitude != 0}
	if !lat.Valid {
		lat = existing.Latitude
	}
	lon := pgtype.Float8{Float64: req.Longitude, Valid: req.Longitude != 0}
	if !lon.Valid {
		lon = existing.Longitude
	}

	err = c.svc.UpdateProfile(ctx.Request.Context(), user.ID, gen.UpdateProfileParams{
		DanceStyles: styles,
		Latitude:    lat,
		Longitude:   lon,
		Visible:     visible,
		Data:        existingData,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to update profile"})
		return
	}

	updated, _ := c.svc.GetProfile(ctx.Request.Context(), user.ID)
	ctx.JSON(http.StatusOK, types.Resp{Data: updated.ToDTO()})
}

func (c *RecommendationController) GetProfilePreview(ctx *gin.Context) {
	targetID, err := utils.StringToUUID(ctx.Param("userId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid user ID"})
		return
	}

	preview, err := c.svc.GetProfilePreview(ctx.Request.Context(), targetID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Profile not found"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: preview.ToDTO()})
}

func (c *RecommendationController) GetPreferences(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	prefs, err := c.svc.GetPreferences(ctx.Request.Context(), user.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Preferences not found"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: prefs.ToDTO()})
}

func (c *RecommendationController) UpdatePreferences(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	// Accept the entire preferences as a JSONB object
	var req types.JSONB
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	prefs, err := c.svc.UpsertPreferences(ctx.Request.Context(), user.ID, gen.UpsertPreferencesParams{
		Data: req,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to update preferences"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: prefs.ToDTO()})
}

func (c *RecommendationController) AddMedia(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	if err := c.svc.AddMediaURL(ctx.Request.Context(), user.ID, req.URL); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to add media"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *RecommendationController) RemoveMedia(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	if err := c.svc.RemoveMediaURL(ctx.Request.Context(), user.ID, req.URL); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to remove media"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *RecommendationController) RegisterRoutes(rg *gin.RouterGroup, userAuth gin.HandlerFunc) {
	rg.Use(userAuth)
	rg.GET("/profile", c.GetProfile)
	rg.PUT("/profile", c.CreateOrUpdateProfile)
	rg.POST("/profile/media", c.AddMedia)
	rg.DELETE("/profile/media", c.RemoveMedia)
	rg.GET("/preferences", c.GetPreferences)
	rg.PUT("/preferences", c.UpdatePreferences)
}
