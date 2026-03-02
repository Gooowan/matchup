package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/email"
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
}

type JWTClaims struct {
	UserID string `json:"sub"`
	Nonce  int32  `json:"nonce"`
	jwt.RegisteredClaims
}

type RegistrationRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`

	ReferralID int64 `json:"referral_id"`

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

	if req.ReferralID <= 0 {
		return nil, fmt.Errorf("referral_id is required")
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

		ProfileData: req.ProfileData,
		Metadata:    req.Metadata,
	}

	inviter, err := s.core.Queries.GetUserByReferralId(ctx, req.ReferralID)
	if err != nil {
		return nil, fmt.Errorf("inviter not found")
	}
	params.InviterID = inviter.ID

	user, err := s.core.Queries.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create core user: %w", err)
	}

	for _, hook := range s.hooks {
		if err := hook.PostRegistration(qtx, ctx, &user, req); err != nil {
			return nil, fmt.Errorf("post-registration hook failed: %w", err)
		}
	}

	if s.emailService != nil {
		if err := s.RequestEmailVerification(ctx, req.Email, emailVerificationToken.String); err != nil {
			return nil, fmt.Errorf("failed to send email verification: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, time.Time, *coregen.User, error) {

	user, err := s.core.Queries.GetUserByEmail(ctx, pgtype.Text{String: strings.ToLower(email), Valid: true})
	if err != nil {
		return "", time.Time{}, nil, ErrInvalidUser
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte(password)); err != nil {
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

	err = s.core.Queries.UpdateUserForgotPasswordToken(ctx, coregen.UpdateUserForgotPasswordTokenParams{
		ForgotPasswordToken: pgtype.Text{String: token, Valid: true},
		UserID:              user.ID,
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
		ForgotPasswordToken: pgtype.Text{},
		UserID:              user.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to clear reset token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *AuthService) RequestEmailVerification(ctx context.Context, address, token string) error {
	if s.emailService == nil {
		return fmt.Errorf("email service not available")
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		return fmt.Errorf("FRONTEND_URL environment variable is required")
	}

	verificationLink := fmt.Sprintf("%s/emailVerify?token=%s", frontendURL, token)

	return s.emailService.SendEmail(ctx, email.EmailRequest{
		To:         address,
		From:       fmt.Sprintf("%s <noreply@%s>", s.emailService.GetSender(), s.emailService.GetDomain()),
		Subject:    "Verify Your Email",
		TemplateID: email.EmailVerifyTemplate,
		TemplateData: types.JSONB{
			"VerifyLink": verificationLink,
		},
	})
}

func (s *AuthService) EmailVerify(ctx context.Context, token string) (*coregen.User, error) {
	user, err := s.core.Queries.GetUserByEmailVerificationToken(ctx, pgtype.Text{String: token, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("invalid verification token: %w", err)
	}

	err = s.core.Queries.UpdateUserEmailVerificationToken(ctx, coregen.UpdateUserEmailVerificationTokenParams{
		EmailVerificationToken: pgtype.Text{},
		UserID:                 user.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update verification token: %w", err)
	}

	return &user, nil
}
