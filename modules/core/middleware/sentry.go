package middleware

import (
	"log/slog"
	"os"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// InitSentry initialises the Sentry SDK from the SENTRY_DSN environment variable.
// If SENTRY_DSN is empty, Sentry is disabled gracefully — no error is returned.
// PII scrubbing: Authorization and Cookie headers are removed from all events.
func InitSentry(logger *slog.Logger) error {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		logger.Info("SENTRY_DSN not set, Sentry error tracking disabled")
		return nil
	}

	env := os.Getenv("GIN_MODE")
	if env == "" {
		env = "development"
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      env,
		TracesSampleRate: 0.1,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			if event.Request != nil {
				event.Request.Cookies = ""
				delete(event.Request.Headers, "Authorization")
				delete(event.Request.Headers, "Cookie")
			}
			return event
		},
	})
	if err != nil {
		return err
	}

	logger.Info("Sentry error tracking initialised", "environment", env)
	return nil
}

// SentryMiddleware returns the sentrygin middleware configured to re-panic after
// capturing so that gin.Recovery() can still handle the panic response.
func SentryMiddleware() gin.HandlerFunc {
	return sentrygin.New(sentrygin.Options{Repanic: true})
}
