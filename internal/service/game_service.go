package service

import (
	"context"

	"go.opentelemetry.io/otel/codes"

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
	ctx, span := tracer.Start(ctx, "GetAllGames")
	defer span.End()
	game, err := s.repo.GetAllGames(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return game, nil
}

func (s *gameService) GetGameByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
	ctx, span := tracer.Start(ctx, "GetGameByID")
	defer span.End()
	game, err := s.repo.GetGameByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return game, nil
}
