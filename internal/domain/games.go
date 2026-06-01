package domain

import (
	"github.com/google/uuid"
)

type Game struct {
	ID uuid.UUID `json:"id"`
	HomeTeamID uuid.UUID `json:"home_team_id"`
	AwayTeamID uuid.UUID `json:"away_team_id"`
	GameDate string `json:"game_date"`
	StartTime string `json:"start_time"`
}