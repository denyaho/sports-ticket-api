package service

import (
	"context"
	"fmt"
	"time"

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
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

func (s *reservationService) CancelReservation(ctx context.Context, reservationID, userID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "ReservationService.CancelReservation")
	defer span.End()
	span.SetAttributes(
		attribute.String("reservation.id", reservationID.String()),
		attribute.String("user.id", userID.String()),
	)

	if err := s.repo.CancelReservation(ctx, reservationID, userID); err != nil {
		span.SetStatus(codes.Error, "failed to cancel reservation")
		return fmt.Errorf("failed to cancel reservation %w", err)
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

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("game.id", reqBody.GameID.String()),
		attribute.String("reservation.expires_at", expiresAt.String()),
		attribute.Int("reservation.total_seats", totalSeats),
	)

	reservation, err := s.repo.CreateReservation(ctx, reqBody, userID, expiresAt)
	if err != nil {
		span.SetStatus(codes.Error, "failed to create reservation")
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	return reservation, nil
}

func (s *reservationService) GetUserReservations(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error) {
	ctx, span := tracer.Start(ctx, "ReservationService.GetUserReservations")
	defer span.End()
	reservations, err := s.repo.GetUserReservations(ctx, userID)
	if err != nil {
		span.SetStatus(codes.Error, "failed to get user reservations")
		return nil, err
	}
	now := s.clock.Now()
	for _, r := range reservations {
		r.Status = r.EffectiveStatus(now)
	}
	return reservations, nil
}

func (s *reservationService) GetReservationByID(
	ctx context.Context,
	reservationID, userID uuid.UUID,
) (*domain.Reservation, error) {
	ctx, span := tracer.Start(ctx, "ReservationService.GetReservationByID")
	defer span.End()
	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("reservation.id", reservationID.String()),
	)

	reservation, err := s.repo.GetReservationByID(ctx, reservationID, userID)
	if err != nil {
		span.SetStatus(codes.Error, "failed to get reservation by ID")
		resErr := fmt.Errorf(
			"failed to get reservation by ID(ID=%s) for user(ID=%s): %w",
			reservationID.String(),
			userID.String(),
			err,
		)
		span.RecordError(resErr)
		return nil, resErr
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
	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("reservation.id", reservationID.String()),
	)
	reservation, err := s.repo.PurchaseReservation(ctx, reservationID, userID)
	if err != nil {
		span.SetStatus(codes.Error, "failed to purchase reservation")
		return nil, fmt.Errorf(
			"failed to purchase reservation(ID=%s) by user(ID=%s): %w",
			reservationID.String(),
			userID.String(),
			err,
		)
	}
	return reservation, nil
}
