package ports

import "context"

type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, func())
}

type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

type SampleMetrics interface {
	IncOperation(operation, status string)
}

type GRPCMetrics interface {
	IncRequest(method, result string)
}
