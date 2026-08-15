package service

import (
	"context"
	"time"

	"42tokyo-road-to-dena-server/internal/apperror"
	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type ReservationService interface {
	CreateReservation(
		ctx context.Context,
		reqBody *domain.ReservationRequest,
		userID uuid.UUID,
	) (*domain.Reservation, error)
	GetUserReservations(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error)
	GetReservationByID(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error)
	PurchaseReservation(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error)
	CancelReservation(ctx context.Context, reservationID, userID uuid.UUID) error
	ExpiredReservations(ctx context.Context) error
}

type reservationService struct {
	repo     repository.ReservationRepository
	clock    domain.Clock
	holdTime time.Duration
	maxSeats int
}

type Option func(*reservationService)

func WithClock(clock domain.Clock) Option {
	return func(s *reservationService) {
		s.clock = clock
	}
}

func WithHoldTime(holdTime time.Duration) Option {
	return func(s *reservationService) {
		s.holdTime = holdTime
	}
}

func WithMaxSeats(maxSeats int) Option {
	return func(s *reservationService) {
		if maxSeats > 0 {
			s.maxSeats = maxSeats
		}
	}
}

func NewReservationService(repo repository.ReservationRepository, opts ...Option) ReservationService {
	s := &reservationService{
		repo:     repo,
		clock:    &domain.RealClock{},
		holdTime: 15 * time.Minute,
		maxSeats: 10,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *reservationService) ExpiredReservations(ctx context.Context) error {
	return s.repo.ExpiredReservations(ctx)
}

func (s *reservationService) CancelReservation(ctx context.Context, reservationID, userID uuid.UUID) error {
	return s.repo.CancelReservation(ctx, reservationID, userID)
}

func (s *reservationService) CreateReservation(
	ctx context.Context,
	reqBody *domain.ReservationRequest,
	userID uuid.UUID,
) (*domain.Reservation, error) {
	totalSeats := 0
	for _, seat := range reqBody.Seats {
		if seat.Quantity <= 0 {
			return nil, fmt.Errorf("invalid seat quantity: %w", apperror.ErrValidation)
		}
		totalSeats += seat.Quantity
	}
	if totalSeats > s.maxSeats {
		return nil, fmt.Errorf("exceeded maximum seat limit: %w", apperror.ErrValidation)
	}
	expiresAt := s.clock.Now().Add(s.holdTime)

	return s.repo.CreateReservation(ctx, reqBody, userID, expiresAt)
}

func (s *reservationService) GetUserReservations(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error) {
	reservations, err := s.repo.GetUserReservations(ctx, userID)
	if err != nil {
		return nil, err
	}
	clock := s.clock.Now()
	for _, r := range reservations {
		r.Status = r.EffectiveStatus(clock)
	}
	return reservations, err
}

func (s *reservationService) GetReservationByID(
	ctx context.Context,
	reservationID, userID uuid.UUID,
) (*domain.Reservation, error) {
	return s.repo.GetReservationByID(ctx, reservationID, userID)
}

func (s *reservationService) PurchaseReservation(
	ctx context.Context,
	reservationID, userID uuid.UUID,
) (*domain.Reservation, error) {
	return s.repo.PurchaseReservation(ctx, reservationID, userID)
}
