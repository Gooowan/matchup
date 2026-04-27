package gen

import (
	"os"
	"strings"

	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
)

// normalizeAvatarURL rewrites legacy http://localhost:9000 avatar URLs to the
// current MINIO_PUBLIC_ENDPOINT so avatars stored before the env-var fix load correctly.
func normalizeAvatarURL(url string) string {
	if url == "" {
		return url
	}
	pub := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	if pub == "" || strings.HasPrefix(url, pub) {
		return url
	}
	for _, legacy := range []string{"http://localhost:9000", "http://minio:9000"} {
		if strings.HasPrefix(url, legacy) {
			return pub + url[len(legacy):]
		}
	}
	return url
}

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
	pd := user.ProfileData
	if avatar, ok := pd["avatar"].(string); ok && avatar != "" {
		pd = make(types.JSONB, len(user.ProfileData))
		for k, v := range user.ProfileData {
			pd[k] = v
		}
		pd["avatar"] = normalizeAvatarURL(avatar)
	}
	return &UserDTO{
		ID:          utils.UUIDToString(user.ID),
		Role:        user.Role,
		Email:       user.Email.String,
		InviterID:   utils.UUIDToString(user.InviterID),
		ProfileData: pd,
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
