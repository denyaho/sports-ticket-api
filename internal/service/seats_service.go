package service

import (
	"context"

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
	defer span.End()
	seats, err := s.repo.GetSeatsByGameID(ctx, gameID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return seats, nil
}
