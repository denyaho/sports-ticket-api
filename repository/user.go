package repository

import (
	"fmt"
	"database/sql"
	"github.com/google/uuid"
	"context"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, name, email, hashedPassword string) (uuid.UUID, error) {

	query := "INSERT INTO users (id, username, email, password_hash) VALUES ($1, $2, $3, $4) RETURNING id"

	newid, err := uuid.NewUUID()
	if err != nil {
		return uuid.Nil, err
	}

	var id uuid.UUID
	if err = r.DB.QueryRowContext(ctx, query, newid, name, email, hashedPassword).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create user: %w", err)
	}
	return id, nil
}