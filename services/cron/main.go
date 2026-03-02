package main

//

// import (
// 	"log"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// 	"github.com/robfig/cron/v3"

// 	core "github.com/Gooowan/matchup/modules/users"
// 	"github.com/Gooowan/matchup/modules/core/db"
// 	"github.com/Gooowan/matchup/services/cron/controllers"
// )

// func main() {
// 	dbpool, err := db.PostgresConnect()
// 	if err != nil {
// 		log.Fatalf("Error connecting to database: %v", err)
// 	}
// 	defer dbpool.Close()

// 	coreService := core.NewCoreService(dbpool)

// 	cronController := controllers.NewCronController(coreService, paymentsService, desimService)

// 	c := cron.New(cron.WithSeconds())

// 	_, err = c.AddFunc("@every 30s", cronController.ExpirePendingInvoices)
// 	if err != nil {
// 		log.Fatalf("[CRON] Error scheduling ExpirePendingInvoices job: %v", err)
// 	}

// 	if _, err = c.AddFunc("@every 40s", cronController.RefreshReferralStuff); err != nil {
// 		log.Fatalf("[CRON] Error scheduling RefreshUserPersonalTurnover job: %v", err)
// 	}

// 	if _, err = c.AddFunc("0 0 0 * * *", cronController.DepositDailyAccruedBalance); err != nil {
// 		log.Fatalf("[CRON] Error scheduling DepositDailyAccruedBalance job: %v", err)
// 	}

// 	c.Start()
// 	log.Println("[CRON] service started")

// 	r := gin.Default()
// 	r.GET("/health", func(c *gin.Context) {
// 		c.String(http.StatusOK, "OK")
// 	})
// 	r.HEAD("/health", func(c *gin.Context) {
// 		c.String(http.StatusOK, "OK")
// 	})

// 	cronController.DepositDailyAccruedBalance()
// 	cronController.RefreshReferralStuff()

// 	if err := r.Run(":8000"); err != nil {
// 		log.Fatalf("Failed to start server: %v", err)
// 	}
// }
