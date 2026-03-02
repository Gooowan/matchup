#!/bin/sh
set -e

echo "running psqldef"
psqldef --enable-drop --skip-view -h $POSTGRES_HOST -U $POSTGRES_USER -p $POSTGRES_PORT $POSTGRES_DB < /schema.sql
echo "recreating views"
psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -p "$POSTGRES_PORT" -d "$POSTGRES_DB" -f /views.sql
