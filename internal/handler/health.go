package handler

import (
	"net/http"
	"time"
)

func (h *Handler) HealthCheck(w http.ResponseWriter, _ *http.Request) error {
	response := map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	}
	h.respondJSON(w, response, http.StatusOK)
	return nil
}
