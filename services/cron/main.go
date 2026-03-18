package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"

	"github.com/Gooowan/matchup/modules/core/db"
	"github.com/Gooowan/matchup/modules/subscriptions"
	"github.com/Gooowan/matchup/services/cron/controllers"
)

func main() {
	dbpool, err := db.PostgresConnect()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer dbpool.Close()

	subscriptionSvc := subscriptions.NewSubscriptionService(dbpool)
	cronController := controllers.NewCronController(subscriptionSvc)

	c := cron.New(cron.WithSeconds())

	// Expire active subscriptions past their expiry date (every 5 minutes)
	if _, err := c.AddFunc("0 */5 * * * *", cronController.ExpireSubscriptions); err != nil {
		log.Fatalf("[CRON] Error scheduling ExpireSubscriptions: %v", err)
	}

	// Check for subscriptions expiring within 1 day (once per hour)
	if _, err := c.AddFunc("0 0 * * * *", cronController.NotifyExpiringSoon); err != nil {
		log.Fatalf("[CRON] Error scheduling NotifyExpiringSoon: %v", err)
	}

	c.Start()
	log.Println("[CRON] service started")

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	r.HEAD("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	if err := r.Run(":8001"); err != nil {
		log.Fatalf("Failed to start cron health server: %v", err)
	}
}
