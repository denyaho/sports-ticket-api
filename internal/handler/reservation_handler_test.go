package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"errors"
	"bytes"
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"io"
	"time"


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

// テスト内容
// 1. コンテキストにユーザーIDが含まれていない場合、ErrUnauthorizedが返されることを確認する。
// 2. コンテキストにreservationIDが含まれていない場合、ErrBadRequestが返されることを確認する。
// 3. StubreservationServiceのFakeCancelReservationが正しく呼び出されることを確認する。
// 4. StubreservationServiceのFakeCancelReservationがエラーを返す場合、HandleCancelReservationがそのエラーを返すことを確認する。
// 5. 成功した場合、StautsNoContentが返されることを確認する。
var (
	validUserID        = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	validReservationID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

func TestHandleCancelReservation(t *testing.T) {
	t.Parallel()
	cancelTests := []struct {
		name          string
		called      bool
		wantErr	   error
		reservationID string
		userID		*uuid.UUID
		serviceErr    error
		wantStatus    int
	}{
		{
			name: "success",
			reservationID: validReservationID.String(),
			userID: &validUserID,
			called: true,
			wantStatus: http.StatusNoContent,
		},
		{
			name: "unAuthorized",
			called: false,
			wantErr: apperror.ErrUnauthorized,
		},
		{
			name: "BadRequest",
			reservationID: "invalid-reservation-id",
			userID: &validUserID,
			called: false,
			wantErr: apperror.ErrBadRequest,
		},
		{
			name: "notFound",
			reservationID: validReservationID.String(),
			userID: &validUserID,
			called: true,
			serviceErr: apperror.ErrNotFound,
			wantErr: apperror.ErrNotFound,
		},
		{
			name: "internalServerError",
			reservationID: validReservationID.String(),
			userID: &validUserID,
			called: true,
			serviceErr: apperror.ErrDatabase,
			wantErr: apperror.ErrDatabase,
		},
	}
	for _, tt := range cancelTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var (
				called bool
				gotIDReservation uuid.UUID
				gotIDUser uuid.UUID
			)
			var ctx context.Context
			if tt.userID != nil {
				ctx = authbundle.SetUserIDInContext(context.Background(), *tt.userID)
			} else {
				ctx = context.Background()
			}
			h := &Handler{
				reservationService: &StubreservationService{
					FakeCancelReservation: func(ctx context.Context, rID, uID uuid.UUID) error {
						called = true
						gotIDReservation = rID
						gotIDUser = uID
						if tt.serviceErr != nil {
							return tt.serviceErr
						}
						return nil
					},
				},
			}
			request := httptest.NewRequestWithContext(
				ctx,
				"DELETE",
				"/api/reservations/"+tt.reservationID,
				nil,
			)
			response := httptest.NewRecorder()
			request.SetPathValue("id", tt.reservationID)
			
			err := h.HandleCancelReservation(response, request)

			// 失敗系
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				if called != tt.called {
					t.Errorf("service called = %v, want %v", called, tt.called)
				}
				return
			}
			// 成功系
			if err != nil {
				t.Errorf("unexpected error = %v", err)
			}
			if !called {
				t.Errorf("expected service to be called, but it was not")
			}
			if response.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
			}
			if tt.called && gotIDUser != *tt.userID {
				t.Errorf("service called with user id = %v, want %v", gotIDUser, *tt.userID)
			}
			wantIDReservation, _ := uuid.Parse(tt.reservationID)
			if tt.called && gotIDReservation != wantIDReservation {
				t.Errorf("service called with reservation id = %v, want %v", gotIDReservation, wantIDReservation)
			}
			if ct := response.Header().Get("Content-Type"); ct != ""{
				t.Errorf("expected no Content-Type header, got %s", ct)
			}
			if diff := response.Body.Len(); diff != 0 {
				t.Errorf("expected empty body, got %d bytes", diff)
			}
		})
	}
}

var (
	validGameID = uuid.MustParse("00000000-0000-0000-0000-000000000003")
	validReqBody = domain.ReservationRequest{
		GameID: validGameID,
		Seats: []domain.SeatRequest{
			{
				SeatGrade: "A",
				Quantity:  2,
			},
		},
	}
	validResponse = &domain.Reservation{
		ID:        uuid.MustParse("00000000-0000-0000-0000-000000000004"),
		GameID:    validGameID,
		Status:   "pending",
		Seats: []domain.SeatRequest{
			{
				SeatGrade: "A",
				Quantity:  2,
			},
		},
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Tickets: []domain.Tickets{
			{
				ID:        uuid.MustParse("00000000-0000-0000-0000-000000000005"),
				SeatGrade: "A",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}
)

func TestHandleCreateReservation(t *testing.T) {
	t.Parallel()
	createTests := []struct {
		name           string
		called	  bool
		userID		*uuid.UUID
		reqBody        domain.ReservationRequest
		rawBody		string
		fakeErr        error
		wantErr        error
		wantStatus int
		wantResponse  *domain.Reservation
	}{
		{
			name:           "success",
			userID: &validUserID,
			called: true,
			reqBody:        validReqBody,
			fakeErr:        nil,
			wantStatus: http.StatusOK,
			wantResponse:  validResponse,
		},
		{
			name:           "unauthorized",
			called: false,
			wantErr: apperror.ErrUnauthorized,
		},
		{
			name:		   "bad request",
			userID: &validUserID,
			rawBody: `{"not-json`,
			called: false,
			wantErr: apperror.ErrBadRequest,
		},
		{
			name:           "internal server error",
			userID: &validUserID,
			reqBody:        validReqBody,
			called: true,
			fakeErr: apperror.ErrDatabase,
			wantErr: apperror.ErrDatabase,
		},
	}
	for _, tt := range createTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var (
				called bool
				gotUserID uuid.UUID
				gotReqBody *domain.ReservationRequest
			)
			var ctx context.Context
			if tt.userID != nil {
				ctx = authbundle.SetUserIDInContext(context.Background(), *tt.userID)
			} else {
				ctx = context.Background()
			}
			h := &Handler{
				reservationService: &StubreservationService{
					FakeCreateReservation: func(_ context.Context, reqBody *domain.ReservationRequest, id uuid.UUID) (*domain.Reservation, error) {
						called = true
						gotReqBody = reqBody
						gotUserID = id
						if tt.fakeErr != nil {
							return nil, tt.fakeErr
						}
						return validResponse, nil
					},
				},
			}
			var bodyReader io.Reader
			var bodyBytes []byte
			if tt.rawBody != "" {
				bodyBytes = []byte(tt.rawBody)
			} else {
				b, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
				bodyBytes = b
			}
			bodyReader = bytes.NewReader(bodyBytes)
			request := httptest.NewRequestWithContext(
				ctx,
				"POST",
				"/api/reservations",
				bodyReader,
			)
			response := httptest.NewRecorder()

			err := h.HandleCreateReservation(response, request)

			// 失敗系
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				if called != tt.called {
					t.Errorf("service called = %v, want %v", called, tt.called)
				}
				return
			}
			if called != tt.called {
				t.Errorf("service called = %v, want %v", called, tt.called)
			}
			// 成功系
			if err != nil {
				t.Errorf("unexpected error = %v", err)
			}
			if response.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
			}
			if tt.called && gotUserID != *tt.userID {
					t.Errorf("service called with user id = %v, want %v", gotUserID, tt.userID)
				}
			if tt.called {
				if diff := cmp.Diff(tt.reqBody, gotReqBody); diff != "" {
					t.Errorf("service called with reqBody mismatch (-want +got):\n%s", diff)
				}
			}
			if ct := response.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", ct)
			}
			var gotResponse domain.Reservation
			if err := json.NewDecoder(response.Body).Decode(&gotResponse); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			opts := cmpopts.IgnoreFields(domain.Reservation{}, "ExpiresAt", "CreatedAt", "UpdatedAt")
			if diff := cmp.Diff(tt.wantResponse, gotResponse, opts); diff != "" {
				t.Errorf("response body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHandleGetUserReservations(t *testing.T) {
	t.Parallel()
	getUserTests := []struct {
		name           string
		userID 	 *uuid.UUID
		serviceErr        error
		wantErr		error
		wantStatus int
		called bool
		wantReservations []*domain.Reservation
	}{
		{
			name:           "success",
			userID: validUserID,
			wantStatus: http.StatusOK,
			called: true,
			wantReservations: validResponse,
		},
		{
			name:           "unauthorized",
			wantErr: apperror.ErrUnauthorized,
			called: false,
		},
		{
			name:           "internal server error",
			userID: validUserID,
			serviceErr:        apperror.ErrDatabase,
			wantErr: apperror.ErrDatabase,
			called: true,
		},
		{
			name:           "not found",
			userID: validUserID,
			serviceErr:		apperror.ErrNotFound,
			wantErr: apperror.ErrNotFound,
			called: true,
		},
	}
	for _, tt := range getUserTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var ctx context.Context
			if tt.userID != nil {
				ctx = authbundle.SetUserIDInContext(context.Background(), *tt.userID)
			} else{
				ctx = context.Background()
			}
			var (
				called bool
				gotUserID uuid.UUID
			)
			h := &Handler{
				reservationService: &StubreservationService{
					FakeGetUserReservations: func(_ context.Context, id uuid.UUID) ([]*domain.Reservation, error) {
						called = true
						gotUserID = id
						if tt.serviceErr != nil {
							return nil, tt.serviceErr
						}
						return tt.wantReservations, nil
					},
				},
			}
			request := httptest.NewRequestWithContext(ctx, "GET", "/api/reservations", nil)
			response := httptest.NewRecorder()

			err := h.HandleGetUserReservations(response, request)

			// 失敗系
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				if called != tt.called {
					t.Errorf("service called = %v, want %v", called, tt.called)
				}
				return
			}
			// 成功系
			if err != nil {
				t.Errorf("unexpected error = %v", err)
			}
			if !called {
				t.Errorf("expected service to be called, but it was not")
			}
			if response.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
			}
			if tt.called && gotUserID != *tt.userID {
					t.Errorf("service called with user id = %v, want %v", gotUserID, *tt.userID)
				}
			var gotReservations []*domain.Reservation
			if err := json.NewDecoder(response.Body).Decode(&gotReservations); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			opts := cmpopts.IgnoreFields(domain.Reservation{}, "ExpiresAt", "CreatedAt", "UpdatedAt")
			if diff := cmp.Diff(tt.wantReservations, gotReservations, opts); diff != "" {
				t.Errorf("response body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHandleGetReservationByID(t *testing.T) {
	t.Parallel()
	reservationID := "f7f7dad8-84fd-4c10-9f95-d2a68d38a46f"
	getReservationTests := []struct {
		name           string
		userID *uuid.UUID
		called bool
		reservationID  string
		serviceErr        error
		wantStatus int
		wantErr error
		wantResponse *domain.Reservation
	}{
		{
			name:           "success",
			userID:         &validUserID,
			reservationID:  validReservationID.String(),
			wantStatus: http.StatusOK,
			called: true,
			wantResponse: validResponse,
		},
		{
			name:           "unauthorized",
			userID:         nil,
			wantStatus: http.StatusUnauthorized,
			called: false,
		},
		{
			name:           "internal server error",
			userID:         &validUserID,
			reservationID:  validReservationID.String(),
			serviceErr:        apperror.ErrDatabase,
			wantErr: apperror.ErrDatabase,
			called: true,
		},
		{
			name:           "not found",
			userID:         &validUserID,
			reservationID:  validReservationID.String(),
			serviceErr:        apperror.ErrNotFound,
			wantErr: apperror.ErrNotFound,
			called: true,
		},
		{
			name:           "bad request",
			userID:         &validUserID,
			reservationID:  "invalid-reservation-uuid",
			serviceErr:        apperror.ErrBadRequest,
			wantErr: apperror.ErrBadRequest,
			called: false,
		},
	}
	for _, tt := range getReservationTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var ctx context.Context
			if tt.userID == nil {
				ctx = context.Background()
			} else {
				ctx = authbundle.SetUserIDInContext(context.Background(), *tt.userID)
			}
			var (
				called bool
				gotUserID uuid.UUID
				gotReservationID uuid.UUID
			)
			h := &Handler{
				reservationService: &StubreservationService{
					FakeGetReservationByID: func(_ context.Context, rID, uID uuid.UUID) (*domain.Reservation, error) {
						called = true
						gotUserID = uID
						gotReservationID = rID
						if tt.serviceErr != nil {
							return nil, tt.serviceErr
						}
						return tt.wantResponse, nil
					},
				},
			}
			request := httptest.NewRequestWithContext(
				ctx,
				"GET",
				"/api/reservations/"+tt.reservationID,
				nil,
			)
			response := httptest.NewRecorder()
			request.SetPathValue("id", tt.reservationID)

			err := h.HandleGetReservationByID(response, request)

			// 失敗系
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				if called != tt.called {
					t.Errorf("service called = %v, want %v", called, tt.called)
				}
				return
			}
			// 成功系
			if err != nil {
				t.Errorf("unexpected error = %v", err)
			}
			if !called {
				t.Errorf("expected service to be called, but it was not")
			}
			if response.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
			}
			if tt.called && gotUserID != *tt.userID {
					t.Errorf("service called with user id = %v, want %v", gotUserID, *tt.userID)
				}
			wantReservationID, _ := uuid.Parse(tt.reservationID)
			if tt.called && gotReservationID != wantReservationID {
				t.Errorf("service called with reservation id = %v, want %v", gotReservationID, wantReservationID)
			}
			var gotResponse domain.Reservation
			if err := json.NewDecoder(response.Body).Decode(&gotResponse); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			opts := cmpopts.IgnoreFields(domain.Reservation{}, "ExpiresAt", "CreatedAt", "UpdatedAt")
			if diff := cmp.Diff(tt.wantResponse, &gotResponse, opts); diff != "" {
				t.Errorf("response body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHandlePurchaseReservation(t *testing.T) {
	t.Parallel()
	purchageTests := []struct {
		name           string
		userID *uuid.UUID
		called bool
		reservationID  string
		serviceErr        error
		wantStatus int
		wantResponse *domain.Reservation
		wantErr error
	}{
		{
			name:           "success",
			userID: 	   &validUserID,
			reservationID:  validReservationID.String(),
			wantStatus: http.StatusOK,
			wantResponse: validResponse,
			called: true,
		},
		{
			name:           "unauthorized",
			wantErr: apperror.ErrUnauthorized,
			called: false,
		},
		{
			name:           "internal server error",
			userID: 	   &validUserID,
			reservationID:  validReservationID.String(),
			serviceErr:        apperror.ErrDatabase,
			wantStatus: http.StatusInternalServerError,
			called: true,
		},
		{
			name:           "not found",
			userID: 	   &validUserID,
			reservationID:  validReservationID.String(),
			serviceErr:        apperror.ErrNotFound,
			wantStatus: http.StatusNotFound,
			called: true,
		},
		{
			name:           "bad request",
			userID: 	   &validUserID,
			reservationID:  "invalid-reservation-uuid",
			serviceErr:        apperror.ErrBadRequest,
			wantErr: apperror.ErrBadRequest,
			called: false,
		},
	}
	for _, tt := range purchageTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var ctx context.Context
			if tt.userID == nil {
				ctx = context.Background()
			} else {
				ctx = authbundle.SetUserIDInContext(context.Background(), *tt.userID)
			}

			var (
				called bool
				gotUserID uuid.UUID
				gotReservationID uuid.UUID
			)
			h := &Handler{
				reservationService: &StubreservationService{
					FakePurchaseReservation: func(_ context.Context, rID, uID uuid.UUID) (*domain.Reservation, error) {
						gotReservationID = rID
						gotUserID = uID
						called = true
						if tt.serviceErr != nil {
							return nil, tt.serviceErr
						}
						return tt.wantResponse, nil
					},
				},
			}
			request := httptest.NewRequestWithContext(
				ctx,
				"POST",
				"/api/reservations/"+tt.reservationID+"/purchase",
				nil,
			)
			response := httptest.NewRecorder()
			request.SetPathValue("id", tt.reservationID)

			err := h.HandlePurchaseReservation(response, request)

			// 失敗系
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				if called != tt.called {
					t.Errorf("service called = %v, want %v", called, tt.called)
				}
				return
			}
			// 成功系
			if err != nil {
				t.Errorf("unexpected error = %v", err)
			}
			if !called {
				t.Errorf("expected service to be called, but it was not")
			}
			if response.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
			}
			if tt.called && gotUserID != *tt.userID {
					t.Errorf("service called with user id = %v, want %v", gotUserID, *tt.userID)
				}
			wantReservationID, _ := uuid.Parse(tt.reservationID)
			if tt.called && gotReservationID != wantReservationID {
				t.Errorf("service called with reservation id = %v, want %v", gotReservationID, wantReservationID)
			}
			var gotResponse domain.Reservation
			if err := json.NewDecoder(response.Body).Decode(&gotResponse); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			opts := cmpopts.IgnoreFields(domain.Reservation{}, "ExpiresAt", "CreatedAt", "UpdatedAt")
			if diff := cmp.Diff(tt.wantResponse, &gotResponse, opts); diff != "" {
				t.Errorf("response body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// func assertStatus(t testing.TB, got, want int) {
// 	t.Helper()
// 	if got != want {
// 		t.Errorf("did not get correct status, got %d, want %d", got, want)
// 	}
// }
