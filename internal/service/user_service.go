package service

import (
	"context"
	"fmt"

	"42tokyo-road-to-dena-server/authbundle"
	"42tokyo-road-to-dena-server/internal/apperror"
	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/repository"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(ctx context.Context, user *domain.User) (uuid.UUID, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	AuthenticateUser(ctx context.Context, user *domain.User) (uuid.UUID, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "UserService.CreateUser")
	defer span.End()
	hashedPassword, err := authbundle.HashPassword(user.Password)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return uuid.Nil, fmt.Errorf("failed to hash password: %w", err)
	}

	id, err := uuid.NewUUID()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return uuid.Nil, fmt.Errorf("failed to generate user ID: %w", err)
	}
	userToSave := &domain.User{
		ID:       id,
		Username: user.Username,
		Email:    user.Email,
		Password: hashedPassword,
	}
	span.SetAttributes(
		attribute.String("user.id", id.String()),
		attribute.String("user.username", user.Username),
		attribute.String("user.email", user.Email),
	)
	id, err = s.repo.CreateUser(ctx, userToSave)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create user")
		return uuid.Nil, fmt.Errorf("failed to create user: %w", err)
	}
	return id, nil
}

func (s *userService) FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	ctx, span := tracer.Start(ctx, "UserService.FindUserByID")
	span.SetAttributes(
		attribute.String("user.id", id.String()),
	)
	defer span.End()
	user, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, "failed to find user by ID")
		return nil, fmt.Errorf("failed to find user by ID (id: %s): %w", id.String(), err)
	}
	return user, nil
}

// CheckPassword compares a plaintext password with a hashed password and returns true if they match.
func CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *userService) AuthenticateUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "UserService.AuthenticateUser")
	defer span.End()
	password := user.Password
	userinfo, err := s.repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		span.SetStatus(codes.Error, "authentication failed")
		return uuid.Nil, fmt.Errorf("authentication failed: %w", err)
	}
	if !CheckPassword(password, userinfo.Password) {
		span.RecordError(apperror.ErrUnauthorized)
		span.SetStatus(codes.Error, apperror.ErrUnauthorized.Error())
		return uuid.Nil, apperror.ErrUnauthorized
	}
	return userinfo.ID, nil
}
