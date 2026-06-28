package handler

import (
	"testing"
	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/apperror"
	"context"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"strings"
)

type StubuserService struct {
	FakeCreateUser func(ctx context.Context, user *domain.User) (uuid.UUID, error)
	FakeFindUserByID func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FakeAuthenticateUser func(ctx context.Context, user *domain.User) (uuid.UUID, error)
}

func (s *StubuserService) CreateUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	return s.FakeCreateUser(ctx, user)
}

func (s *StubuserService) FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.FakeFindUserByID(ctx, id)
}

func (s *StubuserService) AuthenticateUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	return s.FakeAuthenticateUser(ctx, user)
}


func TestHandleUserSignup(t *testing.T) {
	successReqBody := `{
		"name": "testuser",
		"email": "testuser@example.com",
		"password": "password123"
	}`
	failReqBody := `"invalid-json"`

	signupTests := []struct {
		name string
		setupContext func() context.Context
		reqBody string
		fakeErr error
		expectedErr int
	}{	
		{
			name: "InternalServerError",
			setupContext: createContext,
			reqBody: successReqBody,
			fakeErr: apperror.ErrInternal,
			expectedErr: http.StatusInternalServerError,
		},
		{
			name: "BadRequest",
			setupContext: createContext,
			reqBody: failReqBody,
			fakeErr: nil,
			expectedErr: http.StatusBadRequest,
		},
		{
			name: "duplicateEmail",
			setupContext: createContext,
			reqBody: successReqBody,
			fakeErr: apperror.ErrDuplicateEmail,
			expectedErr: http.StatusConflict,
		},
		{
			name: "databaseError",
			setupContext: createContext,
			reqBody: successReqBody,
			fakeErr: apperror.ErrDatabase,
			expectedErr: http.StatusInternalServerError,
		},
		{
			name: "User Not Created",
			setupContext: createContext,
			reqBody: successReqBody,
			fakeErr: apperror.ErrUserNotCreated,
			expectedErr: http.StatusInternalServerError,
		},
	}
	for _, tt := range signupTests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				userservice: &StubuserService{
					FakeCreateUser: func(ctx context.Context, user *domain.User) (uuid.UUID, error) {
						return uuid.New(), tt.fakeErr
					},
				},
			}
			request := httptest.NewRequestWithContext(tt.setupContext(), "POST", "/api/user/signup", strings.NewReader(tt.reqBody))
			response := httptest.NewRecorder()

			h.HandleUserSignup(response, request)
			assertStatus(t, response.Code, tt.expectedErr)
		})
	}
}

func TestHandleGetUser(t *testing.T) {

	mockUserInfo := &domain.User{
		ID: uuid.New(),
		Username: "testuser",
		Email: "testuser@example.com",
	}

	FindUserTests := []struct {
		name string
		setupContext func() context.Context
		userInfo *domain.User
		fakeErr error
		expectedErr int
	}{
		{
			name: "success",
			setupContext: createContext,
			userInfo: mockUserInfo,
			fakeErr: nil,
			expectedErr: http.StatusOK,
		},
		{
			name: "Not authorized",
			setupContext: func() context.Context{
				return context.Background()
			},
			userInfo: nil,
			fakeErr: nil,
			expectedErr: http.StatusUnauthorized,
		},
		{
			name: "user not found",
			setupContext: createContext,
			userInfo: nil,
			fakeErr: apperror.ErrUserNotFound,
			expectedErr: http.StatusNotFound,
		},
		{
			name: "database error",
			setupContext: createContext,
			userInfo: nil,
			fakeErr: apperror.ErrDatabase,
			expectedErr: http.StatusInternalServerError,
		},
	}
	for _, tt := range FindUserTests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				userservice: &StubuserService{
					FakeFindUserByID: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
						return tt.userInfo, tt.fakeErr
					},
				},
			}
			request := httptest.NewRequestWithContext(tt.setupContext(), "GET", "/api/user/me", nil)
			response := httptest.NewRecorder()

			h.HandleGetUser(response, request)
			assertStatus(t, response.Code, tt.expectedErr)
		})

	}
}

func TestHandleUserLogin(t *testing.T) {
	reqBody := `{
		"Email" : "testuser@example.com",
		"Password" : "password123"
	}`

	loginTests := []struct {
		name string
		setupContext func() context.Context
		reqBody string
		fakeErr error
		expectedErr int
	}{
		{
			name: "user not found",
			setupContext: createContext,
			reqBody: reqBody,
			fakeErr: apperror.ErrUserNotFound,
			expectedErr: http.StatusNotFound,
		},
		{
			name: "bad request",
			setupContext: createContext,
			reqBody: `"invalid-request"`,
			fakeErr: nil,
			expectedErr: http.StatusBadRequest,
		},
		{
			name: "database error",
			setupContext: createContext,
			reqBody: reqBody,
			fakeErr: apperror.ErrDatabase,
			expectedErr: http.StatusInternalServerError,
		},
		{
			name: "authentication failed",
			setupContext: createContext,
			reqBody: reqBody,
			fakeErr: apperror.ErrAuthenticationFailed,
			expectedErr: http.StatusUnauthorized,
		},
	}
	for _, tt := range loginTests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				userservice: &StubuserService{
					FakeAuthenticateUser: func(ctx context.Context, user *domain.User) (uuid.UUID, error) {
						return uuid.New(), tt.fakeErr
					},
				},
			}
			request := httptest.NewRequestWithContext(tt.setupContext(), "POST", "/api/user/login", strings.NewReader(tt.reqBody))
			response := httptest.NewRecorder()

			h.HandleUserLogin(response, request)
			assertStatus(t, response.Code, tt.expectedErr)
		})
	}
	
}