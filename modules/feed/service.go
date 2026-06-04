package feed

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Gooowan/matchup/modules/chat"
	"github.com/Gooowan/matchup/modules/clubs"
	"github.com/Gooowan/matchup/modules/core/geocoding"
	"github.com/Gooowan/matchup/modules/core/logging"
	"github.com/Gooowan/matchup/modules/core/metrics"
	"github.com/Gooowan/matchup/modules/core/tracing"
	"github.com/Gooowan/matchup/modules/feed/gen"
	"github.com/Gooowan/matchup/modules/moderation"
	"github.com/Gooowan/matchup/modules/recommendation"
	recgen "github.com/Gooowan/matchup/modules/recommendation/gen"
	recTier1 "github.com/Gooowan/matchup/modules/recommendation/tier1"
	recTier2 "github.com/Gooowan/matchup/modules/recommendation/tier2"
	recTier3 "github.com/Gooowan/matchup/modules/recommendation/tier3"
)

// PushNotifier is a minimal interface so feed doesn't import the push package.
type PushNotifier interface {
	SendToUser(ctx context.Context, userID string, title, body string)
}

type FeedService struct {
	DB                *pgxpool.Pool
	Queries           *gen.Queries
	ChatSvc           *chat.ChatService
	ModerationSvc     *moderation.ModerationService
	RecommendationSvc *recommendation.RecommendationService
	ClubSvc           *clubs.ClubService
	Recommender       RecommendationProvider
	PushSvc           PushNotifier
}

func NewFeedService(
	db *pgxpool.Pool,
	chatSvc *chat.ChatService,
	moderationSvc *moderation.ModerationService,
	recommendationSvc *recommendation.RecommendationService,
	clubSvc *clubs.ClubService,
) *FeedService {
	// Tiers map: Tier3 = mutual-match, Tier2 = one-way filter, Tier1 = pure distance.
	// The Recommender runs them in order: Tier3 → Tier2 → Tier1.
	tier1 := recTier1.NewProvider(recommendationSvc.Queries)
	tier2 := recTier2.NewProvider(recommendationSvc.Queries)
	tier3 := recTier3.NewProvider(recommendationSvc.Queries)
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

// ErrNoProfile is returned by GetFeed when the requesting user has no
// recommendation profile, so callers can surface a "complete your profile"
// prompt instead of a silent empty deck.
var ErrNoProfile = errors.New("user has no profile")

func (s *FeedService) GetFeed(ctx context.Context, userID pgtype.UUID, limit int32) ([]RankedRow, error) {
	ctx, span := tracing.StartDBSpan(ctx, "GetFeed", "profiles")
	defer span.End()

	profile, err := s.RecommendationSvc.Queries.GetProfileByUserID(ctx, userID)
	if err != nil {
		logging.FromContext(ctx).Warn("feed: viewer has no profile, returning empty feed", "user_id", userID.String())
		return nil, ErrNoProfile
	}

	// Reference coordinates are club-based: use the user's primary club coords,
	// then their other clubs' centroid, then city centroid — so distance is club-to-club.
	var lat, lng float64
	var refSet bool
	if profile.PrimaryClubID.Valid {
		if rawClubs, err2 := s.ClubSvc.GetUserClubs(ctx, userID); err2 == nil {
			for _, c := range rawClubs {
				if c.ID == profile.PrimaryClubID {
					lat, lng = c.Latitude, c.Longitude
					refSet = (lat != 0 || lng != 0)
					break
				}
			}
		}
	}
	if !refSet {
		// Fall back to city centroid so the feed always runs.
		country := ""
		city := ""
		if profile.Country.Valid {
			country = profile.Country.String
		}
		if profile.City.Valid {
			city = profile.City.String
		}
		lat, lng = geocoding.CityLatLng(country, city)
	}

	// build exclude list: swiped + blocked
	swipedIDs, err := s.Queries.GetSwipedUserIDs(ctx, userID)
	if err != nil {
		logging.FromContext(ctx).Warn("failed to get swiped user IDs, proceeding without", "error", err)
		swipedIDs = nil
	}
	blockedIDs, err := s.ModerationSvc.Queries.GetBlockedUserIDs(ctx, userID)
	if err != nil {
		logging.FromContext(ctx).Warn("failed to get blocked user IDs, proceeding without", "error", err)
		blockedIDs = nil
	}

	excludeIDs := make([]pgtype.UUID, 0, len(swipedIDs)+len(blockedIDs))
	excludeIDs = append(excludeIDs, swipedIDs...)
	excludeIDs = append(excludeIDs, blockedIDs...)

	prefs, err := s.RecommendationSvc.Queries.GetPreferences(ctx, userID)
	if err != nil {
		logging.FromContext(ctx).Warn("no preferences found, using defaults", "error", err)
	}

	country := ""
	if profile.Country.Valid {
		country = profile.Country.String
	}

	// Collect all clubs the user belongs to (for any future near-club logic).
	var userClubs []recommendation.UserClub
	if s.ClubSvc != nil {
		if rawClubs, err2 := s.ClubSvc.GetUserClubs(ctx, userID); err2 == nil {
			for _, c := range rawClubs {
				userClubs = append(userClubs, recommendation.UserClub{
					ID:        c.ID,
					Latitude:  c.Latitude,
					Longitude: c.Longitude,
				})
			}
		}
	}

	myProfile := &recommendation.MyProfileData{
		Gender:     profile.Gender,
		BirthDate:  profile.BirthDate,
		HeightCm:   profile.HeightCm,
		Goal:       profile.Goal,
		Program:    profile.Program,
		Categories: profile.Categories,
	}
	if profile.City.Valid {
		myProfile.City = profile.City.String
	}
	if profile.Country.Valid {
		myProfile.Country = profile.Country.String
	}

	return s.Recommender.GetFeed(ctx, FeedParams{
		UserID:     userID,
		Latitude:   lat,
		Longitude:  lng,
		Country:    country,
		UserClubs:  userClubs,
		Prefs:      &prefs,
		ExcludeIDs: excludeIDs,
		Limit:      limit,
		MyProfile:  myProfile,
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

	// Create chat on mutual match (best-effort; match rows are already committed).
	// A chat creation failure must not prevent the match from being surfaced to the
	// user — they can still navigate to the chat later via the inbox.
	if result.IsMutualMatch {
		u1, u2 := orderUUIDs(fromUserID, toUserID)
		chatID, chatErr := s.ChatSvc.CreateChat(ctx, u1, u2)
		if chatErr != nil {
			logging.FromContext(ctx).Error("mutual match: failed to create chat (match rows committed)",
				"error", chatErr,
				"user1", u1.String(),
				"user2", u2.String(),
			)
			// Continue without chat_id — frontend will show the match popup
			// and the user can open the chat from the inbox.
		} else {
			result.ChatID = &chatID
		}

		// Notify the other user who liked first.
		if s.PushSvc != nil {
			toID := toUserID.String()
			go s.PushSvc.SendToUser(context.Background(), toID, "New Match!", "You have a new match on MatchUp. Say hi!")
		}
	}

	// Track swipe and match metrics
	metrics.SwipeEventsTotal.WithLabelValues(action, source).Inc()
	if result.IsMutualMatch {
		metrics.MatchEventsTotal.Inc()
	}

	// Log liked profile features for Tier 2/3 recommendation learning.
	if action == "LIKE" {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logging.FromContext(ctx).Error("logLike goroutine panicked", "recover", r)
				}
			}()
			logCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			s.logLike(logCtx, fromUserID, toUserID)
		}()
	}

	return result, nil
}

func (s *FeedService) logLike(ctx context.Context, fromUserID, toUserID pgtype.UUID) {
	profile, err := s.RecommendationSvc.Queries.GetProfileByUserID(ctx, toUserID)
	if err != nil {
		return
	}
	features := recgen.ProfileToFeatures(profile)
	_ = s.RecommendationSvc.Queries.InsertLikeLog(ctx, recgen.InsertLikeLogParams{
		UserID:   fromUserID,
		LikedID:  toUserID,
		Features: features,
	})
}

func orderUUIDs(a, b pgtype.UUID) (pgtype.UUID, pgtype.UUID) {
	if bytes.Compare(a.Bytes[:], b.Bytes[:]) < 0 {
		return a, b
	}
	return b, a
}
