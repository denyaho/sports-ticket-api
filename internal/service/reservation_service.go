package service

import (
	"42tokyo-road-to-dena-server/internal/domain"
	"context"
	"time"
	"github.com/google/uuid"
	"42tokyo-road-to-dena-server/internal/repository"
)

type ReservationService interface {
	CreateReservation(ctx context.Context, reqBody *domain.ReservationRequest, userID uuid.UUID) (*domain.Reservation, error)
	GetUserReservations(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error)
	GetReservationByID(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error)
	PurchaseReservation(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error)

}

type reservationService struct {
	repo repository.ReservationRepository
}

func NewReservationService(repo repository.ReservationRepository) ReservationService {
	return &reservationService{repo: repo}
}

func (s *reservationService) CheckExpiredReservations(ctx context.Context) error {
	return s.repo.CheckExpiredReservations(ctx)
}

func (s *reservationService) CreateReservation(ctx context.Context, reqBody *domain.ReservationRequest, userID uuid.UUID) (*domain.Reservation, error) {

	expires_At := time.Now().Add(15 * time.Minute)

	return s.repo.CreateReservation(ctx, reqBody, userID, expires_At)
}

func (s *reservationService) GetUserReservations(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error) {
	return s.repo.GetUserReservations(ctx, userID)
}

func (s *reservationService) GetReservationByID(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error) {
	return s.repo.GetReservationByID(ctx, reservationID, userID)
}

func (s *reservationService) PurchaseReservation(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error) {
	return s.repo.PurchaseReservation(ctx, reservationID, userID)
}