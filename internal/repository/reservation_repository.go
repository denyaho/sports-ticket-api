package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"42tokyo-road-to-dena-server/internal/apperror"
	"42tokyo-road-to-dena-server/internal/domain"

	"github.com/google/uuid"
)

type ReservationRepository interface {
	CreateReservation(ctx context.Context, reqBody *domain.ReservationRequest, userID uuid.UUID, expiresAt time.Time) (*domain.Reservation, error)
	GetUserReservations(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error)
	GetReservationByID(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error)
	ExpiredReservations(ctx context.Context) error
	PurchaseReservation(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error)
	CancelReservation(ctx context.Context, reservationID, userID uuid.UUID) error
}

type reservationRepository struct {
	DB *sql.DB
}

func NewReservationRepository(db *sql.DB) ReservationRepository {
	return &reservationRepository{DB: db}
}

func (r *reservationRepository) ExpiredReservations(ctx context.Context) error {
	query := `WITH expired_reservations AS (
	UPDATE reservations SET status = 'expired'
	WHERE status = 'pending' AND expires_at < NOW()
	RETURNING id
	)
	UPDATE tickets SET status = 'available', reservation_id = NULL
	WHERE tickets.reservation_id IN (SELECT id FROM expired_reservations)`

	result, err := r.DB.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error updating expired reservations: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting affected rows: %w", err)
	}
	log.Printf("Expired reservations updated, rows affected: %d", rows)
	return nil
}

// func CheckReservation(ctx context.Context, tx *sql.Tx, reservationID, userID uuid.UUID) (bool, error) {
// 	query := `SELECT reservations.status, reservations.user_id, reservations.expires_at, tickets.status FROM reservations JOIN tickets ON tickets.reservation_id = reservations.id WHERE reservations.id = $1 FOR UPDATE`

// 	rows, err := tx.QueryContext(ctx, query, reservationID)
// 	if err != nil {
// 		log.Printf("Error executing query: %v", err)
// 		return false, apperror.ErrDatabase
// 	}
// 	defer rows.Close()

// 	var status string
// 	var ticketStatus string
// 	var expiresAt time.Time
// 	var dbUserID uuid.UUID

// 	for rows.Next() {
// 		err := rows.Scan(&status, &dbUserID, &expiresAt, &ticketStatus)
// 		if err != nil {
// 			log.Printf("Error scanning row: %v", err)
// 			return false, apperror.ErrDatabase
// 		}
// 		if dbUserID != userID {
// 			return false, apperror.ErrNotFound
// 		}
// 	}
// 	if err := rows.Err(); err != nil {
// 		log.Printf("Error iterating rows: %v", err)
// 		return false, apperror.ErrDatabase
// 	}
// 	return true, nil
// }

func (r *reservationRepository) CancelReservation(ctx context.Context, reservationID, userID uuid.UUID) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w: %w", apperror.ErrDatabase, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	query := `WITH c_reservation AS (
			UPDATE reservations SET status = 'cancelled'
			WHERE id = $1 AND user_id = $2
			AND (
				(status = 'pending' AND expires_at > NOW()) 
				OR (status = 'confirmed')
			) 
			RETURNING id
	)
	UPDATE tickets SET status = 'available', reservation_id = NULL 
	FROM c_reservation
	WHERE tickets.reservation_id = c_reservation.id`

	result, err := tx.ExecContext(ctx, query, reservationID, userID)
	if err != nil {
		return fmt.Errorf("error updating reservation and tickets: %w: %w", apperror.ErrDatabase, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting affected rows: %v", err)
		return apperror.ErrDatabase
	}
	if affected == 0 {
		return apperror.ErrNotFound
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w: %w", apperror.ErrDatabase, err)
	}
	return nil
}

func (r *reservationRepository) PurchaseReservation(ctx context.Context, reservationID, userID uuid.UUID) (*domain.Reservation, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w: %w", apperror.ErrDatabase, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `WITH p_reservation AS (
			UPDATE reservations SET status = 'confirmed'
			WHERE status = 'pending' AND expires_at > NOW() AND id = $1 AND user_id = $2
			RETURNING id, game_id, status, expires_at, created_at, updated_at
	)
	UPDATE tickets SET status = 'sold'
	FROM p_reservation
	WHERE tickets.reservation_id = p_reservation.id
	RETURNING p_reservation.id, p_reservation.game_id, p_reservation.status, p_reservation.expires_at, p_reservation.created_at, p_reservation.updated_at, tickets.id, tickets.seat_id, tickets.price, tickets.status, tickets.created_at, tickets.updated_at`

	rows, err := tx.QueryContext(ctx, query, reservationID, userID)
	if err != nil {
		return nil, fmt.Errorf("error updating reservation and tickets: %w: %w", apperror.ErrDatabase, err)
	}

	defer func() {
		_ = rows.Close()
	}()

	reservation := domain.Reservation{}
	for rows.Next() {
		var ticket domain.Tickets
		err = rows.Scan(
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
			return nil, fmt.Errorf("error scanning row: %w: %w", apperror.ErrDatabase, err)
		}
		reservation.Tickets = append(reservation.Tickets, ticket)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w: %w", apperror.ErrDatabase, err)
	}
	if len(reservation.Tickets) == 0 {
		return nil, apperror.ErrNotFound
	}

	_ = rows.Close()
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w: %w", apperror.ErrDatabase, err)
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

	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w: %w", apperror.ErrDatabase, err)
	}

	defer func() {
		_ = rows.Close()
	}()

	reservations := make(map[uuid.UUID]*domain.Reservation)
	order := make([]uuid.UUID, 0)
	for rows.Next() {
		var reservation domain.Reservation
		var ticket domain.Tickets
		err = rows.Scan(
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
			return nil, fmt.Errorf("error scanning row: %w: %w", apperror.ErrDatabase, err)
		}
		resMap, ok := reservations[reservation.ID]
		if !ok {
			if reservation.Status == "pending" && reservation.ExpiresAt.Before(time.Now()) {
				reservation.Status = "expired"
			}
			reservation.Tickets = []domain.Tickets{}
			reservations[reservation.ID] = &reservation
			resMap = &reservation
			order = append(order, reservation.ID)
		}
		resMap.Tickets = append(resMap.Tickets, ticket)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w: %w", apperror.ErrDatabase, err)
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

	rows, err := r.DB.QueryContext(ctx, query, reservationID, userID)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w: %w", apperror.ErrDatabase, err)
	}

	defer func() {
		_ = rows.Close()
	}()

	reservation := domain.Reservation{}

	for rows.Next() {
		var ticket domain.Tickets
		err = rows.Scan(
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
			return nil, fmt.Errorf("error scanning row: %w: %w", apperror.ErrDatabase, err)
		}
		reservation.Tickets = append(reservation.Tickets, ticket)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w: %w", apperror.ErrDatabase, err)
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
	if err := r.ExpiredReservations(ctx); err != nil {
		log.Printf("Error checking expired reservations: %v", err)
		return nil, apperror.ErrDatabase
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return nil, apperror.ErrDatabase
	}

	defer func() {
		_ = tx.Rollback()
	}()

	gameID := reqBody.GameID
	tickets := []uuid.UUID{}

	for _, seat := range reqBody.Seats {
		ticketIDs := []uuid.UUID{}
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
		// skip locked でテストしてみる
		var rows *sql.Rows
		rows, err = tx.QueryContext(ctx, query, gameID, seatGrade, seatQuantity)
		if err != nil {
			return nil, fmt.Errorf("error executing query: %w: %w", apperror.ErrDatabase, err)
		}

		defer func() {
			_ = rows.Close()
		}()

		var ticketID uuid.UUID
		for rows.Next() {
			err = rows.Scan(&ticketID)
			if err != nil {
				return nil, fmt.Errorf("error scanning row: %w: %w", apperror.ErrDatabase, err)
			}
			ticketIDs = append(ticketIDs, ticketID)
		}
		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating rows: %w: %w", apperror.ErrDatabase, err)
		}
		if len(ticketIDs) < seatQuantity {
			return nil, fmt.Errorf("not enough available tickets for seat grade %s: %w", seatGrade, apperror.ErrInsufficientTickets)
		}
		tickets = append(tickets, ticketIDs...)
	}

	var reservationResponse domain.Reservation

	insertQuery := `INSERT INTO reservations
	(id, user_id, game_id, status, expires_at) VALUES ($1, $2, $3, $4, $5) RETURNING id, game_id, status, expires_at, created_at, updated_at`

	if err = tx.QueryRowContext(ctx, insertQuery, uuid.New(), user_ID, gameID, "pending", expiresAt).Scan(
		&reservationResponse.ID,
		&reservationResponse.GameID,
		&reservationResponse.Status,
		&reservationResponse.ExpiresAt,
		&reservationResponse.CreatedAt,
		&reservationResponse.UpdatedAt,
	); err != nil {
		log.Printf("Error inserting reservation: %v", err)
		return nil, apperror.ErrDatabase
	}

	args := make([]any, len(tickets)+1)
	args[0] = reservationResponse.ID
	placeholders := make([]string, len(tickets))

	for i, ticketID := range tickets {
		args[i+1] = ticketID
		placeholders[i] = fmt.Sprintf("$%d", i+2)
	}

	var ticketsResponse []domain.Tickets

	updateQuery := fmt.Sprintf(`UPDATE tickets SET reservation_id = $1, status = 'reserved' WHERE id IN (%s) returning id, seat_id, price, status, created_at, updated_at`, strings.Join(placeholders, ", "))

	rows, err := tx.QueryContext(ctx, updateQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("error updating tickets: %w: %w", apperror.ErrDatabase, err)
	}

	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var ticketInfo domain.Tickets
		err = rows.Scan(
			&ticketInfo.ID,
			&ticketInfo.SeatID,
			&ticketInfo.Price,
			&ticketInfo.Status,
			&ticketInfo.CreatedAt,
			&ticketInfo.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w: %w", apperror.ErrDatabase, err)
		}
		ticketsResponse = append(ticketsResponse, ticketInfo)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w: %w", apperror.ErrDatabase, err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w: %w", apperror.ErrDatabase, err)
	}
	reservationResponse.Tickets = ticketsResponse

	return &reservationResponse, nil
}