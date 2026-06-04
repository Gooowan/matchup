package main

import (
	"context"
	cryptotls "crypto/tls"
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
	"github.com/Gooowan/matchup/modules/core/geocoding"
	"github.com/Gooowan/matchup/modules/core/gmaps"
	"github.com/Gooowan/matchup/modules/push"
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

// metricsAuthMiddleware protects the /metrics endpoint with a bearer token
// supplied via the METRICS_TOKEN environment variable.  If the variable is
// empty the endpoint is disabled (returns 404) so it is never accidentally
// exposed without authentication.
func metricsAuthMiddleware() gin.HandlerFunc {
	token := os.Getenv("METRICS_TOKEN")
	return func(c *gin.Context) {
		if token == "" {
			c.Status(http.StatusNotFound)
			c.Abort()
			return
		}
		auth := c.GetHeader("Authorization")
		if auth != "Bearer "+token {
			c.Header("WWW-Authenticate", `Bearer realm="metrics"`)
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}

// productionGuard refuses to start when dangerous placeholder values are
// detected while running in release mode, preventing accidental use of
// .env.example defaults in production.
func productionGuard(logger interface{ Error(string, ...any) }) {
	if os.Getenv("GIN_MODE") != "release" {
		return
	}
	dangerous := map[string]string{
		"JWT_SECRET":        "CHANGE_ME_use_openssl_rand_hex_32_output_here",
		"POSTGRES_PASSWORD": "CHANGE_ME_strong_db_password",
		"MINIO_ACCESS_KEY":  "CHANGE_ME_minio_access_key",
		"MINIO_SECRET_KEY":  "CHANGE_ME_minio_secret_key",
	}
	for key, placeholder := range dangerous {
		if val := os.Getenv(key); val == placeholder || val == "" {
			logger.Error("refusing to start: unsafe placeholder or empty value in production",
				"env_var", key)
			os.Exit(1)
		}
	}
	// Warn (but do not block) when SSL is explicitly disabled in release mode.
	// Set ALLOW_DB_SSL_DISABLE=true to suppress this warning when a managed
	// proxy (e.g. Cloud SQL Auth Proxy, RDS Proxy) terminates TLS externally.
	if os.Getenv("DB_SSL_MODE") == "disable" && os.Getenv("ALLOW_DB_SSL_DISABLE") != "true" {
		logger.Error("security warning: DB_SSL_MODE=disable in production — database traffic is unencrypted. " +
			"Set DB_SSL_MODE=require or ALLOW_DB_SSL_DISABLE=true to acknowledge.",
			"env_var", "DB_SSL_MODE")
		os.Exit(1)
	}
}

func main() {
	logger := logging.Init()

	// Set Gin mode before anything else so internal framework logging respects it.
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	productionGuard(logger)

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

	redisOpts := &redis.Options{
		Addr: redisAddress,
	}
	if os.Getenv("REDIS_TLS_ENABLED") == "true" {
		redisOpts.TLSConfig = &cryptotls.Config{
			MinVersion: cryptotls.VersionTLS12,
		}
	}
	redisClient := redis.NewClient(redisOpts)
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
	authService.SetOTPService(otpService)

	moderationSvc := moderation.NewModerationService(dbpool)
	recommendationSvc := recommendation.NewRecommendationService(dbpool)
	geocoder := geocoding.NewNominatimGeocoder(
		os.Getenv("NOMINATIM_URL"),
		func() string {
			if ua := os.Getenv("MATCHUP_USER_AGENT"); ua != "" {
				return ua
			}
			return "matchup-server/1.0 (admin@matchup.local)"
		}(),
	)
	defer geocoder.Close()

	clubSvc := clubs.NewClubService(dbpool)
	chatSvc := chat.NewChatService(dbpool, moderationSvc)
	feedSvc := feed.NewFeedService(dbpool, chatSvc, moderationSvc, recommendationSvc, clubSvc)
	pushSvc := push.NewService(dbpool, logger)
	feedSvc.PushSvc = pushSvc
	mapSvc := mapmod.NewMapService(dbpool, recommendationSvc)
	subscriptionSvc := subscriptions.NewSubscriptionService(dbpool)

	authController := auth.NewAuthController(authService)
	authController.SetLockoutRecorder(rlService)
	userController := controllers.NewUserController(coreService)
	userController.SetOTPService(otpService)
	userController.SetAuthService(authService)
	filesController := files.NewFilesController(coreService, fileService)
	fileAdminController := files.NewFileAdminController(fileService)
	adminController := core.NewAdminController(coreService)

	recommendationCtrl := recommendation.NewRecommendationController(recommendationSvc)
	pushCtrl := push.NewController(pushSvc)
	feedCtrl := feed.NewFeedController(feedSvc)
	chatCtrl := chat.NewChatController(chatSvc)
	mapCtrl := mapmod.NewMapController(mapSvc)
	moderationCtrl := moderation.NewModerationController(moderationSvc, coreService)
	subscriptionCtrl := subscriptions.NewSubscriptionController(subscriptionSvc)
	placesClient := gmaps.NewGooglePlacesClient(os.Getenv("GOOGLE_PLACES_API_KEY"))
	clubCtrl := clubs.NewClubController(clubSvc, chatSvc, geocoder, placesClient)

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

	// Protect /metrics with a bearer token.  If METRICS_TOKEN is not set in
	// production the productionGuard check above will already refuse to start;
	// here we fall back to refusing all requests so the endpoint is never open.
	r.GET("/metrics", metricsAuthMiddleware(), gin.WrapH(promhttp.Handler()))

	userAuth := auth.RequireAuth(authService)
	adminAuth := auth.RequireAdmin(authService)

	adminGroup := r.Group("/admin")
	adminController.RegisterRoutes(adminGroup, adminAuth)
	fileAdminController.RegisterRoutes(adminGroup, adminAuth)
	// Chat moderation admin routes (registered after chatCtrl is created below)
	// Stored as a closure to avoid forward reference issues.
	var registerAdminChatRoutes func()

	otpAuthController := auth.NewOTPAuthController(authService, otpService)
	googleAuthController := auth.NewGoogleAuthController(authService)

	authGroup := r.Group("/auth")
	// loginChain: lockout check first, then sliding-window rate limiter.
	authController.RegisterRoutesWithLockout(authGroup,
		rlService.LoginLockoutMiddleware(),
		rlService.LoginRateLimiter,
		rlService.RegisterRateLimiter)
	otpAuthController.RegisterRoutes(authGroup)
	googleAuthController.RegisterRoutes(authGroup)

	userGroup := r.Group("/user")
	userController.RegisterRoutes(userGroup, userAuth, filesController, authController, rlService.UploadRateLimiter)

	meGroup := r.Group("/me")
	recommendationCtrl.RegisterRoutes(meGroup, userAuth)
	pushCtrl.RegisterRoutes(meGroup, userAuth)

	matchupGroup := r.Group("/matchup")
	feedCtrl.RegisterRoutes(matchupGroup, userAuth, rlService.SwipeRateLimiter)

	chatsGroup := r.Group("/chats")
	chatCtrl.RegisterRoutes(chatsGroup, userAuth, rlService.MessageRateLimiter)

	registerAdminChatRoutes = func() {
		adminChatGroup := adminGroup.Group("/chats")
		adminChatGroup.Use(adminAuth)
		chatCtrl.RegisterAdminRoutes(adminChatGroup)
	}
	registerAdminChatRoutes()

	mapGroup := r.Group("/map")
	mapCtrl.RegisterRoutes(mapGroup, userAuth)

	profilesGroup := r.Group("/profiles")
	profilesGroup.Use(userAuth)
	profilesGroup.GET("/:userId/preview", recommendationCtrl.GetProfilePreview)

	moderationCtrl.RegisterRoutes(r, userAuth)
	adminGroup.GET("/reports", moderationCtrl.AdminListReports)
	adminGroup.POST("/users/:userId/ban", moderationCtrl.AdminBanUser)

	subscriptionsGroup := r.Group("/subscriptions")
	subscriptionCtrl.RegisterRoutes(subscriptionsGroup, adminAuth, userAuth)
	subscriptionCtrl.RegisterWebhook(r)

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
