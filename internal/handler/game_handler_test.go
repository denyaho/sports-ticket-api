package handler

import (
	"testing"
	"context"
	"net/http/httptest"
	"net/http"
	"github.com/google/uuid"
	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/apperror"
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
		name string
		setupContext func() context.Context
		fakeErr error
		expectedErr int
	}{
		{
			name: "success",
			setupContext: createContext,
			fakeErr: nil,
			expectedErr: http.StatusOK,
		},
		{
			name: "InternalServerError",
			setupContext: createContext,
			fakeErr: apperror.ErrDatabase,
			expectedErr: http.StatusInternalServerError,
		},
	}

	for _, tt := range getAllGamesTest {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				gameService: &StubgameService{
					FakeGetAllGames: func(ctx context.Context) ([]domain.Game, error) {
						return []domain.Game{}, tt.fakeErr
				},
			},
		}
		request := httptest.NewRequestWithContext(tt.setupContext(), "GET", "/api/games", nil)
		response := httptest.NewRecorder()

		h.HandleGetAllGames(response, request)
		assertStatus(t, response.Code, tt.expectedErr)
		})
	}
}

func TestGetGameByID(t *testing.T) {

	gameID := "ff79a6c3-50e4-11f1-899d-bc2411066ccc"

	getGameByIDTest := []struct {
		name string
		setupContext func() context.Context
		gameID string
		fakeErr error
		expectedErr int
	}{
		{
			name: "success",
			setupContext: createContext,
			gameID: gameID,
			fakeErr: nil,
			expectedErr: http.StatusOK,
		},
		{
			name: "InternalServerError",
			setupContext: createContext,
			gameID: gameID,
			fakeErr: apperror.ErrDatabase,
			expectedErr: http.StatusInternalServerError,
		},
		{
			name: "BadRequest",
			setupContext: createContext,
			gameID: "invalid-uuid",
			fakeErr: nil,
			expectedErr: http.StatusBadRequest,
		},
		{
			name: "NotFound",
			setupContext: createContext,
			gameID: gameID,
			fakeErr: apperror.ErrNotFound,
			expectedErr: http.StatusNotFound,
		},
	}
	for _, tt := range getGameByIDTest {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				gameService: &StubgameService{
					FakeGetGameByID: func(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
						return &domain.Game{}, tt.fakeErr
					},
				},
			}
			request := httptest.NewRequestWithContext(tt.setupContext(), "GET", "/api/games/"+tt.gameID, nil)
			request.SetPathValue("id", tt.gameID)
			response := httptest.NewRecorder()

			h.HandleGetGameByID(response, request)
			assertStatus(t, response.Code, tt.expectedErr)
		})
	}
}

