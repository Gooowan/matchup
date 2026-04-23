package main

import (
	"context"
	"embed"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/valkey-io/valkey-go"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/Gooowan/matchup/modules/chat"
	"github.com/Gooowan/matchup/modules/clubs"
	"github.com/Gooowan/matchup/modules/core/db"
	"github.com/Gooowan/matchup/modules/core/logging"
	coreMetrics "github.com/Gooowan/matchup/modules/core/metrics"
	coreMiddleware "github.com/Gooowan/matchup/modules/core/middleware"
	"github.com/Gooowan/matchup/modules/core/tracing"
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
	logger := logging.Init()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown, err := tracing.Init(ctx, "matchup-api")
	if err != nil {
		logger.Error("failed to initialise tracing", "error", err)
		os.Exit(1)
	}
	defer func() { _ = shutdown(ctx) }()

	if err := coreMiddleware.InitSentry(logger); err != nil {
		logger.Error("failed to initialise Sentry", "error", err)
		// Non-fatal: continue without Sentry
	}

	dbpool, err := db.PostgresConnect()
	if err != nil {
		logger.Error("error connecting to database", "error", err)
		os.Exit(1)
	}
	defer dbpool.Close()
	go coreMetrics.StartDBPoolCollector(ctx, dbpool)

	redisAddress := os.Getenv("REDIS_ADDRESS")
	if redisAddress == "" {
		logger.Error("REDIS_ADDRESS isn't set")
		os.Exit(1)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})
	defer redisClient.Close()

	allowOrigins := os.Getenv("ALLOW_ORIGINS")
	if allowOrigins == "" {
		logger.Error("ALLOW_ORIGINS isn't set")
		os.Exit(1)
	}
	allowedOrigins := strings.Split(allowOrigins, ",")

	coreService := core.NewCoreService(dbpool)
	rlService := ratelimit.NewRLService(redisClient)

	var emailProvider email.EmailProvider
	switch os.Getenv("EMAIL_PROVIDER") {
	case "resend":
		emailProvider, err = providers.NewResendProvider()
		if err != nil {
			logger.Error("error initializing Resend provider", "error", err)
			os.Exit(1)
		}
	case "mailgun":
		emailProvider, err = providers.NewMailgunProvider()
		if err != nil {
			logger.Error("error initializing Mailgun provider", "error", err)
			os.Exit(1)
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
		logger.Error("error initializing email service", "error", err)
		os.Exit(1)
	}

	authService, err := auth.NewAuthService(coreService, emailService)
	if err != nil {
		logger.Error("error initializing auth service", "error", err)
		os.Exit(1)
	}

	fileService, err := files.NewFileService(dbpool)
	if err != nil {
		logger.Error("error initializing file service", "error", err)
		os.Exit(1)
	}

	valkeyClient, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{redisAddress},
	})
	if err != nil {
		logger.Error("error initializing Valkey client", "error", err)
		os.Exit(1)
	}
	defer valkeyClient.Close()

	otpService := otp.NewOTPService(valkeyClient, emailService)

	moderationSvc := moderation.NewModerationService(dbpool)
	recommendationSvc := recommendation.NewRecommendationService(dbpool)
	clubSvc := clubs.NewClubService(dbpool)
	chatSvc := chat.NewChatService(dbpool, moderationSvc)
	feedSvc := feed.NewFeedService(dbpool, chatSvc, moderationSvc, recommendationSvc, clubSvc)
	mapSvc := mapmod.NewMapService(dbpool, recommendationSvc)
	subscriptionSvc := subscriptions.NewSubscriptionService(dbpool)

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

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(coreMiddleware.SentryMiddleware())
	r.Use(coreMiddleware.RequestID())
	r.Use(otelgin.Middleware("matchup-api"))
	r.Use(coreMiddleware.PrometheusMetrics())
	r.Use(coreMiddleware.RequestLogger(logger))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "X-Telegram-Web-App-Data", "Authorization", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.HEAD("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

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

	meGroup := r.Group("/me")
	recommendationCtrl.RegisterRoutes(meGroup, userAuth)

	matchupGroup := r.Group("/matchup")
	feedCtrl.RegisterRoutes(matchupGroup, userAuth)

	chatsGroup := r.Group("/chats")
	chatCtrl.RegisterRoutes(chatsGroup, userAuth)

	mapGroup := r.Group("/map")
	mapCtrl.RegisterRoutes(mapGroup, userAuth)

	profilesGroup := r.Group("/profiles")
	profilesGroup.Use(userAuth)
	profilesGroup.GET("/:userId/preview", recommendationCtrl.GetProfilePreview)

	moderationCtrl.RegisterRoutes(r, userAuth)

	subscriptionsGroup := r.Group("/subscriptions")
	subscriptionCtrl.RegisterRoutes(subscriptionsGroup, adminAuth, userAuth)

	clubCtrl.RegisterRoutes(r, meGroup, adminGroup, userAuth, adminAuth)

	marketingGroup := r.Group("/media")
	marketingGroup.GET("", filesController.ListVisibleMaterials)
	marketingGroup.GET("/:id/download", filesController.GetMaterialDownloadURL)

	logger.Info("starting API server", "port", 8000)
	if err := r.Run(":8000"); err != nil {
		logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
