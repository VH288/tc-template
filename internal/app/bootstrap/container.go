package bootstrap

import (
	"log/slog"

	"tc-template/internal/app/lifecycle"
	"tc-template/internal/config"

	"github.com/jmoiron/sqlx"
)

type Container struct {
	Config *config.Config
	Logger *slog.Logger

	Infra     *Infra
	Lifecycle *lifecycle.Manager
	Domain
}

type Infra struct {
	DB     *sqlx.DB
	Logger *slog.Logger
	Config *config.Config
}
