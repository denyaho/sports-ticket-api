package repository

import (
	"context"
	"database/sql"
	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/apperror"
	"fmt"
	"errors"
	"github.com/google/uuid"	
)

var ErrNotFound = errors.New("Not found")

type GameRepository interface {
	GetAllGames(ctx context.Context) ([]domain.Game, error)
	GetGameByID(ctx context.Context, id uuid.UUID) (*domain.Game, error)
}

type postgreGamesRepository struct {
	DB *sql.DB
}

func NewGameRepository(db *sql.DB) GameRepository {
	return &postgreGamesRepository{DB: db}
}

func (r *postgreGamesRepository) GetAllGames(ctx context.Context) ([]domain.Game, error) {

	query := `SELECT g.id, g.game_date, g.start_time,
	home.id AS home_team_id, home.name AS home_team_name,
	away.id AS away_team_id, away.name AS away_team_name
	FROM games g
	JOIN teams home ON g.home_team_id = home.id
	JOIN teams away ON g.away_team_id = away.id`

	games := []domain.Game{}

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var game domain.Game
		err := rows.Scan(
			&game.ID, &game.GameDate, &game.StartTime,
			&game.HomeTeam.ID, &game.HomeTeam.Name,
			&game.AwayTeam.ID, &game.AwayTeam.Name)
		if err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		games = append(games, game)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return games, nil
}

func (r *postgreGamesRepository) GetGameByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {

	query := `SELECT g.id, g.game_date, g.start_time,
	home.id AS home_team_id, home.name AS home_team_name,
	away.id AS away_team_id, away.name AS away_team_name
	FROM games g
	JOIN teams home ON g.home_team_id = home.id
	JOIN teams away ON g.away_team_id = away.id
	WHERE g.id = $1`

	var game domain.Game
	if err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&game.ID, &game.GameDate, &game.StartTime,
		&game.HomeTeam.ID, &game.HomeTeam.Name,
		&game.AwayTeam.ID, &game.AwayTeam.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("game with ID %s not found: %w", id, apperror.ErrNotFound)
		}
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	return &game, nil
}