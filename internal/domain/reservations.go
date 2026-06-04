package domain

import (
	"github.com/google/uuid"
)

type Reservation struct {
	ID    uuid.UUID `json:"id"`
	GameID uuid.UUID `json:"game_id"`
	Status string    `json:"status"`
	ExpiresAt int64     `json:"expires_at"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
	Ticket []Tickets `json:"ticket"`
}