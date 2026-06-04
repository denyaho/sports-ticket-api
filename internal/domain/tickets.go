package domain

import (
	"github.com/google/uuid"
)

type Tickets struct {
	ID uuid.UUID `json:"id"`
	Seat uuid.UUID `json:"seat_id"`
	Price int `json:"price"`
	status string `json:"status"`
	createdAt string `json:"created_at"`
	updatedAt string `json:"updated_at"`
}