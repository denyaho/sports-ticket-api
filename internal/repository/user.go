package repository

import (
	"fmt"
	"database/sql"
	"github.com/google/uuid"
	"context"
	"42tokyo-road-to-dena-server/internal/domain"
	"github.com/lib/pq"
	"net/http"
	"errors"
)

var ErrDuplicateEmail = errors.New("email already exists")
var ErrUserNotFound = errors.New("user not found")
var ErrDatabase = errors.New("database error")
var ErrUserNotCreated = errors.New("failed to create user")

type UserRepository interface {
	CreateUser(ctx context.Context, *domain.User) (uuid.UUID, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}


type postgreUserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &postgreUserRepository{DB: db}
}

func (r *postgreUserRepository) CreateUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {

	query := "INSERT INTO users (id, username, email, password_hash) VALUES ($1, $2, $3, $4) RETURNING id"

	var id uuid.UUID
	if err := r.DB.QueryRowContext(ctx, query, &user.ID, &user.Username, &user.Email, &user.Password).Scan(&id); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
				case "23505":
					return uuid.Nil, ErrDuplicateEmail
				default:
					return uuid.Nil, ErrDatabase
			}
		}
		return uuid.Nil, ErrUserNotCreated
	}
	return id, nil
}

func (r *postgreUserRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := "SELECT id, username, email, password_hash FROM users WHERE id = $1"

	var user domain.User
	if err := r.DB.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.Email, &user.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, ErrDatabase
	}
	return &user, nil
}