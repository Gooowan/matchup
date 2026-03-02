#!/bin/sh
set -e

echo "running psqldef"

# Check if database has any tables (fresh database detection)
TABLE_COUNT=$(psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -p "$POSTGRES_PORT" -d "$POSTGRES_DB" -tAc \
  "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE';")

if [ "$TABLE_COUNT" = "0" ]; then
  echo "fresh database detected, applying schema directly"
  psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -p "$POSTGRES_PORT" -d "$POSTGRES_DB" -f /schema.sql
else
  echo "existing database, running psqldef diff"
  psqldef --enable-drop --skip-view -h $POSTGRES_HOST -U $POSTGRES_USER -p $POSTGRES_PORT $POSTGRES_DB < /schema.sql
fi

if [ -f /views.sql ]; then
  echo "recreating views"
  psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -p "$POSTGRES_PORT" -d "$POSTGRES_DB" -f /views.sql
else
  echo "no views to apply"
fi
