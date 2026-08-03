package domain

import (
	"time"

	"github.com/google/uuid"
)

type Tickets struct {
	ID        uuid.UUID `json:"id"`
	SeatID    uuid.UUID `json:"seat_id"`
	Price     int       `json:"price"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
