package domain

import (
	"github.com/google/uuid"
)

type Team struct {
	ID uuid.UUID `json:"id"`
	Name string `json:"name"`
}

type Game struct {
	ID uuid.UUID `json:"id"`
	GameDate string `json:"game_date"`
	StartTime string `json:"start_time"`
	HomeTeam Team `json:"home_team"`
	AwayTeam Team `json:"away_team"`

}