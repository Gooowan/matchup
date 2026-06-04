package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"

	corehttp "github.com/Gooowan/matchup/modules/core/http"
	"github.com/Gooowan/matchup/modules/core/types"
	coregen "github.com/Gooowan/matchup/modules/users/gen"

	"github.com/jackc/pgx/v5/pgtype"
)

// GoogleAuthController handles Google ID-token sign-in.
// Flow (web and Capacitor native both produce an ID token):
//  1. Frontend acquires an ID token from Google (GIS or native plugin).
//  2. Frontend POSTs { id_token } to POST /auth/google.
//  3. This handler verifies the token against GOOGLE_CLIENT_ID.
//  4. Finds or creates a user, links the identity, issues the session cookie.
type GoogleAuthController struct {
	authService *AuthService
}

func NewGoogleAuthController(authService *AuthService) *GoogleAuthController {
	return &GoogleAuthController{authService: authService}
}

func (c *GoogleAuthController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/google", c.GoogleLogin)
}

func (c *GoogleAuthController) GoogleLogin(ctx *gin.Context) {
	var req struct {
		IDToken string `json:"id_token" binding:"required"`
	}
	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		ctx.JSON(http.StatusServiceUnavailable, types.Resp{Error: "Google login is not configured"})
		return
	}

	// Validate the ID token. This call fetches Google's public keys and verifies
	// the signature, expiry, issuer, and audience automatically.
	payload, err := idtoken.Validate(ctx.Request.Context(), req.IDToken, clientID)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Invalid Google token"})
		return
	}

	sub := payload.Subject // Google's unique user identifier
	emailRaw, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	firstName, _ := payload.Claims["given_name"].(string)
	lastName, _ := payload.Claims["family_name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	user, err := c.findOrCreateGoogleUser(ctx.Request.Context(), sub, emailRaw, emailVerified, firstName, lastName, picture)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to sign in with Google"})
		return
	}

	token, expiresAt, err := c.authService.CreateJwtToken(ctx.Request.Context(), user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to create session"})
		return
	}

	domain := os.Getenv("COOKIE_DOMAIN")
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
		Data: gin.H{"user": user.ToDTO()},
	})
}

// findOrCreateGoogleUser resolves the Google identity to a MatchUp user:
//  1. Existing identity row → return that user.
//  2. Existing user with matching email (password account) → link and return.
//  3. Neither → create a new user and link the identity.
func (c *GoogleAuthController) findOrCreateGoogleUser(
	ctx context.Context,
	sub, email string,
	emailVerified bool,
	firstName, lastName, picture string,
) (*coregen.User, error) {
	queries := c.authService.core.Queries

	// 1. Check user_identities first.
	userID, err := queries.FindIdentity(ctx, coregen.FindIdentityParams{
		Provider:        "google",
		ProviderSubject: sub,
	})
	if err == nil && userID.Valid {
		user, err := queries.GetUser(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("identity found but user lookup failed: %w", err)
		}
		return &user, nil
	}

	// 2. Try to find by email and link.
	if email != "" && emailVerified {
		existingUser, err := queries.GetUserByEmail(ctx, pgtype.Text{String: strings.ToLower(email), Valid: true})
		if err == nil {
			// Link the Google identity to the existing account.
			_ = queries.CreateIdentity(ctx, coregen.CreateIdentityParams{
				UserID:          existingUser.ID,
				Provider:        "google",
				ProviderSubject: sub,
				Email:           pgtype.Text{String: email, Valid: true},
			})
			return &existingUser, nil
		}
	}

	// 3. Create a new user. Google-verified emails skip OTP verification.
	profileData := types.JSONB{}
	if firstName != "" {
		profileData["first_name"] = firstName
	}
	if lastName != "" {
		profileData["last_name"] = lastName
	}
	if picture != "" {
		profileData["avatar"] = picture
	}

	emailField := pgtype.Text{}
	if email != "" {
		emailField = pgtype.Text{String: strings.ToLower(email), Valid: true}
	}

	// EmailVerificationToken is intentionally left empty — Google already verified the email.
	newUser, err := queries.CreateUser(ctx, coregen.CreateUserParams{
		Email:                  emailField,
		EmailVerificationToken: pgtype.Text{},
		Password:               pgtype.Text{}, // no password for OAuth users
		ProfileData:            profileData,
		Metadata:               types.JSONB{"auth_provider": "google"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Link the identity.
	_ = queries.CreateIdentity(ctx, coregen.CreateIdentityParams{
		UserID:          newUser.ID,
		Provider:        "google",
		ProviderSubject: sub,
		Email:           pgtype.Text{String: email, Valid: email != ""},
	})

	return &newUser, nil
}
