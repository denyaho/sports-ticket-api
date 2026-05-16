package handler

import (
	"42tokyo-road-to-dena-server/authbundle"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func New() *Handler {
	return &Handler{}
}

type Handler struct {
	authBundleService *authbundle.AuthBundle
	authConfig        *authbundle.AuthConfig
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// ルーティング
	mux.HandleFunc("GET /health", h.HealthCheck)
	mux.HandleFunc("POST /api/user/signup", h.UsrSignup)

	// Swagger/OpenAPI 配信
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		repoRoot, _ := os.Getwd()
		p := filepath.Join(repoRoot, "docs", "openapi.yaml")
		w.Header().Set("Content-Type", "application/yaml")
		http.ServeFile(w, r, p)
	})
	mux.HandleFunc("GET /swagger", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("docs", "swagger", "index.html"))
	})
	mux.HandleFunc("GET /swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("docs", "swagger", "index.html"))
	})

	return mux
}

func (h *Handler) respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *Handler) respondError(w http.ResponseWriter, err error, status int) {
	response := map[string]string{
		"error": err.Error(),
	}
	h.respondJSON(w, response, status)
}
