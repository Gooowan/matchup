package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	coregen "github.com/Gooowan/matchup/modules/core/gen"
	"github.com/Gooowan/matchup/modules/core/utils"
)

func (s *AuthService) CreateJwtToken(ctx context.Context, userID pgtype.UUID) (string, time.Time, error) {
	nonce, err := s.core.Queries.IncrementUserNonce(ctx, userID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to increment user nonce: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(s.expiration)

	userIDString := utils.UUIDToString(userID)

	claims := JWTClaims{
		UserID: userIDString,
		Nonce:  nonce,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			Subject:   userIDString,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt, nil
}

func (s *AuthService) ValidateJwtToken(ctx context.Context, tokenString string) (*coregen.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	validAudience := slices.Contains(claims.RegisteredClaims.Audience, s.audience)
	if !validAudience || claims.RegisteredClaims.Issuer != s.issuer {
		return nil, fmt.Errorf("invalid token")
	}

	userUUID, err := utils.StringToUUID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	user, err := s.core.Queries.GetUser(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	if user.AuthNonce > claims.Nonce {
		return nil, fmt.Errorf("invalid nonce")
	}
	if user.AuthNonce != claims.Nonce {
		return nil, fmt.Errorf("invalid nonce")
	}

	return &user, nil
}

func (s *AuthService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

func (s *AuthService) GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func (s *AuthService) UpdateUserPassword(ctx context.Context, user *coregen.User, newPassword string) error {
	hashed, err := s.HashPassword(newPassword)
	if err != nil {
		return ErrInternal
	}

	if err := s.core.Queries.UpdateUserPassword(ctx, coregen.UpdateUserPasswordParams{
		Password: pgtype.Text{String: hashed, Valid: true},
		UserID:   user.ID,
	}); err != nil {
		return ErrInternal
	}

	return nil
}
