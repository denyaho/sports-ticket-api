package handler

import (
	"net/http"
	"encoding/json"
	"errors"
	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/authbundle"
	"42tokyo-road-to-dena-server/internal/apperror"
)


func (h *Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	//リクエストに対する認証
	userID, ok := authbundle.GetUserIDFromContext(r.Context())
	if !ok {
		h.handleError(w, apperror.ErrUnauthorized)
		return
	}

	userInfo, err := h.userservice.FindUserByID(r.Context(), userID)

	if errors.Is(err, apperror.ErrUserNotFound) {
		h.handleError(w, err)
		return
	}
	if errors.Is(err, apperror.ErrDatabase) {
		h.handleError(w, err)
		return
	}
	response := map[string]string{
		"user_id": userInfo.ID.String(),
		"username": userInfo.Username,
		"email": userInfo.Email,
	}
	h.respondJSON(w, response, http.StatusOK)
}

type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) HandleUserLogin(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	var reqBody LoginRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		h.handleError(w, err)
		return
	}
	userInfo := &domain.User{
		Email: reqBody.Email,
		Password: reqBody.Password,
	}
	id, err := h.userservice.AuthenticateUser(ctx, userInfo)

	if errors.Is(err, apperror.ErrUserNotFound) {
		h.handleError(w, err)
		return
	}
	if errors.Is(err, apperror.ErrDatabase) {
		h.handleError(w, err)
		return
	}
	if errors.Is(err, apperror.ErrAuthenticationFailed) {
		h.handleError(w, err)
		return
	}

	accessToken, err := h.authBundleService.GenerateAccessToken(id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	refreshToken, err := h.authBundleService.GenerateRefreshToken(ctx, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	authbundle.SetAuthCookies(w, accessToken, refreshToken, h.authConfig)

	response := map[string]string{
		"user_id": id.String(),
		"access_token": accessToken,
		"refresh_token": refreshToken,
	}
	h.respondJSON(w, response, http.StatusOK)
}

type SignupRequest struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) HandleUserSignup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // リクエストのコンテキストを取得
	var reqBody SignupRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		h.handleError(w, err)
		return
	}
	
	userinfo := &domain.User{
		Username: reqBody.Name,
		Email: reqBody.Email,
		Password: reqBody.Password,
	}
	id, err := h.userservice.CreateUser(ctx, userinfo)
	if errors.Is(err, apperror.ErrDuplicateEmail) {
		h.handleError(w, err)
		return
	}
	if errors.Is(err, apperror.ErrDatabase) {
		h.handleError(w, err)
		return
	}

	accessToken, err := h.authBundleService.GenerateAccessToken(id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	refreshToken, err := h.authBundleService.GenerateRefreshToken(ctx, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	authbundle.SetAuthCookies(w, accessToken, refreshToken, h.authConfig)

	response := map[string]string{
		"user_id": id.String(),
		"access_token": accessToken,
		"refresh_token": refreshToken,
	}
	h.respondJSON(w, response, http.StatusOK)
}

