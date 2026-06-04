package service

import (
	"42go-road-to-dena/internal/handler"
	"42go-road-to-dena/internal/domain"
	"context"
)

type ReservationService interface {
	CreateReservation(ctx context.Context, reqBody *handler.ReservationRequest) ([]domain.Reservation, error)
}

type reservationService struct {
	repo repository.ReservationRepository
}

func NewReservationService(repo repository.ReservationRepository) ReservationService {
	return &reservationService{repo: repo}
}

func (s *reservationService) CreateReservation(ctx context.Context, reqBody *handler.ReservationRequest) ([]domain.Reservation, error) {
	return s.repo.CreateReservation(ctx, reqBody)
}
