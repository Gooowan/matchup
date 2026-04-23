package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"

	"github.com/Gooowan/matchup/modules/core/db"
	"github.com/Gooowan/matchup/modules/core/logging"
	"github.com/Gooowan/matchup/modules/subscriptions"
	"github.com/Gooowan/matchup/services/cron/controllers"
)

func main() {
	logger := logging.Init()

	dbpool, err := db.PostgresConnect()
	if err != nil {
		logger.Error("error connecting to database", "error", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	subscriptionSvc := subscriptions.NewSubscriptionService(dbpool)
	cronController := controllers.NewCronController(subscriptionSvc, logger)

	c := cron.New(cron.WithSeconds())

	if _, err := c.AddFunc("0 */5 * * * *", cronController.ExpireSubscriptions); err != nil {
		logger.Error("error scheduling ExpireSubscriptions", "error", err)
		os.Exit(1)
	}

	if _, err := c.AddFunc("0 0 * * * *", cronController.NotifyExpiringSoon); err != nil {
		logger.Error("error scheduling NotifyExpiringSoon", "error", err)
		os.Exit(1)
	}

	c.Start()
	logger.Info("cron service started")

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	r.HEAD("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	logger.Info("cron health/metrics server starting", "port", 8001)
	if err := r.Run(":8001"); err != nil {
		slog.Error("failed to start cron health server", "error", err)
		os.Exit(1)
	}
}
