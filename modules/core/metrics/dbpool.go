package metrics

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StartDBPoolCollector starts a goroutine that scrapes pgxpool stats every 15 seconds
// and updates the corresponding Prometheus gauges.
// The goroutine exits when ctx is cancelled.
func StartDBPoolCollector(ctx context.Context, pool *pgxpool.Pool) {
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stat := pool.Stat()
				DBPoolAcquiredConns.Set(float64(stat.AcquiredConns()))
				DBPoolIdleConns.Set(float64(stat.IdleConns()))
				DBPoolTotalConns.Set(float64(stat.TotalConns()))
			}
		}
	}()
}
