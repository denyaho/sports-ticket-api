package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"
	"encoding/json"
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

var (
	validID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	homeTeamID = uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	awayTeamID = uuid.MustParse("00000000-0000-0000-0000-0000000000b1")
	wantGame = &domain.Game{
		ID: validID,
		GameDate: "2024-06-01",
		StartTime: "18:00",
		HomeTeam: domain.Team{
			ID:   homeTeamID,
			Name: "Home Team",
		},
		AwayTeam: domain.Team{
			ID:   awayTeamID,
			Name: "Away Team",
		},
	}
)

func TestGetAllGames(t *testing.T) {
	t.Parallel()
	getAllGamesTest := []struct {
		name         string
		games []domain.Game
		wantBody []domain.Game
		serviceErr      error
		wantStatus  int
	}{
		{
			name:         "success",
			games: wantGame,
			wantBody:    wantGame,
			serviceErr:      nil,
			wantStatus:  http.StatusOK,
		},
		{
			name: 	   "empty list",
			games:    []domain.Game{},
			wantBody: []domain.Game{},
			serviceErr: 	nil,
			wantStatus:  http.StatusOK,
		},
		{
			name:         "InternalServerError",
			games:        nil,
			serviceErr:      apperror.ErrDatabase,
			wantStatus:  http.StatusInternalServerError,
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
			h.toHandler(h.HandleGetAllGames)(response, request)

			if response.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if tt.wantBody != nil {
				if ct := response.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
					t.Errorf("Content-Type = %s, want application/json", ct)
				}
				var got []domain.Game
				if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if diff := cmp.Diff(tt.wantBody, got); diff != "" {
					t.Errorf("response body mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

// テスト内容
// 1. 正しいUUIDを渡した場合、ステータスコード200が返ること
// 2. 不正なUUIDを渡した場合、ステータスコード400が返ること
// 3. データベースエラーが発生した場合、ステータスコード500が返ること
// 4. 該当するゲームがない場合、respond bodyにエラーが返ること
// 5. 該当するゲームがない場合、ステータスコード404が返ること
func TestGetGameByID(t *testing.T) {
	t.Parallel()
	
	getGameByIDTest := []struct {
		name         string
		pathID	   string
		game *domain.Game
		serviceErr error
		wantStatus int
		wantCalled bool
		wantBody *domain.Game
	}{
		{
			name:         "success",
			pathID: 	 validID.String(),
			game:        wantGame,
			wantStatus:  http.StatusOK,
			wantCalled:  true,
			wantBody:    wantGame,
		},
		{
			name:         "invalid uuid returns 400 without calling service",
			pathID:       "invalid-uuid",
			wantStatus:  http.StatusBadRequest,
			wantCalled:  false,
		},
		{
			name: "not found",
			pathID: validID.String(),
			serviceErr: apperror.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantCalled: true,
		},
		{
			name:         "database error",
			pathID:       validID.String(),
			serviceErr:   apperror.ErrDatabase,
			wantStatus:   http.StatusInternalServerError,
			wantCalled:   true,
		},
	}
	for _, tt := range getGameByIDTest {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				called bool
				gotID uuid.UUID
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

			h.toHandler(h.HandleGetGameByID)(response, req)
			if response.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if called != tt.wantCalled {
				t.Errorf("service called = %v, want %v", called, tt.wantCalled)
			}
			if tt.wantCalled && gotID != validID {
				t.Errorf("service called with id = %v, want %v", gotID, validID)
			}
			if tt.wantBody != nil {
				if ct := response.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
					t.Errorf("Content-Type = %s, want application/json", ct)
				}
				var got domain.Game
				if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if diff := cmp.Diff(*tt.wantBody, got); diff != "" {
					t.Errorf("response body mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
