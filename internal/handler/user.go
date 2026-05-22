package handler

import (
	"net/http"
	"encoding/json"
	"errors"
	"fmt"
	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/service"
	"42tokyo-road-to-dena-server/internal/repository"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(s service.UserService) *UserHandler {
	return &UserHandler{userService: s}
}

func (h *UserHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, errors.New("invalid user ID"), http.StatusBadRequest)
		return
	}
	user, err := h.userService.FindUserByID(r.Context(), id)

	if errors.Is(err, repository.ErrUserNotFound) {
		h.respondError(w, err, http.StatusNotFound)
		return
	}
	if errors.Is(err, repository.ErrDatabase) {
		h.respondError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	h.respondJSON(w, user, http.StatusOK)
}

func (h *UserHandler) HandleUserSignup(w http.ResponseWriter, r *http.Request) {
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
	id, err := h.userService.CreateUser(ctx, userinfo)
	if errors.Is(err, repository.ErrDuplicateEmail) {
		h.respondError(w, err, http.StatusConflict)
		return
	}
	if errors.Is(err, repository.ErrDatabase) {
		h.respondError(w, err, http.StatusInternalServerError)
		return
	}
	response := map[string]string{
		"user_id": id.String(),
	}
	h.respondJSON(w, response, http.StatusOK)
}

