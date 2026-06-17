package service

import(
	"context"
	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/repository"
	"42tokyo-road-to-dena-server/internal/apperror"
	"42tokyo-road-to-dena-server/authbundle"
	"fmt"
	"github.com/google/uuid"
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

// パスワード検証
func CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *userService) AuthenticateUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {

	password := user.Password
	userinfo, err := s.repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return uuid.Nil, fmt.Errorf("authentication failed: %w", err)
	}
	if !CheckPassword(password, userinfo.Password) {
		return uuid.Nil, apperror.ErrAuthenticationFailed
	}
	return userinfo.ID, nil
}