package handler	

import (
	"net/http"
	"42tokyo-road-to-dena-server/internal/apperror"
	"github.com/google/uuid"
)

func (h *Handler) HandleGetSeatsByGameID(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")

	gameID, err := uuid.Parse(id)
	if err != nil {
		return apperror.ErrBadRequest
	}
	seats, err := h.seatsService.GetSeatsByGameID(r.Context(), gameID)
	if err != nil {
		return err
	}
	h.respondJSON(w, seats, http.StatusOK)
	return nil
}