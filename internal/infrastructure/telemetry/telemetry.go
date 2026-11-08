package telemetry

import (
	"context"
	"io"
	"log/slog"

	sdklogger "github.com/VH288/tc-sdk/logger"
	"github.com/VH288/tc-template/internal/domain/ports"
)

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

var (
	_ ports.Logger = SlogLogger{}
	_ ports.Logger = NoopLogger{}
)

func DiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
