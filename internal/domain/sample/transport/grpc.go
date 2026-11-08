package transport

import (
	"github.com/VH288/tc-template/internal/domain/sample/handler"
	samplev1 "github.com/VH288/tc-template/internal/proto/sample/v1"

	"google.golang.org/grpc"
)

func RegisterGRPC(server *grpc.Server, h *handler.GRPCHandler) {
	samplev1.RegisterSampleServiceServer(server, h)
}
