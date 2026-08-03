package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"42tokyo-road-to-dena-server/internal/apperror"
	"42tokyo-road-to-dena-server/internal/domain"
	"errors"

	"github.com/google/uuid"
)

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
		return nil, fmt.Errorf("execute query: %w: %w", apperror.ErrDatabase, err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	for rows.Next() {
		var game domain.Game
		err = rows.Scan(
			&game.ID, &game.GameDate, &game.StartTime,
			&game.HomeTeam.ID, &game.HomeTeam.Name,
			&game.AwayTeam.ID, &game.AwayTeam.Name)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w: %w", apperror.ErrDatabase, err)
		}
		games = append(games, game)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, apperror.ErrDatabase
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
		switch {
			case errors.Is(err, sql.ErrNoRows):
				return nil, fmt.Errorf("game not found: %w: %w", apperror.ErrNotFound, err)
			case err != nil:
				return nil, fmt.Errorf("get game %s: %w: %w", id, apperror.ErrDatabase, err)
		}
		log.Printf("Error executing query: %v", err)
		return nil, apperror.ErrDatabase
	}
	return &game, nil
}
