package middleware

import (
	"log/slog"
	"time"

	"github.com/Gooowan/matchup/modules/core/logging"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// RequestLogger is a Gin middleware that logs each HTTP request as a structured
// JSON slog entry after the handler completes.
// It builds a child logger enriched with request_id, method, path, client_ip,
// and — if otelgin ran first — trace_id and span_id for log-trace correlation.
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID, _ := c.Get(RequestIDKey)

		reqLogger := logger.With(
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.FullPath(),
			"raw_path", c.Request.URL.Path,
			"client_ip", c.ClientIP(),
		)

		// Store logger in request context so service/controller layers can pick it up.
		ctx := logging.WithLogger(c.Request.Context(), reqLogger)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// After handler — enrich log with response details.
		attrs := []any{
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"bytes", c.Writer.Size(),
		}

		// Inject trace correlation if OTel span is present (requires otelgin before this middleware).
		if span := trace.SpanFromContext(c.Request.Context()); span.SpanContext().IsValid() {
			attrs = append(attrs,
				"trace_id", span.SpanContext().TraceID().String(),
				"span_id", span.SpanContext().SpanID().String(),
			)
		}

		level := slog.LevelInfo
		status := c.Writer.Status()
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		reqLogger.Log(c.Request.Context(), level, "request", attrs...)
	}
}
