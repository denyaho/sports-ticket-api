package service

import(
	"context"
	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/repository"
	"42tokyo-road-to-dena-server/authbundle"
	"fmt"
	"github.com/google/uuid"
)


type UserService interface {
	CreateUser(ctx context.Context, user *domain.User) (uuid.UUID, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	hashedPassword, err := authbundle.HashPassword(user.Password)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to hash password: %w", err)
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to generate user ID: %w", err)
	}
	userToSave := &domain.User{
		ID: id,
		Username: user.Username,
		Email: user.Email,
		Password: hashedPassword,
	}
	return s.repo.CreateUser(ctx, userToSave)
}

func (s *userService) FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repo.FindUserByID(ctx, id)
}


