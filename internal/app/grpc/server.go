package grpc

import (
	samplehandler "github.com/VH288/tc-template/internal/domain/sample/handler"
	sampletransport "github.com/VH288/tc-template/internal/domain/sample/transport"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Dependencies struct {
	Sample *samplehandler.GRPCHandler
}

func NewServer(deps Dependencies) *grpc.Server {
	server := grpc.NewServer()

	if deps.Sample != nil {
		sampletransport.RegisterGRPC(server, deps.Sample)
	}

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("sample.v1.SampleService", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(server)

	return server
}
