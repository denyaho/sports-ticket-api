package repository

import (
	"context"
	"database/sql"
)

type GameRepository interface {
	GetAllGames(ctx context.Context) ([]Game, error)
	GetGameByID(ctx context.Context, id string) (*Game, error)
}

type postgreGamesRepository struct {
	DB *sql.DB
}

func NewGameRepository(db *sql.DB) GameRepository {
	return &postgreGamesRepository{DB: db}
}

func (r *postgreGamesRepository) GetAllGames(ctx context.Context) ([]Game, error) {

	query := "SELECT * FROM Games"

	var games domain.Games
	if err := r.DB.Query

}

func (r *postgreGamesRepository) GetGameByID(ctx context.Context, id string) (*Game, error) {