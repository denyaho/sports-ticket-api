package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"42tokyo-road-to-dena-server/internal/apperror"
	"42tokyo-road-to-dena-server/internal/domain"

	"github.com/google/uuid"
)

type stubGameService struct {
	GetAllGamesFunc func(ctx context.Context) ([]domain.Game, error)
	GetGameByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.Game, error)
}

func (s *stubGameService) GetAllGames(ctx context.Context) ([]domain.Game, error) {
	return s.GetAllGamesFunc(ctx)
}

func (s *stubGameService) GetGameByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
	return s.GetGameByIDFunc(ctx, id)
}

func newWantGame() *domain.Game {
	return &domain.Game{
		ID:        uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		GameDate:  "2024-06-01",
		StartTime: "18:00",
		HomeTeam: domain.Team{
			ID:   uuid.MustParse("00000000-0000-0000-0000-0000000000a1"),
			Name: "Home Team",
		},
		AwayTeam: domain.Team{
			ID:   uuid.MustParse("00000000-0000-0000-0000-0000000000b1"),
			Name: "Away Team",
		},
	}
}

func TestGetAllGames(t *testing.T) {
	t.Parallel()
	getAllGamesTest := []struct {
		name       string
		games      []domain.Game
		wantBody   []domain.Game
		serviceErr error
		wantErr    error
		wantStatus int
	}{
		{
			name:       "success",
			games:      []domain.Game{*newWantGame()},
			wantBody:   []domain.Game{*newWantGame()},
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty list",
			games:      []domain.Game{},
			wantBody:   []domain.Game{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "InternalServerError",
			serviceErr: apperror.ErrDatabase,
			wantErr:    apperror.ErrDatabase,
		},
	}
	for _, tt := range getAllGamesTest {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &Handler{
				gameService: &stubGameService{
					GetAllGamesFunc: func(_ context.Context) ([]domain.Game, error) {
						if tt.serviceErr != nil {
							return nil, tt.serviceErr
						}
						return tt.games, nil
					},
				},
			}
			request := httptest.NewRequestWithContext(context.Background(), "GET", "/api/games", nil)
			response := httptest.NewRecorder()
			err := h.HandleGetAllGames(response, request)
			// 失敗系
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			// 成功系
			if err != nil {
				t.Errorf("unexpected error = %v", err)
			}
			if response.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if ct := response.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Errorf("Content-Type = %s, want application/json", ct)
			}
			var got []domain.Game
			if errDecode := json.NewDecoder(response.Body).Decode(&got); errDecode != nil {
				t.Fatalf("decode body: %v", errDecode)
			}
			if diff := cmp.Diff(tt.wantBody, got); diff != "" {
				t.Errorf("response body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetGameByID(t *testing.T) {
	t.Parallel()

	validID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	getGameByIDTest := []struct {
		name       string
		pathID     string
		game       *domain.Game
		serviceErr error
		wantErr    error
		wantStatus int
		wantCalled bool
		wantBody   *domain.Game
	}{
		{
			name:       "success",
			pathID:     validID.String(),
			game:       newWantGame(),
			wantStatus: http.StatusOK,
			wantCalled: true,
			wantBody:   newWantGame(),
		},
		{
			name:       "invalid uuid returns 400 without calling service",
			pathID:     "invalid-uuid",
			wantErr:    apperror.ErrBadRequest,
			wantCalled: false,
		},
		{
			name:       "not found",
			pathID:     validID.String(),
			serviceErr: apperror.ErrNotFound,
			wantErr:    apperror.ErrNotFound,
			wantCalled: true,
		},
		{
			name:       "database error",
			pathID:     validID.String(),
			serviceErr: apperror.ErrDatabase,
			wantErr:    apperror.ErrDatabase,
			wantCalled: true,
		},
	}
	for _, tt := range getGameByIDTest {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				called bool
				gotID  uuid.UUID
			)
			h := &Handler{
				gameService: &stubGameService{
					GetGameByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.Game, error) {
						called = true
						gotID = id
						if tt.serviceErr != nil {
							return nil, tt.serviceErr
						}
						return tt.game, nil
					},
				},
			}
			req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/games/"+tt.pathID, nil)
			req.SetPathValue("id", tt.pathID)
			response := httptest.NewRecorder()
			err := h.HandleGetGameByID(response, req)

			// 失敗系
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				if called != tt.wantCalled {
					t.Errorf("service called = %v, want %v", called, tt.wantCalled)
				}
				wantID, _ := uuid.Parse(tt.pathID)
				if tt.wantCalled && gotID != wantID {
					t.Errorf("service called with id = %v, want %v", gotID, wantID)
				}
				return
			}
			// 成功系
			if err != nil {
				t.Errorf("unexpected error = %v", err)
			}
			if response.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if !called {
				t.Errorf("expected service to be called, but it was not")
			}
			wantID, _ := uuid.Parse(tt.pathID)
			if tt.wantCalled && gotID != wantID {
				t.Errorf("service called with id = %v, want %v", gotID, wantID)
			}
			if tt.wantBody != nil {
				if ct := response.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
					t.Errorf("Content-Type = %s, want application/json", ct)
				}
				var got domain.Game
				if errDecode := json.NewDecoder(response.Body).Decode(&got); errDecode != nil {
					t.Fatalf("decode body: %v", errDecode)
				}
				if diff := cmp.Diff(*tt.wantBody, got); diff != "" {
					t.Errorf("response body mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
