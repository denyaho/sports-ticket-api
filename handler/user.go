package handler

import (
	"net/http"
	"42tokyo-road-to-dena-server/authbundle"
	"encoding/json"
)

type SignupRequest struct {
	Name string `json:"name"`//struct tag を追加して、JSONのキーを指定
	Email string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) UserSignup(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context() // リクエストのコンテキストを取得
	var reqBody SignupRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		h.respondError(w, err, http.StatusBadRequest)
		return
	}
	hashedPassword, err := authbundle.HashPassword(reqBody.Password)
	if err != nil {
		h.respondError(w, err, http.StatusInternalServerError)
		return
	}
	id, err := h.userRepository.CreateUser(ctx, reqBody.Name, reqBody.Email, hashedPassword)
	if err != nil {
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
	response := map[string]string{
		"user_id": id.String(),
		"access_token": accessToken,
		"refresh_token": refreshToken,
	}
	h.respondJSON(w, response, http.StatusOK)
}