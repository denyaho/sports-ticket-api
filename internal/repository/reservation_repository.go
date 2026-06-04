package repository

import (
	"context"
	"database/sql"
	"42tokyo-road-to-dena-server/internal/domain"
)

type ReservationRepository interface {
	CreateReservation(ctx context.Context, reqBody *handler.ReservationRequest) ([]domain.Reservation, error)
}

type reservationRepository struct {
	DB *sql.DB
}

func NewReservationRepository(db *sql.DB) ReservationRepository {
	return &reservationRepository{DB: db}
}

func (r *reservationRepository) CreateReservation(ctx context.Context, reqBody *handler.ReservationRequest) ([]domain.Reservation, error) {

	fail := func(err error) ([]domain.Reservation, error) {
		return nil, fmt.Errorf("CreateReservation: %w", err)
	}

	tx, err := r.DB.BeginTx(ctx nil)

	if err != nil {
		return fail(err)
	}

	defer tx.Rollback()

	query := 


}