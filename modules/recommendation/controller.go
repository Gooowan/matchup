package recommendation

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Gooowan/matchup/modules/core/logging"
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

	profile, err := c.svc.Queries.GetProfileByUserID(ctx.Request.Context(), user.ID)
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
		DanceStyles     []string `json:"dance_styles"`
		Latitude        float64  `json:"latitude"`
		Longitude       float64  `json:"longitude"`
		Visible         *bool    `json:"visible"`
		Gender          string   `json:"gender"`
		BirthDate       string   `json:"birth_date"`
		HeightCm        *int16   `json:"height_cm"`
		Goal            string   `json:"goal"`
		Program         string   `json:"program"`
		Categories      []string `json:"categories"`
		Country         string   `json:"country"`
		City            string   `json:"city"`
		ReadyToRelocate *bool    `json:"ready_to_relocate"`
		ReadyToFinance  string   `json:"ready_to_finance"`
		// Non-filterable fields stored in metadata JSONB
		Bio string `json:"bio"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	if req.BirthDate != "" {
		if dob, err := time.Parse("2006-01-02", req.BirthDate); err == nil {
			age := time.Now().Year() - dob.Year()
			if time.Now().YearDay() < dob.YearDay() {
				age--
			}
			if age < 18 {
				ctx.JSON(http.StatusBadRequest, types.Resp{Error: "You must be 18 or older to use MatchUp"})
				return
			}
		}
	}

	visible := true
	if req.Visible != nil {
		visible = *req.Visible
	}

	existing, err := c.svc.Queries.GetProfileByUserID(ctx.Request.Context(), user.ID)
	if err != nil {
		// Create
		params := gen.CreateProfileParams{
			UserID:      user.ID,
			DanceStyles: req.DanceStyles,
			Latitude:    pgtype.Float8{Float64: req.Latitude, Valid: req.Latitude != 0},
			Longitude:   pgtype.Float8{Float64: req.Longitude, Valid: req.Longitude != 0},
			Visible:     visible,
			Gender:      req.Gender,
			Goal:        orDefault(req.Goal, "hobby"),
			Program:     orDefault(req.Program, "standard"),
			Categories:  req.Categories,
			Metadata:    types.JSONB{},
		}
		if req.BirthDate != "" {
			if t, err := time.Parse("2006-01-02", req.BirthDate); err == nil {
				params.BirthDate = pgtype.Date{Time: t, Valid: true}
			}
		}
		if req.HeightCm != nil {
			params.HeightCm = pgtype.Int2{Int16: *req.HeightCm, Valid: true}
		}
		if req.Country != "" {
			params.Country = pgtype.Text{String: req.Country, Valid: true}
		}
		if req.City != "" {
			params.City = pgtype.Text{String: req.City, Valid: true}
		}
		if req.ReadyToRelocate != nil {
			params.ReadyToRelocate = pgtype.Bool{Bool: *req.ReadyToRelocate, Valid: true}
		}
		if req.ReadyToFinance != "" {
			params.ReadyToFinance = pgtype.Text{String: req.ReadyToFinance, Valid: true}
		}
		if req.Bio != "" {
			params.Metadata = types.JSONB{"bio": req.Bio}
		}

		profile, err := c.svc.Queries.CreateProfile(ctx.Request.Context(), params)
		if err != nil {
			logging.FromContext(ctx.Request.Context()).Error("failed to create profile", "error", err)
			ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to create profile"})
			return
		}
		ctx.JSON(http.StatusCreated, types.Resp{Data: profile.ToDTO()})
		return
	}

	// Update — keep existing values where new request omits them
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

	// Merge bio into existing metadata
	metadata := existing.Metadata
	if metadata == nil {
		metadata = types.JSONB{}
	}
	if req.Bio != "" {
		metadata["bio"] = req.Bio
	}

	params := gen.UpdateProfileParams{
		UserID:      user.ID,
		DanceStyles: styles,
		Latitude:    lat,
		Longitude:   lon,
		Visible:     visible,
		Gender:      orDefault(req.Gender, existing.Gender),
		Goal:        orDefault(req.Goal, existing.Goal),
		Program:     orDefault(req.Program, existing.Program),
		Categories:  orSlice(req.Categories, existing.Categories),
		Metadata:    metadata,
		Data:        existing.Data,
	}
	if req.BirthDate != "" {
		if t, err := time.Parse("2006-01-02", req.BirthDate); err == nil {
			params.BirthDate = pgtype.Date{Time: t, Valid: true}
		}
	} else {
		params.BirthDate = existing.BirthDate
	}
	if req.HeightCm != nil {
		params.HeightCm = pgtype.Int2{Int16: *req.HeightCm, Valid: true}
	} else {
		params.HeightCm = existing.HeightCm
	}
	if req.Country != "" {
		params.Country = pgtype.Text{String: req.Country, Valid: true}
	} else {
		params.Country = existing.Country
	}
	if req.City != "" {
		params.City = pgtype.Text{String: req.City, Valid: true}
	} else {
		params.City = existing.City
	}
	if req.ReadyToRelocate != nil {
		params.ReadyToRelocate = pgtype.Bool{Bool: *req.ReadyToRelocate, Valid: true}
	} else {
		params.ReadyToRelocate = existing.ReadyToRelocate
	}
	if req.ReadyToFinance != "" {
		params.ReadyToFinance = pgtype.Text{String: req.ReadyToFinance, Valid: true}
	} else {
		params.ReadyToFinance = existing.ReadyToFinance
	}

	if err := c.svc.Queries.UpdateProfile(ctx.Request.Context(), params); err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to update profile", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to update profile"})
		return
	}

	updated, err := c.svc.Queries.GetProfileByUserID(ctx.Request.Context(), user.ID)
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to fetch updated profile", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to fetch updated profile"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: updated.ToDTO()})
}

func (c *RecommendationController) GetProfilePreview(ctx *gin.Context) {
	targetID, err := utils.StringToUUID(ctx.Param("userId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid user ID"})
		return
	}

	preview, err := c.svc.Queries.GetProfilePreview(ctx.Request.Context(), targetID)
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

	prefs, err := c.svc.Queries.GetPreferences(ctx.Request.Context(), user.ID)
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

	var req struct {
		PreferredGender        string   `json:"preferred_gender"`
		AgeMin                 *int16   `json:"age_min"`
		AgeMax                 *int16   `json:"age_max"`
		HeightMin              *int16   `json:"height_min"`
		HeightMax              *int16   `json:"height_max"`
		PreferredGoal          string   `json:"preferred_goal"`
		PreferredProgram       string   `json:"preferred_program"`
		PreferredCategories    []string `json:"preferred_categories"`
		PreferredCountry       string   `json:"preferred_country"`
		PreferredCity          string   `json:"preferred_city"`
		WantsPartnerToRelocate *bool    `json:"wants_partner_to_relocate"`
		WantsPartnerToFinance  string   `json:"wants_partner_to_finance"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	params := gen.UpsertPreferencesParams{
		UserID:              user.ID,
		PreferredCategories: req.PreferredCategories,
		Metadata:            types.JSONB{},
	}
	if req.PreferredGender != "" {
		params.PreferredGender = pgtype.Text{String: req.PreferredGender, Valid: true}
	}
	if req.AgeMin != nil {
		params.AgeMin = pgtype.Int2{Int16: *req.AgeMin, Valid: true}
	}
	if req.AgeMax != nil {
		params.AgeMax = pgtype.Int2{Int16: *req.AgeMax, Valid: true}
	}
	if req.HeightMin != nil {
		params.HeightMin = pgtype.Int2{Int16: *req.HeightMin, Valid: true}
	}
	if req.HeightMax != nil {
		params.HeightMax = pgtype.Int2{Int16: *req.HeightMax, Valid: true}
	}
	if req.PreferredGoal != "" {
		params.PreferredGoal = pgtype.Text{String: req.PreferredGoal, Valid: true}
	}
	if req.PreferredProgram != "" {
		params.PreferredProgram = pgtype.Text{String: req.PreferredProgram, Valid: true}
	}
	if req.PreferredCountry != "" {
		params.PreferredCountry = pgtype.Text{String: req.PreferredCountry, Valid: true}
	}
	if req.PreferredCity != "" {
		params.PreferredCity = pgtype.Text{String: req.PreferredCity, Valid: true}
	}
	if req.WantsPartnerToRelocate != nil {
		params.WantsPartnerToRelocate = pgtype.Bool{Bool: *req.WantsPartnerToRelocate, Valid: true}
	}
	if req.WantsPartnerToFinance != "" {
		params.WantsPartnerToFinance = pgtype.Text{String: req.WantsPartnerToFinance, Valid: true}
	}

	prefs, err := c.svc.Queries.UpsertPreferences(ctx.Request.Context(), params)
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to update preferences", "error", err)
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
		logging.FromContext(ctx.Request.Context()).Error("failed to add media URL", "error", err)
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
		logging.FromContext(ctx.Request.Context()).Error("failed to remove media URL", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to remove media"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *RecommendationController) RegisterRoutes(rg *gin.RouterGroup, userAuth gin.HandlerFunc) {
	rg.Use(userAuth)
	rg.GET("/profile", c.GetProfile)
	rg.PUT("/profile", c.CreateOrUpdateProfile)
	rg.GET("/profile/:userId", c.GetProfilePreview)
	rg.POST("/profile/media", c.AddMedia)
	rg.DELETE("/profile/media", c.RemoveMedia)
	rg.GET("/preferences", c.GetPreferences)
	rg.PUT("/preferences", c.UpdatePreferences)
}

// orDefault returns val if non-empty, otherwise fallback.
func orDefault(val, fallback string) string {
	if val != "" {
		return val
	}
	return fallback
}

// orSlice returns val if non-nil/non-empty, otherwise fallback.
func orSlice(val, fallback []string) []string {
	if len(val) > 0 {
		return val
	}
	return fallback
}
