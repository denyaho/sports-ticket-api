package domain

import "github.com/google/uuid"

type Games struct {
	ID uuid.UUID
	home_team_id uuid.UUID
	away_team_id uuid.UUID
	game_date time.Time
	start_time time.Time
	venue string

}