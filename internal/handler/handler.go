package handler

import (
	"42tokyo-road-to-dena-server/authbundle"
	"42tokyo-road-to-dena-server/internal/service"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func New(authbundle *authbundle.AuthBundle,
	authConfig *authbundle.AuthConfig,
	userservice service.UserService,
	gameService service.GameService,
	seatsService service.SeatsService) *Handler {
	return &Handler{
		authBundleService: authbundle,
		authConfig:        authConfig,
		userservice:       userservice,
		gameService:		gameService,
		seatsService:      seatsService,
	}
}

type Handler struct {
	authBundleService *authbundle.AuthBundle
	authConfig        *authbundle.AuthConfig	
	userservice service.UserService
	gameService service.GameService
	seatsService service.SeatsService
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// ルーティング
	mux.HandleFunc("GET /health", h.HealthCheck)
	mux.HandleFunc("POST /api/user/signup", h.HandleUserSignup)
	// 認証が必要なルートはミドルウェアで保護
	mux.Handle("GET /api/user/me", h.AuthRequired(http.HandlerFunc(h.HandleGetUser)))

	mux.HandleFunc("GET /api/games", h.HandleGetAllGames)
	mux.HandleFunc("GET /api/games/{id}", h.HandleGetGameByID)

	mux.HandleFunc("GET /api/games/{id}/seats", h.HandleGetSeatsByGameID)

	mux.Handle("POST /api/reservations", h.AuthRequired(http.HandlerFunc(h.HandleCreateReservation)))

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
