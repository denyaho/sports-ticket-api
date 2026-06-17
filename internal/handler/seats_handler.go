package handler	

import (
	"net/http"
	"42tokyo-road-to-dena-server/internal/apperror"
	"github.com/google/uuid"
)

func (h *Handler) HandleGetSeatsByGameID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	seatID, err := uuid.Parse(id)
	if err != nil {
		h.handleError(w, apperror.ErrBadRequest)
		return
	}
	seats, err := h.seatsService.GetSeatsByGameID(r.Context(), seatID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, seats, http.StatusOK)
}