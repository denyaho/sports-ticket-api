package service

import (
	"context"
	"time"

	"fmt"

	"42tokyo-road-to-dena-server/internal/apperror"
	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/repository"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
	ctx, span := tracer.Start(ctx, "ReservationService.ExpiredReservations")

	defer span.End()
	if err := s.repo.ExpiredReservations(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

func (s *reservationService) CancelReservation(ctx context.Context, reservationID, userID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "ReservationService.CancelReservation")
	defer span.End()

	if err := s.repo.CancelReservation(ctx, reservationID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

func (s *reservationService) CreateReservation(
	ctx context.Context,
	reqBody *domain.ReservationRequest,
	userID uuid.UUID,
) (*domain.Reservation, error) {
	ctx, span := tracer.Start(ctx, "ReservationService.CreateReservation")
	defer span.End()
	totalSeats := 0
	for _, seat := range reqBody.Seats {
		if seat.Quantity <= 0 {
			var errInvalidSeatQuantity = fmt.Errorf("invalid seat quantity: %w", apperror.ErrValidation)
			span.RecordError(errInvalidSeatQuantity)
			span.SetStatus(codes.Error, errInvalidSeatQuantity.Error())
			return nil, errInvalidSeatQuantity
		}
		totalSeats += seat.Quantity
	}
	if totalSeats > s.maxSeats {
		var errCreateReservationExceeded = fmt.Errorf("exceeded maximum seat limit: %w", apperror.ErrValidation)
		span.RecordError(errCreateReservationExceeded)
		span.SetStatus(codes.Error, errCreateReservationExceeded.Error())
		span.SetAttributes(attribute.Int("ReservationService.total_seats", totalSeats))
		return nil, errCreateReservationExceeded
	}

	expiresAt := s.clock.Now().Add(s.holdTime)

	return s.repo.CreateReservation(ctx, reqBody, userID, expiresAt)
}

func (s *reservationService) GetUserReservations(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error) {
	ctx, span := tracer.Start(ctx, "ReservationService.GetUserReservations")
	defer span.End()
	reservations, err := s.repo.GetUserReservations(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	now := s.clock.Now()
	for _, r := range reservations {
		r.Status = r.EffectiveStatus(now)
	}
	return reservations, err
}

func (s *reservationService) GetReservationByID(
	ctx context.Context,
	reservationID, userID uuid.UUID,
) (*domain.Reservation, error) {
	ctx, span := tracer.Start(ctx, "ReservationService.GetReservationByID")
	defer span.End()
	reservation, err := s.repo.GetReservationByID(ctx, reservationID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return reservation, nil
	// Removed redundant code as it is now handled above
}

func (s *reservationService) PurchaseReservation(
	ctx context.Context,
	reservationID, userID uuid.UUID,
) (*domain.Reservation, error) {
	ctx, span := tracer.Start(ctx, "ReservationService.PurchaseReservation")
	defer span.End()
	reservation, err := s.repo.PurchaseReservation(ctx, reservationID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return reservation, nil
}
