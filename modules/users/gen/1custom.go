package gen

import (
	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
)

type UserDTO struct {
	ID          string      `json:"id"`
	Role        string      `json:"role"`
	Email       string      `json:"email"`
	InviterID   string      `json:"inviter_id"`
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
		ID:          utils.UUIDToString(user.ID),
		Role:        user.Role,
		Email:       user.Email.String,
		InviterID:   utils.UUIDToString(user.InviterID),
		ProfileData: user.ProfileData,
		CreatedAt:   user.CreatedAt.Time.UnixMilli(),
	}
}

type AdminUserDTO struct {
	ID                     string      `json:"id"`
	Email                  string      `json:"email"`
	InviterID              string      `json:"inviter_id"`
	Metadata               types.JSONB `json:"metadata"`
	ProfileData            types.JSONB `json:"profile_data"`
	CreatedAt              int64       `json:"created_at"`
	Role                   string      `json:"role"`
	AuthNonce              int32       `json:"auth_nonce"`
	ForgotPasswordToken    *string     `json:"forgot_password_token"`
	EmailVerificationToken *string     `json:"email_verification_token"`
}

func (user AdminSearchUsersRow) ToAdminDTO() *AdminUserDTO {
	return &AdminUserDTO{
		ID:          utils.UUIDToString(user.ID),
		Email:       user.Email.String,
		InviterID:   utils.UUIDToString(user.InviterID),
		Metadata:    user.Metadata,
		ProfileData: user.ProfileData,
		CreatedAt:   user.CreatedAt.Time.UnixMilli(),
		Role:        user.Role,
		AuthNonce:   user.AuthNonce,
	}
}

func (user AdminGetUserRow) ToAdminDTO() *AdminUserDTO {
	dto := &AdminUserDTO{
		ID:          utils.UUIDToString(user.ID),
		Email:       user.Email.String,
		InviterID:   utils.UUIDToString(user.InviterID),
		Metadata:    user.Metadata,
		ProfileData: user.ProfileData,
		CreatedAt:   user.CreatedAt.Time.UnixMilli(),
		Role:        user.Role,
		AuthNonce:   user.AuthNonce,
	}

	if user.ForgotPasswordToken.Valid {
		dto.ForgotPasswordToken = &user.ForgotPasswordToken.String
	}
	if user.EmailVerificationToken.Valid {
		dto.EmailVerificationToken = &user.EmailVerificationToken.String
	}

	return dto
}
