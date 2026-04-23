package logging

import (
	"context"
	"log/slog"
	"os"
)

type contextKey string

const loggerKey contextKey = "logger"

// Init initialises the global slog logger with a JSON handler.
// Level is DEBUG in dev (GIN_MODE != "release") and INFO in production.
// Call once at process startup before anything else.
func Init() *slog.Logger {
	level := slog.LevelInfo
	if os.Getenv("GIN_MODE") != "release" {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

// WithLogger stores a child logger in the context.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves the child logger from context,
// falling back to slog.Default() if none is set.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
