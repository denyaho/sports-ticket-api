package handler	

import (
	"net/http"
)

func (h *Handler) HandleGetSeatsByGameID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	seats, err := h.seatsService.GetSeatsByGameID(r.Context(), id)
	if err != nil {
		h.respondError(w, err, http.StatusInternalServerError)
		return
	}
	h.respondJSON(w, seats, http.StatusOK)
}