package apm_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/VH288/tc-template/internal/infrastructure/apm"
)

func TestProvider_NilShutdown(t *testing.T) {
	var p *apm.Provider
	assert.NoError(t, p.Shutdown(context.Background()))
}

func TestTracerAndStart(t *testing.T) {
	tr := apm.Tracer("test")
	assert.NotNil(t, tr)

	ctx, span := apm.Start(context.Background(), "test.span")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}
