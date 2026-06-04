package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	gen "github.com/Gooowan/matchup/modules/chat/gen"
	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
	"github.com/Gooowan/matchup/modules/moderation"
	modgen "github.com/Gooowan/matchup/modules/moderation/gen"
)

// Sentinel errors used by controller for typed HTTP responses.
var (
	ErrChatNotFound   = errors.New("chat not found")
	ErrAccessDenied   = errors.New("access denied")
	ErrUserBlocked    = errors.New("cannot send message to blocked user")
	ErrContentBlocked = errors.New("content blocked")
	ErrNoMutualMatch  = errors.New("no mutual match with this user")
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

// CreateChatForUsers creates (or retrieves) a chat between callerID and targetID.
// Trainers may be messaged directly without a mutual match; dancers require one.
func (s *ChatService) CreateChatForUsers(ctx context.Context, callerID, targetID pgtype.UUID) (pgtype.UUID, error) {
	// Check if target is a trainer — trainers allow open DMs.
	var targetAccountType string
	err := s.DB.QueryRow(ctx,
		`SELECT account_type FROM profiles WHERE user_id = $1`, targetID).Scan(&targetAccountType)
	if err != nil {
		targetAccountType = "dancer" // default to requiring mutual match on lookup failure
	}

	if targetAccountType != "trainer" {
		var isMutual bool
		err := s.DB.QueryRow(ctx, `
			SELECT (
				EXISTS(SELECT 1 FROM matches WHERE from_user_id=$1 AND to_user_id=$2 AND action='LIKE')
				AND
				EXISTS(SELECT 1 FROM matches WHERE from_user_id=$2 AND to_user_id=$1 AND action='LIKE')
			)`, callerID, targetID).Scan(&isMutual)
		if err != nil {
			return pgtype.UUID{}, fmt.Errorf("mutual match check: %w", err)
		}
		if !isMutual {
			return pgtype.UUID{}, ErrNoMutualMatch
		}
	}

	u1, u2 := orderUUIDs(callerID, targetID)
	return s.CreateChat(ctx, u1, u2)
}

// CreateClubChat creates (or retrieves) a chat between a dancer and a club.
// Works even for unclaimed clubs — the club side is answered by whoever owns it.
func (s *ChatService) CreateClubChat(ctx context.Context, userID, clubID pgtype.UUID) (pgtype.UUID, error) {
	chat, err := s.Queries.CreateClubChat(ctx, gen.CreateClubChatParams{
		User1ID: userID,
		ClubID:  clubID,
	})
	if err != nil {
		return pgtype.UUID{}, err
	}
	return chat.ID, nil
}

// userCanAccess reports whether userID may read/write a chat: they are one of the
// two DM participants, or they own the chat's club (club-side of a club chat).
func (s *ChatService) userCanAccess(ctx context.Context, chat gen.Chat, userID pgtype.UUID) bool {
	if chat.User1ID == userID {
		return true
	}
	if chat.User2ID.Valid && chat.User2ID == userID {
		return true
	}
	if chat.ClubID.Valid {
		var owns bool
		if err := s.DB.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM clubs WHERE id = $1 AND owner_user_id = $2)`,
			chat.ClubID, userID).Scan(&owns); err == nil {
			return owns
		}
	}
	return false
}

// ListChats returns all chats for userID enriched with peer user info,
// unread counts, last activity, and whether the peer is a club owner.
func (s *ChatService) ListChats(ctx context.Context, userID pgtype.UUID) ([]gen.ChatDTO, error) {
	// Includes both DMs (user1/user2) and club chats: those the caller started
	// (user1) and those addressed to clubs the caller owns. For a club chat the
	// "other_uid" is NULL on the dancer side (we show the club) and the dancer on
	// the owner side.
	const q = `
		WITH base AS (
			SELECT c.id, c.user1_id, c.user2_id, c.club_id, c.created_at,
				CASE
					WHEN c.club_id IS NULL THEN (CASE WHEN c.user1_id = $1 THEN c.user2_id ELSE c.user1_id END)
					WHEN c.user1_id = $1 THEN NULL
					ELSE c.user1_id
				END AS other_uid
			FROM chats c
			WHERE c.user1_id = $1
			   OR c.user2_id = $1
			   OR c.club_id IN (SELECT id FROM clubs WHERE owner_user_id = $1)
		)
		SELECT
			b.id,
			b.other_uid                                    AS other_user_id,
			u.profile_data                                 AS other_profile_data,
			b.created_at,
			COALESCE(lm.last_msg_at, b.created_at)          AS last_activity_at,
			COALESCE(unread.cnt, 0)::bigint                 AS unread_count,
			(b.club_id IS NOT NULL)                         AS is_club_chat,
			b.club_id                                       AS club_id,
			cl.name                                         AS club_name,
			cl.slug                                         AS club_slug,
			cl.phone                                        AS club_phone,
			cl.metadata                                     AS club_metadata
		FROM base b
		LEFT JOIN users u ON u.id = b.other_uid
		LEFT JOIN clubs cl ON cl.id = b.club_id
		LEFT JOIN LATERAL (
			SELECT created_at AS last_msg_at
			FROM messages
			WHERE chat_id = b.id
			ORDER BY created_at DESC
			LIMIT 1
		) lm ON true
		LEFT JOIN chat_reads cr ON cr.chat_id = b.id AND cr.user_id = $1
		LEFT JOIN LATERAL (
			SELECT COUNT(*)::bigint AS cnt
			FROM messages
			WHERE chat_id  = b.id
			  AND sender_id != $1
			  AND created_at > COALESCE(cr.last_read_at, b.created_at - INTERVAL '1 millisecond')
		) unread ON true
		ORDER BY COALESCE(lm.last_msg_at, b.created_at) DESC
		LIMIT 100`

	rows, err := s.DB.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dtos []gen.ChatDTO
	for rows.Next() {
		var (
			chatID         pgtype.UUID
			otherUserID    pgtype.UUID
			rawProfileData []byte
			createdAt      pgtype.Timestamp
			lastActivityAt pgtype.Timestamp
			unreadCount    int64
			isClubChat     bool
			clubID         pgtype.UUID
			clubName       pgtype.Text
			clubSlug       pgtype.Text
			clubPhone      pgtype.Text
			clubMeta       []byte
		)
		if err := rows.Scan(
			&chatID,
			&otherUserID,
			&rawProfileData,
			&createdAt,
			&lastActivityAt,
			&unreadCount,
			&isClubChat,
			&clubID,
			&clubName,
			&clubSlug,
			&clubPhone,
			&clubMeta,
		); err != nil {
			return nil, err
		}

		dto := gen.ChatDTO{
			ID:           utils.UUIDToString(chatID),
			UnreadCount:  unreadCount,
			IsClubChat:   isClubChat,
			IsClubOwner:  isClubChat, // deprecated alias kept for backward compatibility
			CreatedAt:    createdAt.Time.UnixMilli(),
			LastActivity: lastActivityAt.Time.UnixMilli(),
		}

		// Peer user (DM, or the dancer when an owner views a club chat).
		if otherUserID.Valid {
			var profileData types.JSONB
			if len(rawProfileData) > 0 {
				_ = json.Unmarshal(rawProfileData, &profileData)
			}
			if profileData == nil {
				profileData = types.JSONB{}
			}
			normalizeAvatarInJSONB(profileData)
			dto.OtherUserID = utils.UUIDToString(otherUserID)
			dto.OtherUser = &gen.ChatPeerDTO{
				ID:          utils.UUIDToString(otherUserID),
				ProfileData: slimProfileData(profileData),
			}
		}

		// Club identity for club chats (used to brand the thread on the dancer side).
		if clubID.Valid {
			dto.Club = &gen.ChatClubDTO{
				ID:    utils.UUIDToString(clubID),
				Name:  clubName.String,
				Slug:  clubSlug.String,
				Phone: clubPhone.String,
				Logo:  firstClubPhoto(clubMeta),
			}
		}

		// Attach last message preview.
		msg, msgErr := s.Queries.GetLatestMessage(ctx, chatID)
		if msgErr == nil {
			m := msg.ToDTOFor(userID)
			dto.LastMessage = &m
		}

		dtos = append(dtos, dto)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if dtos == nil {
		dtos = []gen.ChatDTO{}
	}
	return dtos, nil
}

// GetMessages returns paginated messages for chatID.
//
// When afterTime is non-zero, returns messages created strictly after that
// timestamp in ascending order (used by the polling loop).
// Otherwise uses cursor (messages before that timestamp, ascending order) for
// initial page load and backward pagination.
func (s *ChatService) GetMessages(ctx context.Context, chatID, userID pgtype.UUID, cursor time.Time, afterTime time.Time, limit int32) ([]gen.MessageDTO, error) {
	chat, err := s.Queries.GetChat(ctx, chatID)
	if err != nil {
		return nil, ErrChatNotFound
	}
	if !s.userCanAccess(ctx, chat, userID) {
		return nil, ErrAccessDenied
	}

	var msgs []gen.Message

	if !afterTime.IsZero() {
		// Polling mode: fetch new messages after afterTime, ascending.
		// Exclude hidden/deleted messages to keep the chat view clean.
		rows, err := s.DB.Query(ctx, `
			SELECT id, chat_id, sender_id, type, content, moderation_status, deleted_at, created_at
			FROM messages
			WHERE chat_id = $1
			  AND created_at > $2
			  AND deleted_at IS NULL
			  AND (moderation_status IS NULL OR moderation_status != 'hidden')
			ORDER BY created_at ASC
			LIMIT $3`,
			chatID, afterTime, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var m gen.Message
			if err := rows.Scan(&m.ID, &m.ChatID, &m.SenderID, &m.Type, &m.Content, &m.ModerationStatus, &m.DeletedAt, &m.CreatedAt); err != nil {
				return nil, err
			}
			msgs = append(msgs, m)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	} else {
		// Pagination mode: fetch messages before cursor, then reverse to ASC.
		raw, err := s.Queries.ListMessages(ctx, gen.ListMessagesParams{
			ChatID:     chatID,
			CursorTime: pgtype.Timestamp{Time: cursor, Valid: true},
			LimitVal:   limit,
		})
		if err != nil {
			return nil, err
		}
		// Reverse DESC → ASC.
		msgs = make([]gen.Message, len(raw))
		for i, m := range raw {
			msgs[len(raw)-1-i] = m
		}
	}

	dtos := make([]gen.MessageDTO, len(msgs))
	for i, m := range msgs {
		dtos[i] = m.ToDTOFor(userID)
	}
	return dtos, nil
}

// SendMessage validates access, checks blocks, filters content, and creates the message.
func (s *ChatService) SendMessage(ctx context.Context, chatID, senderID pgtype.UUID, msgType, content string) (*gen.MessageDTO, error) {
	chat, err := s.Queries.GetChat(ctx, chatID)
	if err != nil {
		return nil, ErrChatNotFound
	}
	if !s.userCanAccess(ctx, chat, senderID) {
		return nil, ErrAccessDenied
	}

	// Block checks only apply to user<->user DMs; club chats have no peer user.
	if !chat.ClubID.Valid {
		var otherID pgtype.UUID
		if chat.User1ID == senderID {
			otherID = chat.User2ID
		} else {
			otherID = chat.User1ID
		}

		if otherID.Valid {
			blocked, bErr := s.ModerationSvc.Queries.IsBlocked(ctx, modgen.IsBlockedParams{
				User1ID: senderID,
				User2ID: otherID,
			})
			if bErr == nil && blocked {
				return nil, ErrUserBlocked
			}
		}
	}

	if ok, reason := isMessageAllowed(content); !ok {
		return nil, fmt.Errorf("%w: %s", ErrContentBlocked, reason)
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

	dto := msg.ToDTOFor(senderID)
	return &dto, nil
}

// GetChatMeta returns the peer's public profile data for a chat thread,
// so the chat detail screen can render the partner's avatar and name without
// the caller having to round-trip through the full inbox.
func (s *ChatService) GetChatMeta(ctx context.Context, chatID, userID pgtype.UUID) (*gen.ChatPeerDTO, error) {
	chat, err := s.Queries.GetChat(ctx, chatID)
	if err != nil {
		return nil, ErrChatNotFound
	}
	if !s.userCanAccess(ctx, chat, userID) {
		return nil, ErrAccessDenied
	}

	// Club chat viewed by the dancer → brand the thread as the club itself.
	if chat.ClubID.Valid && chat.User1ID == userID {
		var (
			name, slug string
			phone      pgtype.Text
			rawMeta    []byte
		)
		if err := s.DB.QueryRow(ctx,
			`SELECT name, slug, phone, metadata FROM clubs WHERE id = $1`,
			chat.ClubID).Scan(&name, &slug, &phone, &rawMeta); err != nil {
			return nil, err
		}
		return &gen.ChatPeerDTO{
			ID:          utils.UUIDToString(chat.ClubID),
			ProfileData: types.JSONB{},
			Kind:        "club",
			ClubName:    name,
			ClubSlug:    slug,
			ClubPhone:   phone.String,
			ClubLogo:    firstClubPhoto(rawMeta),
		}, nil
	}

	// DM, or club chat viewed by the owner → return the peer user (the dancer).
	otherID := chat.User1ID
	if chat.User1ID == userID {
		otherID = chat.User2ID
	}

	var rawProfileData []byte
	err = s.DB.QueryRow(ctx, `SELECT profile_data FROM users WHERE id = $1`, otherID).Scan(&rawProfileData)
	if err != nil {
		return nil, err
	}

	var profileData types.JSONB
	if len(rawProfileData) > 0 {
		_ = json.Unmarshal(rawProfileData, &profileData)
	}
	if profileData == nil {
		profileData = types.JSONB{}
	}
	normalizeAvatarInJSONB(profileData)

	return &gen.ChatPeerDTO{
		ID:          utils.UUIDToString(otherID),
		ProfileData: slimProfileData(profileData),
	}, nil
}

// slimProfileData returns a JSONB with only the fields the chat UI needs,
// avoiding transmitting the full profile blob over the wire.
func slimProfileData(pd types.JSONB) types.JSONB {
	slim := types.JSONB{}
	for _, key := range []string{"first_name", "last_name", "avatar"} {
		if v, ok := pd[key]; ok {
			slim[key] = v
		}
	}
	return slim
}

// ReportMessage creates a message_reports row so an admin can review it.
// The reporting user must be a participant in the chat.
func (s *ChatService) ReportMessage(ctx context.Context, chatID, messageID, reporterID pgtype.UUID, category, comment string) (*gen.MessageReport, error) {
	chat, err := s.Queries.GetChat(ctx, chatID)
	if err != nil {
		return nil, ErrChatNotFound
	}
	if !s.userCanAccess(ctx, chat, reporterID) {
		return nil, ErrAccessDenied
	}

	msg, err := s.Queries.GetMessageByID(ctx, messageID)
	if err != nil || msg.ChatID != chatID {
		return nil, ErrChatNotFound
	}

	// A user cannot report their own messages.
	if msg.SenderID == reporterID {
		return nil, fmt.Errorf("cannot report your own message")
	}

	commentField := pgtype.Text{}
	if comment != "" {
		commentField = pgtype.Text{String: comment, Valid: true}
	}

	report, err := s.Queries.CreateMessageReport(ctx, gen.CreateMessageReportParams{
		MessageID:       messageID,
		ChatID:          chatID,
		ReporterID:      reporterID,
		ReportedUserID:  msg.SenderID,
		Category:        category,
		Comment:         commentField,
		ContentSnapshot: msg.Content,
	})
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// AdminHideMessage soft-deletes a message (sets moderation_status=hidden, deleted_at=now).
func (s *ChatService) AdminHideMessage(ctx context.Context, messageID pgtype.UUID) error {
	return s.Queries.HideMessage(ctx, messageID)
}

// AdminListMessageReports returns open message reports for the admin queue.
func (s *ChatService) AdminListMessageReports(ctx context.Context, status string, limit, offset int32) ([]gen.ListMessageReportsRow, error) {
	return s.Queries.ListMessageReports(ctx, gen.ListMessageReportsParams{
		Status:    status,
		LimitVal:  limit,
		OffsetVal: offset,
	})
}

// AdminResolveMessageReport marks a report as resolved or dismissed.
func (s *ChatService) AdminResolveMessageReport(ctx context.Context, reportID, adminID pgtype.UUID, status string) error {
	return s.Queries.ResolveMessageReport(ctx, gen.ResolveMessageReportParams{
		ID:         reportID,
		Status:     status,
		ResolvedBy: adminID,
	})
}

// MarkChatRead updates (or inserts) the last_read_at timestamp for the user in a chat.
func (s *ChatService) MarkChatRead(ctx context.Context, chatID, userID pgtype.UUID) error {
	chat, err := s.Queries.GetChat(ctx, chatID)
	if err != nil {
		return ErrChatNotFound
	}
	if !s.userCanAccess(ctx, chat, userID) {
		return ErrAccessDenied
	}

	_, err = s.DB.Exec(ctx, `
		INSERT INTO chat_reads(chat_id, user_id, last_read_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (chat_id, user_id) DO UPDATE SET last_read_at = EXCLUDED.last_read_at`,
		chatID, userID)
	return err
}

// orderUUIDs returns the two UUIDs in a deterministic order (smaller bytes first).
func orderUUIDs(a, b pgtype.UUID) (pgtype.UUID, pgtype.UUID) {
	for i := 0; i < 16; i++ {
		if a.Bytes[i] < b.Bytes[i] {
			return a, b
		}
		if a.Bytes[i] > b.Bytes[i] {
			return b, a
		}
	}
	return a, b
}

// firstClubPhoto extracts a usable logo URL from a club's metadata JSONB,
// checking logo_url, logo, then the first entry of a "photos" array.
func firstClubPhoto(rawMeta []byte) string {
	if len(rawMeta) == 0 {
		return ""
	}
	var meta map[string]any
	if err := json.Unmarshal(rawMeta, &meta); err != nil {
		return ""
	}
	// Prefer explicit logo_url (set by Google import).
	if lu, ok := meta["logo_url"].(string); ok && lu != "" {
		return normalizeURL(lu)
	}
	if logo, ok := meta["logo"].(string); ok && logo != "" {
		return normalizeURL(logo)
	}
	if photos, ok := meta["photos"].([]any); ok {
		for _, p := range photos {
			if u, ok := p.(string); ok && u != "" {
				return normalizeURL(u)
			}
		}
	}
	return ""
}

// normalizeAvatarInJSONB rewrites legacy MinIO avatar URLs in-place.
func normalizeAvatarInJSONB(pd types.JSONB) {
	if pd == nil {
		return
	}
	if avatar, ok := pd["avatar"].(string); ok && avatar != "" {
		pd["avatar"] = normalizeURL(avatar)
	}
}

// normalizeURL rewrites legacy MinIO host prefixes to the current public endpoint.
func normalizeURL(url string) string {
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
