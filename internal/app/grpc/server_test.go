package grpc_test

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	grpcapp "github.com/VH288/tc-template/internal/app/grpc"
	"github.com/VH288/tc-template/internal/domain/sample/entity"
	"github.com/VH288/tc-template/internal/domain/sample/handler"
	"github.com/VH288/tc-template/internal/infrastructure/telemetry"
	apperrors "github.com/VH288/tc-template/internal/pkg/errors"
	samplev1 "github.com/VH288/tc-template/internal/proto/sample/v1"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	reflectionpb "google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)
type memoryUsecase struct {
	mu    sync.Mutex
	items map[string]*entity.Sample
}

func newMemoryUsecase() *memoryUsecase {
	return &memoryUsecase{
		items: make(map[string]*entity.Sample),
	}
}

func (m *memoryUsecase) Create(ctx context.Context, name, description, status string) (*entity.Sample, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	id := uuid.NewString()
	s := &entity.Sample{
		ID:          id,
		Name:        name,
		Description: description,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	m.items[id] = s
	return s, nil
}

func (m *memoryUsecase) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.items[id]
	if !ok {
		return nil, apperrors.New(apperrors.CodeNotFound, "sample not found")
	}
	return s, nil
}

func (m *memoryUsecase) List(ctx context.Context, page, perPage int) ([]entity.Sample, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	list := make([]entity.Sample, 0, len(m.items))
	for _, item := range m.items {
		list = append(list, *item)
	}
	return list, int64(len(list)), nil
}

func (m *memoryUsecase) Update(ctx context.Context, id, name, description, status string) (*entity.Sample, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.items[id]
	if !ok {
		return nil, apperrors.New(apperrors.CodeNotFound, "sample not found")
	}

	s.Name = name
	s.Description = description
	s.Status = status
	s.UpdatedAt = time.Now().UTC()
	return s, nil
}

func (m *memoryUsecase) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.items[id]; !ok {
		return apperrors.New(apperrors.CodeNotFound, "sample not found")
	}
	delete(m.items, id)
	return nil
}

func TestGRPCServer_EndToEnd(t *testing.T) {
	bufferSize := 1024 * 1024
	lis := bufconn.Listen(bufferSize)

	uc := newMemoryUsecase()
	grpcHandler := handler.NewGRPCHandler(uc, telemetry.NewGRPCMetrics())

	server := grpcapp.NewServer(grpcapp.Dependencies{
		Sample: grpcHandler,
	})

	go func() {
		if err := server.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			t.Errorf("server error: %v", err)
		}
	}()
	defer server.GracefulStop()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Verify Health Check
	healthClient := healthpb.NewHealthClient(conn)
	healthResp, err := healthClient.Check(ctx, &healthpb.HealthCheckRequest{Service: ""})
	require.NoError(t, err)
	require.Equal(t, healthpb.HealthCheckResponse_SERVING, healthResp.GetStatus())

	serviceHealthResp, err := healthClient.Check(ctx, &healthpb.HealthCheckRequest{Service: "sample.v1.SampleService"})
	require.NoError(t, err)
	require.Equal(t, healthpb.HealthCheckResponse_SERVING, serviceHealthResp.GetStatus())

	// 2. Verify Reflection Server
	reflectionClient := reflectionpb.NewServerReflectionClient(conn)
	stream, err := reflectionClient.ServerReflectionInfo(ctx)
	require.NoError(t, err)
	err = stream.Send(&reflectionpb.ServerReflectionRequest{
		MessageRequest: &reflectionpb.ServerReflectionRequest_ListServices{
			ListServices: "*",
		},
	})
	require.NoError(t, err)
	refResp, err := stream.Recv()
	require.NoError(t, err)
	serviceList := refResp.GetListServicesResponse().GetService()
	serviceNames := make(map[string]bool)
	for _, s := range serviceList {
		serviceNames[s.GetName()] = true
	}
	require.True(t, serviceNames["sample.v1.SampleService"])
	require.True(t, serviceNames["grpc.health.v1.Health"])

	// 3. CRUD End-to-End
	client := samplev1.NewSampleServiceClient(conn)

	// CreateSample
	createResp, err := client.CreateSample(ctx, &samplev1.CreateSampleRequest{
		Name:        "Test Concert",
		Description: "VIP Ticket",
		Status:      "active",
	})
	require.NoError(t, err)
	require.NotEmpty(t, createResp.GetId())
	require.Equal(t, "Test Concert", createResp.GetName())
	require.Equal(t, "VIP Ticket", createResp.GetDescription())
	require.Equal(t, "active", createResp.GetStatus())
	require.NotNil(t, createResp.GetCreatedAt())
	require.NotNil(t, createResp.GetUpdatedAt())

	createdID := createResp.GetId()

	// GetSample
	getResp, err := client.GetSample(ctx, &samplev1.GetSampleRequest{Id: createdID})
	require.NoError(t, err)
	require.Equal(t, createdID, getResp.GetId())
	require.Equal(t, "Test Concert", getResp.GetName())

	// ListSamples
	listResp, err := client.ListSamples(ctx, &samplev1.ListSamplesRequest{Page: 1, PerPage: 10})
	require.NoError(t, err)
	require.Equal(t, int32(1), listResp.GetTotal())
	require.Len(t, listResp.GetItems(), 1)
	require.Equal(t, createdID, listResp.GetItems()[0].GetId())

	// UpdateSample
	updateResp, err := client.UpdateSample(ctx, &samplev1.UpdateSampleRequest{
		Id:          createdID,
		Name:        "Updated Concert",
		Description: "Standard",
		Status:      "archived",
	})
	require.NoError(t, err)
	require.Equal(t, createdID, updateResp.GetId())
	require.Equal(t, "Updated Concert", updateResp.GetName())
	require.Equal(t, "Standard", updateResp.GetDescription())
	require.Equal(t, "archived", updateResp.GetStatus())

	// DeleteSample
	deleteResp, err := client.DeleteSample(ctx, &samplev1.DeleteSampleRequest{Id: createdID})
	require.NoError(t, err)
	require.True(t, deleteResp.GetSuccess())

	// GetSample after delete -> NotFound
	deletedGetResp, err := client.GetSample(ctx, &samplev1.GetSampleRequest{Id: createdID})
	require.Nil(t, deletedGetResp)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
	require.Equal(t, "sample not found", st.Message())
}
