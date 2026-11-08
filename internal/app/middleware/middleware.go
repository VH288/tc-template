package middleware

import (
	"log/slog"
	"net/http"
	"time"

	sdklogger "github.com/VH288/tc-sdk/logger"
	"github.com/VH288/tc-template/internal/pkg/constant"
	apperrors "github.com/VH288/tc-template/internal/pkg/errors"
	"github.com/VH288/tc-template/internal/pkg/formatter"
	"github.com/VH288/tc-template/internal/pkg/helper"

	"github.com/google/uuid"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(constant.HeaderRequestID)
		if reqID == "" {
			reqID = uuid.NewString()
		}
		w.Header().Set(constant.HeaderRequestID, reqID)
		ctx := helper.WithRequestID(r.Context(), reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Logging(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			reqID := helper.RequestIDFromContext(r.Context())
			reqLogger := base.With(
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)
			ctx := sdklogger.WithContext(r.Context(), reqLogger)

			next.ServeHTTP(sw, r.WithContext(ctx))

			reqLogger.Info("request completed",
				slog.Int("status", sw.status),
				slog.Duration("latency", time.Since(start)),
			)
		})
	}
}

func Recoverer(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					base.Error("panic recovered", slog.Any("panic", rec))
					formatter.Fail(w, apperrors.ErrInternal)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
