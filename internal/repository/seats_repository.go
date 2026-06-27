package repository

import (
	"context"
	"database/sql"
	"log"
	"github.com/google/uuid"
	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/apperror"
)

type SeatsRepository interface {
	GetSeatsByGameID(ctx context.Context, gameID uuid.UUID) ([]domain.Seat, error)
}

type postgreSeatsRepository struct {
	DB *sql.DB
}

func NewSeatsRepository(db *sql.DB) SeatsRepository {
	return &postgreSeatsRepository{DB: db}
}

func (r *postgreSeatsRepository) GetSeatsByGameID(ctx context.Context, gameID uuid.UUID) ([]domain.Seat, error) {

	query := `SELECT 
	seats.grade, seats.price,
	COUNT(*) AS total_seats,
	COUNT(*) FILTER (WHERE status = 'available') AS available_seats,
	COUNT(*) FILTER (WHERE status = 'reserved') AS reserved_seats,
	COUNT(*) FILTER (WHERE status = 'sold') AS sold_seats
	FROM tickets
	JOIN seats ON tickets.seat_id = seats.id
	WHERE game_id = $1
	GROUP BY seats.grade, seats.price
	`

	var seats []domain.Seat

	rows, err := r.DB.QueryContext(ctx, query, gameID)
	if err != nil {
		log.Printf("Error querying seats by game ID: %v", err)
		return nil, apperror.ErrDatabase

	}
	defer rows.Close()

	for rows.Next() {
		var seat domain.Seat
		err := rows.Scan(&seat.Grade, &seat.Price, &seat.Total, &seat.Available, &seat.Reserved, &seat.Sold)
		if err != nil {
			log.Printf("Error scanning seat: %v", err)
			return nil, apperror.ErrDatabase
		}
		seats = append(seats, seat)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, apperror.ErrDatabase
	}
	return seats, nil
}