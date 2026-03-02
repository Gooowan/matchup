package controllers

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Gooowan/matchup/modules/core"
	"github.com/Gooowan/matchup/modules/core/utils"
)

type CronController struct {
	core *core.CoreService
}

func NewCronController(coreService *core.CoreService) *CronController {
	return &CronController{
		core: coreService,
	}
}

func (c *CronController) DepositDailyAccruedBalance() {

	log.Printf("[CRON] Starting daily deposit")

	start := time.Now()
	ctx := context.Background()
	const batchSize = 100

	totalCount, err := c.desim.Queries.CountUserActiveProducts(ctx)
	if err != nil {
		log.Printf("[CRON] Failed to count active user products: %v", err)
		return
	}

	log.Printf("[CRON] Found %d active user products to process", totalCount)

	if totalCount == 0 {
		log.Println("[CRON] No active user products to process")
		return
	}

	var totalProcessed, totalFailed int
	batches := (totalCount + batchSize - 1) / batchSize

	for batch := int64(0); batch < batches; batch++ {
		offset := batch * batchSize
		log.Printf("[CRON] Processing batch %d/%d (offset: %d, limit: %d)", batch+1, batches, offset, batchSize)

		processed, failed, err := c.processDailyDepositBatch(ctx, int32(batchSize), int32(offset))
		if err != nil {
			log.Printf("[CRON] Failed to process batch %d: %v", batch+1, err)
			totalFailed += batchSize
			continue
		}

		totalProcessed += processed
		totalFailed += failed

		log.Printf("[CRON] Batch %d completed. Processed: %d, Failed: %d",
			batch+1, processed, failed)
	}

	log.Printf("[CRON] Daily deposit completed. Total - Processed: %d, Failed: %d",
		totalProcessed, totalFailed)
	log.Printf("[CRON] Daily deposit completed in %s", time.Since(start))

}

func (c *CronController) processDailyDepositBatch(ctx context.Context, limit int32, offset int32) (processed int, failed int, err error) {

	tx, err := c.desim.DB.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to start transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	qtx := c.desim.Queries.WithTx(tx)

	activeUserProducts, err := qtx.GetUserActiveProductsPaginated(ctx, gen.GetUserActiveProductsPaginatedParams{
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		return 0, 0, fmt.Errorf("[CRON] failed to get active user products: %w", err)
	}

	if len(activeUserProducts) == 0 {
		return 0, 0, nil
	}

	for _, product := range activeUserProducts {
		maxTokens := product.Amount * 2 / desim.TokenFloorPrice
		dailyAccrual := (product.Amount / desim.TokenFloorPrice) / 30
		accuredBalance := math.Min(product.AccruedBalance+dailyAccrual, maxTokens)

		if err = qtx.AccureUserDesimProduct(ctx, gen.AccureUserDesimProductParams{
			ID:             product.ID,
			IsActive:       accuredBalance < maxTokens,
			AccruedBalance: accuredBalance,
		}); err != nil {
			log.Printf("[CRON] Failed to process user product %d: %v", product.ID, err)
			failed++
			continue
		}

		processed++
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return processed, failed, nil
}

func (c *CronController) RefreshReferralStuff() {
	ctx := context.Background()
	start := time.Now()

	utils.DebugPrint("refreshing personal turnover")
	if err := c.desim.Queries.RefreshPersonalTurnover(ctx); err != nil {
		log.Printf("[CRON] Failed to refresh user personal turnover: %v", err)
	}
	utils.DebugPrint("refreshed personal turnover in %s", time.Since(start))

	startStats := time.Now()
	utils.DebugPrint("refreshing referral stats")
	if err := c.desim.Queries.RefreshReferralStats(ctx); err != nil {
		log.Printf("[CRON] Failed to refresh user referral stats: %v", err)
	}
	utils.DebugPrint("refreshed referral stats in %s", time.Since(startStats))

	// Worker and batch size per user count
	// - 1,000-10,000:     4 workers, 256 batch size (current)
	// - 10,000-100,000:   6-8 workers, 512 batch size
	// - 100,000+:         8-12 workers, 512-1024 batch size
	// Keep workers < 50% of DB MaxConns (24) to not kill other stuff while updating
	startRanks := time.Now()
	utils.DebugPrint("recalculating user ranks, workers: %d, batch size: %d", 4, 256)
	if err := c.desim.RecalculateAllRanksParallel(ctx, 4, 256); err != nil {
		log.Printf("[CRON] Failed to recalculate user ranks: %v", err)
	}
	utils.DebugPrint("recalculated user ranks in %s", time.Since(startRanks))
	utils.DebugPrint("cron completed in %s", time.Since(start))
}
