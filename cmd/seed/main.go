package main

import (
	"database/sql"
	"42tokyo-road-to-dena-server/config"
	"context"
	"fmt"
	"log"
	"strconv"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var teamVenues = map[string]string {
	"Team A": "Stadium A",
	"Team B": "Stadium B",
	"Team C": "Stadium C",
	"Team D": "Stadium D",
}


var teamNames = []string{
	"Team A",
	"Team B",
	"Team C",
	"Team D",
}

func _seedTeams(ctx context.Context, db *sql.DB) error {
	query := "INSERT INTO Teams (id, name) VALUES ($1, $2)"	
	for _, name := range teamNames {
		uuid, err := uuid.NewUUID()
		if err != nil {
			return err
		}
		_, err = db.ExecContext(ctx, query, uuid.String(),name)
		if err != nil {
			return err
		}
	}
	return nil
}

var games = []struct {
    home      string
    away      string
    gameDate  string
    startTime string
    venue     string
}{
    {"Team A", "Team B", "2024-04-01", "2024-04-01 13:00:00", "Stadium A"},
    {"Team C", "Team D", "2024-04-01", "2024-04-01 16:00:00", "Stadium C"},
    {"Team A", "Team C", "2024-04-08", "2024-04-08 13:00:00", "Stadium A"},
    {"Team B", "Team D", "2024-04-08", "2024-04-08 16:00:00", "Stadium B"},
	{"Team D", "Team A", "2024-04-15", "2024-04-15 13:00:00", "Stadium D"},
	{"Team B", "Team C", "2024-04-15", "2024-04-15 16:00:00", "Stadium B"},
}

func getTeamIDByName(ctx context.Context, db *sql.DB, name string) (string, error) {
	query := "SELECT id FROM Teams WHERE name = $1"
	var id string
	if err := db.QueryRowContext(ctx, query, name).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

func _seedGames(ctx context.Context, db *sql.DB) error {
	query := "INSERT INTO Games (id, home_team_id, away_team_id, game_date, start_time, venue) VALUES ($1, $2, $3, $4, $5, $6)"
	for _, game := range games {
		homeTeamID, err := getTeamIDByName(ctx, db, game.home)
		if err != nil {
			return err
		}
		awayTeamID, err := getTeamIDByName(ctx, db, game.away)
		if err != nil {
			return err
		}
		uuid, err := uuid.NewUUID()
		if err != nil {
			return err
		}
		_, err = db.ExecContext(ctx, query, uuid.String(), homeTeamID, awayTeamID, game.gameDate, game.startTime, game.venue)
		if err != nil {
			return err
		}
	}
	return nil
}


var seatGrade = [][2]string{
	{"SS", "10000"},
	{"S", "8000"},
	{"A", "5000"},
	{"B", "3000"},
	{"C", "1000"},
}

func _seedSeats(ctx context.Context, db *sql.DB) error {
	query := "INSERT INTO Seats (id, grade, price) VALUES ($1, $2, $3)"
	seatsPerGrade := 3
	for _, grade := range seatGrade {
		for i := 0; i < seatsPerGrade; i++ {
			uuid, err := uuid.NewUUID()
			if err != nil {
				return err
			}
			price, err := strconv.Atoi(grade[1])
			if err != nil {
				return err
			}
			_ , err = db.ExecContext(ctx, query, uuid.String(), grade[0], price)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getSeatInfo(ctx context.Context, db *sql.DB) (map[string]int, error) {
	query := "SELECT id, price FROM Seats"
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seatInfo := make(map[string]int)

	for rows.Next() {
		var id string
		var price int
		if err := rows.Scan(&id, &price); err != nil {
			return nil, err
		}
		seatInfo[id] = price
	}
	return seatInfo, nil
}

func getGameIDs(ctx context.Context, db *sql.DB) ([]string, error) {
	query := "SELECT id FROM Games"
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gameIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		gameIDs = append(gameIDs, id)
	}
	return gameIDs, nil
}

func _seedTickets(ctx context.Context, db *sql.DB) error {
	query := "INSERT INTO Tickets (id, seat_id, game_id, reservation_id, price, status) VALUES ($1, $2, $3, NULL, $4, 'available')"
	seatInfo, err := getSeatInfo(ctx, db)
	if err != nil {
		return err
	}
	gameIDs, err := getGameIDs(ctx, db)
	if err != nil {
		return err
	}
	for _, gameID := range gameIDs {
		for seatID, price := range seatInfo {
			uuid, err := uuid.NewUUID()
			if err != nil {
				return err
			}
			_, err = db.ExecContext(ctx, query, uuid.String(), seatID, gameID, price)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func _clearTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "TRUNCATE TABLE users, games, seats, teams, tickets, reservations, refresh_tokens CASCADE")
	if err != nil {
		return err
	}
	return nil
}




func main() {
	dbDriver := "postgres"
	cfg, err := config.Load()
	if err != nil {	
		panic(err)
	}
	DBcfg := cfg.Database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",DBcfg.Host, DBcfg.Port, DBcfg.User, DBcfg.Password, DBcfg.Name)
	
	db, err := sql.Open(dbDriver, dsn)
	fmt.Printf("Connecting to database with DSN: %s\n", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = _clearTables(context.Background(), db)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	if err := _seedTeams(ctx, db); err != nil {
		log.Fatal(err)
	}
	if err := _seedGames(ctx, db); err != nil {
		log.Fatal(err)
	}
	if err := _seedSeats(ctx, db); err != nil {
		log.Fatal(err)
	}

	if err := _seedTickets(ctx, db); err != nil {
		log.Fatal(err)
	}
	log.Println("Seeding completed successfully.")
}