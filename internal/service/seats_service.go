package service

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/repository"

	"github.com/google/uuid"
)

type SeatsService interface {
	GetSeatsByGameID(ctx context.Context, gameID uuid.UUID) ([]domain.Seat, error)
}

type seatsservice struct {
	repo repository.SeatsRepository
}

func NewSeatsService(repo repository.SeatsRepository) SeatsService {
	return &seatsservice{repo: repo}
}

func (s *seatsservice) GetSeatsByGameID(ctx context.Context, gameID uuid.UUID) ([]domain.Seat, error) {
	ctx, span := tracer.Start(ctx, "GetSeatsByGameID")
	span.SetAttributes(
		attribute.String("game.id", gameID.String()),
	)
	defer span.End()
	seats, err := s.repo.GetSeatsByGameID(ctx, gameID)
	if err != nil {
		span.SetStatus(codes.Error, "failed to get seats by game")
		return nil, fmt.Errorf("failed to get seats by game (ID: %s): %w", gameID.String(), err)
	}
	return seats, nil
}
