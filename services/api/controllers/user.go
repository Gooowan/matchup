package controllers

import (
	"net/http"
	"os"
	"slices"

	"github.com/gin-gonic/gin"

	"github.com/Gooowan/matchup/modules/core/types"
	corehttp "github.com/Gooowan/matchup/modules/core/http"
	"github.com/Gooowan/matchup/modules/files"
	"github.com/Gooowan/matchup/modules/otp"
	core "github.com/Gooowan/matchup/modules/users"
	"github.com/Gooowan/matchup/modules/users/auth"
	coregen "github.com/Gooowan/matchup/modules/users/gen"
)

const MAX_DEPTH = 999

const otpPurposePasswordChange = "password_change"

type UserController struct {
	core       *core.UserService
	otpService *otp.OTPService
	authSvc    *auth.AuthService
}

func NewUserController(coreService *core.UserService) *UserController {
	return &UserController{
		core: coreService,
	}
}

func (c *UserController) SetOTPService(svc *otp.OTPService) { c.otpService = svc }
func (c *UserController) SetAuthService(svc *auth.AuthService) { c.authSvc = svc }

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

	if !corehttp.BindJSON(ctx, &req) {
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
		LastName  string `json:"last_name"  binding:"omitempty,min=1,max=50"`
	}

	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	profileData := types.JSONB{"first_name": req.FirstName}
	if req.LastName != "" {
		profileData["last_name"] = req.LastName
	}
	err := c.core.Queries.UpdateUserProfileData(ctx.Request.Context(), coregen.UpdateUserProfileDataParams{
		ProfileData: profileData,
		UserID:      user.ID,
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

// POST /user/password/change-otp/request
// Sends a 5-digit OTP to the authenticated user's email for in-app password change.
func (c *UserController) RequestPasswordChangeOTP(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}
	if c.otpService == nil {
		ctx.JSON(http.StatusServiceUnavailable, types.Resp{Error: "OTP service unavailable"})
		return
	}
	email := user.Email.String
	if err := c.otpService.CreateAndSendOTP(ctx.Request.Context(), user.ID.String(), email, otpPurposePasswordChange, map[string]interface{}{}); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Не вдалося надіслати код. Спробуй ще раз."})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"email": email}})
}

// POST /user/password/change-otp/confirm
// Verifies the OTP and sets the new password.
func (c *UserController) ConfirmPasswordChangeOTP(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}
	if c.otpService == nil || c.authSvc == nil {
		ctx.JSON(http.StatusServiceUnavailable, types.Resp{Error: "Service unavailable"})
		return
	}

	var req struct {
		Code        string `json:"code"         binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	if err := c.otpService.VerifyOTP(ctx.Request.Context(), user.ID.String(), otpPurposePasswordChange, req.Code); err != nil {
		switch err {
		case otp.ErrOTPExpired:
			ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Код застарів, надішли новий"})
		case otp.ErrTooManyAttempts:
			ctx.JSON(http.StatusTooManyRequests, types.Resp{Error: "Забагато спроб, надішли новий код"})
		default:
			ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Невірний код"})
		}
		return
	}

	if err := c.authSvc.UpdateUserPassword(ctx.Request.Context(), user, req.NewPassword); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Не вдалося змінити пароль"})
		return
	}
	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *UserController) RegisterRoutes(rg *gin.RouterGroup, userAuthMiddleware gin.HandlerFunc, filesController *files.FilesController, authController *auth.AuthController, uploadRL ...gin.HandlerFunc) {
	rg.Use(userAuthMiddleware)

	rg.POST("/password/change", authController.ChangePassword)
	rg.POST("/password/change-otp/request", c.RequestPasswordChangeOTP)
	rg.POST("/password/change-otp/confirm", c.ConfirmPasswordChangeOTP)
	avatarHandlers := []gin.HandlerFunc{filesController.UploadAvatar}
	if len(uploadRL) > 0 {
		avatarHandlers = append([]gin.HandlerFunc{uploadRL[0]}, avatarHandlers...)
	}
	rg.POST("/files/avatar", avatarHandlers...)
	rg.POST("/files/photo", filesController.UploadPhoto)
	rg.POST("/locale", c.SetUserLocale)
	rg.POST("/profile/update", c.UpdateUserProfile)

	rg.GET("/profile", c.GetUserProfile)
	rg.GET("/inviter", c.GetUserInviterProfile)
}
