package repository

import (
	"context"
	"database/sql"
	"42tokyo-road-to-dena-server/internal/domain"
	"errors"
)

var ErrNotFound = errors.New("Not found")

type GameRepository interface {
	GetAllGames(ctx context.Context) ([]domain.Game, error)
	GetGameByID(ctx context.Context, id string) (*domain.Game, error)
}

type postgreGamesRepository struct {
	DB *sql.DB
}

func NewGameRepository(db *sql.DB) GameRepository {
	return &postgreGamesRepository{DB: db}
}

func (r *postgreGamesRepository) GetAllGames(ctx context.Context) ([]domain.Game, error) {

	query := "SELECT id, game_date, start_time, home_team_id, away_team_id FROM games"

	games := []domain.Game{}

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var game domain.Game
	for rows.Next() {
		err := rows.Scan(&game.ID, &game.GameDate, &game.StartTime, &game.HomeTeamID, &game.AwayTeamID)
		if err != nil {
			return nil, err
		}
		games = append(games, game)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return games, nil
}

func (r *postgreGamesRepository) GetGameByID(ctx context.Context, id string) (*domain.Game, error) {
	query := "SELECT id, game_date, start_time, home_team_id, away_team_id FROM games WHERE id = $1"

	var game domain.Game
	if err := r.DB.QueryRowContext(ctx, query, id).Scan(&game.ID, &game.GameDate, &game.StartTime, &game.HomeTeamID, &game.AwayTeamID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &game, nil
}