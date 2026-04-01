package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/otp"
)

const otpEmailVerifyPurpose = "email_verify"

// OTPAuthController handles OTP-based email verification.
type OTPAuthController struct {
	auth *AuthService
	otp  *otp.OTPService
}

func NewOTPAuthController(auth *AuthService, otpSvc *otp.OTPService) *OTPAuthController {
	return &OTPAuthController{auth: auth, otp: otpSvc}
}

// RegisterRoutes attaches OTP routes to the given group (e.g. /auth).
func (c *OTPAuthController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/otp/send", c.SendEmailOTP)
	rg.POST("/otp/verify", c.VerifyEmailOTP)
}

// SendEmailOTP godoc
// POST /auth/otp/send
// Body: { "email": "user@example.com" }
func (c *OTPAuthController) SendEmailOTP(ctx *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	email := strings.ToLower(req.Email)
	user, err := c.auth.core.Queries.GetUserByEmail(ctx.Request.Context(), pgtype.Text{String: email, Valid: true})
	if err != nil {
		// Don't reveal whether user exists
		ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
		return
	}

	userIDStr := user.ID.String()

	if err := c.otp.CreateAndSendEmailVerifyOTP(ctx.Request.Context(), userIDStr, email); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to send verification code"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: "ok"})
}

// VerifyEmailOTP godoc
// POST /auth/otp/verify
// Body: { "email": "user@example.com", "code": "12345678" }
func (c *OTPAuthController) VerifyEmailOTP(ctx *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	email := strings.ToLower(req.Email)
	user, err := c.auth.core.Queries.GetUserByEmail(ctx.Request.Context(), pgtype.Text{String: email, Valid: true})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid email or code"})
		return
	}

	userIDStr := user.ID.String()

	if err := c.otp.VerifyOTP(ctx.Request.Context(), userIDStr, otpEmailVerifyPurpose, req.Code); err != nil {
		switch err {
		case otp.ErrOTPExpired:
			ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Code has expired, please request a new one"})
		case otp.ErrTooManyAttempts:
			ctx.JSON(http.StatusTooManyRequests, types.Resp{Error: "Too many attempts, please request a new code"})
		default:
			ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid verification code"})
		}
		return
	}

	if _, err := c.auth.EmailVerifyByUserID(ctx.Request.Context(), userIDStr); err != nil {
		// Non-fatal: token may already be cleared
		_ = err
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"message": "Email verified successfully"}})
}
