package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Gooowan/matchup/modules/core/utils"
	gen "github.com/Gooowan/matchup/modules/matchup/gen"
)

type ChatService struct {
	DB      *pgxpool.Pool
	Queries *gen.Queries
}

func NewChatService(db *pgxpool.Pool, queries *gen.Queries) *ChatService {
	return &ChatService{DB: db, Queries: queries}
}

func (s *ChatService) ListChats(ctx context.Context, userID pgtype.UUID) ([]gen.ChatDTO, error) {
	rows, err := s.Queries.ListUserChats(ctx, userID)
	if err != nil {
		return nil, err
	}

	dtos := make([]gen.ChatDTO, len(rows))
	for i, row := range rows {
		otherID := gen.ParseOtherUserID(row.OtherUserID)

		dto := gen.ChatDTO{
			ID:          utils.UUIDToString(row.ID),
			OtherUserID: utils.UUIDToString(otherID),
			CreatedAt:   row.CreatedAt.Time.UnixMilli(),
		}

		// attach latest message preview
		msg, err := s.Queries.GetLatestMessage(ctx, row.ID)
		if err == nil {
			msgDTO := msg.ToDTO()
			dto.LastMessage = &msgDTO
		}

		dtos[i] = dto
	}

	return dtos, nil
}

func (s *ChatService) GetMessages(ctx context.Context, chatID, userID pgtype.UUID, cursor time.Time, limit int32) ([]gen.MessageDTO, error) {
	// verify user belongs to chat
	chat, err := s.Queries.GetChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("chat not found")
	}
	if chat.User1ID != userID && chat.User2ID != userID {
		return nil, fmt.Errorf("access denied")
	}

	msgs, err := s.Queries.ListMessages(ctx, gen.ListMessagesParams{
		ChatID:     chatID,
		CursorTime: pgtype.Timestamp{Time: cursor, Valid: true},
		LimitVal:   limit,
	})
	if err != nil {
		return nil, err
	}

	dtos := make([]gen.MessageDTO, len(msgs))
	for i, m := range msgs {
		dtos[i] = m.ToDTO()
	}
	return dtos, nil
}

func (s *ChatService) SendMessage(ctx context.Context, chatID, senderID pgtype.UUID, msgType, content string) (*gen.MessageDTO, error) {
	// verify sender belongs to chat
	chat, err := s.Queries.GetChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("chat not found")
	}
	if chat.User1ID != senderID && chat.User2ID != senderID {
		return nil, fmt.Errorf("access denied")
	}

	// check block status
	var otherID pgtype.UUID
	if chat.User1ID == senderID {
		otherID = chat.User2ID
	} else {
		otherID = chat.User1ID
	}

	blocked, err := s.Queries.IsBlocked(ctx, gen.IsBlockedParams{
		User1ID: senderID,
		User2ID: otherID,
	})
	if err == nil && blocked {
		return nil, fmt.Errorf("cannot send message to blocked user")
	}

	msg, err := s.Queries.CreateMessage(ctx, gen.CreateMessageParams{
		ChatID:   chatID,
		SenderID: senderID,
		Type:     msgType,
		Content:  content,
	})
	if err != nil {
		return nil, err
	}

	dto := msg.ToDTO()
	return &dto, nil
}
