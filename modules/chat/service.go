package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	gen "github.com/Gooowan/matchup/modules/chat/gen"
	"github.com/Gooowan/matchup/modules/core/utils"
	"github.com/Gooowan/matchup/modules/moderation"
)

type ChatService struct {
	DB            *pgxpool.Pool
	Queries       *gen.Queries
	ModerationSvc *moderation.ModerationService
}

func NewChatService(db *pgxpool.Pool, moderationSvc *moderation.ModerationService) *ChatService {
	return &ChatService{DB: db, Queries: gen.New(db), ModerationSvc: moderationSvc}
}

func (s *ChatService) CreateChat(ctx context.Context, user1ID, user2ID pgtype.UUID) (pgtype.UUID, error) {
	chat, err := s.Queries.CreateChat(ctx, gen.CreateChatParams{
		User1ID: user1ID,
		User2ID: user2ID,
	})
	if err != nil {
		return pgtype.UUID{}, err
	}
	return chat.ID, nil
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

	blocked, err := s.ModerationSvc.IsBlocked(ctx, senderID, otherID)
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
