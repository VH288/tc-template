package handler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/VH288/tc-template/internal/domain/sample/entity"
	"github.com/VH288/tc-template/internal/domain/sample/handler"
	apperrors "github.com/VH288/tc-template/internal/pkg/errors"
	samplev1 "github.com/VH288/tc-template/internal/proto/sample/v1"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockGRPCMetrics struct {
	calls []metricCall
}

type metricCall struct {
	method string
	result string
}

func (m *mockGRPCMetrics) IncRequest(method, result string) {
	m.calls = append(m.calls, metricCall{method: method, result: result})
}

type configurableUsecase struct {
	createFn  func(ctx context.Context, name, description, status string) (*entity.Sample, error)
	getByIDFn func(ctx context.Context, id string) (*entity.Sample, error)
	listFn    func(ctx context.Context, page, perPage int) ([]entity.Sample, int64, error)
	updateFn  func(ctx context.Context, id, name, description, status string) (*entity.Sample, error)
	deleteFn  func(ctx context.Context, id string) error
}

func (m *configurableUsecase) Create(ctx context.Context, name, description, status string) (*entity.Sample, error) {
	if m.createFn != nil {
		return m.createFn(ctx, name, description, status)
	}
	return nil, nil
}

func (m *configurableUsecase) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *configurableUsecase) List(ctx context.Context, page, perPage int) ([]entity.Sample, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, page, perPage)
	}
	return nil, 0, nil
}

func (m *configurableUsecase) Update(ctx context.Context, id, name, description, status string) (*entity.Sample, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, name, description, status)
	}
	return nil, nil
}

func (m *configurableUsecase) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func TestGRPCHandler_CreateSample(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		metrics := &mockGRPCMetrics{}
		now := time.Now().Truncate(time.Second)
		uc := &configurableUsecase{
			createFn: func(ctx context.Context, name, description, s string) (*entity.Sample, error) {
				return &entity.Sample{
					ID:          "sample-1",
					Name:        name,
					Description: description,
					Status:      s,
					CreatedAt:   now,
					UpdatedAt:   now,
				}, nil
			},
		}

		h := handler.NewGRPCHandler(uc, metrics)
		resp, err := h.CreateSample(context.Background(), &samplev1.CreateSampleRequest{
			Name:        "Test Sample",
			Description: "Test Description",
			Status:      "active",
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, "sample-1", resp.GetId())
		require.Equal(t, "Test Sample", resp.GetName())
		require.Equal(t, "Test Description", resp.GetDescription())
		require.Equal(t, "active", resp.GetStatus())
		require.Equal(t, now.Unix(), resp.GetCreatedAt().GetSeconds())
		require.Equal(t, now.Unix(), resp.GetUpdatedAt().GetSeconds())

		require.Len(t, metrics.calls, 1)
		require.Equal(t, "CreateSample", metrics.calls[0].method)
		require.Equal(t, "ok", metrics.calls[0].result)
	})

	t.Run("failure maps error", func(t *testing.T) {
		metrics := &mockGRPCMetrics{}
		uc := &configurableUsecase{
			createFn: func(ctx context.Context, name, description, s string) (*entity.Sample, error) {
				return nil, apperrors.New(apperrors.CodeValidation, "invalid input")
			},
		}

		h := handler.NewGRPCHandler(uc, metrics)
		resp, err := h.CreateSample(context.Background(), &samplev1.CreateSampleRequest{})

		require.Nil(t, resp)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
		require.Equal(t, "invalid input", st.Message())

		require.Len(t, metrics.calls, 1)
		require.Equal(t, "CreateSample", metrics.calls[0].method)
		require.Equal(t, "error", metrics.calls[0].result)
	})
}

func TestGRPCHandler_GetSample(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		metrics := &mockGRPCMetrics{}
		uc := &configurableUsecase{
			getByIDFn: func(ctx context.Context, id string) (*entity.Sample, error) {
				return &entity.Sample{
					ID:     id,
					Name:   "Found Sample",
					Status: "active",
				}, nil
			},
		}

		h := handler.NewGRPCHandler(uc, metrics)
		resp, err := h.GetSample(context.Background(), &samplev1.GetSampleRequest{Id: "sample-123"})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, "sample-123", resp.GetId())
		require.Equal(t, "Found Sample", resp.GetName())

		require.Len(t, metrics.calls, 1)
		require.Equal(t, "GetSample", metrics.calls[0].method)
		require.Equal(t, "ok", metrics.calls[0].result)
	})

	t.Run("not found", func(t *testing.T) {
		metrics := &mockGRPCMetrics{}
		uc := &configurableUsecase{
			getByIDFn: func(ctx context.Context, id string) (*entity.Sample, error) {
				return nil, apperrors.New(apperrors.CodeNotFound, "sample not found")
			},
		}

		h := handler.NewGRPCHandler(uc, metrics)
		resp, err := h.GetSample(context.Background(), &samplev1.GetSampleRequest{Id: "missing"})

		require.Nil(t, resp)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.NotFound, st.Code())
		require.Equal(t, "sample not found", st.Message())

		require.Len(t, metrics.calls, 1)
		require.Equal(t, "GetSample", metrics.calls[0].method)
		require.Equal(t, "error", metrics.calls[0].result)
	})
}

func TestGRPCHandler_ListSamples(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		metrics := &mockGRPCMetrics{}
		uc := &configurableUsecase{
			listFn: func(ctx context.Context, page, perPage int) ([]entity.Sample, int64, error) {
				return []entity.Sample{
					{ID: "s1", Name: "Sample 1"},
					{ID: "s2", Name: "Sample 2"},
				}, 2, nil
			},
		}

		h := handler.NewGRPCHandler(uc, metrics)
		resp, err := h.ListSamples(context.Background(), &samplev1.ListSamplesRequest{
			Page:    1,
			PerPage: 10,
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.GetItems(), 2)
		require.Equal(t, int32(2), resp.GetTotal())
		require.Equal(t, int32(1), resp.GetPage())
		require.Equal(t, int32(10), resp.GetPerPage())
		require.Equal(t, "s1", resp.GetItems()[0].GetId())

		require.Len(t, metrics.calls, 1)
		require.Equal(t, "ListSamples", metrics.calls[0].method)
		require.Equal(t, "ok", metrics.calls[0].result)
	})

	t.Run("error", func(t *testing.T) {
		metrics := &mockGRPCMetrics{}
		uc := &configurableUsecase{
			listFn: func(ctx context.Context, page, perPage int) ([]entity.Sample, int64, error) {
				return nil, 0, apperrors.New(apperrors.CodeInternal, "db failure")
			},
		}

		h := handler.NewGRPCHandler(uc, metrics)
		resp, err := h.ListSamples(context.Background(), &samplev1.ListSamplesRequest{
			Page:    1,
			PerPage: 10,
		})

		require.Nil(t, resp)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.Internal, st.Code())
	})
}

func TestGRPCHandler_UpdateSample(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		metrics := &mockGRPCMetrics{}
		uc := &configurableUsecase{
			updateFn: func(ctx context.Context, id, name, description, s string) (*entity.Sample, error) {
				return &entity.Sample{
					ID:          id,
					Name:        name,
					Description: description,
					Status:      s,
				}, nil
			},
		}

		h := handler.NewGRPCHandler(uc, metrics)
		resp, err := h.UpdateSample(context.Background(), &samplev1.UpdateSampleRequest{
			Id:          "s1",
			Name:        "Updated Name",
			Description: "Updated Desc",
			Status:      "archived",
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, "s1", resp.GetId())
		require.Equal(t, "Updated Name", resp.GetName())
		require.Equal(t, "Updated Desc", resp.GetDescription())
		require.Equal(t, "archived", resp.GetStatus())

		require.Len(t, metrics.calls, 1)
		require.Equal(t, "UpdateSample", metrics.calls[0].method)
		require.Equal(t, "ok", metrics.calls[0].result)
	})

	t.Run("conflict error", func(t *testing.T) {
		metrics := &mockGRPCMetrics{}
		uc := &configurableUsecase{
			updateFn: func(ctx context.Context, id, name, description, s string) (*entity.Sample, error) {
				return nil, apperrors.New(apperrors.CodeConflict, "version conflict")
			},
		}

		h := handler.NewGRPCHandler(uc, metrics)
		resp, err := h.UpdateSample(context.Background(), &samplev1.UpdateSampleRequest{
			Id: "s1",
		})

		require.Nil(t, resp)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.AlreadyExists, st.Code())
		require.Equal(t, "version conflict", st.Message())
	})
}

func TestGRPCHandler_DeleteSample(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		metrics := &mockGRPCMetrics{}
		uc := &configurableUsecase{
			deleteFn: func(ctx context.Context, id string) error {
				return nil
			},
		}

		h := handler.NewGRPCHandler(uc, metrics)
		resp, err := h.DeleteSample(context.Background(), &samplev1.DeleteSampleRequest{Id: "s1"})

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.GetSuccess())

		require.Len(t, metrics.calls, 1)
		require.Equal(t, "DeleteSample", metrics.calls[0].method)
		require.Equal(t, "ok", metrics.calls[0].result)
	})

	t.Run("not found", func(t *testing.T) {
		metrics := &mockGRPCMetrics{}
		uc := &configurableUsecase{
			deleteFn: func(ctx context.Context, id string) error {
				return apperrors.New(apperrors.CodeNotFound, "sample not found")
			},
		}

		h := handler.NewGRPCHandler(uc, metrics)
		resp, err := h.DeleteSample(context.Background(), &samplev1.DeleteSampleRequest{Id: "s1"})

		require.Nil(t, resp)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.NotFound, st.Code())
		require.Equal(t, "sample not found", st.Message())
	})
}

func TestGRPCHandler_ErrorMapping(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode codes.Code
		expectedMsg  string
	}{
		{
			name:         "not found",
			err:          apperrors.New(apperrors.CodeNotFound, "not found item"),
			expectedCode: codes.NotFound,
			expectedMsg:  "not found item",
		},
		{
			name:         "validation error",
			err:          apperrors.New(apperrors.CodeValidation, "validation failed"),
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "validation failed",
		},
		{
			name:         "bad request",
			err:          apperrors.New(apperrors.CodeBadRequest, "bad request input"),
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "bad request input",
		},
		{
			name:         "conflict",
			err:          apperrors.New(apperrors.CodeConflict, "already exists"),
			expectedCode: codes.AlreadyExists,
			expectedMsg:  "already exists",
		},
		{
			name:         "unauthorized",
			err:          apperrors.New(apperrors.CodeUnauthorized, "unauthenticated"),
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "unauthenticated",
		},
		{
			name:         "forbidden",
			err:          apperrors.New(apperrors.CodeForbidden, "permission denied"),
			expectedCode: codes.PermissionDenied,
			expectedMsg:  "permission denied",
		},
		{
			name:         "unavailable",
			err:          apperrors.New(apperrors.CodeUnavailable, "service down"),
			expectedCode: codes.Unavailable,
			expectedMsg:  "service down",
		},
		{
			name:         "timeout",
			err:          apperrors.New(apperrors.CodeTimeout, "deadline exceeded"),
			expectedCode: codes.DeadlineExceeded,
			expectedMsg:  "deadline exceeded",
		},
		{
			name:         "internal custom code",
			err:          apperrors.New(apperrors.CodeInternal, "database crash"),
			expectedCode: codes.Internal,
			expectedMsg:  "database crash",
		},
		{
			name:         "non-apperror standard error",
			err:          errors.New("raw unexpected failure"),
			expectedCode: codes.Internal,
			expectedMsg:  "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &configurableUsecase{
				getByIDFn: func(ctx context.Context, id string) (*entity.Sample, error) {
					return nil, tt.err
				},
			}
			h := handler.NewGRPCHandler(uc, nil)
			_, err := h.GetSample(context.Background(), &samplev1.GetSampleRequest{Id: "test"})
			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.expectedCode, st.Code())
			require.Equal(t, tt.expectedMsg, st.Message())
		})
	}
}
