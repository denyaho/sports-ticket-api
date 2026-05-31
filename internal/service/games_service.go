package service

import (
	"context"
)

type GameService interface {
	GetAllGames(ctx context.Context) ([]Game, error)
	GetGameByID(ctx context.Context, id string) (*Game, error)
}

type gameService struct {
	repo repository.GameRepository
}

func NewGameService(repo repository.GameRepository) GameService {
}

