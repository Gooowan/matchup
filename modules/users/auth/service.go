package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
	"github.com/Gooowan/matchup/modules/email"
	"github.com/Gooowan/matchup/modules/otp"
	core "github.com/Gooowan/matchup/modules/users"
	coregen "github.com/Gooowan/matchup/modules/users/gen"
)

var (
	ErrInternal         = errors.New("errors.internal")
	ErrInvalidUser      = errors.New("errors.invalid_user")
	ErrInvalidPassword  = errors.New("errors.invalid_password")
	ErrEmailNotVerified = errors.New("errors.email_not_verified")
)

type AuthService struct {
	core         *core.UserService
	hooks        []RegistrationHook
	secretKey    []byte
	issuer       string
	audience     string
	expiration   time.Duration
	emailService *email.EmailService
	otpService   *otp.OTPService
}

func (s *AuthService) SetOTPService(svc *otp.OTPService) {
	s.otpService = svc
}

type JWTClaims struct {
	UserID string `json:"sub"`
	Nonce  int32  `json:"nonce"`
	jwt.RegisteredClaims
}

type RegistrationRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`

	InviterID string `json:"inviter_id"`

	ProfileData types.JSONB `json:"profile_data"`
	Metadata    types.JSONB `json:"metadata,omitempty"`

	// Module-specific data
	Extensions map[string]any `json:"extensions,omitempty"`
}

type RegistrationHook interface {
	// Validate additional fields in registration request
	ValidateRegistration(qtx *coregen.Queries, ctx context.Context, req *RegistrationRequest) error

	// Process after core user is created
	PostRegistration(qtx *coregen.Queries, ctx context.Context, user *coregen.User, req *RegistrationRequest) error
}

func NewAuthService(coreService *core.UserService, emailService *email.EmailService) (*AuthService, error) {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable not set")
	}

	issuer := os.Getenv("APP_URL")
	if issuer == "" {
		return nil, fmt.Errorf("APP_URL environment variable not set")
	}

	audience := os.Getenv("JWT_AUDIENCE")
	if audience == "" {
		return nil, fmt.Errorf("JWT_AUDIENCE environment variable not set")
	}

	expirationStr := os.Getenv("JWT_EXPIRATION_TIME")
	expiration := 8 * time.Hour
	if expirationStr != "" {
		if seconds, err := strconv.Atoi(expirationStr); err == nil {
			expiration = time.Duration(seconds) * time.Second
		}
	}

	return &AuthService{
		core:         coreService,
		emailService: emailService,
		secretKey:    []byte(secretKey),
		issuer:       issuer,
		audience:     audience,
		expiration:   expiration,
		hooks:        make([]RegistrationHook, 0),
	}, nil
}

func (s *AuthService) AddHook(hook RegistrationHook) {
	s.hooks = append(s.hooks, hook)
}

func (s *AuthService) CheckEmailAvailability(ctx context.Context, email string) bool {
	emailText := pgtype.Text{String: strings.ToLower(email), Valid: true}
	_, err := s.core.Queries.GetUserByEmail(ctx, emailText)
	return err != nil
}

func (s *AuthService) Register(ctx context.Context, req *RegistrationRequest) (*coregen.User, error) {
	var err error

	// inviter_id is optional for MatchUp
	var inviterUUID pgtype.UUID
	if req.InviterID != "" {
		parsed, parseErr := utils.StringToUUID(req.InviterID)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid inviter_id format")
		}
		// Validate inviter exists
		_, findErr := s.core.Queries.GetUser(ctx, parsed)
		if findErr != nil {
			return nil, fmt.Errorf("inviter not found")
		}
		inviterUUID = parsed
	}

	tx, err := s.core.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.core.Queries.WithTx(tx)

	for _, hook := range s.hooks {
		if err := hook.ValidateRegistration(qtx, ctx, req); err != nil {
			return nil, fmt.Errorf("registration validation failed: %w", err)
		}
	}

	var emailVerificationToken pgtype.Text

	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	email := pgtype.Text{String: strings.ToLower(req.Email), Valid: true}

	shouldVerifyEmail := s.emailService != nil && !s.emailService.IsMockProvider()

	token, err := s.GenerateSecureToken(32) // 64 char hex string
	if err != nil {
		return nil, fmt.Errorf("failed to generate email verification token: %w", err)
	}
	emailVerificationToken = pgtype.Text{String: token, Valid: shouldVerifyEmail}

	params := coregen.CreateUserParams{
		Email:                  email,
		EmailVerificationToken: emailVerificationToken,
		Password:               pgtype.Text{String: hashedPassword, Valid: true},
		InviterID:              inviterUUID,
		ProfileData:            req.ProfileData,
		Metadata:               req.Metadata,
	}

	user, err := qtx.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create core user: %w", err)
	}

	for _, hook := range s.hooks {
		if err := hook.PostRegistration(qtx, ctx, &user, req); err != nil {
			return nil, fmt.Errorf("post-registration hook failed: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Send OTP after commit so a transient email failure never orphans the user row.
	// The verify-email page has a "resend" button that covers this case.
	if s.emailService != nil && s.otpService != nil && !s.emailService.IsMockProvider() {
		if err := s.otpService.CreateAndSendEmailVerifyOTP(ctx, user.ID.String(), req.Email); err != nil {
			// Non-fatal: user is persisted; they can request a new code on the verify page.
			_ = err
		}
	}

	return &user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, time.Time, *coregen.User, error) {

	user, err := s.core.Queries.GetUserByEmail(ctx, pgtype.Text{String: strings.ToLower(email), Valid: true})
	if err != nil {
		slog.WarnContext(ctx, "security: auth_failure — unknown email",
			"event", "auth_failure",
			"reason", "unknown_email",
			"email", strings.ToLower(email),
		)
		return "", time.Time{}, nil, ErrInvalidUser
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte(password)); err != nil {
		slog.WarnContext(ctx, "security: auth_failure — wrong password",
			"event", "auth_failure",
			"reason", "wrong_password",
			"user_id", user.ID.String(),
		)
		return "", time.Time{}, nil, ErrInvalidPassword
	}

	// Skip email verification check for mock provider
	requireEmailVerification := s.emailService != nil && !s.emailService.IsMockProvider()
	if requireEmailVerification && user.EmailVerificationToken.Valid && user.EmailVerificationToken.String != "" {
		return "", time.Time{}, nil, ErrEmailNotVerified
	}

	token, expiresAt, err := s.CreateJwtToken(ctx, user.ID)
	if err != nil {
		return "", time.Time{}, nil, ErrInternal
	}

	return token, expiresAt, &user, nil
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, address string) error {
	if s.emailService == nil {
		return fmt.Errorf("email service not available")
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		return fmt.Errorf("FRONTEND_URL environment variable is required")
	}

	user, err := s.core.Queries.GetUserByEmail(ctx, pgtype.Text{String: strings.ToLower(address), Valid: true})
	if err != nil {
		return fmt.Errorf("failed to get user by email: %w", err)
	}

	token, err := s.GenerateSecureToken(32) // 64 char hex string
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	expiresAt := time.Now().Add(1 * time.Hour)
	err = s.core.Queries.UpdateUserForgotPasswordToken(ctx, coregen.UpdateUserForgotPasswordTokenParams{
		ForgotPasswordToken:          pgtype.Text{String: token, Valid: true},
		ForgotPasswordTokenExpiresAt: pgtype.Timestamp{Time: expiresAt, Valid: true},
		UserID:                       user.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	resetLink := fmt.Sprintf("%s/resetPassword?token=%s", frontendURL, token)

	err = s.emailService.SendEmail(ctx, email.EmailRequest{
		To:         address,
		From:       fmt.Sprintf("%s <noreply@%s>", s.emailService.GetSender(), s.emailService.GetDomain()),
		Subject:    "Reset Your Password",
		TemplateID: email.PasswordResetTemplate,
		TemplateData: types.JSONB{
			"ResetLink": resetLink,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

func (s *AuthService) PasswordReset(ctx context.Context, token, newPassword string) error {
	tx, err := s.core.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.core.Queries.WithTx(tx)
	user, err := qtx.GetUserByForgotPasswordToken(ctx, pgtype.Text{String: token, Valid: true})
	if err != nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	hashed, err := s.HashPassword(newPassword)
	if err != nil {
		return ErrInternal
	}

	if err := qtx.UpdateUserPassword(ctx, coregen.UpdateUserPasswordParams{
		Password: pgtype.Text{String: hashed, Valid: true},
		UserID:   user.ID,
	}); err != nil {
		return ErrInternal
	}

	err = qtx.UpdateUserForgotPasswordToken(ctx, coregen.UpdateUserForgotPasswordTokenParams{
		ForgotPasswordToken:          pgtype.Text{},
		ForgotPasswordTokenExpiresAt: pgtype.Timestamp{},
		UserID:                       user.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to clear reset token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// IssueToken creates and returns a signed JWT for the given user ID.
// It is a thin alias over CreateJwtToken exposed for use by other controllers.
func (s *AuthService) IssueToken(ctx context.Context, userID pgtype.UUID) (string, time.Time, error) {
	return s.CreateJwtToken(ctx, userID)
}

// EmailVerifyByUserID marks a user's email as verified by clearing the verification token.
func (s *AuthService) EmailVerifyByUserID(ctx context.Context, userIDStr string) (*coregen.User, error) {
	userUUID, err := utils.StringToUUID(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.core.Queries.GetUser(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if err := s.core.Queries.UpdateUserEmailVerificationToken(ctx, coregen.UpdateUserEmailVerificationTokenParams{
		EmailVerificationToken: pgtype.Text{},
		UserID:                 user.ID,
	}); err != nil {
		return nil, fmt.Errorf("failed to update verification token: %w", err)
	}

	return &user, nil
}

