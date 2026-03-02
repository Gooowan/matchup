package gen

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Gooowan/matchup/modules/core/utils"
)

type ChatDTO struct {
	ID          string      `json:"id"`
	OtherUserID string      `json:"other_user_id"`
	CreatedAt   int64       `json:"created_at"`
	LastMessage *MessageDTO `json:"last_message,omitempty"`
}

type MessageDTO struct {
	ID        string `json:"id"`
	ChatID    string `json:"chat_id"`
	SenderID  string `json:"sender_id"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

func (m Message) ToDTO() MessageDTO {
	return MessageDTO{
		ID:        utils.UUIDToString(m.ID),
		ChatID:    utils.UUIDToString(m.ChatID),
		SenderID:  utils.UUIDToString(m.SenderID),
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
