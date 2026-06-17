package handler

import (
	"errors"
	"net/http"
	"42tokyo-road-to-dena-server/internal/repository"
	"github.com/google/uuid"
	"42tokyo-road-to-dena-server/internal/apperror"
)


func (h *Handler) HandleGetAllGames(w http.ResponseWriter, r *http.Request) {
	games, err := h.gameService.GetAllGames(r.Context())
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, games, http.StatusOK)
}

func (h *Handler) HandleGetGameByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	gameID, err := uuid.Parse(id)
	if err != nil {
		h.handleError(w, apperror.ErrBadRequest)
		return
	}
	game, err := h.gameService.GetGameByID(r.Context(), gameID)
	if errors.Is(err, repository.ErrNotFound) {
		h.respondError(w, err, http.StatusNotFound)
		return
	}
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, game, http.StatusOK)
}