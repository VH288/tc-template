package bootstrap

import (
	"log/slog"

	"github.com/VH288/tc-template/internal/app/lifecycle"
	"github.com/VH288/tc-template/internal/config"
	"github.com/VH288/tc-template/internal/infrastructure/apm"

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
	APM    *apm.Provider
	Logger *slog.Logger
	Config *config.Config
}
