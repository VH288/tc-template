package bootstrap

import (
	"context"

	"github.com/VH288/tc-template/internal/app/lifecycle"
)

// wireLifecycle registers graceful shutdown hooks (LIFO).
func wireLifecycle(infra *Infra) *lifecycle.Manager {
	lc := lifecycle.New(infra.Logger, infra.Config.HTTP.ShutdownTimeout)

	if infra.APM != nil {
		lc.Add(infra.APM.Shutdown)
	}

	lc.Add(func(ctx context.Context) error { return infra.DB.Close() })
	return lc
}
