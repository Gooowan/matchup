package clubs

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Gooowan/matchup/modules/chat"
	"github.com/Gooowan/matchup/modules/core/logging"
	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
	"github.com/Gooowan/matchup/modules/users/auth"
)

type ClubController struct {
	svc     *ClubService
	chatSvc *chat.ChatService
}

func NewClubController(svc *ClubService, chatSvc *chat.ChatService) *ClubController {
	return &ClubController{svc: svc, chatSvc: chatSvc}
}

// --- Request types ---

type clubRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Description string  `json:"description"`
	Country     string  `json:"country"     binding:"required"`
	City        string  `json:"city"        binding:"required"`
	Address     string  `json:"address"`
	Latitude    float64 `json:"latitude"    binding:"required"`
	Longitude   float64 `json:"longitude"   binding:"required"`
	Website     string  `json:"website"`
	Phone       string  `json:"phone"`
}

func (r *clubRequest) toParams(verified bool) CreateClubParams {
	return CreateClubParams{
		Name:        r.Name,
		Description: r.Description,
		Country:     r.Country,
		City:        r.City,
		Address:     r.Address,
		Latitude:    r.Latitude,
		Longitude:   r.Longitude,
		Website:     r.Website,
		Phone:       r.Phone,
		IsVerified:  verified,
	}
}

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
		ctx.Query("country"), ctx.Query("city"), limit, offset)
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to list clubs", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to list clubs"})
		return
	}
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
	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"club": club, "member_count": count}})
}

// GET /clubs/:id/members
func (c *ClubController) ListMembers(ctx *gin.Context) {
	club, err := c.svc.GetClubBySlug(ctx.Request.Context(), ctx.Param("slug"))
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Club not found"})
		return
	}
	limit, offset := pageParams(ctx)
	members, err := c.svc.ListClubMembers(ctx.Request.Context(), club.ID, limit, offset)
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to list club members", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to list members"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: members})
}

// POST /clubs/register — public self-registration (creates unverified club)
func (c *ClubController) RegisterClub(ctx *gin.Context) {
	var req clubRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}
	club, err := c.svc.RegisterClub(ctx.Request.Context(), req.toParams(false))
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to register club", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to register club: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, types.Resp{Data: club})
}

// --- Authenticated user endpoints ---

// POST /clubs/:id/join
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
	if err := c.svc.JoinClub(ctx.Request.Context(), club.ID, user.ID); err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to join club", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to join club"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

// DELETE /clubs/:slug/join
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
	if err := c.svc.LeaveClub(ctx.Request.Context(), club.ID, user.ID); err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to leave club", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to leave club"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

// GET /me/clubs
func (c *ClubController) GetMyClubs(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}
	clubs, err := c.svc.GetUserClubs(ctx.Request.Context(), user.ID)
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to get user clubs", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to get clubs"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: clubs})
}

// --- Business owner endpoints ---

// POST /clubs/:id/claim
func (c *ClubController) ClaimClub(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}
	clubID, err := utils.StringToUUID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid club ID"})
		return
	}
	club, err := c.svc.ClaimClub(ctx.Request.Context(), clubID, user.ID)
	if err != nil {
		ctx.JSON(http.StatusConflict, types.Resp{Error: "Club is already claimed or not found"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: club})
}

type manageClubRequest struct {
	Description  string         `json:"description"`
	Address      string         `json:"address"`
	Phone        string         `json:"phone"`
	Website      string         `json:"website"`
	WorkingHours map[string]any `json:"working_hours"`
}

// PUT /clubs/:id/manage
func (c *ClubController) ManageClub(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}
	clubID, err := utils.StringToUUID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid club ID"})
		return
	}
	var req manageClubRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}
	if err := c.svc.ManageClub(ctx.Request.Context(), clubID, user.ID, ManageClubParams{
		Description:  req.Description,
		Address:      req.Address,
		Phone:        req.Phone,
		Website:      req.Website,
		WorkingHours: types.JSONB(req.WorkingHours),
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

// POST /clubs/:id/chat — create or find a DM between current user and club owner.
func (c *ClubController) ChatWithClub(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}
	clubID, err := utils.StringToUUID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid club ID"})
		return
	}
	club, err := c.svc.GetClubByID(ctx.Request.Context(), clubID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "Club not found"})
		return
	}
	if !club.OwnerUserID.Valid {
		ctx.JSON(http.StatusUnprocessableEntity, types.Resp{Error: "This club has no business owner yet"})
		return
	}

	// Order UUIDs to match the UNIQUE(user1_id, user2_id) constraint in chats
	u1, u2 := club.OwnerUserID, user.ID
	if bytes.Compare(u1.Bytes[:], u2.Bytes[:]) > 0 {
		u1, u2 = u2, u1
	}
	chatID, err := c.chatSvc.CreateChat(ctx.Request.Context(), u1, u2)
	if err != nil {
		// Chat may already exist — fetch it
		logging.FromContext(ctx.Request.Context()).Info("club chat already exists or error creating", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to create business chat"})
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
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
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
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
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
	public.POST("/register", c.RegisterClub)

	// Authenticated user routes
	auth := r.Group("/clubs")
	auth.Use(userAuth)
	auth.POST("/:slug/join", c.JoinClub)
	auth.DELETE("/:slug/join", c.LeaveClub)
	auth.POST("/:id/claim", c.ClaimClub)
	auth.PUT("/:id/manage", c.ManageClub)
	auth.POST("/:id/chat", c.ChatWithClub)

	// /me/clubs and /me/owned-clubs
	meGroup.GET("/clubs", c.GetMyClubs)
	meGroup.GET("/owned-clubs", c.GetOwnedClubs)

	// Admin routes
	adminGroup.GET("/clubs", c.AdminListClubs)
	adminGroup.POST("/clubs", c.AdminCreateClub)
	adminGroup.PUT("/clubs/:id", c.AdminUpdateClub)
	adminGroup.POST("/clubs/:id/verify", c.AdminVerifyClub)
	adminGroup.DELETE("/clubs/:id", c.AdminDeactivateClub)
}
