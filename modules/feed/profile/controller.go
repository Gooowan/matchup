// package profile

// import (
// 	"net/http"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/jackc/pgx/v5/pgtype"

// 	"github.com/Gooowan/matchup/modules/core/types"
// 	"github.com/Gooowan/matchup/modules/core/utils"
// 	gen "github.com/Gooowan/matchup/modules/matchup/gen"
// 	"github.com/Gooowan/matchup/modules/users/auth"
// )

// type ProfileController struct {
// 	svc *ProfileService
// }

// func NewProfileController(svc *ProfileService) *ProfileController {
// 	return &ProfileController{svc: svc}
// }

// func (c *ProfileController) GetProfile(ctx *gin.Context) {
// 	user, ok := auth.GetUserFromContext(ctx)
// 	if !ok {
// 		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
// 		return
// 	}

// 	profile, err := c.svc.GetProfile(ctx.Request.Context(), user.ID)
// 	if err != nil {
// 		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Profile not found"})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, types.Resp{Data: profile.ToDTO()})
// }

// func (c *ProfileController) CreateOrUpdateProfile(ctx *gin.Context) {
// 	user, ok := auth.GetUserFromContext(ctx)
// 	if !ok {
// 		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
// 		return
// 	}

// 	var req struct {
// 		DanceStyles []string `json:"dance_styles"`
// 		DanceRole   string   `json:"dance_role"`
// 		DanceLevel  string   `json:"dance_level"`
// 		HeightCm    int32    `json:"height_cm"`
// 		Bio         string   `json:"bio"`
// 		BirthDate   string   `json:"birth_date"`
// 		Gender      string   `json:"gender"`
// 		City        string   `json:"city"`
// 		Latitude    float64  `json:"latitude"`
// 		Longitude   float64  `json:"longitude"`
// 		Visible     *bool    `json:"visible"`
// 	}
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
// 		return
// 	}

// 	visible := true
// 	if req.Visible != nil {
// 		visible = *req.Visible
// 	}

// 	var birthDate pgtype.Date
// 	if req.BirthDate != "" {
// 		parsed, err := parseDate(req.BirthDate)
// 		if err != nil {
// 			ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid birth_date format, use YYYY-MM-DD"})
// 			return
// 		}
// 		birthDate = pgtype.Date{Time: parsed, Valid: true}
// 	}

// 	// Try update first, create if not exists
// 	existing, err := c.svc.GetProfile(ctx.Request.Context(), user.ID)
// 	if err != nil {
// 		// Create
// 		profile, err := c.svc.CreateProfile(ctx.Request.Context(), user.ID, gen.CreateProfileParams{
// 			DanceStyles: req.DanceStyles,
// 			DanceRole:   pgtype.Text{String: req.DanceRole, Valid: req.DanceRole != ""},
// 			DanceLevel:  pgtype.Text{String: req.DanceLevel, Valid: req.DanceLevel != ""},
// 			HeightCm:    pgtype.Int4{Int32: req.HeightCm, Valid: req.HeightCm > 0},
// 			Bio:         pgtype.Text{String: req.Bio, Valid: req.Bio != ""},
// 			BirthDate:   birthDate,
// 			Gender:      pgtype.Text{String: req.Gender, Valid: req.Gender != ""},
// 			City:        pgtype.Text{String: req.City, Valid: req.City != ""},
// 			Latitude:    pgtype.Float8{Float64: req.Latitude, Valid: req.Latitude != 0},
// 			Longitude:   pgtype.Float8{Float64: req.Longitude, Valid: req.Longitude != 0},
// 			Visible:     visible,
// 		})
// 		if err != nil {
// 			ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to create profile"})
// 			return
// 		}
// 		ctx.JSON(http.StatusCreated, types.Resp{Data: profile.ToDTO()})
// 		return
// 	}

// 	// Update — merge with existing
// 	styles := req.DanceStyles
// 	if styles == nil {
// 		styles = existing.DanceStyles
// 	}

// 	err = c.svc.UpdateProfile(ctx.Request.Context(), user.ID, gen.UpdateProfileParams{
// 		DanceStyles: styles,
// 		DanceRole:   textOrExisting(req.DanceRole, existing.DanceRole),
// 		DanceLevel:  textOrExisting(req.DanceLevel, existing.DanceLevel),
// 		HeightCm:    int4OrExisting(req.HeightCm, existing.HeightCm),
// 		Bio:         textOrExisting(req.Bio, existing.Bio),
// 		BirthDate:   dateOrExisting(birthDate, existing.BirthDate),
// 		Gender:      textOrExisting(req.Gender, existing.Gender),
// 		City:        textOrExisting(req.City, existing.City),
// 		Latitude:    float8OrExisting(req.Latitude, existing.Latitude),
// 		Longitude:   float8OrExisting(req.Longitude, existing.Longitude),
// 		Visible:     visible,
// 	})
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to update profile"})
// 		return
// 	}

// 	updated, _ := c.svc.GetProfile(ctx.Request.Context(), user.ID)
// 	ctx.JSON(http.StatusOK, types.Resp{Data: updated.ToDTO()})
// }

// func (c *ProfileController) GetProfilePreview(ctx *gin.Context) {
// 	targetID, err := utils.StringToUUID(ctx.Param("userId"))
// 	if err != nil {
// 		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid user ID"})
// 		return
// 	}

// 	preview, err := c.svc.GetProfilePreview(ctx.Request.Context(), targetID)
// 	if err != nil {
// 		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Profile not found"})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, types.Resp{Data: preview.ToDTO()})
// }

// func (c *ProfileController) GetPreferences(ctx *gin.Context) {
// 	user, ok := auth.GetUserFromContext(ctx)
// 	if !ok {
// 		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
// 		return
// 	}

// 	prefs, err := c.svc.GetPreferences(ctx.Request.Context(), user.ID)
// 	if err != nil {
// 		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Preferences not found"})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, types.Resp{Data: prefs.ToDTO()})
// }

// func (c *ProfileController) UpdatePreferences(ctx *gin.Context) {
// 	user, ok := auth.GetUserFromContext(ctx)
// 	if !ok {
// 		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
// 		return
// 	}

// 	var req struct {
// 		PreferredStyles  []string `json:"preferred_styles"`
// 		PreferredRole    string   `json:"preferred_role"`
// 		MinLevel         string   `json:"min_level"`
// 		MaxLevel         string   `json:"max_level"`
// 		MinHeightCm      int32    `json:"min_height_cm"`
// 		MaxHeightCm      int32    `json:"max_height_cm"`
// 		MinAge           int32    `json:"min_age"`
// 		MaxAge           int32    `json:"max_age"`
// 		MaxDistanceKm    float64  `json:"max_distance_km"`
// 		GenderPreference string   `json:"gender_preference"`
// 	}
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
// 		return
// 	}

// 	prefs, err := c.svc.UpsertPreferences(ctx.Request.Context(), user.ID, gen.UpsertPreferencesParams{
// 		PreferredStyles:  req.PreferredStyles,
// 		PreferredRole:    pgtype.Text{String: req.PreferredRole, Valid: req.PreferredRole != ""},
// 		MinLevel:         pgtype.Text{String: req.MinLevel, Valid: req.MinLevel != ""},
// 		MaxLevel:         pgtype.Text{String: req.MaxLevel, Valid: req.MaxLevel != ""},
// 		MinHeightCm:      pgtype.Int4{Int32: req.MinHeightCm, Valid: req.MinHeightCm > 0},
// 		MaxHeightCm:      pgtype.Int4{Int32: req.MaxHeightCm, Valid: req.MaxHeightCm > 0},
// 		MinAge:           pgtype.Int4{Int32: req.MinAge, Valid: req.MinAge > 0},
// 		MaxAge:           pgtype.Int4{Int32: req.MaxAge, Valid: req.MaxAge > 0},
// 		MaxDistanceKm:    pgtype.Float8{Float64: req.MaxDistanceKm, Valid: req.MaxDistanceKm > 0},
// 		GenderPreference: pgtype.Text{String: req.GenderPreference, Valid: req.GenderPreference != ""},
// 	})
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to update preferences"})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, types.Resp{Data: prefs.ToDTO()})
// }

// func (c *ProfileController) AddMedia(ctx *gin.Context) {
// 	user, ok := auth.GetUserFromContext(ctx)
// 	if !ok {
// 		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
// 		return
// 	}

// 	var req struct {
// 		URL string `json:"url" binding:"required"`
// 	}
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
// 		return
// 	}

// 	if err := c.svc.AddMediaURL(ctx.Request.Context(), user.ID, req.URL); err != nil {
// 		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to add media"})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
// }

// func (c *ProfileController) RemoveMedia(ctx *gin.Context) {
// 	user, ok := auth.GetUserFromContext(ctx)
// 	if !ok {
// 		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
// 		return
// 	}

// 	var req struct {
// 		URL string `json:"url" binding:"required"`
// 	}
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
// 		return
// 	}

// 	if err := c.svc.RemoveMediaURL(ctx.Request.Context(), user.ID, req.URL); err != nil {
// 		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to remove media"})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
// }

// func (c *ProfileController) RegisterRoutes(rg *gin.RouterGroup, userAuth gin.HandlerFunc) {
// 	rg.Use(userAuth)
// 	rg.GET("/profile", c.GetProfile)
// 	rg.PUT("/profile", c.CreateOrUpdateProfile)
// 	rg.POST("/profile/media", c.AddMedia)
// 	rg.DELETE("/profile/media", c.RemoveMedia)
// 	rg.GET("/preferences", c.GetPreferences)
// 	rg.PUT("/preferences", c.UpdatePreferences)
// }

// // helpers

// func textOrExisting(val string, existing pgtype.Text) pgtype.Text {
// 	if val != "" {
// 		return pgtype.Text{String: val, Valid: true}
// 	}
// 	return existing
// }

// func int4OrExisting(val int32, existing pgtype.Int4) pgtype.Int4 {
// 	if val > 0 {
// 		return pgtype.Int4{Int32: val, Valid: true}
// 	}
// 	return existing
// }

// func float8OrExisting(val float64, existing pgtype.Float8) pgtype.Float8 {
// 	if val != 0 {
// 		return pgtype.Float8{Float64: val, Valid: true}
// 	}
// 	return existing
// }

// func dateOrExisting(val pgtype.Date, existing pgtype.Date) pgtype.Date {
// 	if val.Valid {
// 		return val
// 	}
// 	return existing
// }

// func parseDate(s string) (time.Time, error) {
// 	return time.Parse("2006-01-02", s)
// }
