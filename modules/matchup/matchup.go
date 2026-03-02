package matchup

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Gooowan/matchup/modules/matchup/chat"
	"github.com/Gooowan/matchup/modules/matchup/feed"
	gen "github.com/Gooowan/matchup/modules/matchup/gen"
	mapmod "github.com/Gooowan/matchup/modules/matchup/map"
	"github.com/Gooowan/matchup/modules/matchup/moderation"
	"github.com/Gooowan/matchup/modules/matchup/profile"
)

type MatchupModule struct {
	profileCtrl    *profile.ProfileController
	feedCtrl       *feed.FeedController
	chatCtrl       *chat.ChatController
	mapCtrl        *mapmod.MapController
	moderationCtrl *moderation.ModerationController
}

func NewMatchupModule(db *pgxpool.Pool) *MatchupModule {
	queries := gen.New(db)

	profileSvc := profile.NewProfileService(db, queries)
	feedSvc := feed.NewFeedService(db, queries)
	chatSvc := chat.NewChatService(db, queries)
	mapSvc := mapmod.NewMapService(queries)
	moderationSvc := moderation.NewModerationService(queries)

	return &MatchupModule{
		profileCtrl:    profile.NewProfileController(profileSvc),
		feedCtrl:       feed.NewFeedController(feedSvc),
		chatCtrl:       chat.NewChatController(chatSvc),
		mapCtrl:        mapmod.NewMapController(mapSvc),
		moderationCtrl: moderation.NewModerationController(moderationSvc),
	}
}

func (m *MatchupModule) RegisterRoutes(r *gin.Engine, userAuth gin.HandlerFunc) {
	// Profile & preferences: /me/...
	meGroup := r.Group("/me")
	m.profileCtrl.RegisterRoutes(meGroup, userAuth)

	// Feed & swipe: /matchup/...
	matchupGroup := r.Group("/matchup")
	m.feedCtrl.RegisterRoutes(matchupGroup, userAuth)

	// Chats: /chats/...
	chatsGroup := r.Group("/chats")
	m.chatCtrl.RegisterRoutes(chatsGroup, userAuth)

	// Map: /map/...
	mapGroup := r.Group("/map")
	m.mapCtrl.RegisterRoutes(mapGroup, userAuth)

	// Profile preview (authenticated, but views other users)
	profilesGroup := r.Group("/profiles")
	profilesGroup.Use(userAuth)
	profilesGroup.GET("/:userId/preview", m.profileCtrl.GetProfilePreview)

	// Moderation: /users/:userId/block, /users/:userId/report
	m.moderationCtrl.RegisterRoutes(r, userAuth)
}
