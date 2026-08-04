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

type StubgameService struct {
	FakeGetAllGames func(ctx context.Context) ([]domain.Game, error)
	FakeGetGameByID func(ctx context.Context, id uuid.UUID) (*domain.Game, error)
}

func (s *StubgameService) GetAllGames(ctx context.Context) ([]domain.Game, error) {
	return s.FakeGetAllGames(ctx)
}

func (s *StubgameService) GetGameByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
	return s.FakeGetGameByID(ctx, id)
}

func TestGetAllGames(t *testing.T) {
	getAllGamesTest := []struct {
		name         string
		setupContext func() context.Context
		fakeErr      error
		expectedErr  int
	}{
		{
			name:         "success",
			setupContext: CreateContext,
			fakeErr:      nil,
			expectedErr:  http.StatusOK,
		},
		{
			name:         "InternalServerError",
			setupContext: CreateContext,
			fakeErr:      apperror.ErrDatabase,
			expectedErr:  http.StatusInternalServerError,
		},
	}
	t.Parallel()

	for _, tt := range getAllGamesTest {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &Handler{
				gameService: &StubgameService{
					FakeGetAllGames: func(_ context.Context) ([]domain.Game, error) {
						return []domain.Game{}, tt.fakeErr
					},
				},
			}
			request := httptest.NewRequestWithContext(tt.setupContext(), "GET", "/api/games", nil)
			response := httptest.NewRecorder()

			h.toHandler(h.HandleGetAllGames)(response, request)
			assertStatus(t, response.Code, tt.expectedErr)
		})
	}
}

func TestGetGameByID(t *testing.T) {
	gameID := "ff79a6c3-50e4-11f1-899d-bc2411066ccc"

	getGameByIDTest := []struct {
		name         string
		setupContext func() context.Context
		gameID       string
		fakeErr      error
		expectedErr  int
	}{
		{
			name:         "success",
			setupContext: CreateContext,
			gameID:       gameID,
			fakeErr:      nil,
			expectedErr:  http.StatusOK,
		},
		{
			name:         "InternalServerError",
			setupContext: CreateContext,
			gameID:       gameID,
			fakeErr:      apperror.ErrDatabase,
			expectedErr:  http.StatusInternalServerError,
		},
		{
			name:         "BadRequest",
			setupContext: CreateContext,
			gameID:       "invalid-uuid",
			fakeErr:      nil,
			expectedErr:  http.StatusBadRequest,
		},
		{
			name:         "NotFound",
			setupContext: CreateContext,
			gameID:       gameID,
			fakeErr:      apperror.ErrNotFound,
			expectedErr:  http.StatusNotFound,
		},
	}
	t.Parallel()
	for _, tt := range getGameByIDTest {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &Handler{
				gameService: &StubgameService{
					FakeGetGameByID: func(_ context.Context, _ uuid.UUID) (*domain.Game, error) {
						return &domain.Game{}, tt.fakeErr
					},
				},
			}
			request := httptest.NewRequestWithContext(tt.setupContext(), "GET", "/api/games/"+tt.gameID, nil)
			request.SetPathValue("id", tt.gameID)
			response := httptest.NewRecorder()

			h.toHandler(h.HandleGetGameByID)(response, request)
			assertStatus(t, response.Code, tt.expectedErr)
		})
	}
}
