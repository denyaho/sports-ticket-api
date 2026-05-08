package handler

import (
	"net/http"
	"time"
)

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	}
	h.respondJSON(w, response, http.StatusOK)
}
