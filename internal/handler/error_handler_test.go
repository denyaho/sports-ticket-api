package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"42tokyo-road-to-dena-server/internal/apperror"
)

func newTestHandler() *Handler {
	return &Handler{
		logger: slog.New(slog.DiscardHandler),
	}
}

func withRequestState(r *http.Request) (*http.Request, *RequestState) {
	state := &RequestState{}
	ctx := context.WithValue(r.Context(), requestStateKey, state)
	return r.WithContext(ctx), state
}

func TestHandleError_StatusMapping(t *testing.T) {
	t.Parallel()

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	errorTests := []struct {
		name           string
		fakeErr        error
		requestCtx     context.Context
		expectedStatus int
	}{
		{
			name:           "MaxBytesError -> 413",
			fakeErr:        &http.MaxBytesError{},
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:           "json.SyntaxError -> 400",
			fakeErr:        &json.SyntaxError{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "json.UnmarshalTypeError -> 400",
			fakeErr:        &json.UnmarshalTypeError{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "context.Canceled with canceled ctx -> 499",
			fakeErr:        context.Canceled,
			requestCtx:     canceledCtx,
			expectedStatus: 499,
		},
		{
			name:           "context.Canceled with live ctx -> 500",
			fakeErr:        context.Canceled,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "context.DeadlineExceeded -> 504",
			fakeErr:        context.DeadlineExceeded,
			expectedStatus: http.StatusGatewayTimeout,
		},
		{
			name:           "ErrConflict -> 409",
			fakeErr:        apperror.ErrConflict,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "ErrValidation -> 400",
			fakeErr:        apperror.ErrValidation,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ErrRetryable -> 503",
			fakeErr:        apperror.ErrRetryable,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "ErrTimeout -> 504",
			fakeErr:        apperror.ErrTimeout,
			expectedStatus: http.StatusGatewayTimeout,
		},
		{
			name:           "ErrUnavailable -> 503",
			fakeErr:        apperror.ErrUnavailable,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "ErrNotFound -> 404",
			fakeErr:        apperror.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "ErrInsufficientTickets -> 409",
			fakeErr:        apperror.ErrInsufficientTickets,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "ErrUnauthorized -> 401",
			fakeErr:        apperror.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "ErrInternal -> 500",
			fakeErr:        apperror.ErrInternal,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "ErrBadRequest -> 400",
			fakeErr:        apperror.ErrBadRequest,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unknown error -> 500 (default)",
			fakeErr:        errors.New("something unexpected"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "wrapped ErrConflict -> 409",
			fakeErr:        fmt.Errorf("repo: %w", apperror.ErrConflict),
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "wrapped json.SyntaxError -> 400",
			fakeErr:        fmt.Errorf("decode: %w", &json.SyntaxError{}),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := newTestHandler()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			if tt.requestCtx != nil {
				r = r.WithContext(tt.requestCtx)
			}
			r, state := withRequestState(r)

			h.HandleError(w, r, tt.fakeErr)

			if w.Code != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, w.Code)
			}

			if got := w.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type: want application/json, got %q", got)
			}

			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("response body is not JSON: %v (body=%q)", err, w.Body.String())
			}
			if want, got := http.StatusText(tt.expectedStatus), body["error"]; want != got {
				t.Errorf("body.error: want %q, got %q", want, got)
			}

			if !errors.Is(state.Err, tt.fakeErr) && !errors.Is(tt.fakeErr, state.Err) {
				t.Errorf("state.Err: want %v, got %v", tt.fakeErr, state.Err)
			}
		})
	}
}

func TestHandleError_NoRequestStateInContext(t *testing.T) {
	t.Parallel()

	h := newTestHandler()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/test", nil)

	h.HandleError(w, r, apperror.ErrNotFound)

	if w.Code != http.StatusNotFound {
		t.Errorf("status: want %d, got %d", http.StatusNotFound, w.Code)
	}
	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type: want application/json, got %q", got)
	}
}
