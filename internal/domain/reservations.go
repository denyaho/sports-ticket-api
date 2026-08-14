package domain

import (
	"time"

	"github.com/google/uuid"
)

type Reservation struct {
	ID        uuid.UUID `json:"id"`
	GameID    uuid.UUID `json:"game_id"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Tickets   []Tickets `json:"tickets"`
}

type SeatInfo struct {
	Grade    string `json:"seat_grade"`
	Quantity int    `json:"quantity"`
}

type ReservationRequest struct {
	GameID uuid.UUID  `json:"game_id"`
	Seats  []SeatInfo `json:"seats"`
}

var (
	ReservationStatusReserved = "reserved"
	ReservationStatusPurchased = "purchased"
	ReservationStatusCanceled  = "canceled"
	ReservationStatusExpired   = "expired"
	ReservationStatusPending   = "pending"
)

func (r *Reservation) EffectiveStatus(now time.Time) string {
	if r.Status == ReservationStatusPending && now.After(r.ExpiresAt) {
		return ReservationStatusExpired
	}
	return r.Status
}

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (rc *RealClock) Now() time.Time {
	return time.Now()
}