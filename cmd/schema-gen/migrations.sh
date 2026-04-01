#!/bin/sh
set -e

PSQL="psql -h $POSTGRES_HOST -U $POSTGRES_USER -p $POSTGRES_PORT -d $POSTGRES_DB"

# Ensure the migrations tracking table exists (idempotent)
$PSQL -c "
  CREATE TABLE IF NOT EXISTS schema_migrations (
    filename  TEXT PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
  );
"

# Detect fresh vs existing database (exclude schema_migrations itself)
TABLE_COUNT=$($PSQL -tAc "
  SELECT count(*) FROM information_schema.tables
  WHERE table_schema = 'public'
    AND table_type = 'BASE TABLE'
    AND table_name != 'schema_migrations';
")

if [ "$TABLE_COUNT" = "0" ]; then
  echo "fresh database — applying combined schema"
  $PSQL -f /schema.sql

  # Mark all migration files as already applied (schema covers them)
  for f in /migrations/*.sql; do
    name=$(basename "$f")
    $PSQL -c "INSERT INTO schema_migrations (filename) VALUES ('$name') ON CONFLICT DO NOTHING;"
  done
else
  echo "existing database — applying pending migrations"
  for f in $(ls /migrations/*.sql | sort); do
    name=$(basename "$f")
    already=$($PSQL -tAc "SELECT count(*) FROM schema_migrations WHERE filename = '$name';")
    if [ "$already" = "0" ]; then
      echo "  applying $name"
      $PSQL -v ON_ERROR_STOP=1 -f "$f"
      $PSQL -c "INSERT INTO schema_migrations (filename) VALUES ('$name');"
    else
      echo "  skipping $name (already applied)"
    fi
  done
fi

if [ -f /views.sql ]; then
  echo "recreating views"
  $PSQL -f /views.sql
fi

echo "migrations complete"
