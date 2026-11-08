package telemetry

import (
	"context"
	"io"
	"log/slog"

	sdklogger "github.com/VH288/tc-sdk/logger"
	"github.com/VH288/tc-template/internal/domain/ports"
	"github.com/VH288/tc-template/internal/infrastructure/apm"
	"github.com/VH288/tc-template/internal/infrastructure/metrics"
)

type Tracer struct{}

func NewTracer() *Tracer { return &Tracer{} }

func (Tracer) Start(ctx context.Context, name string) (context.Context, func()) {
	ctx, span := apm.Start(ctx, name)
	return ctx, func() {
		span.End()
	}
}

type SlogLogger struct{}

func NewSlogLogger() *SlogLogger {
	return &SlogLogger{}
}

func (SlogLogger) Info(ctx context.Context, msg string, args ...any) {
	sdklogger.FromContext(ctx).Info(msg, args...)
}

func (SlogLogger) Warn(ctx context.Context, msg string, args ...any) {
	sdklogger.FromContext(ctx).Warn(msg, args...)
}

func (SlogLogger) Error(ctx context.Context, msg string, args ...any) {
	sdklogger.FromContext(ctx).Error(msg, args...)
}

type NoopLogger struct{}

func (NoopLogger) Info(context.Context, string, ...any)  {}
func (NoopLogger) Warn(context.Context, string, ...any)  {}
func (NoopLogger) Error(context.Context, string, ...any) {}

type GRPCMetrics struct{}

func NewGRPCMetrics() *GRPCMetrics { return &GRPCMetrics{} }
func (GRPCMetrics) IncRequest(method, result string) {
	metrics.GRPCRequestsTotal.WithLabelValues(method, result).Inc()
}

type SampleMetrics struct{}

func NewSampleMetrics() *SampleMetrics { return &SampleMetrics{} }

func (SampleMetrics) IncOperation(operation, status string) {
	metrics.SampleOperationsTotal.WithLabelValues(operation, status).Inc()
}

type NoopTracer struct{}

func (NoopTracer) Start(ctx context.Context, _ string) (context.Context, func()) {
	return ctx, func() {}
}

type NoopSampleMetrics struct{}

func (NoopSampleMetrics) IncOperation(string, string) {}
type NoopGRPCMetrics struct{}

func (NoopGRPCMetrics) IncRequest(string, string) {}

var (
	_ ports.Tracer        = Tracer{}
	_ ports.Tracer        = NoopTracer{}
	_ ports.Logger        = SlogLogger{}
	_ ports.Logger        = NoopLogger{}
	_ ports.SampleMetrics = SampleMetrics{}
	_ ports.SampleMetrics = NoopSampleMetrics{}
	_ ports.GRPCMetrics   = GRPCMetrics{}
	_ ports.GRPCMetrics   = NoopGRPCMetrics{}
)

func DiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
