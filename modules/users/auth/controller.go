package auth

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/Gooowan/matchup/modules/core/types"
)

type AuthController struct {
	authService *AuthService
}

func NewAuthController(authService *AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	token, expiresAt, user, err := c.authService.Login(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case ErrInvalidUser, ErrInvalidPassword:
			ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid email or password"})
			return
		case ErrEmailNotVerified:
			ctx.JSON(http.StatusForbidden, types.Resp{Error: "Please verify your email before logging in"})
			return
		default:
			ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Login failed"})
			return
		}
	}

	// Set HttpOnly cookie for cross-subdomain authentication
	// For cross-domain setup, we need to set the domain to the parent domain
	// e.g., if backend is api.yourdomain.com and frontend is id.yourdomain.com
	// we set domain to .yourdomain.com
	domain := os.Getenv("COOKIE_DOMAIN") // e.g., ".yourdomain.com"
	if domain == "" {
		// If no domain is set, use empty string (current domain only)
		domain = ""
	}
	ctx.SetCookie("auth_token", token, int(time.Until(expiresAt).Seconds()), "/", domain, true, true)

	if user.ProfileData != nil {
		if localeValue, exists := user.ProfileData["locale"]; exists {
			if locale, ok := localeValue.(string); ok && locale != "" {
				ctx.SetCookie("locale", locale, 60*60*24*365, "/", domain, false, false)
			}
		}
	}

	ctx.JSON(http.StatusOK, types.Resp{
		Data: gin.H{
			"user": user.ToDTO(),
		},
	})
}

func (c *AuthController) Register(ctx *gin.Context) {
	var req RegistrationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	user, err := c.authService.Register(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, types.Resp{Data: gin.H{"user": user.ToDTO()}})
}

func (c *AuthController) VerifyEmail(ctx *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	user, err := c.authService.EmailVerify(ctx.Request.Context(), req.Token)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid verification token"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"user": user.ToDTO()}})
}

func (c *AuthController) CheckEmail(ctx *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	if available := c.authService.CheckEmailAvailability(ctx.Request.Context(), req.Email); available {
		ctx.JSON(http.StatusOK, types.Resp{
			Data: gin.H{
				"available": true,
			},
		})
	} else {
		ctx.JSON(http.StatusOK, types.Resp{
			Data: gin.H{
				"available": false,
			},
		})
	}
}

func (c *AuthController) CheckInviter(ctx *gin.Context) {
	var req struct {
		ReferralID int64 `json:"referral_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	user, err := c.authService.core.Queries.GetUserByReferralId(ctx.Request.Context(), req.ReferralID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "User not found"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"inviter": gin.H{"profile_data": user.ProfileData, "referral_id": user.ReferralID}}})
}

func (c *AuthController) ForgotPassword(ctx *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	err := c.authService.RequestPasswordReset(ctx.Request.Context(), req.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: fmt.Sprintf("Failed to process password reset request: %s", err.Error())})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *AuthController) ChangePassword(ctx *gin.Context) {
	var req struct {
		Password    string `json:"password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	user, exists := GetUserFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "User not authenticated"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte(req.Password)); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Current password is incorrect"})
		return
	}

	err := c.authService.UpdateUserPassword(ctx.Request.Context(), user, req.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to update password"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *AuthController) ResetPassword(ctx *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	err := c.authService.PasswordReset(ctx.Request.Context(), req.Token, req.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *AuthController) Logout(ctx *gin.Context) {
	// Clear the auth cookie by setting it to expire immediately
	domain := os.Getenv("COOKIE_DOMAIN") // e.g., ".yourdomain.com"
	if domain == "" {
		// If no domain is set, use empty string (current domain only)
		domain = ""
	}
	ctx.SetCookie("auth_token", "", -1, "/", domain, true, true)

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *AuthController) RegisterRoutes(rg *gin.RouterGroup, authRateLimit gin.HandlerFunc) {
	rg.POST("/login", authRateLimit, c.Login)
	rg.POST("/register", c.Register)
	rg.POST("/logout", c.Logout)

	rg.POST("/check/email", c.CheckEmail)
	rg.POST("/check/inviter", c.CheckInviter)

	rg.POST("/verify/email", c.VerifyEmail)

	rg.POST("/password/forgot", c.ForgotPassword)
	rg.POST("/password/reset", c.ResetPassword)
}
