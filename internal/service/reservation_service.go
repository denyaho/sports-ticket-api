package service

import (
	"context"
	"time"

	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/repository"

	"github.com/google/uuid"
)

type ReservationService interface {
	CreateReservation(ctx context.Context, reqBody *domain.ReservationRequest, userID uuid.UUID) (*domain.Reservation, error)
	GetUserReservations(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error)
	GetReservationByID(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error)
	PurchaseReservation(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error)
	CancelReservation(ctx context.Context, reservationID, userID uuid.UUID) error
	ExpiredReservations(ctx context.Context) error
}

type reservationService struct {
	repo repository.ReservationRepository
}

func NewReservationService(repo repository.ReservationRepository) ReservationService {
	return &reservationService{repo: repo}
}

func (s *reservationService) ExpiredReservations(ctx context.Context) error {
	return s.repo.ExpiredReservations(ctx)
}

func (s *reservationService) CancelReservation(ctx context.Context, reservationID, userID uuid.UUID) error {
	return s.repo.CancelReservation(ctx, reservationID, userID)
}

func (s *reservationService) CreateReservation(ctx context.Context, reqBody *domain.ReservationRequest, userID uuid.UUID) (*domain.Reservation, error) {

	expiresAt := time.Now().Add(15 * time.Minute)

	return s.repo.CreateReservation(ctx, reqBody, userID, expiresAt)
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
