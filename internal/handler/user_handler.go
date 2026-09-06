package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"go.opentelemetry.io/otel/codes"

	"go.opentelemetry.io/otel/attribute"
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
	ctx, span := tracer.Start(r.Context(), "POST /api/user/refresh", trace.WithAttributes(
		attribute.String("http.request.method", r.Method),
		attribute.String("url.path", r.URL.Path),
		attribute.String("url.scheme", r.URL.Scheme),
	))
	defer span.End()
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		span.RecordError(apperror.ErrUnauthorized)
		span.SetStatus(codes.Error, apperror.ErrUnauthorized.Error())
		return apperror.ErrUnauthorized
	}
	tokenData, err := h.authBundleService.ValidateRefreshToken(r.Context(), cookie.Value)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	userID := tokenData.UserID
	SetUserIDInRequestState(ctx, userID)

	newRefreshToken, err := h.authBundleService.RotateRefreshToken(r.Context(), cookie.Value)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	newAccessToken, err := h.authBundleService.GenerateAccessToken(userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	authbundle.SetAuthCookies(w, newAccessToken, newRefreshToken, h.authConfig)

	span.SetStatus(codes.Ok, "Token refresh successful")

	h.respondJSON(w, AuthResponse{
		UserID:       userID.String(),
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, http.StatusOK)
	return nil
}

func (h *Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) error {
	ctx, span := tracer.Start(r.Context(), "GET /api/user/me", trace.WithAttributes(
		attribute.String("http.request.method", r.Method),
		attribute.String("http.request.url", r.URL.String()),
		attribute.String("url.path", r.URL.Path),
	))
	defer span.End()
	// リクエストに対する認証
	userID, ok := authbundle.GetUserIDFromContext(ctx)
	if !ok {
		span.RecordError(apperror.ErrUnauthorized)
		span.SetStatus(codes.Error, apperror.ErrUnauthorized.Error())
		return apperror.ErrUnauthorized
	}

	userInfo, err := h.userService.FindUserByID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	SetUserIDInRequestState(ctx, userID)

	span.SetStatus(codes.Ok, "User retrieval successful")

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
	ctx, span := tracer.Start(r.Context(), "POST /api/user/login", trace.WithAttributes(
		attribute.String("http.request.method", r.Method),
		attribute.String("url.path", r.URL.Path),
		attribute.String("url.scheme", r.URL.Scheme),
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
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	SetUserIDInRequestState(ctx, id)

	accessToken, err := h.authBundleService.GenerateAccessToken(id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	refreshToken, err := h.authBundleService.GenerateRefreshToken(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	authbundle.SetAuthCookies(w, accessToken, refreshToken, h.authConfig)
	span.SetStatus(codes.Ok, "User login successful")

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
	ctx, span := tracer.Start(r.Context(), "POST /api/user/signup", trace.WithAttributes(
		attribute.String("http.request.method", r.Method),
		attribute.String("url.path", r.URL.Path),
		attribute.String("url.scheme", r.URL.Scheme),
	))
	defer span.End()

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
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	SetUserIDInRequestState(ctx, id)

	accessToken, err := h.authBundleService.GenerateAccessToken(id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	refreshToken, err := h.authBundleService.GenerateRefreshToken(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	authbundle.SetAuthCookies(w, accessToken, refreshToken, h.authConfig)
	span.SetStatus(codes.Ok, "User signup successful")

	h.respondJSON(w, AuthResponse{
		UserID:       id.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, http.StatusOK)
	return nil
}
