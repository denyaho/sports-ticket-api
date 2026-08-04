package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"42tokyo-road-to-dena-server/authbundle"
	"42tokyo-road-to-dena-server/internal/apperror"
	"42tokyo-road-to-dena-server/internal/domain"

	"github.com/google/uuid"
)

type StubreservationService struct {
	FakeCancelReservation   func(ctx context.Context, reservationID, userID uuid.UUID) error
	FakeCreateReservation   func(ctx context.Context, reqBody *domain.ReservationRequest, userID uuid.UUID) (*domain.Reservation, error)
	FakeGetUserReservations func(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error)
	FakeGetReservationByID  func(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error)
	FakePurchaseReservation func(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error)
	FakeExpiredReservations func(ctx context.Context) error
}

func (m *StubreservationService) CancelReservation(ctx context.Context, reservationID, userID uuid.UUID) error {
	return m.FakeCancelReservation(ctx, reservationID, userID)
}

func (m *StubreservationService) CreateReservation(
	ctx context.Context,
	reqBody *domain.ReservationRequest,
	userID uuid.UUID,
) (*domain.Reservation, error) {
	return m.FakeCreateReservation(ctx, reqBody, userID)
}

func (m *StubreservationService) GetUserReservations(
	ctx context.Context,
	userID uuid.UUID,
) ([]*domain.Reservation, error) {
	return m.FakeGetUserReservations(ctx, userID)
}

func (m *StubreservationService) GetReservationByID(
	ctx context.Context,
	reservationID, userID uuid.UUID,
) (*domain.Reservation, error) {
	return m.FakeGetReservationByID(ctx, reservationID, userID)
}

func (m *StubreservationService) PurchaseReservation(
	ctx context.Context,
	reservationID, userID uuid.UUID,
) (*domain.Reservation, error) {
	return m.FakePurchaseReservation(ctx, reservationID, userID)
}

func (m *StubreservationService) ExpiredReservations(ctx context.Context) error {
	return m.FakeExpiredReservations(ctx)
}

func createContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, authbundle.UserIDKey, uuid.New())
	return ctx
}

func TestHandleCancelReservation(t *testing.T) {
	var reservationID = "f7f7dad8-84fd-4c10-9f95-d2a68d38a46f"

	cancelTests := []struct {
		name           string
		setupContext   func() context.Context
		reservationID  string
		fakeErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			setupContext:   createContext,
			reservationID:  reservationID,
			fakeErr:        nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "not found",
			setupContext:   createContext,
			reservationID:  reservationID,
			fakeErr:        apperror.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "unauthorized",
			setupContext:   context.Background,
			reservationID:  reservationID,
			fakeErr:        nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "internal server error",
			setupContext:   createContext,
			reservationID:  reservationID,
			fakeErr:        apperror.ErrDatabase,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "bad request",
			setupContext:   createContext,
			reservationID:  "invalid-uuid",
			fakeErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "reservation conflict",
			setupContext:   createContext,
			reservationID:  reservationID,
			fakeErr:        apperror.ErrReservationConflict,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "reservation not pending",
			setupContext:   createContext,
			reservationID:  reservationID,
			fakeErr:        apperror.ErrReservationNotPending,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "reservation expired",
			setupContext:   createContext,
			reservationID:  reservationID,
			fakeErr:        apperror.ErrReservationExpired,
			expectedStatus: http.StatusGone,
		},
	}
	t.Parallel()
	for _, tt := range cancelTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &Handler{
				reservationService: &StubreservationService{
					FakeCancelReservation: func(_ context.Context, _, _ uuid.UUID) error {
						return tt.fakeErr
					},
				},
			}
			request := httptest.NewRequestWithContext(
				tt.setupContext(),
				"DELETE",
				"/api/reservations/"+tt.reservationID,
				nil,
			)
			response := httptest.NewRecorder()
			request.SetPathValue("id", tt.reservationID)

			h.toHandler(h.HandleCancelReservation)(response, request)

			assertStatus(t, response.Code, tt.expectedStatus)
		})
	}
}

func TestHandleCreateReservation(t *testing.T) {
	successReqBody := `{
		"game_id": "f7f7dad8-84fd-4c10-9f95-d2a68d38a46f",
		"seats": [
			{
				"seat_grade": "A",
				"quantity": 2
			}
		]
	}`
	failReqBody := `"invalid-json"`

	createTests := []struct {
		name           string
		setupContext   func() context.Context
		reqBody        string
		fakeErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			setupContext:   createContext,
			reqBody:        successReqBody,
			fakeErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			setupContext:   context.Background,
			reqBody:        successReqBody,
			fakeErr:        nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "internal server error",
			setupContext:   createContext,
			reqBody:        successReqBody,
			fakeErr:        apperror.ErrDatabase,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "bad request",
			setupContext:   createContext,
			reqBody:        failReqBody,
			fakeErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Insufficient seats available",
			setupContext:   createContext,
			reqBody:        successReqBody,
			fakeErr:        apperror.ErrReservationConflict,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "not found",
			setupContext:   createContext,
			reqBody:        successReqBody,
			fakeErr:        apperror.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "reservation expired",
			setupContext:   createContext,
			reqBody:        successReqBody,
			fakeErr:        apperror.ErrReservationExpired,
			expectedStatus: http.StatusGone,
		},
		{
			name:           "reservation not pending",
			setupContext:   createContext,
			reqBody:        successReqBody,
			fakeErr:        apperror.ErrReservationNotPending,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "reservation conflict",
			setupContext:   createContext,
			reqBody:        successReqBody,
			fakeErr:        apperror.ErrReservationConflict,
			expectedStatus: http.StatusConflict,
		},
	}
	t.Parallel()
	for _, tt := range createTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &Handler{
				reservationService: &StubreservationService{
					FakeCreateReservation: func(_ context.Context, _ *domain.ReservationRequest, _ uuid.UUID) (*domain.Reservation, error) {
						return &domain.Reservation{}, tt.fakeErr
					},
				},
			}
			request := httptest.NewRequestWithContext(
				tt.setupContext(),
				"POST",
				"/api/reservations",
				strings.NewReader(tt.reqBody),
			)
			response := httptest.NewRecorder()

			h.toHandler(h.HandleCreateReservation)(response, request)

			assertStatus(t, response.Code, tt.expectedStatus)
		})
	}
}

func TestHandleGetUserReservations(t *testing.T) {
	getUserTests := []struct {
		name           string
		setupContext   func() context.Context
		fakeErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			setupContext:   createContext,
			fakeErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			setupContext:   context.Background,
			fakeErr:        nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "internal server error",
			setupContext:   createContext,
			fakeErr:        apperror.ErrDatabase,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "not found",
			setupContext:   createContext,
			fakeErr:        apperror.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
	}
	t.Parallel()
	for _, tt := range getUserTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &Handler{
				reservationService: &StubreservationService{
					FakeGetUserReservations: func(_ context.Context, _ uuid.UUID) ([]*domain.Reservation, error) {
						return []*domain.Reservation{}, tt.fakeErr
					},
				},
			}
			request := httptest.NewRequestWithContext(tt.setupContext(), "GET", "/api/reservations", nil)
			response := httptest.NewRecorder()

			h.toHandler(h.HandleGetUserReservations)(response, request)

			assertStatus(t, response.Code, tt.expectedStatus)
		})
	}
}

func TestHandleGetReservationByID(t *testing.T) {
	reservationID := "f7f7dad8-84fd-4c10-9f95-d2a68d38a46f"
	getReservationTests := []struct {
		name           string
		setupContext   func() context.Context
		reservationID  string
		fakeErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			setupContext:   createContext,
			reservationID:  reservationID,
			fakeErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			setupContext:   context.Background,
			reservationID:  reservationID,
			fakeErr:        nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "internal server error",
			setupContext:   createContext,
			reservationID:  reservationID,
			fakeErr:        apperror.ErrDatabase,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "not found",
			setupContext:   createContext,
			reservationID:  reservationID,
			fakeErr:        apperror.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "bad request",
			setupContext:   createContext,
			reservationID:  "invalid-uuid",
			fakeErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
	}
	t.Parallel()
	for _, tt := range getReservationTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &Handler{
				reservationService: &StubreservationService{
					FakeGetReservationByID: func(_ context.Context, _, _ uuid.UUID) (*domain.Reservation, error) {
						return &domain.Reservation{}, tt.fakeErr
					},
				},
			}
			request := httptest.NewRequestWithContext(
				tt.setupContext(),
				"GET",
				"/api/reservations/"+tt.reservationID,
				nil,
			)
			response := httptest.NewRecorder()
			request.SetPathValue("id", tt.reservationID)

			h.toHandler(h.HandleGetReservationByID)(response, request)

			assertStatus(t, response.Code, tt.expectedStatus)
		})
	}
}

func TestHandlePurchaseReservation(t *testing.T) {
	reservationID := "f7f7dad8-84fd-4c10-9f95-d2a68d38a46f"
	purchageTests := []struct {
		name           string
		setupContext   func() context.Context
		reservationID  string
		fakeErr        error
		expectedStatus int
	}{
		{
			name:           "success",
			setupContext:   createContext,
			reservationID:  reservationID,
			fakeErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			setupContext:   context.Background,
			reservationID:  reservationID,
			fakeErr:        nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "internal server error",
			setupContext:   createContext,
			reservationID:  reservationID,
			fakeErr:        apperror.ErrDatabase,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "not found",
			setupContext:   createContext,
			reservationID:  reservationID,
			fakeErr:        apperror.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "bad request",
			setupContext:   createContext,
			reservationID:  "invalid-uuid",
			fakeErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
	}
	t.Parallel()
	for _, tt := range purchageTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &Handler{
				reservationService: &StubreservationService{
					FakePurchaseReservation: func(_ context.Context, _, _ uuid.UUID) (*domain.Reservation, error) {
						return &domain.Reservation{}, tt.fakeErr
					},
				},
			}
			request := httptest.NewRequestWithContext(
				tt.setupContext(),
				"POST",
				"/api/reservations/"+tt.reservationID+"/purchase",
				nil,
			)
			response := httptest.NewRecorder()
			request.SetPathValue("id", tt.reservationID)

			h.toHandler(h.HandlePurchaseReservation)(response, request)

			assertStatus(t, response.Code, tt.expectedStatus)
		})
	}
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}
