package auth

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/Gooowan/matchup/modules/core/types"
	corehttp "github.com/Gooowan/matchup/modules/core/http"
	"github.com/Gooowan/matchup/modules/core/utils"
)

// LockoutRecorder is implemented by ratelimit.RLService; kept as an interface
// here so the auth package does not import the ratelimit package.
type LockoutRecorder interface {
	RecordLoginFailure(ctx context.Context, email string)
	ClearLoginFailures(ctx context.Context, email string)
}

type AuthController struct {
	authService     *AuthService
	lockoutRecorder LockoutRecorder
}

func NewAuthController(authService *AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// SetLockoutRecorder wires in the Redis-backed lockout tracker after construction
// (avoids an import cycle between auth and ratelimit).
func (c *AuthController) SetLockoutRecorder(r LockoutRecorder) {
	c.lockoutRecorder = r
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	token, expiresAt, user, err := c.authService.Login(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		if c.lockoutRecorder != nil {
			c.lockoutRecorder.RecordLoginFailure(ctx.Request.Context(), req.Email)
		}
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
	if c.lockoutRecorder != nil {
		c.lockoutRecorder.ClearLoginFailures(ctx.Request.Context(), req.Email)
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
	ctx.SetSameSite(http.SameSiteLaxMode)
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
	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	user, err := c.authService.Register(ctx.Request.Context(), &req)
	if err != nil {
		// Surface "email already in use" as 409 so the frontend can show a useful message.
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "duplicate") {
			ctx.JSON(http.StatusConflict, types.Resp{Error: "Ця адреса вже використовується"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to create account"})
		return
	}

	ctx.JSON(http.StatusCreated, types.Resp{Data: gin.H{"user": user.ToDTO()}})
}

func (c *AuthController) CheckEmail(ctx *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if !corehttp.BindJSON(ctx, &req) {
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
		InviterID string `json:"inviter_id" binding:"required"`
	}

	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	inviterUUID, err := utils.StringToUUID(req.InviterID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid inviter ID"})
		return
	}

	user, err := c.authService.core.Queries.GetUser(ctx.Request.Context(), inviterUUID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.Resp{Error: "User not found"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"inviter": gin.H{"profile_data": user.ProfileData}}})
}

func (c *AuthController) ForgotPassword(ctx *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	err := c.authService.RequestPasswordReset(ctx.Request.Context(), req.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{ErrorCode: "RESET_FAILED", Error: "Failed to process password reset request"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

func (c *AuthController) ChangePassword(ctx *gin.Context) {
	var req struct {
		Password    string `json:"password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if !corehttp.BindJSON(ctx, &req) {
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

	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	err := c.authService.PasswordReset(ctx.Request.Context(), req.Token, req.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{ErrorCode: "INVALID_RESET_TOKEN", Error: "Invalid or expired reset link"})
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
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("auth_token", "", -1, "/", domain, true, true)

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

// RegisterRoutesWithLockout registers routes with a separate lockout middleware
// and a per-minute rate limiter both applied to /login.
func (c *AuthController) RegisterRoutesWithLockout(rg *gin.RouterGroup, lockout, authRateLimit gin.HandlerFunc, registerRateLimit ...gin.HandlerFunc) {
	rg.POST("/login", lockout, authRateLimit, c.Login)
	regHandlers := []gin.HandlerFunc{c.Register}
	if len(registerRateLimit) > 0 {
		regHandlers = append([]gin.HandlerFunc{registerRateLimit[0]}, regHandlers...)
	}
	rg.POST("/register", regHandlers...)
	rg.POST("/logout", c.Logout)

	rg.POST("/check/email", c.CheckEmail)
	rg.POST("/check/inviter", c.CheckInviter)

	rg.POST("/password/forgot", c.ForgotPassword)
	rg.POST("/password/reset", c.ResetPassword)
}

func (c *AuthController) RegisterRoutes(rg *gin.RouterGroup, authRateLimit gin.HandlerFunc, registerRateLimit ...gin.HandlerFunc) {
	rg.POST("/login", authRateLimit, c.Login)
	regHandlers := []gin.HandlerFunc{c.Register}
	if len(registerRateLimit) > 0 {
		regHandlers = append([]gin.HandlerFunc{registerRateLimit[0]}, regHandlers...)
	}
	rg.POST("/register", regHandlers...)
	rg.POST("/logout", c.Logout)

	rg.POST("/check/email", c.CheckEmail)
	rg.POST("/check/inviter", c.CheckInviter)

	rg.POST("/password/forgot", c.ForgotPassword)
	rg.POST("/password/reset", c.ResetPassword)
}
