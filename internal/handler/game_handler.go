package handler

import (
	"errors"
	"net/http"
	"42tokyo-road-to-dena-server/internal/repository"
)


func (h *Handler) HandleGetAllGames(w http.ResponseWriter, r *http.Request) {
	games, err := h.gameService.GetAllGames(r.Context())
	if err != nil {
		h.respondError(w, err, http.StatusInternalServerError)
		return
	}
	h.respondJSON(w, games, http.StatusOK)
}

func (h *Handler) HandleGetGameByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	game, err := h.gameService.GetGameByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		h.respondError(w, err, http.StatusNotFound)
		return
	}
	if err != nil {
		h.respondError(w, err, http.StatusInternalServerError)
		return
	}
	h.respondJSON(w, game, http.StatusOK)
}