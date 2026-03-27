package clubs

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
	"github.com/Gooowan/matchup/modules/users/auth"
)

type ClubController struct {
	svc *ClubService
}

func NewClubController(svc *ClubService) *ClubController {
	return &ClubController{svc: svc}
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
	clubID, err := utils.StringToUUID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid club ID"})
		return
	}
	limit, offset := pageParams(ctx)
	members, err := c.svc.ListClubMembers(ctx.Request.Context(), clubID, limit, offset)
	if err != nil {
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
	clubID, err := utils.StringToUUID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid club ID"})
		return
	}
	if err := c.svc.JoinClub(ctx.Request.Context(), clubID, user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to join club"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

// DELETE /clubs/:id/join
func (c *ClubController) LeaveClub(ctx *gin.Context) {
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
	if err := c.svc.LeaveClub(ctx.Request.Context(), clubID, user.ID); err != nil {
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
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to get clubs"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: clubs})
}

// --- Admin endpoints ---

// GET /admin/clubs
func (c *ClubController) AdminListClubs(ctx *gin.Context) {
	limit, offset := pageParams(ctx)
	clubs, err := c.svc.AdminListClubs(ctx.Request.Context(), limit, offset)
	if err != nil {
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
	public.GET("/:id/members", c.ListMembers)
	public.POST("/register", c.RegisterClub)

	// Authenticated user routes
	auth := r.Group("/clubs")
	auth.Use(userAuth)
	auth.POST("/:id/join", c.JoinClub)
	auth.DELETE("/:id/join", c.LeaveClub)

	// /me/clubs
	meGroup.GET("/clubs", c.GetMyClubs)

	// Admin routes
	adminGroup.GET("/clubs", c.AdminListClubs)
	adminGroup.POST("/clubs", c.AdminCreateClub)
	adminGroup.PUT("/clubs/:id", c.AdminUpdateClub)
	adminGroup.POST("/clubs/:id/verify", c.AdminVerifyClub)
	adminGroup.DELETE("/clubs/:id", c.AdminDeactivateClub)
}
