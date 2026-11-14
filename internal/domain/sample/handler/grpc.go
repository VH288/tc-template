package handler

import (
	"context"

	"github.com/VH288/tc-sdk/grpcerrors"
	"github.com/VH288/tc-template/internal/domain/ports"
	"github.com/VH288/tc-template/internal/domain/sample"
	"github.com/VH288/tc-template/internal/domain/sample/entity"
	samplev1 "github.com/VH288/tc-template/internal/proto/sample/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCHandler struct {
	samplev1.UnimplementedSampleServiceServer
	usecase sample.Usecase
	metrics ports.GRPCMetrics
}

func NewGRPCHandler(usecase sample.Usecase, metrics ports.GRPCMetrics) *GRPCHandler {
	return &GRPCHandler{
		usecase: usecase,
		metrics: metrics,
	}
}

func (h *GRPCHandler) CreateSample(ctx context.Context, req *samplev1.CreateSampleRequest) (*samplev1.SampleResponse, error) {
	result, err := h.usecase.Create(ctx, req.GetName(), req.GetDescription(), req.GetStatus())
	h.recordMetrics("CreateSample", err)
	if err != nil {
		return nil, grpcerrors.FromError(err)
	}
	return toProtoSample(result), nil
}

func (h *GRPCHandler) GetSample(ctx context.Context, req *samplev1.GetSampleRequest) (*samplev1.SampleResponse, error) {
	result, err := h.usecase.GetByID(ctx, req.GetId())
	h.recordMetrics("GetSample", err)
	if err != nil {
		return nil, grpcerrors.FromError(err)
	}
	return toProtoSample(result), nil
}

func (h *GRPCHandler) ListSamples(ctx context.Context, req *samplev1.ListSamplesRequest) (*samplev1.ListSamplesResponse, error) {
	page := int(req.GetPage())
	perPage := int(req.GetPerPage())

	samples, total, err := h.usecase.List(ctx, page, perPage)
	h.recordMetrics("ListSamples", err)
	if err != nil {
		return nil, grpcerrors.FromError(err)
	}

	items := make([]*samplev1.SampleResponse, len(samples))
	for i := range samples {
		items[i] = toProtoSample(&samples[i])
	}

	return &samplev1.ListSamplesResponse{
		Items:   items,
		Total:   int32(total),
		Page:    req.GetPage(),
		PerPage: req.GetPerPage(),
	}, nil
}

func (h *GRPCHandler) UpdateSample(ctx context.Context, req *samplev1.UpdateSampleRequest) (*samplev1.SampleResponse, error) {
	result, err := h.usecase.Update(ctx, req.GetId(), req.GetName(), req.GetDescription(), req.GetStatus())
	h.recordMetrics("UpdateSample", err)
	if err != nil {
		return nil, grpcerrors.FromError(err)
	}
	return toProtoSample(result), nil
}

func (h *GRPCHandler) DeleteSample(ctx context.Context, req *samplev1.DeleteSampleRequest) (*samplev1.DeleteSampleResponse, error) {
	err := h.usecase.Delete(ctx, req.GetId())
	h.recordMetrics("DeleteSample", err)
	if err != nil {
		return nil, grpcerrors.FromError(err)
	}
	return &samplev1.DeleteSampleResponse{Success: true}, nil
}

func (h *GRPCHandler) recordMetrics(method string, err error) {
	if h.metrics == nil {
		return
	}
	result := "ok"
	if err != nil {
		result = "error"
	}
	h.metrics.IncRequest(method, result)
}

func toProtoSample(s *entity.Sample) *samplev1.SampleResponse {
	if s == nil {
		return nil
	}
	return &samplev1.SampleResponse{
		Id:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Status:      s.Status,
		CreatedAt:   timestamppb.New(s.CreatedAt),
		UpdatedAt:   timestamppb.New(s.UpdatedAt),
	}
}
