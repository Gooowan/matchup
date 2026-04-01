package main

import (
	"embed"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/valkey-io/valkey-go"

	"github.com/Gooowan/matchup/modules/chat"
	"github.com/Gooowan/matchup/modules/clubs"
	"github.com/Gooowan/matchup/modules/core/db"
	"github.com/Gooowan/matchup/modules/email"
	"github.com/Gooowan/matchup/modules/email/providers"
	"github.com/Gooowan/matchup/modules/feed"
	"github.com/Gooowan/matchup/modules/files"
	mapmod "github.com/Gooowan/matchup/modules/map"
	"github.com/Gooowan/matchup/modules/moderation"
	"github.com/Gooowan/matchup/modules/recommendation"
	"github.com/Gooowan/matchup/modules/subscriptions"
	core "github.com/Gooowan/matchup/modules/users"
	"github.com/Gooowan/matchup/modules/users/auth"
	"github.com/Gooowan/matchup/services/api/controllers"

	"github.com/Gooowan/matchup/modules/otp"
	"github.com/Gooowan/matchup/modules/ratelimit"
)

//go:embed templates/*.html
var emailTemplates embed.FS

func main() {
	dbpool, err := db.PostgresConnect()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer dbpool.Close()

	redisAddress := os.Getenv("REDIS_ADDRESS")
	if redisAddress == "" {
		log.Fatal("REDIS_ADDRESS isn't set")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})
	defer redisClient.Close()

	allowOrigins := os.Getenv("ALLOW_ORIGINS")
	if allowOrigins == "" {
		log.Fatal("ALLOW_ORIGINS isn't set")
	}
	allowedOrigins := strings.Split(allowOrigins, ",")

	coreService := core.NewCoreService(dbpool)
	rlService := ratelimit.NewRLService(redisClient)

	var emailProvider email.EmailProvider
	switch os.Getenv("EMAIL_PROVIDER") {
	case "resend":
		emailProvider, err = providers.NewResendProvider()
		if err != nil {
			log.Fatalf("Error initializing Resend provider: %v", err)
		}
	case "mailgun":
		emailProvider, err = providers.NewMailgunProvider()
		if err != nil {
			log.Fatalf("Error initializing Mailgun provider: %v", err)
		}
	default:
		emailProvider = providers.NewMockProvider()
	}

	emailService, err := email.NewEmailService(
		emailProvider,
		&emailTemplates,
		"MatchUp",
		[]email.TemplateID{
			email.EmailVerifyTemplate,
			email.PasswordResetTemplate,
			email.OTPCodeTemplate,
		})
	if err != nil {
		log.Fatalf("Error initializing email service: %v", err)
	}

	authService, err := auth.NewAuthService(coreService, emailService)
	if err != nil {
		log.Fatalf("Error initializing auth service: %v", err)
	}

	fileService, err := files.NewFileService(dbpool)
	if err != nil {
		log.Fatalf("Error initializing file service: %v", err)
	}

	// Initialize Valkey client for OTP service (Valkey is Redis-compatible)
	valkeyClient, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{redisAddress},
	})
	if err != nil {
		log.Fatalf("Error initializing Valkey client: %v", err)
	}
	defer valkeyClient.Close()

	// Initialize OTP service
	otpService := otp.NewOTPService(valkeyClient, emailService)

	// Initialize module services
	moderationSvc := moderation.NewModerationService(dbpool)
	recommendationSvc := recommendation.NewRecommendationService(dbpool)
	clubSvc := clubs.NewClubService(dbpool)
	chatSvc := chat.NewChatService(dbpool, moderationSvc)
	feedSvc := feed.NewFeedService(dbpool, chatSvc, moderationSvc, recommendationSvc, clubSvc)
	mapSvc := mapmod.NewMapService(dbpool, recommendationSvc)
	subscriptionSvc := subscriptions.NewSubscriptionService(dbpool)

	// Initialize controllers
	authController := auth.NewAuthController(authService)
	userController := controllers.NewUserController(coreService)
	filesController := files.NewFilesController(coreService, fileService)
	fileAdminController := files.NewFileAdminController(fileService)
	adminController := core.NewAdminController(coreService)

	recommendationCtrl := recommendation.NewRecommendationController(recommendationSvc)
	feedCtrl := feed.NewFeedController(feedSvc)
	chatCtrl := chat.NewChatController(chatSvc)
	mapCtrl := mapmod.NewMapController(mapSvc)
	moderationCtrl := moderation.NewModerationController(moderationSvc)
	subscriptionCtrl := subscriptions.NewSubscriptionController(subscriptionSvc)
	clubCtrl := clubs.NewClubController(clubSvc)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "X-Telegram-Web-App-Data", "Authorization"},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.HEAD("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	userAuth := auth.RequireAuth(authService)
	adminAuth := auth.RequireAdmin(authService)

	adminGroup := r.Group("/admin")
	adminController.RegisterRoutes(adminGroup, adminAuth)
	fileAdminController.RegisterRoutes(adminGroup, adminAuth)

	otpAuthController := auth.NewOTPAuthController(authService, otpService)

	authGroup := r.Group("/auth")
	authController.RegisterRoutes(authGroup, rlService.LoginRateLimiter)
	otpAuthController.RegisterRoutes(authGroup)

	userGroup := r.Group("/user")
	userController.RegisterRoutes(userGroup, userAuth, filesController, authController)

	// Profile & preferences: /me/...
	meGroup := r.Group("/me")
	recommendationCtrl.RegisterRoutes(meGroup, userAuth)

	// Feed & swipe: /matchup/...
	matchupGroup := r.Group("/matchup")
	feedCtrl.RegisterRoutes(matchupGroup, userAuth)

	// Chats: /chats/...
	chatsGroup := r.Group("/chats")
	chatCtrl.RegisterRoutes(chatsGroup, userAuth)

	// Map: /map/...
	mapGroup := r.Group("/map")
	mapCtrl.RegisterRoutes(mapGroup, userAuth)

	// Profile preview (authenticated, but views other users)
	profilesGroup := r.Group("/profiles")
	profilesGroup.Use(userAuth)
	profilesGroup.GET("/:userId/preview", recommendationCtrl.GetProfilePreview)

	// Moderation: /users/:userId/block, /users/:userId/report
	moderationCtrl.RegisterRoutes(r, userAuth)

	// Subscriptions: /subscriptions/...
	subscriptionsGroup := r.Group("/subscriptions")
	subscriptionCtrl.RegisterRoutes(subscriptionsGroup, adminAuth, userAuth)

	// Clubs: /clubs/..., /me/clubs, /admin/clubs/...
	clubCtrl.RegisterRoutes(r, meGroup, adminGroup, userAuth, adminAuth)

	// Public marketing materials routes (no authentication required)
	marketingGroup := r.Group("/media")
	marketingGroup.GET("", filesController.ListVisibleMaterials)
	marketingGroup.GET("/:id/download", filesController.GetMaterialDownloadURL)

	if err := r.Run(":8000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
