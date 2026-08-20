package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	"42tokyo-road-to-dena-server/internal/apperror"
	"42tokyo-road-to-dena-server/internal/domain"

	"github.com/google/uuid"
)

type StubseatsService struct {
	FakeGetSeatsByGameID func(ctx context.Context, gameID uuid.UUID) ([]domain.Seat, error)
}

func (s *StubseatsService) GetSeatsByGameID(ctx context.Context, gameID uuid.UUID) ([]domain.Seat, error) {
	return s.FakeGetSeatsByGameID(ctx, gameID)
}

func newValidSeats() []domain.Seat {
	return []domain.Seat{
		{
			Grade:     "A",
			Price:     1000,
			Total:     100,
			Available: 100,
			Reserved:  0,
			Sold:      0,
		},
	}
}

func TestGetSeatsByGameID(t *testing.T) {
	t.Parallel()

	validGameID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	seatsTests := []struct {
		name       string
		wantGame   string
		serviceErr error
		wantErr    error
		wantStatus int
		wantBody   []domain.Seat
		call       bool
	}{
		{
			name:       "success",
			wantGame:   validGameID.String(),
			wantStatus: http.StatusOK,
			wantBody:   newValidSeats(),
			call:       true,
		},
		{
			name:       "InternalServerError",
			wantGame:   validGameID.String(),
			serviceErr: apperror.ErrInternal,
			wantErr:    apperror.ErrInternal,
			call:       true,
		},
		{
			name:     "BadRequest",
			wantGame: "invalid-uuid",
			wantErr:  apperror.ErrBadRequest,
		},
	}
	for _, tt := range seatsTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var called bool
			h := &Handler{
				seatsService: &StubseatsService{
					FakeGetSeatsByGameID: func(_ context.Context, _ uuid.UUID) ([]domain.Seat, error) {
						called = true
						if tt.serviceErr != nil {
							return nil, tt.serviceErr
						}
						return tt.wantBody, nil
					},
				},
			}
			request := httptest.NewRequestWithContext(
				context.Background(),
				"GET",
				"/api/games/"+tt.wantGame+"/seats",
				nil,
			)
			request.SetPathValue("id", tt.wantGame)
			response := httptest.NewRecorder()

			err := h.HandleGetSeatsByGameID(response, request)

			// 失敗系
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				if called != tt.call {
					t.Errorf("service called = %v, want %v", called, tt.call)
				}
				return
			}
			// 成功系
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if response.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, response.Code)
			}
			if !called {
				t.Errorf("service was not called")
			}
			if tt.call {
				var gotSeats []domain.Seat
				if errDecode := json.NewDecoder(response.Body).Decode(&gotSeats); errDecode != nil {
					t.Fatalf("failed to decode response body: %v", errDecode)
				}
				if diff := cmp.Diff(tt.wantBody, gotSeats); diff != "" {
					t.Errorf("response body mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
