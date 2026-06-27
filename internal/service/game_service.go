package service

import (
	"context"
	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/repository"
	"github.com/google/uuid"
)

type GameService interface {
	GetAllGames(ctx context.Context) ([]domain.Game, error)
	GetGameByID(ctx context.Context, id uuid.UUID) (*domain.Game, error)
}

type gameService struct {
	repo repository.GameRepository
}

func NewGameService(repo repository.GameRepository) GameService {
	return &gameService{repo: repo}
}

func (s *gameService) GetAllGames(ctx context.Context) ([]domain.Game, error) {
	return s.repo.GetAllGames(ctx)
}

func (s *gameService) GetGameByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
	return s.repo.GetGameByID(ctx, id)
}

