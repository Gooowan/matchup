package otp

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/email"
	"github.com/valkey-io/valkey-go"
)

const (
	OTPLength     = 8
	OTPExpiration = 15 * time.Minute
	MaxAttempts   = 5
	OTPKeyPrefix  = "otp:user"
)

const (
	PurposeWithdraw = "withdraw"
)

type OTPData struct {
	Code      string    `json:"code"`
	Attempts  int       `json:"attempts"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

var (
	ErrOTPNotFound     = errors.New("OTP not found")
	ErrOTPExpired      = errors.New("OTP has expired")
	ErrOTPInvalid      = errors.New("invalid OTP code")
	ErrTooManyAttempts = errors.New("too many OTP attempts")
	ErrOTPGeneration   = errors.New("failed to generate OTP")
)

type OTPService struct {
	valkey       valkey.Client
	emailService *email.EmailService
}

func NewOTPService(valkey valkey.Client, emailService *email.EmailService) *OTPService {
	return &OTPService{
		valkey:       valkey,
		emailService: emailService,
	}
}

func (s *OTPService) CreateAndSendOTP(ctx context.Context, userID string, userEmail string, purpose string, templateData map[string]interface{}) error {
	code, err := s.generateOTPCode()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOTPGeneration, err)
	}

	now := time.Now()
	otpData := OTPData{
		Code:      code,
		Attempts:  0,
		CreatedAt: now,
		ExpiresAt: now.Add(OTPExpiration),
	}

	data, err := json.Marshal(otpData)
	if err != nil {
		return fmt.Errorf("failed to marshal OTP data: %v", err)
	}

	key := s.buildOTPKey(userID, purpose)
	cmd := s.valkey.Do(ctx, s.valkey.B().Set().Key(key).Value(string(data)).Ex(OTPExpiration).Build())
	if err := cmd.Error(); err != nil {
		return fmt.Errorf("failed to store OTP in Valkey: %v", err)
	}

	// Merge templateData with default values
	emailTemplateData := types.JSONB{
		"Code":      code,
		"ExpiresIn": int(OTPExpiration.Minutes()),
	}
	for k, v := range templateData {
		if k != "Code" && k != "ExpiresIn" {
			emailTemplateData[k] = v
		}
	}

	if err := s.emailService.SendEmail(ctx, email.EmailRequest{
		To:           userEmail,
		From:         fmt.Sprintf("%s <noreply@%s>", s.emailService.GetSender(), s.emailService.GetDomain()),
		Subject:      "Your Verification Code",
		TemplateID:   email.OTPCodeTemplate,
		TemplateData: emailTemplateData,
	}); err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	return nil
}

func (s *OTPService) ValidateOTP(ctx context.Context, userID string, purpose string, code string) error {
	key := s.buildOTPKey(userID, purpose)

	cmd := s.valkey.Do(ctx, s.valkey.B().Get().Key(key).Build())
	if err := cmd.Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			return ErrOTPNotFound
		}
		return fmt.Errorf("failed to get OTP from Valkey: %v", err)
	}

	data, err := cmd.ToString()
	if err != nil {
		return fmt.Errorf("failed to convert OTP data to string: %v", err)
	}

	var otpData OTPData
	if err := json.Unmarshal([]byte(data), &otpData); err != nil {
		return fmt.Errorf("failed to unmarshal OTP data: %v", err)
	}

	if time.Now().After(otpData.ExpiresAt) {
		s.deleteOTP(ctx, key)
		return ErrOTPExpired
	}

	if otpData.Attempts >= MaxAttempts {
		s.deleteOTP(ctx, key)
		return ErrTooManyAttempts
	}

	otpData.Attempts++

	if otpData.Code != code {
		updatedData, _ := json.Marshal(otpData)
		remainingTTL := time.Until(otpData.ExpiresAt)
		s.valkey.Do(ctx, s.valkey.B().Set().Key(key).Value(string(updatedData)).Ex(remainingTTL).Build())

		return ErrOTPInvalid
	}

	s.deleteOTP(ctx, key)
	return nil
}

func (s *OTPService) generateOTPCode() (string, error) {
	min := int64(10000000)
	max := int64(99999999)

	n, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%08d", n.Int64()+min), nil
}

func (s *OTPService) buildOTPKey(userID string, purpose string) string {
	return fmt.Sprintf("%s:%s:%s", OTPKeyPrefix, userID, purpose)
}

func (s *OTPService) deleteOTP(ctx context.Context, key string) {
	s.valkey.Do(ctx, s.valkey.B().Del().Key(key).Build())
}
