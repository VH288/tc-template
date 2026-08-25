package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"tc-template/internal/config"
	"tc-template/internal/infrastructure/database"
)

func wireInfra(ctx context.Context, cfg *config.Config) (*Infra, error) {
	// temp
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	db, err := database.NewPostgres(cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	return &Infra{
		DB:     db,
		Logger: logger,
		Config: cfg,
	}, nil
}
