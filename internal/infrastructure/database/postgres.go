package database

import (
	"context"
	"fmt"
	"time"

	"github.com/VH288/tc-template/internal/config"
	"github.com/jmoiron/sqlx"
)

func NewPostgres(cfg config.PostgresConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

func Ping(ctx context.Context, db *sqlx.DB) error {
	return db.PingContext(ctx)
}
