package middleware

import (
	"strconv"
	"time"

	"github.com/Gooowan/matchup/modules/core/metrics"
	"github.com/gin-gonic/gin"
)

// PrometheusMetrics is a Gin middleware that records HTTP request count and
// latency into Prometheus metrics.
//
// IMPORTANT: uses c.FullPath() (route template, e.g. /user/:id) rather than
// c.Request.URL.Path (which contains real UUIDs) to avoid label cardinality explosion.
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unknown" // 404 routes have no template
		}

		status := strconv.Itoa(c.Writer.Status())
		elapsed := time.Since(start).Seconds()

		metrics.HTTPRequestsTotal.WithLabelValues(c.Request.Method, route, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(c.Request.Method, route, status).Observe(elapsed)
	}
}
