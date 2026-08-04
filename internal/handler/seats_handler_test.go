package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestGetSeatsByGameID(t *testing.T) {
	gameID := "00000000-0000-0000-0000-000000000001"

	seatsTests := []struct {
		name         string
		setupContext func() context.Context
		gameID       string
		fakeErr      error
		expectedErr  int
	}{
		{
			name:         "success",
			setupContext: createContext,
			gameID:       gameID,
			fakeErr:      nil,
			expectedErr:  http.StatusOK,
		},
		{
			name:         "InternalServerError",
			setupContext: createContext,
			gameID:       gameID,
			fakeErr:      apperror.ErrDatabase,
			expectedErr:  http.StatusInternalServerError,
		},
		{
			name:         "BadRequest",
			setupContext: createContext,
			gameID:       "invalid-uuid",
			fakeErr:      apperror.ErrBadRequest,
			expectedErr:  http.StatusBadRequest,
		},
	}
	t.Parallel()
	for _, tt := range seatsTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &Handler{
				seatsService: &StubseatsService{
					FakeGetSeatsByGameID: func(_ context.Context, _ uuid.UUID) ([]domain.Seat, error) {
						return []domain.Seat{}, tt.fakeErr
					},
				},
			}
			request := httptest.NewRequestWithContext(tt.setupContext(), "GET", "/api/games/"+tt.gameID+"/seats", nil)
			request.SetPathValue("id", tt.gameID)
			response := httptest.NewRecorder()

			h.toHandler(h.HandleGetSeatsByGameID)(response, request)
			assertStatus(t, response.Code, tt.expectedErr)
		})
	}
}
