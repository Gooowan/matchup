package gen

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
)

// ChatPeerDTO holds the peer's public identity for the chat detail screen.
// For DMs this is the other user. For club chats (Kind == "club") it is the
// club: name/logo/phone/slug, so the UI can brand the thread as the club and
// surface a call button when a phone is set.
type ChatPeerDTO struct {
	ID          string      `json:"id"`
	ProfileData types.JSONB `json:"profile_data"`
	Kind        string      `json:"kind,omitempty"`
	ClubName    string      `json:"club_name,omitempty"`
	ClubLogo    string      `json:"club_logo,omitempty"`
	ClubPhone   string      `json:"club_phone,omitempty"`
	ClubSlug    string      `json:"club_slug,omitempty"`
}

// ChatClubDTO is the club identity attached to a club chat in the inbox.
type ChatClubDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Logo  string `json:"logo,omitempty"`
	Slug  string `json:"slug"`
	Phone string `json:"phone,omitempty"`
}

type ChatDTO struct {
	ID           string       `json:"id"`
	OtherUserID  string       `json:"other_user_id"`
	OtherUser    *ChatPeerDTO `json:"other_user,omitempty"`
	Club         *ChatClubDTO `json:"club,omitempty"`
	UnreadCount  int64        `json:"unread_count"`
	// IsClubChat is true when this chat is a user-to-club thread (club_id IS NOT NULL).
	// Named is_club_chat in JSON (previously is_club_owner; both are emitted for compatibility).
	IsClubChat   bool         `json:"is_club_chat"`
	IsClubOwner  bool         `json:"is_club_owner"` // deprecated alias kept for backward compatibility
	CreatedAt    int64        `json:"created_at"`
	LastActivity int64        `json:"last_activity"`
	LastMessage  *MessageDTO  `json:"last_message,omitempty"`
}

type MessageDTO struct {
	ID        string `json:"id"`
	ChatID    string `json:"chat_id"`
	SenderID  string `json:"sender_id"`
	IsOwn     bool   `json:"is_own"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

func (m Message) ToDTO() MessageDTO {
	return MessageDTO{
		ID:        utils.UUIDToString(m.ID),
		ChatID:    utils.UUIDToString(m.ChatID),
		SenderID:  utils.UUIDToString(m.SenderID),
		IsOwn:     false,
		Type:      m.Type,
		Content:   m.Content,
		CreatedAt: m.CreatedAt.Time.UnixMilli(),
	}
}

// ToDTOFor is like ToDTO but sets IsOwn based on whether m.SenderID == callerID.
func (m Message) ToDTOFor(callerID pgtype.UUID) MessageDTO {
	return MessageDTO{
		ID:        utils.UUIDToString(m.ID),
		ChatID:    utils.UUIDToString(m.ChatID),
		SenderID:  utils.UUIDToString(m.SenderID),
		IsOwn:     m.SenderID.Valid && callerID.Valid && m.SenderID.Bytes == callerID.Bytes,
		Type:      m.Type,
		Content:   m.Content,
		CreatedAt: m.CreatedAt.Time.UnixMilli(),
	}
}

// ParseOtherUserID extracts the UUID from the CASE expression result
func ParseOtherUserID(val interface{}) pgtype.UUID {
	if b, ok := val.([16]byte); ok {
		return pgtype.UUID{Bytes: b, Valid: true}
	}
	return pgtype.UUID{}
}
