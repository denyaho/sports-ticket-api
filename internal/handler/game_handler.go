package handler

import (
	"fmt"
	"net/http"

	"42tokyo-road-to-dena-server/internal/apperror"

	"github.com/google/uuid"
)

func (h *Handler) HandleGetAllGames(w http.ResponseWriter, r *http.Request) error {
	games, err := h.gameService.GetAllGames(r.Context())
	if err != nil {
		return err
	}
	h.respondJSON(w, games, http.StatusOK)
	return nil
}

func (h *Handler) HandleGetGameByID(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	gameID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", apperror.ErrBadRequest)
	}
	game, err := h.gameService.GetGameByID(r.Context(), gameID)
	if err != nil {
		return err
	}
	h.respondJSON(w, game, http.StatusOK)
	return nil
}
