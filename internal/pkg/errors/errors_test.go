package errors_test

import (
	"errors"
	"net/http"
	"testing"

	apperrors "github.com/VH288/tc-template/internal/pkg/errors"

	"github.com/stretchr/testify/require"
)

func TestAppError_Codes(t *testing.T) {
	require.Equal(t, http.StatusNotFound, apperrors.ErrNotFound.HTTPStatus)
	require.Equal(t, http.StatusConflict, apperrors.ErrConflict.HTTPStatus)
	require.Equal(t, http.StatusUnprocessableEntity, apperrors.ErrValidation.HTTPStatus)
	require.Equal(t, http.StatusBadRequest, apperrors.ErrBadRequest.HTTPStatus)
	require.Equal(t, http.StatusUnauthorized, apperrors.ErrUnauthorized.HTTPStatus)
	require.Equal(t, http.StatusForbidden, apperrors.ErrForbidden.HTTPStatus)
	require.Equal(t, http.StatusServiceUnavailable, apperrors.ErrUnavailable.HTTPStatus)
	require.Equal(t, http.StatusGatewayTimeout, apperrors.ErrTimeout.HTTPStatus)
	require.Equal(t, http.StatusInternalServerError, apperrors.ErrInternal.HTTPStatus)

	wrapped := apperrors.Wrap(apperrors.ErrNotFound, apperrors.CodeInternal, "boom")
	require.Contains(t, wrapped.Error(), "boom")
	require.Equal(t, apperrors.ErrNotFound, wrapped.Unwrap())

	got, ok := apperrors.As(wrapped)
	require.True(t, ok)
	require.Equal(t, apperrors.CodeInternal, got.Code)

	var stdErr = errors.New("standard error")
	_, okStd := apperrors.As(stdErr)
	require.False(t, okStd)

	withDet := apperrors.WithDetails(apperrors.ErrValidation, map[string]string{"field": "bad"})
	require.Equal(t, "bad", withDet.Details["field"])

	withMsg := apperrors.WithMessage(apperrors.ErrNotFound, "custom not found")
	require.Equal(t, "custom not found", withMsg.Message)

	require.True(t, apperrors.Is(wrapped, apperrors.ErrNotFound))
}
