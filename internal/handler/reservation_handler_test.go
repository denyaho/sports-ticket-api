package handler

import (
	"strings"
	"net/http"
	"net/http/httptest"
	"testing"
	"42tokyo-road-to-dena-server/authbundle"
	"context"
	"github.com/google/uuid"
	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/apperror"
)

type StubreservationService struct {
	FakeCancelReservation func(ctx context.Context, reservationID, userID uuid.UUID) error
	FakeCreateReservation func(ctx context.Context, reqBody *domain.ReservationRequest, userID uuid.UUID) (*domain.Reservation, error)
	FakeGetUserReservations func(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error)
	FakeGetReservationByID func(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error)
	FakePurchaseReservation func(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error)
	FakeCheckExpiredReservations func(ctx context.Context) error

}

func (m *StubreservationService) CancelReservation(ctx context.Context, reservationID, userID uuid.UUID) error {
	return m.FakeCancelReservation(ctx, reservationID, userID)
}

func (m *StubreservationService) CreateReservation(ctx context.Context, reqBody *domain.ReservationRequest, userID uuid.UUID) (*domain.Reservation, error) {
	return m.FakeCreateReservation(ctx, reqBody, userID)
}

func (m *StubreservationService) GetUserReservations(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error) {
	return m.FakeGetUserReservations(ctx, userID)
}

func (m *StubreservationService) GetReservationByID(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error) {
	return m.FakeGetReservationByID(ctx, reservationID, userID)
}

func (m *StubreservationService) PurchaseReservation(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error) {
	return m.FakePurchaseReservation(ctx, reservationID, userID)
}

func (m *StubreservationService) CheckExpiredReservations(ctx context.Context) error {
	return m.FakeCheckExpiredReservations(ctx)
}




func TestHandleCancelReservation(t *testing.T) {

	userID := uuid.MustParse("a7c3d0fe-6743-11f1-9249-523f294cde2a")
	var reservationID = "f7f7dad8-84fd-4c10-9f95-d2a68d38a46f"

	cancelTests := []struct {
		name string
		setupContext func() context.Context
		reservationID string
		fakeErr error
		expectedStatus int
	}{
		{
			name: "success",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, authbundle.UserIDKey, userID)
				return ctx
			},
			reservationID: reservationID,
			fakeErr: nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "not found",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, authbundle.UserIDKey, userID)
				return ctx
			},
			reservationID: reservationID,
			fakeErr: apperror.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "unauthorized",	
			setupContext: func() context.Context {
				return context.Background()
			},
			reservationID: reservationID,
			fakeErr: nil,
			expectedStatus: http.StatusUnauthorized,	
		},
		{
			name: "internal server error",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, authbundle.UserIDKey, userID)
				return ctx
			},
			reservationID: reservationID,
			fakeErr: apperror.ErrDatabase,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "bad request",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, authbundle.UserIDKey, userID)
				return ctx
			},
			reservationID: "invalid-uuid",
			fakeErr: nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "reservation conflict",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, authbundle.UserIDKey, userID)
				return ctx
			},
			reservationID: reservationID,
			fakeErr: apperror.ErrReservationConflict,
			expectedStatus: http.StatusConflict,
		},
		{
			name: "reservation not pending",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, authbundle.UserIDKey, userID)
				return ctx
			},
			reservationID: reservationID,
			fakeErr: apperror.ErrReservationNotPending,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "reservation expired",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, authbundle.UserIDKey, userID)
				return ctx
			},
			reservationID: reservationID,
			fakeErr: apperror.ErrReservationExpired,
			expectedStatus: http.StatusGone,
		},
	}

	for _, tt := range cancelTests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				reservationService: &StubreservationService{
					FakeCancelReservation: func(ctx context.Context, reservationID, userID uuid.UUID) error {
						return tt.fakeErr
					},
				},
			}				
			request := httptest.NewRequestWithContext(tt.setupContext(), "DELETE", "/api/reservations/"+ tt.reservationID, nil)
			response := httptest.NewRecorder()
			request.SetPathValue("id", tt.reservationID)

			h.HandleCancelReservation(response, request)

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

	createContext := func() context.Context {
		ctx := context.Background()
		ctx = context.WithValue(ctx, authbundle.UserIDKey, uuid.New())
		return ctx
	}

	createTests := []struct {
		name string
		setupContext func() context.Context
		reqBody string
		fakeErr error
		expectedStatus int
	}{
		{
			name: "success",
			setupContext: createContext,
			reqBody: successReqBody,
			fakeErr: nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "unauthorized",
			setupContext: func() context.Context {
				return context.Background()
			},
			reqBody: successReqBody,
			fakeErr: nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "internal server error",
			setupContext: createContext,
			reqBody: successReqBody,
			fakeErr: apperror.ErrDatabase,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "bad request",
			setupContext: createContext,
			reqBody: failReqBody,
			fakeErr: nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Insufficient seats available",
			setupContext: createContext,
			reqBody: successReqBody,
			fakeErr: apperror.ErrReservationConflict,
			expectedStatus: http.StatusConflict,
		},
		{
			name: "not found",
			setupContext: createContext,
			reqBody: successReqBody,
			fakeErr: apperror.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "reservation expired",
			setupContext: createContext,
			reqBody: successReqBody,
			fakeErr: apperror.ErrReservationExpired,
			expectedStatus: http.StatusGone,
		},
		{
			name: "reservation not pending",
			setupContext: createContext,
			reqBody: successReqBody,
			fakeErr: apperror.ErrReservationNotPending,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "reservation conflict",
			setupContext: createContext,
			reqBody: successReqBody,
			fakeErr: apperror.ErrReservationConflict,
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range createTests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				reservationService: &StubreservationService{
					FakeCreateReservation: func(ctx context.Context, reqBody *domain.ReservationRequest, userID uuid.UUID) (*domain.Reservation, error) {
						return &domain.Reservation{}, tt.fakeErr
					},
				},
			}
			request := httptest.NewRequestWithContext(tt.setupContext(),"POST", "/api/reservations", strings.NewReader(tt.reqBody))
			response := httptest.NewRecorder()

			h.HandleCreateReservation(response, request)

			assertStatus(t, response.Code, tt.expectedStatus)			
		})
	}
}

func TestHandleGetUserReservations(t *testing.T) {


	createContext := func() context.Context {
		ctx := context.Background()
		ctx = context.WithValue(ctx, authbundle.UserIDKey, uuid.New())
		return ctx
	}
	getUserTests := []struct {
		name string
		setupContext func() context.Context
		fakeErr error
		expectedStatus int
	}{
		{
			name: "success",
			setupContext: createContext,
			fakeErr: nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "unauthorized",
			setupContext: func() context.Context {
				return context.Background()
			},
			fakeErr: nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "internal server error",
			setupContext: createContext,
			fakeErr: apperror.ErrDatabase,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "not found",
			setupContext: createContext,
			fakeErr: apperror.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
	}
	for _, tt := range getUserTests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				reservationService: &StubreservationService{
					FakeGetUserReservations: func(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error) {
						return []*domain.Reservation{}, tt.fakeErr
					},
				},
			}
			request := httptest.NewRequestWithContext(tt.setupContext(), "GET", "/api/reservations", nil)
			response := httptest.NewRecorder()

			h.HandleGetUserReservations(response, request)

			assertStatus(t, response.Code, tt.expectedStatus)
		})
	}
}