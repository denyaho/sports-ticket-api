package handler	

import (
	"net/http"
	"42tokyo-road-to-dena-server/internal/apperror"
	"github.com/google/uuid"
)

func (h *Handler) HandleGetSeatsByGameID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	gameID, err := uuid.Parse(id)
	if err != nil {
		h.HandleError(w, apperror.ErrBadRequest)
		return
	}
	seats, err := h.seatsService.GetSeatsByGameID(r.Context(), gameID)
	if err != nil {
		h.HandleError(w, err)
		return
	}
	h.respondJSON(w, seats, http.StatusOK)
}