package feed

import (
	"bytes"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	gen "github.com/Gooowan/matchup/modules/matchup/gen"
)

type FeedService struct {
	DB          *pgxpool.Pool
	Queries     *gen.Queries
	Recommender RecommendationProvider
}

func NewFeedService(db *pgxpool.Pool, queries *gen.Queries) *FeedService {
	primary := NewNearestCandidatesProvider(queries)
	fallback := NewRandomFallbackProvider(queries)

	return &FeedService{
		DB:      db,
		Queries: queries,
		Recommender: &FallbackProvider{
			Primary:  primary,
			Fallback: fallback,
		},
	}
}

func (s *FeedService) GetFeed(ctx context.Context, userID pgtype.UUID, limit int32) ([]gen.FindNearbyVisibleProfilesRow, error) {
	profile, err := s.Queries.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("profile required to get feed: %w", err)
	}

	if !profile.Latitude.Valid || !profile.Longitude.Valid {
		return nil, fmt.Errorf("profile location required")
	}

	// build exclude list: swiped + blocked
	swipedIDs, _ := s.Queries.GetSwipedUserIDs(ctx, userID)
	blockedIDs, _ := s.Queries.GetBlockedUserIDs(ctx, userID)

	excludeIDs := make([]pgtype.UUID, 0, len(swipedIDs)+len(blockedIDs))
	excludeIDs = append(excludeIDs, swipedIDs...)
	excludeIDs = append(excludeIDs, blockedIDs...)

	prefs, _ := s.Queries.GetPreferences(ctx, userID)

	return s.Recommender.GetFeed(ctx, FeedParams{
		UserID:     userID,
		Latitude:   profile.Latitude.Float64,
		Longitude:  profile.Longitude.Float64,
		Prefs:      &prefs,
		ExcludeIDs: excludeIDs,
		Limit:      limit,
	})
}

type SwipeResult struct {
	IsMutualMatch bool
	ChatID        *pgtype.UUID
}

func (s *FeedService) Swipe(ctx context.Context, fromUserID, toUserID pgtype.UUID, action string) (*SwipeResult, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.Queries.WithTx(tx)

	_, err = qtx.CreateMatch(ctx, gen.CreateMatchParams{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Action:     action,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save swipe: %w", err)
	}

	result := &SwipeResult{}

	if action == "LIKE" {
		isMutual, err := qtx.CheckMutualMatch(ctx, gen.CheckMutualMatchParams{
			FromUserID: fromUserID,
			ToUserID:   toUserID,
		})
		if err == nil && isMutual {
			result.IsMutualMatch = true

			// order UUIDs for consistent UNIQUE constraint
			u1, u2 := orderUUIDs(fromUserID, toUserID)
			chat, err := qtx.CreateChat(ctx, gen.CreateChatParams{
				User1ID: u1,
				User2ID: u2,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create chat: %w", err)
			}
			result.ChatID = &chat.ID
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return result, nil
}

func orderUUIDs(a, b pgtype.UUID) (pgtype.UUID, pgtype.UUID) {
	if bytes.Compare(a.Bytes[:], b.Bytes[:]) < 0 {
		return a, b
	}
	return b, a
}
