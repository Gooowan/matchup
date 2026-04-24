package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func PostgresConnect() (*pgxpool.Pool, error) {
	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		return nil, fmt.Errorf("POSTGRES_USER environment variable not set")
	}
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("POSTGRES_PASSWORD environment variable not set")
	}
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		return nil, fmt.Errorf("POSTGRES_HOST environment variable not set")
	}
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		return nil, fmt.Errorf("POSTGRES_PORT environment variable not set")
	}
	dbname := os.Getenv("POSTGRES_DB")
	if dbname == "" {
		return nil, fmt.Errorf("POSTGRES_DB environment variable not set")
	}

	sslMode := os.Getenv("DB_SSL_MODE")
	if sslMode == "" {
		sslMode = "require"
	}
	databaseURL := "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbname + "?sslmode=" + sslMode

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("PGX failed parsing database URL: %w", err)
	}

	config.MaxConns = 24
	config.MinConns = 4
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30

	return pgxpool.NewWithConfig(context.Background(), config)
}
