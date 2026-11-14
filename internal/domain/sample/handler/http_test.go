package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/VH288/tc-template/internal/domain/sample"
	"github.com/VH288/tc-template/internal/domain/sample/dto"
	"github.com/VH288/tc-template/internal/domain/sample/entity"
	"github.com/VH288/tc-template/internal/domain/sample/handler"
	apperrors "github.com/VH288/tc-sdk/errors"
	"github.com/VH288/tc-sdk/formatter"
	"github.com/VH288/tc-template/internal/pkg/testutil"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

type mockUsecase struct {
	items map[string]*entity.Sample
}

func newMockUsecase() *mockUsecase {
	return &mockUsecase{items: make(map[string]*entity.Sample)}
}

var _ sample.Usecase = (*mockUsecase)(nil)

func (m *mockUsecase) Create(ctx context.Context, name, description, status string) (*entity.Sample, error) {
	if status == "" {
		status = entity.StatusActive
	}
	s := &entity.Sample{
		ID:          "sample-1",
		Name:        name,
		Description: description,
		Status:      status,
	}
	m.items[s.ID] = s
	return s, nil
}

func (m *mockUsecase) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	s, ok := m.items[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return s, nil
}

func (m *mockUsecase) List(ctx context.Context, page, perPage int) ([]entity.Sample, int64, error) {
	out := make([]entity.Sample, 0, len(m.items))
	for _, s := range m.items {
		out = append(out, *s)
	}
	return out, int64(len(out)), nil
}

func (m *mockUsecase) Update(ctx context.Context, id, name, description, status string) (*entity.Sample, error) {
	s, ok := m.items[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	s.Name = name
	s.Description = description
	s.Status = status
	return s, nil
}

func (m *mockUsecase) Delete(ctx context.Context, id string) error {
	if _, ok := m.items[id]; !ok {
		return apperrors.ErrNotFound
	}
	delete(m.items, id)
	return nil
}

func setupTestRouter(h *handler.HTTPHandler) http.Handler {
	r := chi.NewRouter()
	r.Route("/samples", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
	return r
}

func TestSampleHTTPHandler_CRUD(t *testing.T) {
	uc := newMockUsecase()
	h := handler.NewHTTPHandler(uc)
	router := setupTestRouter(h)

	// 1. Create - Success (201)
	createReq := dto.CreateSampleRequest{
		Name:        "Concert Ticket",
		Description: "VIP Pass",
		Status:      "active",
	}
	rr := testutil.PerformRequest(t, router, http.MethodPost, "/samples", createReq)
	testutil.AssertStatus(t, rr, http.StatusCreated)

	var createResp formatter.Response
	testutil.DecodeJSON(t, rr, &createResp)
	require.True(t, createResp.Success)
	require.Equal(t, "sample created", createResp.Message)

	// 2. Create - Validation Error (422)
	invalidReq := map[string]string{
		"name": "", // required min 2
	}
	rr = testutil.PerformRequest(t, router, http.MethodPost, "/samples", invalidReq)
	testutil.AssertStatus(t, rr, http.StatusUnprocessableEntity)

	var errResp formatter.Response
	testutil.DecodeJSON(t, rr, &errResp)
	require.False(t, errResp.Success)
	require.NotNil(t, errResp.Error)
	require.Equal(t, string(apperrors.CodeValidation), errResp.Error.Code)

	// 3. GetByID - Success (200)
	rr = testutil.PerformRequest(t, router, http.MethodGet, "/samples/sample-1", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)

	var getResp formatter.Response
	testutil.DecodeJSON(t, rr, &getResp)
	require.True(t, getResp.Success)

	// 4. GetByID - Not Found (404)
	rr = testutil.PerformRequest(t, router, http.MethodGet, "/samples/unknown", nil)
	testutil.AssertStatus(t, rr, http.StatusNotFound)

	// 5. List - Success (200)
	rr = testutil.PerformRequest(t, router, http.MethodGet, "/samples?page=1&per_page=10", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)

	var listResp formatter.Response
	testutil.DecodeJSON(t, rr, &listResp)
	require.True(t, listResp.Success)
	require.NotNil(t, listResp.Meta)
	require.Equal(t, 1, listResp.Meta.Page)
	require.Equal(t, 10, listResp.Meta.PerPage)
	require.Equal(t, int64(1), listResp.Meta.Total)

	// 6. Update - Success (200)
	updateReq := dto.UpdateSampleRequest{
		Name:        "Concert Ticket Updated",
		Description: "Updated VIP Pass",
		Status:      "inactive",
	}
	rr = testutil.PerformRequest(t, router, http.MethodPut, "/samples/sample-1", updateReq)
	testutil.AssertStatus(t, rr, http.StatusOK)

	var updateResp formatter.Response
	testutil.DecodeJSON(t, rr, &updateResp)
	require.True(t, updateResp.Success)
	require.Equal(t, "sample updated", updateResp.Message)

	// 7. Update - Validation Failure (422)
	badUpdateReq := map[string]string{
		"name":   "Valid Name",
		"status": "invalid_status",
	}
	rr = testutil.PerformRequest(t, router, http.MethodPut, "/samples/sample-1", badUpdateReq)
	testutil.AssertStatus(t, rr, http.StatusUnprocessableEntity)

	// 8. Delete - Success (200)
	rr = testutil.PerformRequest(t, router, http.MethodDelete, "/samples/sample-1", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)

	var deleteResp formatter.Response
	testutil.DecodeJSON(t, rr, &deleteResp)
	require.True(t, deleteResp.Success)
	require.Equal(t, "sample deleted", deleteResp.Message)

	// 9. Delete - Not Found (404)
	rr = testutil.PerformRequest(t, router, http.MethodDelete, "/samples/sample-1", nil)
	testutil.AssertStatus(t, rr, http.StatusNotFound)
}
