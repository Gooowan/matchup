package clubs

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	gen "github.com/Gooowan/matchup/modules/clubs/gen"
	"github.com/Gooowan/matchup/modules/chat"
	"github.com/Gooowan/matchup/modules/core/geocoding"
	"github.com/Gooowan/matchup/modules/core/gmaps"
	corehttp "github.com/Gooowan/matchup/modules/core/http"
	"github.com/Gooowan/matchup/modules/core/logging"
	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
	"github.com/Gooowan/matchup/modules/users/auth"
)

type ClubController struct {
	svc      *ClubService
	chatSvc  *chat.ChatService
	geocoder geocoding.Geocoder
	places   *gmaps.GooglePlacesClient
}

func NewClubController(svc *ClubService, chatSvc *chat.ChatService, geocoder geocoding.Geocoder, places *gmaps.GooglePlacesClient) *ClubController {
	return &ClubController{svc: svc, chatSvc: chatSvc, geocoder: geocoder, places: places}
}

// --- Request types ---

// clubRequest is used for admin create/update where lat/lng are required,
// and also for public RegisterClub where they are optional (geocoded server-side).
type clubRequest struct {
	Name         string         `json:"name"          binding:"required,min=2,max=255"`
	Description  string         `json:"description"   binding:"omitempty,max=2000"`
	Country      string         `json:"country"       binding:"required,max=100"`
	City         string         `json:"city"          binding:"required,max=100"`
	Address      string         `json:"address"       binding:"omitempty,max=500"`
	Latitude     float64        `json:"latitude"      binding:"omitempty,latitude"`
	Longitude    float64        `json:"longitude"     binding:"omitempty,longitude"`
	Website      string         `json:"website"       binding:"omitempty,url,max=500"`
	Phone        string         `json:"phone"         binding:"omitempty,max=50"`
	Photos       []string       `json:"photos"`
	WorkingHours map[string]any `json:"working_hours"`
}

func (r *clubRequest) toParams(verified bool) CreateClubParams {
	meta := types.JSONB{}
	if len(r.Photos) > 0 {
		meta = types.JSONB{
			"photos":   r.Photos,
			"logo_url": r.Photos[0], // first photo becomes the club avatar
		}
	}
	return CreateClubParams{
		Name:         r.Name,
		Description:  r.Description,
		Country:      r.Country,
		City:         r.City,
		Address:      r.Address,
		Latitude:     r.Latitude,
		Longitude:    r.Longitude,
		Website:      r.Website,
		Phone:        r.Phone,
		IsVerified:   verified,
		Metadata:     meta,
		WorkingHours: types.JSONB(r.WorkingHours),
	}
}

// parseGmapsRequest is the body for POST /clubs/parse-gmaps.
type parseGmapsRequest struct {
	URL string `json:"url" binding:"required"`
}

// POST /me/clubs/parse-gmaps — resolve a Google Maps link to club data.
func (c *ClubController) ParseGoogleMapsLink(ctx *gin.Context) {
	var req parseGmapsRequest
	if !corehttp.BindJSON(ctx, &req) {
		return
	}
	if c.places == nil || !c.places.IsConfigured() {
		ctx.JSON(http.StatusServiceUnavailable, types.Resp{Error: "Google Places not configured"})
		return
	}
	place, err := c.places.LookupURL(ctx.Request.Context(), req.URL)
	if err != nil {
		if errors.Is(err, gmaps.ErrQuotaExceeded) {
			ctx.JSON(http.StatusTooManyRequests, types.Resp{Error: "Daily import limit reached, try again tomorrow"})
			return
		}
		logging.FromContext(ctx.Request.Context()).Warn("gmaps lookup failed", "error", err)
		ctx.JSON(http.StatusUnprocessableEntity, types.Resp{Error: "Could not extract place data from that URL"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: place})
}

// maxClubsPerUser is the maximum number of clubs any single user can be associated with.
const maxClubsPerUser = 5

// --- Helpers ---

func pageParams(ctx *gin.Context) (limit, offset int32) {
	limit = 20
	offset = 0
	if l, err := strconv.Atoi(ctx.Query("limit")); err == nil && l > 0 && l <= 100 {
		limit = int32(l)
	}
	if p, err := strconv.Atoi(ctx.Query("page")); err == nil && p > 1 {
		offset = int32((p - 1)) * limit
	}
	return
}

// --- Public endpoints ---

// GET /clubs
func (c *ClubController) ListClubs(ctx *gin.Context) {
	limit, offset := pageParams(ctx)
	clubs, err := c.svc.ListClubs(ctx.Request.Context(),
		ctx.Query("country"), ctx.Query("city"), ctx.Query("q"), limit, offset)
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to list clubs", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to list clubs"})
		return
	}
	ctx.Header("Cache-Control", "public, max-age=60")
	ctx.JSON(http.StatusOK, types.Resp{Data: clubs})
}

// GET /clubs/:slug
func (c *ClubController) GetClub(ctx *gin.Context) {
	club, err := c.svc.GetClubBySlug(ctx.Request.Context(), ctx.Param("slug"))
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Club not found"})
		return
	}
	count, _ := c.svc.GetMemberCount(ctx.Request.Context(), club.ID)
	ctx.Header("Cache-Control", "public, max-age=60")
	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"club": club, "member_count": count}})
}

// slimTrainerCard is the public DTO for club trainer rows — no full JSONB blobs.
type slimTrainerCard struct {
	TrainerUserID string   `json:"trainer_user_id"`
	FirstName     string   `json:"first_name"`
	LastName      string   `json:"last_name,omitempty"`
	Avatar        string   `json:"avatar,omitempty"`
	Gender        string   `json:"gender"`
	City          string   `json:"city,omitempty"`
	Categories    []string `json:"categories"`
}

// GET /clubs/:slug/trainers
func (c *ClubController) ListClubTrainers(ctx *gin.Context) {
	club, err := c.svc.GetClubBySlug(ctx.Request.Context(), ctx.Param("slug"))
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Club not found"})
		return
	}
	tLimit, tOffset := pageParams(ctx)
	if tLimit == 0 || tLimit > 50 {
		tLimit = 50
	}
	rows, err := c.svc.Queries.ListClubTrainers(ctx.Request.Context(), gen.ListClubTrainersParams{
		ClubID:    club.ID,
		LimitVal:  tLimit,
		OffsetVal: tOffset,
	})
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to list club trainers", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to list trainers"})
		return
	}
	out := make([]slimTrainerCard, 0, len(rows))
	for _, r := range rows {
		pd := map[string]any(r.ProfileData)
		meta := map[string]any(r.Metadata)
		firstName, _ := pd["first_name"].(string)
		lastName, _ := pd["last_name"].(string)
		avatar, _ := meta["avatar"].(string)
		if avatar == "" {
			avatar, _ = pd["avatar"].(string)
		}
		city, _ := meta["city"].(string)
		out = append(out, slimTrainerCard{
			TrainerUserID: utils.UUIDToString(r.TrainerUserID),
			FirstName:     firstName,
			LastName:      lastName,
			Avatar:        avatar,
			Gender:        r.Gender,
			City:          city,
			Categories:    r.Categories,
		})
	}
	ctx.Header("Cache-Control", "public, max-age=60")
	ctx.JSON(http.StatusOK, types.Resp{Data: out})
}

// GET /clubs/:slug/members
// Accepts filter query params: gender, goal, program, age_min, age_max, city.
func (c *ClubController) ListMembers(ctx *gin.Context) {
	club, err := c.svc.GetClubBySlug(ctx.Request.Context(), ctx.Param("slug"))
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Club not found"})
		return
	}
	// Fetch all members (up to a reasonable ceiling) then filter in Go.
	// The query already restricts to visible profiles.
	members, err := c.svc.ListClubMembers(ctx.Request.Context(), club.ID, 200, 0)
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to list club members", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to list members"})
		return
	}

	// Apply optional server-side filters from query params.
	gender := ctx.Query("gender")
	goal := ctx.Query("goal")
	program := ctx.Query("program")
	city := ctx.Query("city")
	ageMinStr := ctx.Query("age_min")
	ageMaxStr := ctx.Query("age_max")

	if gender != "" || goal != "" || program != "" || city != "" || ageMinStr != "" || ageMaxStr != "" {
		var ageMin, ageMax int
		if ageMinStr != "" {
			if v, err := strconv.Atoi(ageMinStr); err == nil {
				ageMin = v
			}
		}
		if ageMaxStr != "" {
			if v, err := strconv.Atoi(ageMaxStr); err == nil {
				ageMax = v
			}
		}

		filtered := members[:0]
		for _, m := range members {
			if gender != "" && m.Gender != gender {
				continue
			}
			if goal != "" && m.Goal != goal {
				continue
			}
			if program != "" && m.Program != program {
				continue
			}
			if city != "" && m.City.String != city {
				continue
			}
			if ageMin > 0 || ageMax > 0 {
				if m.BirthDate.Valid {
					age := int(time.Since(m.BirthDate.Time).Hours() / (365.25 * 24))
					if ageMin > 0 && age < ageMin {
						continue
					}
					if ageMax > 0 && age > ageMax {
						continue
					}
				}
			}
			filtered = append(filtered, m)
		}
		members = filtered
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: members})
}

// POST /clubs/register — public self-registration (creates unverified club)
func (c *ClubController) RegisterClub(ctx *gin.Context) {
	var req clubRequest
	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	var geocodedFromCentroid bool

	// Always attempt server-side geocoding when coordinates are missing or zero.
	// The client must not pass pre-geocoded coords unless they came from a trusted
	// source (e.g. Google Maps import) — if the client sends (0,0) the geocoder runs.
	if req.Latitude == 0 && req.Longitude == 0 {
		if ng, ok := c.geocoder.(*geocoding.NominatimGeocoder); ok {
			result, err := ng.GeocodeWithResult(ctx.Request.Context(), req.Country, req.City, req.Address)
			if err != nil {
				logging.FromContext(ctx.Request.Context()).Warn("geocoding failed for new club",
					"name", req.Name, "city", req.City, "error", err)
				ctx.JSON(http.StatusUnprocessableEntity, types.Resp{
					Error: "Не вдалося визначити координати клубу. Уточніть адресу або скористайтесь імпортом з Google Maps.",
				})
				return
			}
			req.Latitude, req.Longitude = result.Lat, result.Lng
			geocodedFromCentroid = result.IsCentroidFallback
			logging.FromContext(ctx.Request.Context()).Info("club geocoded",
				"name", req.Name, "lat", result.Lat, "lng", result.Lng,
				"centroid_fallback", result.IsCentroidFallback)
		} else if c.geocoder != nil {
			lat, lng, err := c.geocoder.Geocode(ctx.Request.Context(), req.Country, req.City, req.Address)
			if err != nil {
				logging.FromContext(ctx.Request.Context()).Warn("geocoding failed for new club",
					"name", req.Name, "city", req.City, "error", err)
				ctx.JSON(http.StatusUnprocessableEntity, types.Resp{
					Error: "Не вдалося визначити координати клубу. Уточніть адресу або скористайтесь імпортом з Google Maps.",
				})
				return
			}
			req.Latitude, req.Longitude = lat, lng
		}
	}

	// Validate coordinates are within plausible range (catches Google Maps coords too).
	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		ctx.JSON(http.StatusUnprocessableEntity, types.Resp{Error: "Coordinates out of valid range"})
		return
	}

	// Final guard: refuse to persist a club at the null island (0,0).
	if req.Latitude == 0 && req.Longitude == 0 {
		ctx.JSON(http.StatusUnprocessableEntity, types.Resp{
			Error: "Не вдалося визначити координати клубу. Уточніть адресу або скористайтесь імпортом з Google Maps.",
		})
		return
	}

	club, err := c.svc.RegisterClub(ctx.Request.Context(), req.toParams(false))
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to register club", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to register club: " + err.Error()})
		return
	}

	resp := gin.H{"club": club}
	if geocodedFromCentroid {
		resp["geocode_warning"] = "Адресу не знайдено точно — клуб розміщено в центрі міста. Уточніть адресу пізніше в налаштуваннях."
	}
	ctx.JSON(http.StatusCreated, types.Resp{Data: resp})
}

// --- Authenticated user endpoints ---

// POST /clubs/:slug/join
// Dancers/parents are inserted into club_members; trainers into club_trainers.
// Both capped at maxClubsPerUser (5).
func (c *ClubController) JoinClub(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}
	club, err := c.svc.GetClubBySlug(ctx.Request.Context(), ctx.Param("slug"))
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Club not found"})
		return
	}

	accountType, _ := user.ProfileData["account_type"].(string)
	if accountType == "trainer" {
		count, err := c.svc.Queries.CountTrainerClubs(ctx.Request.Context(), user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to check club count"})
			return
		}
		if int(count) >= maxClubsPerUser {
			ctx.JSON(http.StatusConflict, types.Resp{Error: "You can join at most 5 clubs"})
			return
		}
		if err := c.svc.AddClubTrainer(ctx.Request.Context(), club.ID, user.ID); err != nil {
			logging.FromContext(ctx.Request.Context()).Error("failed to add club trainer", "error", err)
			ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to join club"})
			return
		}
	} else {
		count, err := c.svc.Queries.CountUserClubMemberships(ctx.Request.Context(), user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to check club count"})
			return
		}
		if int(count) >= maxClubsPerUser {
			ctx.JSON(http.StatusConflict, types.Resp{Error: "You can join at most 5 clubs"})
			return
		}
		if err := c.svc.JoinClub(ctx.Request.Context(), club.ID, user.ID); err != nil {
			logging.FromContext(ctx.Request.Context()).Error("failed to join club", "error", err)
			ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to join club"})
			return
		}
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

// DELETE /clubs/:slug/join
// Removes the user from club_trainers or club_members based on their account_type.
func (c *ClubController) LeaveClub(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}
	club, err := c.svc.GetClubBySlug(ctx.Request.Context(), ctx.Param("slug"))
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Club not found"})
		return
	}

	accountType, _ := user.ProfileData["account_type"].(string)
	if accountType == "trainer" {
		if err := c.svc.RemoveClubTrainer(ctx.Request.Context(), club.ID, user.ID); err != nil {
			logging.FromContext(ctx.Request.Context()).Error("failed to remove club trainer", "error", err)
			ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to leave club"})
			return
		}
	} else {
		if err := c.svc.LeaveClub(ctx.Request.Context(), club.ID, user.ID); err != nil {
			logging.FromContext(ctx.Request.Context()).Error("failed to leave club", "error", err)
			ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to leave club"})
			return
		}
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

// GET /me/clubs
// Trainers get their club_trainers entries; dancers/parents get club_members entries.
func (c *ClubController) GetMyClubs(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	accountType, _ := user.ProfileData["account_type"].(string)
	var clubs interface{}
	var err error
	if accountType == "trainer" {
		clubs, err = c.svc.Queries.ListTrainerClubs(ctx.Request.Context(), user.ID)
	} else {
		clubs, err = c.svc.GetUserClubs(ctx.Request.Context(), user.ID)
	}
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to get user clubs", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to get clubs"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: clubs})
}

// --- Business owner endpoints ---

// POST /clubs/:slug/claim
func (c *ClubController) ClaimClub(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}
	club, err := c.svc.GetClubBySlug(ctx.Request.Context(), ctx.Param("slug"))
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Club not found"})
		return
	}
	claimed, err := c.svc.ClaimClub(ctx.Request.Context(), club.ID, user.ID)
	if err != nil {
		ctx.JSON(http.StatusConflict, types.Resp{Error: "Club is already claimed or not found"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: claimed})
}

type manageClubRequest struct {
	Name         string         `json:"name"         binding:"omitempty,min=2,max=255"`
	Description  string         `json:"description"  binding:"omitempty,max=2000"`
	Address      string         `json:"address"      binding:"omitempty,max=500"`
	Phone        string         `json:"phone"        binding:"omitempty,max=50"`
	Website      string         `json:"website"      binding:"omitempty,url,max=500"`
	WorkingHours map[string]any `json:"working_hours"`
	LogoUrl      string         `json:"logo_url"     binding:"omitempty,url,max=500"`
}

// PUT /clubs/:slug/manage
func (c *ClubController) ManageClub(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}
	club, err := c.svc.GetClubBySlug(ctx.Request.Context(), ctx.Param("slug"))
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Club not found"})
		return
	}
	var req manageClubRequest
	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	// If the owner changed the address, always re-geocode (even if GMaps coords exist
	// in the DB) so the pin stays in sync with the text.
	var newLat, newLng float64
	storedAddr := ""
	if club.Address.Valid {
		storedAddr = club.Address.String
	}
	if req.Address != "" && req.Address != storedAddr && c.geocoder != nil {
		if ng, ok := c.geocoder.(*geocoding.NominatimGeocoder); ok {
			result, gErr := ng.GeocodeWithResult(ctx.Request.Context(), club.Country, club.City, req.Address)
			if gErr == nil && (result.Lat != 0 || result.Lng != 0) {
				newLat, newLng = result.Lat, result.Lng
				if result.IsCentroidFallback {
					logging.FromContext(ctx.Request.Context()).Warn("re-geocoded club address to centroid fallback",
						"slug", club.Slug, "address", req.Address)
				} else {
					logging.FromContext(ctx.Request.Context()).Info("re-geocoded club address",
						"slug", club.Slug, "address", req.Address, "lat", result.Lat, "lng", result.Lng)
				}
			}
		} else {
			lat, lng, gErr := c.geocoder.Geocode(ctx.Request.Context(), club.Country, club.City, req.Address)
			if gErr == nil && (lat != 0 || lng != 0) {
				newLat, newLng = lat, lng
			}
		}
	}

	if err := c.svc.ManageClub(ctx.Request.Context(), club.ID, user.ID, ManageClubParams{
		Name:         req.Name,
		Description:  req.Description,
		Address:      req.Address,
		Phone:        req.Phone,
		Website:      req.Website,
		WorkingHours: types.JSONB(req.WorkingHours),
		LogoUrl:      req.LogoUrl,
		Latitude:     newLat,
		Longitude:    newLng,
	}); err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to manage club", "error", err)
		ctx.JSON(http.StatusForbidden, types.Resp{Error: "Not the club owner or club not found"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

// GET /me/owned-clubs
func (c *ClubController) GetOwnedClubs(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}
	clubs, err := c.svc.ListOwnedClubs(ctx.Request.Context(), user.ID)
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to list owned clubs", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to get owned clubs"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: clubs})
}

// POST /clubs/:slug/chat — create or find a chat between the current user and the
// club. Works for unclaimed clubs too; the club side is answered by its owner.
func (c *ClubController) ChatWithClub(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}
	club, err := c.svc.GetClubBySlug(ctx.Request.Context(), ctx.Param("slug"))
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Club not found"})
		return
	}

	chatID, err := c.chatSvc.CreateClubChat(ctx.Request.Context(), user.ID, club.ID)
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to create club chat", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to create club chat"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: map[string]string{"chat_id": utils.UUIDToString(chatID)}})
}

// --- Admin endpoints ---

// GET /admin/clubs
func (c *ClubController) AdminListClubs(ctx *gin.Context) {
	limit, offset := pageParams(ctx)
	clubs, err := c.svc.AdminListClubs(ctx.Request.Context(), limit, offset)
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("admin: failed to list clubs", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to list clubs"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: clubs})
}

// POST /admin/clubs
func (c *ClubController) AdminCreateClub(ctx *gin.Context) {
	var req clubRequest
	if !corehttp.BindJSON(ctx, &req) {
		return
	}
	// Apply geocoding for admin creation too so (0,0) clubs cannot appear on the map.
	if req.Latitude == 0 && req.Longitude == 0 && c.geocoder != nil {
		lat, lng, gErr := c.geocoder.Geocode(ctx.Request.Context(), req.Country, req.City, req.Address)
		if gErr == nil && (lat != 0 || lng != 0) {
			req.Latitude, req.Longitude = lat, lng
		}
	}
	if req.Latitude == 0 && req.Longitude == 0 {
		ctx.JSON(http.StatusUnprocessableEntity, types.Resp{Error: "Could not determine club coordinates; provide lat/lng or a valid address"})
		return
	}
	club, err := c.svc.CreateClub(ctx.Request.Context(), req.toParams(true))
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("admin: failed to create club", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to create club"})
		return
	}
	ctx.JSON(http.StatusCreated, types.Resp{Data: club})
}

// PUT /admin/clubs/:id
func (c *ClubController) AdminUpdateClub(ctx *gin.Context) {
	clubID, err := utils.StringToUUID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid club ID"})
		return
	}
	var req clubRequest
	if !corehttp.BindJSON(ctx, &req) {
		return
	}
	if err := c.svc.UpdateClub(ctx.Request.Context(), clubID, req.toParams(false)); err != nil {
		logging.FromContext(ctx.Request.Context()).Error("admin: failed to update club", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to update club"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

// POST /admin/clubs/:id/verify
func (c *ClubController) AdminVerifyClub(ctx *gin.Context) {
	clubID, err := utils.StringToUUID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid club ID"})
		return
	}
	if err := c.svc.VerifyClub(ctx.Request.Context(), clubID); err != nil {
		logging.FromContext(ctx.Request.Context()).Error("admin: failed to verify club", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to verify club"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

// DELETE /admin/clubs/:id
func (c *ClubController) AdminDeactivateClub(ctx *gin.Context) {
	clubID, err := utils.StringToUUID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid club ID"})
		return
	}
	if err := c.svc.DeactivateClub(ctx.Request.Context(), clubID); err != nil {
		logging.FromContext(ctx.Request.Context()).Error("admin: failed to deactivate club", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to deactivate club"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

// GET /images/gphoto?ref=<photo_reference>&w=<width>
// Proxies a Google Places photo with the server-side API key so the key is
// never exposed to clients. Responds with long cache headers; no MinIO write.
func (c *ClubController) ProxyGooglePhoto(ctx *gin.Context) {
	ref := ctx.Query("ref")
	if ref == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}
	if c.places == nil || !c.places.IsConfigured() {
		ctx.Status(http.StatusServiceUnavailable)
		return
	}

	w := ctx.Query("w")
	maxWidth := "800"
	switch w {
	case "200", "400", "800":
		maxWidth = w
	}

	photoURL := fmt.Sprintf(
		"https://maps.googleapis.com/maps/api/place/photo?maxwidth=%s&photo_reference=%s&key=%s",
		maxWidth, url.QueryEscape(ref), c.places.APIKey(),
	)

	req, err := http.NewRequestWithContext(ctx.Request.Context(), http.MethodGet, photoURL, nil)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		ctx.Status(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ctx.Status(resp.StatusCode)
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}
	ctx.Header("Cache-Control", "public, max-age=2592000") // 30 days
	ctx.Header("Content-Type", contentType)
	ctx.Status(http.StatusOK)
	_, _ = io.Copy(ctx.Writer, resp.Body)
}

// RegisterRoutes wires all club routes.
func (c *ClubController) RegisterRoutes(
	r *gin.Engine,
	meGroup *gin.RouterGroup,
	adminGroup *gin.RouterGroup,
	userAuth gin.HandlerFunc,
	adminAuth gin.HandlerFunc,
) {
	// Public routes
	public := r.Group("/clubs")
	public.GET("", c.ListClubs)
	public.GET("/:slug", c.GetClub)
	public.GET("/:slug/members", c.ListMembers)
	public.GET("/:slug/trainers", c.ListClubTrainers)
	public.POST("/register", c.RegisterClub)

	// Image proxy (public, long-cached, no auth needed)
	r.GET("/images/gphoto", c.ProxyGooglePhoto)

	// Authenticated user routes
	auth := r.Group("/clubs")
	auth.Use(userAuth)
	auth.POST("/:slug/join", c.JoinClub)
	auth.DELETE("/:slug/join", c.LeaveClub)
	auth.POST("/:slug/claim", c.ClaimClub)
	auth.PUT("/:slug/manage", c.ManageClub)
	auth.POST("/:slug/chat", c.ChatWithClub)

	// /me/clubs, /me/owned-clubs, and Google Maps import
	meGroup.GET("/clubs", c.GetMyClubs)
	meGroup.GET("/owned-clubs", c.GetOwnedClubs)
	meGroup.POST("/clubs/parse-gmaps", c.ParseGoogleMapsLink)

	// Admin routes
	adminGroup.GET("/clubs", c.AdminListClubs)
	adminGroup.POST("/clubs", c.AdminCreateClub)
	adminGroup.PUT("/clubs/:id", c.AdminUpdateClub)
	adminGroup.POST("/clubs/:id/verify", c.AdminVerifyClub)
	adminGroup.DELETE("/clubs/:id", c.AdminDeactivateClub)
}
