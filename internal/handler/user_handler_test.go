package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	"42tokyo-road-to-dena-server/authbundle"
	"42tokyo-road-to-dena-server/internal/apperror"
	"42tokyo-road-to-dena-server/internal/domain"

	"github.com/google/uuid"
)

type StubUserService struct {
	FakeCreateUser       func(ctx context.Context, user *domain.User) (uuid.UUID, error)
	FakeFindUserByID     func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FakeAuthenticateUser func(ctx context.Context, user *domain.User) (uuid.UUID, error)
}

func (s *StubUserService) CreateUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	return s.FakeCreateUser(ctx, user)
}

func (s *StubUserService) FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.FakeFindUserByID(ctx, id)
}

func (s *StubUserService) AuthenticateUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	return s.FakeAuthenticateUser(ctx, user)
}

type StubAuthbundleService struct {
	FakeGenerateAccessToken  func(userID uuid.UUID) (string, error)
	FakeGenerateRefreshToken func(ctx context.Context, userID uuid.UUID) (string, error)
	FakeValidateRefreshToken func(ctx context.Context, token string) (*authbundle.RefreshToken, error)
	FakeValidateAccessToken  func(token string) (*authbundle.AuthClaims, error)
	FakeRotateRefreshToken   func(ctx context.Context, oldToken string) (string, error)
}

func (sa *StubAuthbundleService) GenerateAccessToken(userID uuid.UUID) (string, error) {
	return sa.FakeGenerateAccessToken(userID)
}

func (sa *StubAuthbundleService) GenerateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	return sa.FakeGenerateRefreshToken(ctx, userID)
}

func (sa *StubAuthbundleService) ValidateRefreshToken(
	ctx context.Context,
	token string,
) (*authbundle.RefreshToken, error) {
	return sa.FakeValidateRefreshToken(ctx, token)
}

func (sa *StubAuthbundleService) ValidateAccessToken(token string) (*authbundle.AuthClaims, error) {
	return sa.FakeValidateAccessToken(token)
}

func (sa *StubAuthbundleService) RotateRefreshToken(ctx context.Context, oldToken string) (string, error) {
	return sa.FakeRotateRefreshToken(ctx, oldToken)
}

func newValidRequestBody() SignupRequest {
	return SignupRequest{
		Name:     "testuser",
		Email:    "testuser@example.com",
		Password: "password123",
	}
}

func TestHandleUserSignup(t *testing.T) {
	t.Parallel()

	validUserID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	validAccessToken := "validAccessToken"
	validRefreshToken := "validRefreshToken"
	signupTests := []struct {
		name             string
		reqBody          SignupRequest
		serviceErr       error
		wantErr          error
		wantStatus       int
		rawBody          string
		wantUserID       uuid.UUID
		wantAccessToken  string
		wantRefreshToken string
		wantResponseBody AuthResponse
		call             bool
		callAccessToken  bool
		callRefreshToken bool
	}{
		{
			name:       "InternalServerError",
			reqBody:    newValidRequestBody(),
			serviceErr: apperror.ErrInternal,
			wantErr:    apperror.ErrInternal,
			call:       true,
		},
		{
			name:    "BadRequest",
			rawBody: `{"not-json`,
			wantErr: apperror.ErrBadRequest,
			call:    false,
		},
		{
			name:             "success",
			reqBody:          newValidRequestBody(),
			wantStatus:       http.StatusOK,
			call:             true,
			callAccessToken:  true,
			callRefreshToken: true,
			wantUserID:       validUserID,
			wantAccessToken:  validAccessToken,
			wantRefreshToken: validRefreshToken,
			wantResponseBody: AuthResponse{
				UserID:       validUserID.String(),
				AccessToken:  validAccessToken,
				RefreshToken: validRefreshToken,
			},
		},
		{
			name:            "failed with invalid accessToken",
			reqBody:         newValidRequestBody(),
			serviceErr:      apperror.ErrInternal,
			wantErr:         apperror.ErrInternal,
			call:            true,
			callAccessToken: true,
			wantUserID:      validUserID,
		},
		{
			name:             "failed with invalid refreshToken",
			reqBody:          newValidRequestBody(),
			serviceErr:       apperror.ErrInternal,
			wantErr:          apperror.ErrInternal,
			call:             true,
			callAccessToken:  true,
			callRefreshToken: true,
			wantAccessToken:  validAccessToken,
			wantUserID:       validUserID,
		},
	}
	for _, tt := range signupTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			var (
				called             bool
				calledAccessToken  bool
				calledRefreshToken bool
				gotUserInfo        *domain.User
			)
			h := &Handler{
				userService: &StubUserService{
					FakeCreateUser: func(_ context.Context, user *domain.User) (uuid.UUID, error) {
						called = true
						gotUserInfo = user
						if tt.serviceErr != nil && !tt.callAccessToken {
							return uuid.Nil, tt.serviceErr
						}
						return tt.wantUserID, nil
					},
				},
				authBundleService: &StubAuthbundleService{
					FakeGenerateAccessToken: func(_ uuid.UUID) (string, error) {
						calledAccessToken = true
						if tt.serviceErr != nil && !tt.callRefreshToken {
							return "", tt.serviceErr
						}
						return tt.wantAccessToken, nil
					},
					FakeGenerateRefreshToken: func(_ context.Context, _ uuid.UUID) (string, error) {
						calledRefreshToken = true
						if tt.serviceErr != nil {
							return "", tt.serviceErr
						}
						return tt.wantRefreshToken, nil
					},
				},
				authConfig: &authbundle.AuthConfig{},
				logger:     slog.New(slog.DiscardHandler),
			}
			var bodyReader io.Reader
			var bodyBytes []byte
			if tt.rawBody != "" {
				bodyBytes = []byte(tt.rawBody)
			} else {
				b, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
				bodyBytes = b
			}
			bodyReader = bytes.NewReader(bodyBytes)
			request := httptest.NewRequestWithContext(
				ctx,
				"POST",
				"/api/user/signup",
				bodyReader,
			)
			response := httptest.NewRecorder()

			err := h.HandleUserSignup(response, request)
			if tt.call {
				wantUserInfo := &domain.User{
					Username: tt.reqBody.Name,
					Email:    tt.reqBody.Email,
					Password: tt.reqBody.Password,
				}
				if diff := cmp.Diff(wantUserInfo, gotUserInfo); diff != "" {
					t.Errorf("user info passed to service differs: %s", diff)
				}
			}
			// 失敗系
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				if tt.call != called {
					t.Errorf("service called = %v, want %v", called, tt.call)
				}
				if tt.callAccessToken != calledAccessToken {
					t.Errorf("access token service called = %v, want %v", calledAccessToken, tt.callAccessToken)
				}
				if tt.callRefreshToken != calledRefreshToken {
					t.Errorf("refresh token service called = %v, want %v", calledRefreshToken, tt.callRefreshToken)
				}
				return
			}
			// 成功系
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if response.Code != tt.wantStatus {
				t.Errorf("status code = %v, want %v", response.Code, tt.wantStatus)
			}
			if !called {
				t.Errorf("service was not called")
			}
			if !calledAccessToken {
				t.Errorf("access token service was not called")
			}
			if !calledRefreshToken {
				t.Errorf("refresh token service was not called")
			}
			var gotResponse AuthResponse
			if errDecode := json.NewDecoder(response.Body).Decode(&gotResponse); errDecode != nil {
				t.Fatalf("failed to decode response body: %v", errDecode)
			}
			if gotResponse != tt.wantResponseBody {
				t.Errorf("response body = %v, want %v", gotResponse, tt.wantResponseBody)
			}
		})
	}
}

func newValidStubReturn() *domain.User {
	validUserID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	return &domain.User{
		ID:       validUserID,
		Username: "testuser",
		Email:    "testuser@example.com",
		Password: "hashedpassword",
	}
}

func newValidUserInfo() *UserResponse {
	validUserID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	return &UserResponse{
		UserID:   validUserID.String(),
		Username: "testuser",
		Email:    "testuser@example.com",
	}
}

func TestHandleGetUser(t *testing.T) {
	t.Parallel()

	validUserID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	findUserTests := []struct {
		name       string
		wantUserID *uuid.UUID
		call       bool
		serviceErr error
		wantErr    error
		wantStatus int
		wantBody   *UserResponse
		stubReturn *domain.User
	}{
		{
			name:       "unauthorized",
			wantUserID: nil,
			wantErr:    apperror.ErrUnauthorized,
			call:       false,
		},
		{
			name:       "success",
			wantUserID: &validUserID,
			call:       true,
			stubReturn: newValidStubReturn(),
			wantBody:   newValidUserInfo(),
			serviceErr: nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "internal server error",
			wantUserID: &validUserID,
			call:       true,
			serviceErr: apperror.ErrDatabase,
			wantErr:    apperror.ErrDatabase,
		},
	}
	for _, tt := range findUserTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var ctx context.Context
			if tt.wantUserID != nil {
				ctx = authbundle.SetUserIDInContext(context.Background(), *tt.wantUserID)
			} else {
				ctx = context.Background()
			}
			var (
				called bool
			)
			h := &Handler{
				userService: &StubUserService{
					FakeFindUserByID: func(_ context.Context, _ uuid.UUID) (*domain.User, error) {
						called = true
						if tt.serviceErr != nil {
							return nil, tt.serviceErr
						}
						return tt.stubReturn, nil
					},
				},
			}
			request := httptest.NewRequestWithContext(ctx, "GET", "/api/user/me", nil)
			response := httptest.NewRecorder()

			err := h.HandleGetUser(response, request)

			// 失敗系
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				if called != tt.call {
					t.Errorf("service called = %v, want %v", called, tt.call)
				}
				return
			}
			// 成功系
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if response.Code != tt.wantStatus {
				t.Errorf("status code = %v, want %v", response.Code, tt.wantStatus)
			}
			if !called {
				t.Errorf("service was not called")
			}
			var gotUser UserResponse
			if errDecode := json.NewDecoder(response.Body).Decode(&gotUser); errDecode != nil {
				t.Fatalf("failed to decode response body: %v", errDecode)
			}
			if gotUser != *tt.wantBody {
				t.Errorf("response body = %v, want %v", gotUser, *tt.wantBody)
			}
		})
	}
}

func TestHandleUserLogin(t *testing.T) {
	t.Parallel()

	validUserID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	validAccessToken := "validAccessToken"
	validRefreshToken := "validRefreshToken"

	loginTests := []struct {
		name             string
		reqBody          SignupRequest
		serviceErr       error
		wantErr          error
		wantStatus       int
		wantUserID       uuid.UUID
		wantAccessToken  string
		wantRefreshToken string
		wantResponseBody AuthResponse
		call             bool
		rawBody          string
		callAccessToken  bool
		callRefreshToken bool
	}{
		{
			name:    "bad request",
			rawBody: `{"not-json`,
			wantErr: apperror.ErrBadRequest,
			call:    false,
		},
		{
			name:       "authentication failed",
			reqBody:    newValidRequestBody(),
			serviceErr: apperror.ErrUnauthorized,
			wantErr:    apperror.ErrUnauthorized,
			call:       true,
		},
		{
			name:       "internal server error",
			reqBody:    newValidRequestBody(),
			wantErr:    apperror.ErrInternal,
			serviceErr: apperror.ErrInternal,
			call:       true,
		},
		{
			name:             "success",
			reqBody:          newValidRequestBody(),
			wantStatus:       http.StatusOK,
			call:             true,
			callAccessToken:  true,
			callRefreshToken: true,
			wantUserID:       validUserID,
			wantAccessToken:  validAccessToken,
			wantRefreshToken: validRefreshToken,
			wantResponseBody: AuthResponse{
				UserID:       validUserID.String(),
				AccessToken:  validAccessToken,
				RefreshToken: validRefreshToken,
			},
		},
		{
			name:            "failed with invalid accessToken",
			reqBody:         newValidRequestBody(),
			serviceErr:      apperror.ErrInternal,
			wantErr:         apperror.ErrInternal,
			call:            true,
			callAccessToken: true,
			wantUserID:      validUserID,
		},
		{
			name:             "failed with invalid refreshToken",
			reqBody:          newValidRequestBody(),
			serviceErr:       apperror.ErrInternal,
			wantErr:          apperror.ErrInternal,
			call:             true,
			callAccessToken:  true,
			callRefreshToken: true,
			wantAccessToken:  validAccessToken,
			wantUserID:       validUserID,
		},
	}
	for _, tt := range loginTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			var (
				called             bool
				calledAccessToken  bool
				calledRefreshToken bool
			)
			h := &Handler{
				userService: &StubUserService{
					FakeAuthenticateUser: func(_ context.Context, _ *domain.User) (uuid.UUID, error) {
						called = true
						if tt.serviceErr != nil && !tt.callAccessToken {
							return uuid.Nil, tt.serviceErr
						}
						return tt.wantUserID, nil
					},
				},
				authBundleService: &StubAuthbundleService{
					FakeGenerateAccessToken: func(_ uuid.UUID) (string, error) {
						calledAccessToken = true
						if tt.serviceErr != nil && !tt.callRefreshToken {
							return "", tt.serviceErr
						}
						return tt.wantAccessToken, nil
					},
					FakeGenerateRefreshToken: func(_ context.Context, _ uuid.UUID) (string, error) {
						calledRefreshToken = true
						if tt.serviceErr != nil {
							return "", tt.serviceErr
						}
						return tt.wantRefreshToken, nil
					},
				},
				authConfig: &authbundle.AuthConfig{},
				logger:     slog.New(slog.DiscardHandler),
			}
			var bodyReader io.Reader
			var bodyBytes []byte
			if tt.rawBody != "" {
				bodyBytes = []byte(tt.rawBody)
			} else {
				b, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
				bodyBytes = b
			}
			bodyReader = bytes.NewReader(bodyBytes)
			request := httptest.NewRequestWithContext(
				ctx,
				"POST",
				"/api/user/login",
				bodyReader,
			)
			response := httptest.NewRecorder()

			err := h.HandleUserLogin(response, request)

			// 失敗系
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				if tt.call != called {
					t.Errorf("service called = %v, want %v", called, tt.call)
				}
				if tt.callAccessToken != calledAccessToken {
					t.Errorf("access token service called = %v, want %v", calledAccessToken, tt.callAccessToken)
				}
				if tt.callRefreshToken != calledRefreshToken {
					t.Errorf("refresh token service called = %v, want %v", calledRefreshToken, tt.callRefreshToken)
				}
				return
			}
			// 成功系
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if response.Code != tt.wantStatus {
				t.Errorf("status code = %v, want %v", response.Code, tt.wantStatus)
			}
			if !called {
				t.Errorf("service was not called")
			}
			if !calledAccessToken {
				t.Errorf("access token service was not called")
			}
			if !calledRefreshToken {
				t.Errorf("refresh token service was not called")
			}
			var gotResponse AuthResponse
			if errDecode := json.NewDecoder(response.Body).Decode(&gotResponse); errDecode != nil {
				t.Fatalf("failed to decode response body: %v", errDecode)
			}
			if gotResponse != tt.wantResponseBody {
				t.Errorf("response body = %v, want %v", gotResponse, tt.wantResponseBody)
			}
		})
	}
}
