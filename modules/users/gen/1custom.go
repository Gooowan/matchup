package gen

import (
	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
)

type UserDTO struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	// TelegramID   int64       `json:"telegram_id"`
	Email      string `json:"email"`
	ReferralID int64  `json:"referral_id"`
	InviterID  string `json:"inviter_id"`
	// TelegramData types.JSONB `json:"telegram_data"`
	ProfileData types.JSONB `json:"profile_data"`
	CreatedAt   int64       `json:"created_at"`
}

type Users []User

func (users Users) ToDTO() []*UserDTO {
	if len(users) == 0 {
		return nil
	}

	dtos := make([]*UserDTO, len(users))
	for i, user := range users {
		dtos[i] = user.ToDTO()
	}
	return dtos
}

func (user User) ToDTO() *UserDTO {
	return &UserDTO{
		ID: utils.UUIDToString(user.ID),
		// TelegramID:   user.TelegramID.Int64,
		Role:       user.Role,
		Email:      user.Email.String,
		ReferralID: user.ReferralID,
		InviterID:  utils.UUIDToString(user.InviterID),
		// TelegramData: user.TelegramData,
		ProfileData: user.ProfileData,
		CreatedAt:   user.CreatedAt.Time.UnixMilli(),
	}
}

type AdminUserDTO struct {
	ID                     string      `json:"id"`
	TelegramID             *int64      `json:"telegram_id"`
	Email                  string      `json:"email"`
	ReferralID             int64       `json:"referral_id"`
	InviterID              string      `json:"inviter_id"`
	Metadata               types.JSONB `json:"metadata"`
	ProfileData            types.JSONB `json:"profile_data"`
	TelegramData           types.JSONB `json:"telegram_data"`
	CreatedAt              int64       `json:"created_at"`
	Role                   string      `json:"role"`
	AuthNonce              int32       `json:"auth_nonce"`
	ForgotPasswordToken    *string     `json:"forgot_password_token"`
	EmailVerificationToken *string     `json:"email_verification_token"`
}

func (user AdminSearchUsersRow) ToAdminDTO() *AdminUserDTO {
	dto := &AdminUserDTO{
		ID:           utils.UUIDToString(user.ID),
		Email:        user.Email.String,
		ReferralID:   user.ReferralID,
		InviterID:    utils.UUIDToString(user.InviterID),
		Metadata:     user.Metadata,
		ProfileData:  user.ProfileData,
		TelegramData: user.TelegramData,
		CreatedAt:    user.CreatedAt.Time.UnixMilli(),
		Role:         user.Role,
		AuthNonce:    user.AuthNonce,
	}

	if user.TelegramID.Valid {
		dto.TelegramID = &user.TelegramID.Int64
	}

	return dto
}

func (user AdminGetUserRow) ToAdminDTO() *AdminUserDTO {
	dto := &AdminUserDTO{
		ID:           utils.UUIDToString(user.ID),
		Email:        user.Email.String,
		ReferralID:   user.ReferralID,
		InviterID:    utils.UUIDToString(user.InviterID),
		Metadata:     user.Metadata,
		ProfileData:  user.ProfileData,
		TelegramData: user.TelegramData,
		CreatedAt:    user.CreatedAt.Time.UnixMilli(),
		Role:         user.Role,
		AuthNonce:    user.AuthNonce,
	}

	if user.TelegramID.Valid {
		dto.TelegramID = &user.TelegramID.Int64
	}
	if user.ForgotPasswordToken.Valid {
		dto.ForgotPasswordToken = &user.ForgotPasswordToken.String
	}
	if user.EmailVerificationToken.Valid {
		dto.EmailVerificationToken = &user.EmailVerificationToken.String
	}

	return dto
}
