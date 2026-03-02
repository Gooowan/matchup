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

	"github.com/Gooowan/matchup/modules/core"
	"github.com/Gooowan/matchup/modules/core/auth"
	"github.com/Gooowan/matchup/modules/core/db"
	"github.com/Gooowan/matchup/modules/email"
	"github.com/Gooowan/matchup/modules/email/providers"
	"github.com/Gooowan/matchup/modules/files"
	"github.com/Gooowan/matchup/services/api/controllers"

	matchupmod "github.com/Gooowan/matchup/modules/matchup"

	// "github.com/Gooowan/matchup/modules/otp"
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
	// otpService := otp.NewOTPService(valkeyClient, emailService)

	matchupModule := matchupmod.NewMatchupModule(dbpool)

	authController := auth.NewAuthController(authService)
	userController := controllers.NewUserController(coreService)
	filesController := files.NewFilesController(coreService, fileService)
	fileAdminController := files.NewFileAdminController(fileService)
	adminController := core.NewAdminController(coreService)

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

	authGroup := r.Group("/auth")
	authController.RegisterRoutes(authGroup, rlService.LoginRateLimiter)

	userGroup := r.Group("/user")
	userController.RegisterRoutes(userGroup, userAuth, filesController, authController)

	matchupModule.RegisterRoutes(r, userAuth)

	// Public marketing materials routes (no authentication required)
	marketingGroup := r.Group("/marketing")
	marketingGroup.GET("", filesController.ListVisibleMaterials)
	marketingGroup.GET("/:id/download", filesController.GetMaterialDownloadURL)

	if err := r.Run(":8000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
