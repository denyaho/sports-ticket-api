package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	
	"go.opentelemetry.io/otel/trace"

	"42tokyo-road-to-dena-server/authbundle"
	"42tokyo-road-to-dena-server/internal/apperror"
	"42tokyo-road-to-dena-server/internal/domain"
)

type AuthResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (h *Handler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		return apperror.ErrUnauthorized
	}
	tokenData, err := h.authBundleService.ValidateRefreshToken(r.Context(), cookie.Value)
	if err != nil {
		return err
	}
	userID := tokenData.UserID
	SetUserIDInRequestState(r.Context(), userID)

	newRefreshToken, err := h.authBundleService.RotateRefreshToken(r.Context(), cookie.Value)
	if err != nil {
		return err
	}
	newAccessToken, err := h.authBundleService.GenerateAccessToken(userID)
	if err != nil {
		return err
	}

	authbundle.SetAuthCookies(w, newAccessToken, newRefreshToken, h.authConfig)

	h.respondJSON(w, AuthResponse{
		UserID:       userID.String(),
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, http.StatusOK)
	return nil
}

func (h *Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) error {
	// リクエストに対する認証
	userID, ok := authbundle.GetUserIDFromContext(r.Context())
	if !ok {
		return apperror.ErrUnauthorized
	}

	userInfo, err := h.userService.FindUserByID(r.Context(), userID)
	if err != nil {
		return err
	}
	SetUserIDInRequestState(r.Context(), userID)

	h.respondJSON(w, UserResponse{
		UserID:   userInfo.ID.String(),
		Username: userInfo.Username,
		Email:    userInfo.Email,
	}, http.StatusOK)
	return nil
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) HandleUserLogin(w http.ResponseWriter, r *http.Request) error {
	ctx, span := tracer.Start(r.Context(), "handler HandleUserLogin", trace.WithAttributes(
	))
	defer span.End()
	// ctx := r.Context() // This line is redundant because ctx is already defined above with the span.
	var reqBody LoginRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		return errors.Join(apperror.ErrBadRequest, err)
	}
	userInfo := &domain.User{
		Email:    reqBody.Email,
		Password: reqBody.Password,
	}
	id, err := h.userService.AuthenticateUser(ctx, userInfo)
	if err != nil {
		span.RecordError(err)
		return err
	}

	SetUserIDInRequestState(r.Context(), id)

	accessToken, err := h.authBundleService.GenerateAccessToken(id)
	if err != nil {
		return err
	}
	refreshToken, err := h.authBundleService.GenerateRefreshToken(ctx, id)
	if err != nil {
		return err
	}

	authbundle.SetAuthCookies(w, accessToken, refreshToken, h.authConfig)

	h.respondJSON(w, AuthResponse{
		UserID:       id.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, http.StatusOK)
	return nil
}

type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) HandleUserSignup(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context() // リクエストのコンテキストを取得
	var reqBody SignupRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		return errors.Join(apperror.ErrBadRequest, err)
	}

	userInfo := &domain.User{
		Username: reqBody.Name,
		Email:    reqBody.Email,
		Password: reqBody.Password,
	}
	id, err := h.userService.CreateUser(ctx, userInfo)
	if err != nil {
		return err
	}

	SetUserIDInRequestState(r.Context(), id)

	accessToken, err := h.authBundleService.GenerateAccessToken(id)
	if err != nil {
		return err
	}
	refreshToken, err := h.authBundleService.GenerateRefreshToken(ctx, id)
	if err != nil {
		return err
	}

	authbundle.SetAuthCookies(w, accessToken, refreshToken, h.authConfig)

	h.respondJSON(w, AuthResponse{
		UserID:       id.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, http.StatusOK)
	return nil
}
