package recommendation

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	clubsgen "github.com/Gooowan/matchup/modules/clubs/gen"
	"github.com/Gooowan/matchup/modules/core/geocoding"
	"github.com/Gooowan/matchup/modules/core/logging"
	"github.com/Gooowan/matchup/modules/core/types"
	corehttp "github.com/Gooowan/matchup/modules/core/http"
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
		DanceStyles    []string `json:"dance_styles"`
		Latitude       float64  `json:"latitude"  binding:"omitempty,latitude"`
		Longitude      float64  `json:"longitude" binding:"omitempty,longitude"`
		Visible        *bool    `json:"visible"`
		Gender         string   `json:"gender"     binding:"omitempty,max=20"`
		BirthDate      string   `json:"birth_date" binding:"omitempty,datetime=2006-01-02"`
		HeightCm       *int16   `json:"height_cm"  binding:"omitempty,min=100,max=250"`
		Goal           string   `json:"goal"       binding:"omitempty,max=50"`
		Program        string   `json:"program"    binding:"omitempty,max=50"`
		Categories     []string `json:"categories"`
		Country        string   `json:"country"    binding:"omitempty,max=100"`
		City           string   `json:"city"       binding:"omitempty,max=100"`
		PrimaryClubID  *string  `json:"primary_club_id"`
		ReadyToRelocate *bool   `json:"ready_to_relocate"`
		ReadyToFinance  string  `json:"ready_to_finance" binding:"omitempty,max=20"`
		// Non-filterable fields stored in metadata JSONB
		Bio         string `json:"bio"          binding:"omitempty,max=1000"`
		AccountType string `json:"account_type" binding:"omitempty,max=20"`
		Role        string `json:"role"         binding:"omitempty,max=20"`
	}
	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	if req.BirthDate != "" && req.AccountType != "parent" {
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

	// visible defaults to true only for new profiles;
	// for updates, the existing value is preserved when the field is omitted.
	// Set a sentinel; resolved below once we know if this is create or update.
	var visibleOverride *bool = req.Visible

	categoriesProvided := req.Categories != nil
	if req.Categories == nil {
		req.Categories = []string{}
	}
	if req.DanceStyles == nil {
		req.DanceStyles = []string{}
	}

	// Resolve primary club and derive city/country + coordinates from it.
	// Coordinates are always locked to the club (or city centroid) — never client-supplied.
	var resolvedPrimaryClubID pgtype.UUID
	var clubLat, clubLng float64
	var clubResolved bool
	if req.PrimaryClubID != nil && *req.PrimaryClubID != "" {
		if clubID, err := utils.StringToUUID(*req.PrimaryClubID); err == nil {
			clubQueries := clubsgen.New(c.svc.DB)
			if club, err := clubQueries.GetClubByID(ctx.Request.Context(), clubID); err == nil {
				req.Country = club.Country
				req.City = club.City
				resolvedPrimaryClubID = clubID
				if club.Latitude != 0 || club.Longitude != 0 {
					clubLat, clubLng = club.Latitude, club.Longitude
					clubResolved = true
				}
			}
		}
	}
	// When no club (or club has no coords), fall back to city centroid.
	if !clubResolved {
		country := req.Country
		city := req.City
		clubLat, clubLng = geocoding.CityLatLng(country, city)
	}

	existing, err := c.svc.Queries.GetProfileByUserID(ctx.Request.Context(), user.ID)
	if err != nil {
		// Create: default visible = true when not specified.
		createVisible := true
		if visibleOverride != nil {
			createVisible = *visibleOverride
		}
		params := gen.CreateProfileParams{
			UserID:        user.ID,
			AccountType:   orDefault(req.AccountType, "dancer"),
			DanceStyles:   req.DanceStyles,
			Latitude:      pgtype.Float8{Float64: clubLat, Valid: true},
			Longitude:     pgtype.Float8{Float64: clubLng, Valid: true},
			Visible:       createVisible,
			Gender:        req.Gender,
			Goal:          orDefault(req.Goal, "hobby"),
			Program:       orDefault(req.Program, "standard"),
			Categories:    req.Categories,
			PrimaryClubID: resolvedPrimaryClubID,
			Metadata:      types.JSONB{},
			Data:          types.JSONB{},
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
			params.Metadata["bio"] = req.Bio
		}
		// Keep metadata in sync for guard checks that still read from JSONB.
		params.Metadata["account_type"] = params.AccountType
		if req.Role != "" {
			params.Metadata["role"] = req.Role
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

	// Update — keep existing values where new request omits them.
	// Coordinates are always re-derived from the club (or city centroid), never from client.
	styles := req.DanceStyles
	if styles == nil {
		styles = existing.DanceStyles
	}
	// Re-compute coords; if the club/city hasn't changed they'll be the same value.
	lat := pgtype.Float8{Float64: clubLat, Valid: true}
	lon := pgtype.Float8{Float64: clubLng, Valid: true}

	// Merge bio into existing metadata
	metadata := existing.Metadata
	if metadata == nil {
		metadata = types.JSONB{}
	}
	if req.Bio != "" {
		metadata["bio"] = req.Bio
	}
	if req.AccountType != "" {
		metadata["account_type"] = req.AccountType
	}
	if req.Role != "" {
		metadata["role"] = req.Role
	}

	// Keep existing primary_club_id if not explicitly updated.
	primaryClubID := resolvedPrimaryClubID
	if !primaryClubID.Valid && req.PrimaryClubID == nil {
		primaryClubID = existing.PrimaryClubID
	}

	// Update: preserve existing visibility when the field is omitted in the request.
	updateVisible := existing.Visible
	if visibleOverride != nil {
		updateVisible = *visibleOverride
	}

	// Resolve account_type: use request value if provided, fall back to existing column.
	updatedAccountType := existing.AccountType
	if req.AccountType != "" {
		updatedAccountType = req.AccountType
	}
	metadata["account_type"] = updatedAccountType

	params := gen.UpdateProfileParams{
		UserID:        user.ID,
		AccountType:   updatedAccountType,
		DanceStyles:   styles,
		Latitude:      lat,
		Longitude:     lon,
		Visible:       updateVisible,
		Gender:        orDefault(req.Gender, existing.Gender),
		Goal:          orDefault(req.Goal, existing.Goal),
		Program:       orDefault(req.Program, existing.Program),
		Categories:    func() []string { if categoriesProvided { return req.Categories }; return existing.Categories }(),
		PrimaryClubID: primaryClubID,
		Metadata:      metadata,
		Data:          existing.Data,
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
	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	// Load existing preferences so we only override supplied fields (prevents
	// a partial PUT from wiping previously saved age/height/etc.).
	existing, _ := c.svc.Queries.GetPreferences(ctx.Request.Context(), user.ID)
	params := gen.UpsertPreferencesParams{
		UserID:                 user.ID,
		PreferredGender:        existing.PreferredGender,
		AgeMin:                 existing.AgeMin,
		AgeMax:                 existing.AgeMax,
		HeightMin:              existing.HeightMin,
		HeightMax:              existing.HeightMax,
		PreferredGoal:          existing.PreferredGoal,
		PreferredProgram:       existing.PreferredProgram,
		PreferredCategories:    existing.PreferredCategories,
		PreferredCountry:       existing.PreferredCountry,
		PreferredCity:          existing.PreferredCity,
		WantsPartnerToRelocate: existing.WantsPartnerToRelocate,
		WantsPartnerToFinance:  existing.WantsPartnerToFinance,
		Metadata:               types.JSONB{},
		Data:                   types.JSONB{},
	}

	// Override only the fields explicitly provided in the request.
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
	if req.PreferredCategories != nil {
		params.PreferredCategories = req.PreferredCategories
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

// ResetPreferences clears all filter columns to NULL and re-seeds the locked
// default city (Київ), so the feed shows the full proximity pool again.
func (c *RecommendationController) ResetPreferences(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	params := gen.UpsertPreferencesParams{
		UserID: user.ID,
		// All filter columns intentionally left as zero-values (NULL) to remove filters.
		// Re-seed the locked city so the feed doesn't fall back to a global pool.
		PreferredCity: pgtype.Text{String: "Київ", Valid: true},
		Metadata:      types.JSONB{},
		Data:          types.JSONB{},
	}

	prefs, err := c.svc.Queries.UpsertPreferences(ctx.Request.Context(), params)
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to reset preferences", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to reset preferences"})
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
	if !corehttp.BindJSON(ctx, &req) {
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
	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	if err := c.svc.RemoveMediaURL(ctx.Request.Context(), user.ID, req.URL); err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to remove media URL", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to remove media"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *RecommendationController) ListTrainers(ctx *gin.Context) {
	limit := int32(20)
	offset := int32(0)
	if l := ctx.Query("limit"); l != "" {
		if v, err := utils.ParseInt32(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := ctx.Query("offset"); o != "" {
		if v, err := utils.ParseInt32(o); err == nil && v >= 0 {
			offset = v
		}
	}

	trainers, err := c.svc.Queries.ListTrainers(ctx.Request.Context(), gen.ListTrainersParams{
		LimitVal:  limit,
		OffsetVal: offset,
	})
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to list trainers", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to list trainers"})
		return
	}
	dtos := make([]gen.TrainerCardDTO, 0, len(trainers))
	for _, t := range trainers {
		dtos = append(dtos, t.ToTrainerCardDTO())
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: dtos})
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
	rg.DELETE("/preferences", c.ResetPreferences)
	rg.GET("/trainers", c.ListTrainers)
}

// orDefault returns val if non-empty, otherwise fallback.
func orDefault(val, fallback string) string {
	if val != "" {
		return val
	}
	return fallback
}

