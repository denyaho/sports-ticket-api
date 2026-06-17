package domain

import (
	"github.com/google/uuid"
	"time"
)

type Reservation struct {
	ID    uuid.UUID `json:"id"`
	GameID uuid.UUID `json:"game_id"`
	Status string    `json:"status"`
	ExpiresAt time.Time    `json:"expires_at"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Tickets []Tickets `json:"tickets"`
}

type SeatInfo struct {
	Grade string `json:"seat_grade"`
	Quantity int `json:"quantity"`
}

type ReservationRequest struct {
	GameID string `json:"game_id"`
	Seats []SeatInfo `json:"seats"`
}