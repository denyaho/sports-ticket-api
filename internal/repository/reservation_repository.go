package repository

import (
	"strings"
	"time"
	"fmt"
	"context"
	"database/sql"
	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/apperror"
	"github.com/google/uuid"
	"log"
)

type ReservationRepository interface {
	CreateReservation(ctx context.Context, reqBody *domain.ReservationRequest, userID uuid.UUID, expiresAt time.Time) (*domain.Reservation, error)
	GetUserReservations(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error)
	GetReservationByID(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error)
	CheckExpiredReservations(ctx context.Context) error
	PurchaseReservation(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error)
	CancelReservation(ctx context.Context, reservationID, userID uuid.UUID) error
}

type reservationRepository struct {
	DB *sql.DB
}

func NewReservationRepository(db *sql.DB) ReservationRepository {
	return &reservationRepository{DB: db}
}

func (r *reservationRepository) CheckExpiredReservations(ctx context.Context) error {

	query := `WITH expired_reservations AS (
	UPDATE reservations SET status = 'expired'
	WHERE status = 'pending' AND expires_at < NOW()
	RETURNING id
	)
	UPDATE tickets SET status = 'available', reservation_id = NULL
	WHERE tickets.reservation_id IN (SELECT id FROM expired_reservations)`

	_, err := r.DB.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error updating expired reservations: %w", err)
	}
	return nil
}

func CheckReservation(ctx context.Context, tx *sql.Tx, reservationID, userID uuid.UUID) (bool, error) {
	query := `SELECT reservations.status, reservations.user_id, reservations.expires_at, tickets.status FROM reservations JOIN tickets ON tickets.reservation_id = reservations.id WHERE reservations.id = $1 FOR UPDATE`

	rows, err := tx.QueryContext(ctx, query, reservationID)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return false, apperror.ErrDatabase
	}
	defer rows.Close()

	var status string
	var ticketStatus string
	var expiresAt time.Time
	var dbUserID uuid.UUID

	for rows.Next() {
		err := rows.Scan(&status, &dbUserID, &expiresAt, &ticketStatus)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return false, apperror.ErrDatabase
		}
		if dbUserID != userID {
			return false, apperror.ErrNotFound
		}
		if status != "pending" {
			return false, apperror.ErrReservationNotPending
		}
		if expiresAt.Before(time.Now()) {
			return false, apperror.ErrReservationExpired
		}
		if ticketStatus != "reserved" {
			return false, apperror.ErrReservationConflict
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return false, apperror.ErrDatabase
	}
	return true, nil
}

func (r *reservationRepository) CancelReservation(ctx context.Context, reservationID, userID uuid.UUID) error {

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return apperror.ErrDatabase
	}
	defer tx.Rollback()
	
	if _, err := CheckReservation(ctx, tx, reservationID, userID); err != nil {
		return err
	}

	query := `WITH c_reservation AS (
	UPDATE reservations SET status = 'cancelled'
	WHERE status = 'pending' AND expires_at > NOW() AND id = $1 AND user_id = $2 RETURNING id)
	UPDATE tickets SET status = 'available', reservation_id = NULL WHERE tickets.reservation_id = c_reservation.id`

	_, err = tx.ExecContext(ctx, query, reservationID, userID)
	if err != nil {
		log.Printf("Error updating reservation and tickets: %v", err)
		return apperror.ErrDatabase
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return apperror.ErrDatabase
	}
	return nil
}




func (r *reservationRepository) PurchaseReservation(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return nil, apperror.ErrDatabase
	}
	defer tx.Rollback()


	if _, err := CheckReservation(ctx, tx, reservationID, userID); err != nil {
		return nil, err
	}

	query := `WITH p_reservation AS (
	UPDATE reservations SET status = 'confirmed'
	WHERE status = 'pending' AND expires_at > NOW() AND id = $1 AND user_id = $2
	RETURNING id, game_id, status, expires_at, created_at, updated_at)
	UPDATE tickets SET status = 'sold'
	WHERE tickets.reservation_id = p_reservation.id
	RETURNING p_reservation.id, p_reservation.game_id, p_reservation.status, p_reservation.expires_at, p_reservation.created_at, p_reservation.updated_at, tickets.id, tickets.seat_id, tickets.price, tickets.status, tickets.created_at, tickets.updated_at`

	rows, err := tx.QueryContext(ctx, query, reservationID, userID)
	if err != nil {
		log.Printf("Error updating reservation and tickets: %v", err)
		return nil, apperror.ErrDatabase
	}
	defer rows.Close()

	reservation := domain.Reservation{}
	for rows.Next() {
		var ticket domain.Tickets
		err := rows.Scan(
			&reservation.ID,
			&reservation.GameID,
			&reservation.Status,
			&reservation.ExpiresAt,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
			&ticket.ID,
			&ticket.SeatID,
			&ticket.Price,
			&ticket.Status,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		reservation.Tickets = append(reservation.Tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	if len(reservation.Tickets) == 0 {
		return nil, apperror.ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return nil, apperror.ErrDatabase
	}

	return &reservation, nil
}

func (r *reservationRepository) GetUserReservations(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error) {

	query := `SELECT reservations.id, reservations.game_id, reservations.status, reservations.expires_at, reservations.created_at, reservations.updated_at,
	tickets.id, tickets.seat_id, tickets.price, tickets.status, tickets.created_at, tickets.updated_at
	FROM reservations
	JOIN tickets ON reservations.id = tickets.reservation_id
	WHERE reservations.user_id = $1
	ORDER BY reservations.id`


	rows ,err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, apperror.ErrDatabase
	}
	defer rows.Close()

	reservations := make(map[uuid.UUID]*domain.Reservation)
	order := make([]uuid.UUID, 0)
	for rows.Next() {
		var reservation domain.Reservation
		var ticket domain.Tickets
		err := rows.Scan(
			&reservation.ID,
			&reservation.GameID,
			&reservation.Status,
			&reservation.ExpiresAt,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
			&ticket.ID,
			&ticket.SeatID,
			&ticket.Price,
			&ticket.Status,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, apperror.ErrDatabase
		}
		resMap, ok := reservations[reservation.ID]
		order = append(order, reservation.ID)
		if !ok {
			if reservation.Status == "pending" && reservation.ExpiresAt.Before(time.Now()) {
				reservation.Status = "expired"
			}
			reservation.Tickets = []domain.Tickets{}
			reservations[reservation.ID] = &reservation
			resMap = &reservation
		}
		resMap.Tickets = append(resMap.Tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, apperror.ErrDatabase
	}
	if len(reservations) == 0 {
		return nil, apperror.ErrNotFound
	}
	reservationList := make([]*domain.Reservation, 0, len(reservations))
	for _, id := range order {
		reservationList = append(reservationList, reservations[id])
	}
	return reservationList, nil
}

func (r *reservationRepository) GetReservationByID(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error) {

	query := `
	SELECT reservations.id, reservations.game_id, reservations.status, reservations.expires_at, reservations.created_at, reservations.updated_at,
	tickets.id, tickets.seat_id, tickets.price, tickets.status, tickets.created_at, tickets.updated_at
	FROM reservations
	JOIN tickets ON reservations.id = tickets.reservation_id
	WHERE reservations.id = $1 AND reservations.user_id = $2
	`

	rows ,err := r.DB.QueryContext(ctx, query, reservationID, userID)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, apperror.ErrDatabase
	}
	defer rows.Close()

	reservation := domain.Reservation{}

	for rows.Next() {
		var ticket domain.Tickets
		err := rows.Scan(
			&reservation.ID,
			&reservation.GameID,
			&reservation.Status,
			&reservation.ExpiresAt,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
			&ticket.ID,
			&ticket.SeatID,
			&ticket.Price,
			&ticket.Status,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, apperror.ErrDatabase
		}
		reservation.Tickets = append(reservation.Tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, apperror.ErrDatabase
	}
	if reservation.ID == uuid.Nil {
		return nil, apperror.ErrNotFound
	}
	if reservation.Status == "pending" && reservation.ExpiresAt.Before(time.Now()) {
		reservation.Status = "expired"
	}

	return &reservation, nil
}

func (r *reservationRepository) CreateReservation(ctx context.Context, reqBody *domain.ReservationRequest, user_ID uuid.UUID, expiresAt time.Time) (*domain.Reservation, error) {
	if err := r.CheckExpiredReservations(ctx); err != nil {
		log.Printf("Error checking expired reservations: %v", err)
		return nil, apperror.ErrDatabase
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return nil, apperror.ErrDatabase
	}
	defer tx.Rollback()

	game_ID := reqBody.GameID
	tickets := []uuid.UUID{}

	for _, seat := range reqBody.Seats {
		ticke_IDs := []uuid.UUID{}
		seatGrade := seat.Grade
		seatQuantity := seat.Quantity

		query := `SELECT tickets.id FROM tickets JOIN seats ON tickets.seat_id = seats.id 
		WHERE (tickets.game_id = $1 AND seats.grade = $2 
		AND (
			tickets.status = 'available'
			OR (
				tickets.status = 'reserved'
				AND EXISTS (
					SELECT 1 FROM reservations
					WHERE reservations.id = tickets.reservation_id
					AND reservations.status = 'pending'
					AND reservations.expires_at < NOW()
				))))LIMIT $3 FOR UPDATE OF tickets`


		rows, err := tx.QueryContext(ctx, query, game_ID, seatGrade, seatQuantity)
		if err != nil {
			log.Printf("Error executing query: %v", err)
			return nil, apperror.ErrDatabase
		}

		var ticket_ID uuid.UUID
		for rows.Next() {
			err := rows.Scan(&ticket_ID)
			if err != nil {
				rows.Close()
				log.Printf("Error scanning row: %v", err)
				return nil, apperror.ErrDatabase
			}
			ticke_IDs = append(ticke_IDs, ticket_ID)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			log.Printf("Error iterating rows: %v", err)
			return nil, apperror.ErrDatabase
		}
		if len(ticke_IDs) < seatQuantity {
			rows.Close()
			log.Printf("Not enough available tickets for seat grade %s", seatGrade)
			return nil, apperror.ErrInsufficientTickets
		}
		tickets = append(tickets, ticke_IDs...)
		rows.Close()
	}

	var reservation_response domain.Reservation

	insertQuery := `INSERT INTO reservations
	(id, user_id, game_id, status, expires_at) VALUES ($1, $2, $3, $4, $5) RETURNING id, game_id, status, expires_at, created_at, updated_at`

	if err := tx.QueryRowContext(ctx, insertQuery, uuid.New(), user_ID, game_ID, "pending", expiresAt).Scan(
		&reservation_response.ID, 
		&reservation_response.GameID, 
		&reservation_response.Status, 
		&reservation_response.ExpiresAt, 
		&reservation_response.CreatedAt, 
		&reservation_response.UpdatedAt,
	); err != nil {
		log.Printf("Error inserting reservation: %v", err)
		return nil, apperror.ErrDatabase
	}

	args := make([]interface{}, len(tickets)+1)
	args[0] = reservation_response.ID
	placeholders := make([]string, len(tickets))

	for i, ticket_ID := range tickets {
		args[i+1] = ticket_ID
		placeholders[i] = fmt.Sprintf("$%d", i+2)
	}

	var tickets_response []domain.Tickets

	updateQuery := fmt.Sprintf(`UPDATE tickets SET reservation_id = $1, status = 'reserved' WHERE id IN (%s) returning id, seat_id, price, status, created_at, updated_at`, strings.Join(placeholders, ", "))

	rows, err := tx.QueryContext(ctx, updateQuery, args...)
	if err != nil {
		log.Printf("Error updating tickets: %v", err)
		return nil, apperror.ErrDatabase
	}
	defer rows.Close()
	for rows.Next() {
		var ticket_info domain.Tickets
		err := rows.Scan(
			&ticket_info.ID,
			&ticket_info.SeatID,
			&ticket_info.Price,
			&ticket_info.Status,
			&ticket_info.CreatedAt,
			&ticket_info.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, apperror.ErrDatabase
		}
		tickets_response = append(tickets_response, ticket_info)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, apperror.ErrDatabase
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return nil, apperror.ErrDatabase
	}
	reservation_response.Tickets = tickets_response

	return &reservation_response, nil

}