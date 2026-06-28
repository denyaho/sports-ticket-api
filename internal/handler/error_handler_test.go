package handler

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"context"
	"errors"
	"database/sql"
	"42tokyo-road-to-dena-server/internal/apperror"
)

func TestHandleError(t *testing.T) {
	errorTests := []struct {
		name string
		fakeErr error
		expectedStatus int
	}{
		{
			name: "Request Timeout",
			fakeErr: context.Canceled,
			expectedStatus: http.StatusRequestTimeout,
		},
		{
			name: "Gateway Timeout",
			fakeErr: context.DeadlineExceeded,
			expectedStatus: http.StatusGatewayTimeout,
		},
		{
			name: "Internal Server Error",
			fakeErr: sql.ErrConnDone,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "No Rows",
			fakeErr: sql.ErrNoRows,
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Not Found",
			fakeErr: apperror.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Insufficient Tickets",
			fakeErr: apperror.ErrInsufficientTickets,
			expectedStatus: http.StatusConflict,
		},
		{
			name: "User Not Found",
			fakeErr: apperror.ErrUserNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Database Error",
			fakeErr: apperror.ErrDatabase,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Duplicate Email",
			fakeErr: apperror.ErrDuplicateEmail,
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Err internal",
			fakeErr: apperror.ErrInternal,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Unauthorized",
			fakeErr: apperror.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Reservation Expired",
			fakeErr: apperror.ErrReservationExpired,
			expectedStatus: http.StatusGone,
		},
		{
			name: "Reservation Conflict",
			fakeErr: apperror.ErrReservationConflict,
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Reservation Not Pending",
			fakeErr: apperror.ErrReservationNotPending,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Bad Request",
			fakeErr: apperror.ErrBadRequest,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Input",
			fakeErr: apperror.ErrInvalidInput,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Authentication Failed",
			fakeErr: apperror.ErrAuthenticationFailed,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Forbidden",
			fakeErr: apperror.ErrForbidden,
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Default Case",
			fakeErr: errors.New("some unknown error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{}
			w := httptest.NewRecorder()
			h.HandleError(w, tt.fakeErr)
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}