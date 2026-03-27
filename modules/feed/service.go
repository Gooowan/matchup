package feed

import (
	"bytes"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Gooowan/matchup/modules/chat"
	"github.com/Gooowan/matchup/modules/clubs"
	"github.com/Gooowan/matchup/modules/feed/gen"
	"github.com/Gooowan/matchup/modules/moderation"
	"github.com/Gooowan/matchup/modules/recommendation"
	recgen "github.com/Gooowan/matchup/modules/recommendation/gen"
	recTier1 "github.com/Gooowan/matchup/modules/recommendation/tier1"
	recTier2 "github.com/Gooowan/matchup/modules/recommendation/tier2"
	recTier3 "github.com/Gooowan/matchup/modules/recommendation/tier3"
)

type FeedService struct {
	DB                *pgxpool.Pool
	Queries           *gen.Queries
	ChatSvc           *chat.ChatService
	ModerationSvc     *moderation.ModerationService
	RecommendationSvc *recommendation.RecommendationService
	ClubSvc           *clubs.ClubService
	Recommender       RecommendationProvider
}

func NewFeedService(
	db *pgxpool.Pool,
	chatSvc *chat.ChatService,
	moderationSvc *moderation.ModerationService,
	recommendationSvc *recommendation.RecommendationService,
	clubSvc *clubs.ClubService,
) *FeedService {
	tier1 := recTier1.NewProvider(recommendationSvc.Queries)
	tier2 := recTier2.NewProvider()
	tier3 := recTier3.NewProvider()
	rec := recommendation.NewRecommender(tier1, tier2, tier3)

	return &FeedService{
		DB:                db,
		Queries:           gen.New(db),
		ChatSvc:           chatSvc,
		ModerationSvc:     moderationSvc,
		RecommendationSvc: recommendationSvc,
		ClubSvc:           clubSvc,
		Recommender:       NewTierRecommendationProvider(rec, recommendationSvc),
	}
}

func (s *FeedService) GetFeed(ctx context.Context, userID pgtype.UUID, limit int32) ([]recgen.FindNearbyVisibleProfilesRow, error) {
	profile, err := s.RecommendationSvc.Queries.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("profile required to get feed: %w", err)
	}

	if !profile.Latitude.Valid || !profile.Longitude.Valid {
		return nil, fmt.Errorf("profile location required")
	}

	// build exclude list: swiped + blocked
	swipedIDs, _ := s.Queries.GetSwipedUserIDs(ctx, userID)
	blockedIDs, _ := s.ModerationSvc.Queries.GetBlockedUserIDs(ctx, userID)

	excludeIDs := make([]pgtype.UUID, 0, len(swipedIDs)+len(blockedIDs))
	excludeIDs = append(excludeIDs, swipedIDs...)
	excludeIDs = append(excludeIDs, blockedIDs...)

	prefs, _ := s.RecommendationSvc.Queries.GetPreferences(ctx, userID)

	country := ""
	if profile.Country.Valid {
		country = profile.Country.String
	}

	return s.Recommender.GetFeed(ctx, FeedParams{
		UserID:     userID,
		Latitude:   profile.Latitude.Float64,
		Longitude:  profile.Longitude.Float64,
		Country:    country,
		Prefs:      &prefs,
		ExcludeIDs: excludeIDs,
		Limit:      limit,
	})
}

type SwipeResult struct {
	IsMutualMatch bool
	ChatID        *pgtype.UUID
}

func (s *FeedService) Swipe(ctx context.Context, fromUserID, toUserID pgtype.UUID, action, source string) (*SwipeResult, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.Queries.WithTx(tx)

	sourceVal := pgtype.Text{}
	if source != "" {
		sourceVal = pgtype.Text{String: source, Valid: true}
	}

	_, err = qtx.CreateMatch(ctx, gen.CreateMatchParams{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Action:     action,
		Source:     sourceVal,
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
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	// Create chat on mutual match (separate from match transaction)
	if result.IsMutualMatch {
		u1, u2 := orderUUIDs(fromUserID, toUserID)
		chatID, err := s.ChatSvc.CreateChat(ctx, u1, u2)
		if err != nil {
			return nil, fmt.Errorf("failed to create chat: %w", err)
		}
		result.ChatID = &chatID
	}

	return result, nil
}

func orderUUIDs(a, b pgtype.UUID) (pgtype.UUID, pgtype.UUID) {
	if bytes.Compare(a.Bytes[:], b.Bytes[:]) < 0 {
		return a, b
	}
	return b, a
}
