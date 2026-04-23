package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var dbTracer = otel.Tracer("matchup/db")

// StartDBSpan creates a span for a database operation.
// Caller must call span.End() when done (typically via defer).
//
//	ctx, span := tracing.StartDBSpan(ctx, "GetFeed", "profiles")
//	defer span.End()
func StartDBSpan(ctx context.Context, operation, table string) (context.Context, trace.Span) {
	return dbTracer.Start(ctx, "db."+operation,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.sql.table", table),
		),
	)
}
