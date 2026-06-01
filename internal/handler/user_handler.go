package handler

import (
	"net/http"
	"encoding/json"
	"errors"
	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/repository"
	"42tokyo-road-to-dena-server/authbundle"
)


func (h *Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	//リクエストに対する認証
	userID, ok := authbundle.GetUserIDFromContext(r.Context())
	if !ok {
		h.respondError(w, authbundle.ErrUnauthorized, http.StatusUnauthorized)
		return
	}

	userInfo, err := h.userservice.FindUserByID(r.Context(), userID)

	if errors.Is(err, repository.ErrUserNotFound) {
		h.respondError(w, err, http.StatusNotFound)
		return
	}
	if errors.Is(err, repository.ErrDatabase) {
		h.respondError(w, err, http.StatusInternalServerError)
		return
	}
	response := map[string]string{
		"user_id": userInfo.ID.String(),
		"username": userInfo.Username,
		"email": userInfo.Email,
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
		h.respondError(w, err, http.StatusBadRequest)
		return
	}
	
	userinfo := &domain.User{
		Username: reqBody.Name,
		Email: reqBody.Email,
		Password: reqBody.Password,
	}
	id, err := h.userservice.CreateUser(ctx, userinfo)
	if errors.Is(err, repository.ErrDuplicateEmail) {
		h.respondError(w, err, http.StatusConflict)
		return
	}
	if errors.Is(err, repository.ErrDatabase) {
		h.respondError(w, err, http.StatusInternalServerError)
		return
	}

	accessToken, err := h.authBundleService.GenerateAccessToken(id)
	if err != nil {
		h.respondError(w, err, http.StatusInternalServerError)
		return
	}
	refreshToken, err := h.authBundleService.GenerateRefreshToken(ctx, id)
	if err != nil {
		h.respondError(w, err, http.StatusInternalServerError)
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

