package controllers

import (
	"net/http"
	"os"
	"slices"

	"github.com/gin-gonic/gin"

	core "github.com/Gooowan/matchup/modules/users"
	"github.com/Gooowan/matchup/modules/users/auth"
	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/files"
	coregen "github.com/Gooowan/matchup/modules/users/gen"
)

const MAX_DEPTH = 999

type UserController struct {
	core *core.UserService
}

func NewUserController(coreService *core.UserService) *UserController {
	return &UserController{
		core: coreService,
	}
}

func (c *UserController) GetUserProfile(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"user": user.ToDTO()}})
}

func (c *UserController) GetUserInviterProfile(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	inviter, err := c.core.Queries.GetUser(ctx.Request.Context(), user.InviterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to get user profile"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"inviter": inviter.ToDTO()}})
}

func (c *UserController) SetUserLocale(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	var req struct {
		Locale string `json:"locale" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	validLocales := []string{"en", "uk", "es"}
	isValid := slices.Contains(validLocales, req.Locale)

	if !isValid {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid locale"})
		return
	}

	err := c.core.Queries.UpdateUserProfileData(ctx.Request.Context(), coregen.UpdateUserProfileDataParams{
		ProfileData: types.JSONB{
			"locale": req.Locale,
		},
		UserID: user.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to update locale"})
		return
	}

	domain := os.Getenv("COOKIE_DOMAIN")
	if domain == "" {
		domain = ""
	}
	ctx.SetCookie("locale", req.Locale, 60*60*24*365, "/", domain, false, false)

	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"locale": req.Locale}})
}

func (c *UserController) UpdateUserProfile(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	var req struct {
		FirstName string `json:"first_name" binding:"required,min=2,max=50"`
		LastName  string `json:"last_name" binding:"required,min=2,max=50"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	err := c.core.Queries.UpdateUserProfileData(ctx.Request.Context(), coregen.UpdateUserProfileDataParams{
		ProfileData: types.JSONB{
			"first_name": req.FirstName,
			"last_name":  req.LastName,
		},
		UserID: user.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to update profile"})
		return
	}

	// Get updated user data
	updatedUser, err := c.core.Queries.GetUser(ctx.Request.Context(), user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to get updated profile"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"user": updatedUser.ToDTO()}})
}

func (c *UserController) RegisterRoutes(rg *gin.RouterGroup, userAuthMiddleware gin.HandlerFunc, filesController *files.FilesController, authController *auth.AuthController) {
	rg.Use(userAuthMiddleware)

	rg.POST("/password/change", authController.ChangePassword)
	rg.POST("/files/avatar", filesController.UploadAvatar)
	rg.POST("/locale", c.SetUserLocale)
	rg.POST("/profile/update", c.UpdateUserProfile)

	rg.GET("/profile", c.GetUserProfile)
	rg.GET("/inviter", c.GetUserInviterProfile)
}
