package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"42tokyo-road-to-dena-server/internal/apperror"
)

func TestHandleError(t *testing.T) {
	errorTests := []struct {
		name           string
		fakeErr        error
		expectedStatus int
	}{
		{
			name:           "Request Entity Too Large",
			fakeErr:        &http.MaxBytesError{},
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:           "Bad Request - Syntax Error",
			fakeErr:        apperror.ErrBadRequest,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Bad Request - Unmarshal Type Error",
			fakeErr:        &json.UnmarshalTypeError{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Request Canceled",
			fakeErr:        context.Canceled,
			expectedStatus: 499,
		},
		{
			name:           "Deadline Exceeded",
			fakeErr:        context.DeadlineExceeded,
			expectedStatus: http.StatusGatewayTimeout,
		},
		{
			name:           "Conflict",
			fakeErr:        apperror.ErrConflict,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Validation Error",
			fakeErr:        apperror.ErrValidation,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Retryable Error",
			fakeErr:        apperror.ErrRetryable,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "Timeout Error",
			fakeErr:        apperror.ErrTimeout,
			expectedStatus: http.StatusGatewayTimeout,
		},
		{
			name:           "Unavailable Error",
			fakeErr:        apperror.ErrUnavailable,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "Not Found Error",
			fakeErr:        apperror.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Insufficient Tickets Error",
			fakeErr:        apperror.ErrInsufficientTickets,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Unauthorized Error",
			fakeErr:        apperror.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Internal Server Error",
			fakeErr:        apperror.ErrInternal,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Bad Request Error",
			fakeErr:        apperror.ErrBadRequest,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Default Case",
			fakeErr:        errors.New("some unknown error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	t.Parallel()
	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &Handler{}
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			h.HandleError(w, r, tt.fakeErr)
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
