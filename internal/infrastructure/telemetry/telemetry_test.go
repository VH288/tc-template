package telemetry_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/VH288/tc-template/internal/infrastructure/telemetry"
)

func TestTelemetryAdapters(t *testing.T) {
	ctx := context.Background()

	// Tracer
	tr := telemetry.NewTracer()
	ctx2, end := tr.Start(ctx, "test-op")
	assert.NotNil(t, ctx2)
	assert.NotNil(t, end)
	end()

	// NoopTracer
	noopTr := telemetry.NoopTracer{}
	ctx3, noopEnd := noopTr.Start(ctx, "test-op")
	assert.Equal(t, ctx, ctx3)
	noopEnd()

	// SampleMetrics
	sm := telemetry.NewSampleMetrics()
	sm.IncOperation("create", "success")

	noopSM := telemetry.NoopSampleMetrics{}
	noopSM.IncOperation("create", "success")

	// GRPCMetrics
	gm := telemetry.NewGRPCMetrics()
	gm.IncRequest("method", "ok")

	noopGM := telemetry.NoopGRPCMetrics{}
	noopGM.IncRequest("method", "ok")

	// SlogLogger & NoopLogger
	logger := telemetry.NewSlogLogger()
	assert.NotNil(t, logger)

	noopLogger := telemetry.NoopLogger{}
	noopLogger.Info(ctx, "test")
	noopLogger.Warn(ctx, "test")
	noopLogger.Error(ctx, "test")
}
