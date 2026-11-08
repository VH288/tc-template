package bootstrap

import (
	samplehandler "github.com/VH288/tc-template/internal/domain/sample/handler"
	sampleusecase "github.com/VH288/tc-template/internal/domain/sample/usecase"
	samplepersistence "github.com/VH288/tc-template/internal/infrastructure/persistence/sample"
	"github.com/VH288/tc-template/internal/infrastructure/telemetry"
)

type Sample struct {
	HTTP *samplehandler.HTTPHandler
	GRPC *samplehandler.GRPCHandler
}

func wireSample(infra *Infra) *Sample {
	repo := samplepersistence.NewPostgresRepository(infra.DB)
	log := telemetry.NewSlogLogger()
	trace := telemetry.NewTracer()
	sampleMetrics := telemetry.NewSampleMetrics()
	uc := sampleusecase.New(repo, log, trace, sampleMetrics)
	return &Sample{
		HTTP: samplehandler.NewHTTPHandler(uc),
		GRPC: samplehandler.NewGRPCHandler(uc, telemetry.NewGRPCMetrics()),
	}
}
