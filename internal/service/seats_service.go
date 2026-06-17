package service

import (
	"context"
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
	return s.repo.GetSeatsByGameID(ctx, gameID)
}
