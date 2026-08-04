package handler

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"42tokyo-road-to-dena-server/authbundle"
	"42tokyo-road-to-dena-server/internal/service"
)

func New(authbundle *authbundle.AuthBundle,
	authConfig *authbundle.AuthConfig,
	userservice service.UserService,
	gameService service.GameService,
	seatsService service.SeatsService,
	reservationService service.ReservationService,
	logger *slog.Logger) *Handler {
	return &Handler{
		authBundleService:  authbundle,
		authConfig:         authConfig,
		userservice:        userservice,
		gameService:        gameService,
		seatsService:       seatsService,
		reservationService: reservationService,
		logger:             logger,
	}
}

type Handler struct {
	authBundleService  *authbundle.AuthBundle
	authConfig         *authbundle.AuthConfig
	userservice        service.UserService
	gameService        service.GameService
	seatsService       service.SeatsService
	reservationService service.ReservationService
	logger             *slog.Logger
}

func (h *Handler) toHandler(fn func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			h.HandleError(w, r, err)
		}
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// ルーティング
	mux.HandleFunc("GET /health", h.toHandler(h.HealthCheck))
	mux.HandleFunc("POST /api/user/signup", h.toHandler(h.HandleUserSignup))
	mux.HandleFunc("POST /api/user/login", h.toHandler(h.HandleUserLogin))
	mux.HandleFunc("POST /api/user/refresh", h.toHandler(h.HandleRefreshToken))
	// 認証が必要なルートはミドルウェアで保護
	mux.Handle("GET /api/user/me", h.AuthRequired(h.toHandler(h.HandleGetUser)))

	mux.HandleFunc("GET /api/games", h.toHandler(h.HandleGetAllGames))
	mux.HandleFunc("GET /api/games/{id}", h.toHandler(h.HandleGetGameByID))

	mux.HandleFunc("GET /api/games/{id}/seats", h.toHandler(h.HandleGetSeatsByGameID))

	mux.Handle("POST /api/reservations", h.AuthRequired(h.toHandler(h.HandleCreateReservation)))

	mux.Handle("GET /api/reservations", h.AuthRequired(h.toHandler(h.HandleGetUserReservations)))

	mux.Handle("GET /api/reservations/{id}", h.AuthRequired(h.toHandler(h.HandleGetReservationByID)))

	mux.Handle("PUT /api/reservations/{id}/purchase", h.AuthRequired(h.toHandler(h.HandlePurchaseReservation)))

	mux.Handle("DELETE /api/reservations/{id}", h.AuthRequired(h.toHandler(h.HandleCancelReservation)))

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

	return h.Logging(mux)
}

func (h *Handler) respondJSON(w http.ResponseWriter, data any, status int) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		h.logger.Error("encode response failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(buf.Bytes()); err != nil {
		h.logger.Error("write response failed", "error", err)
	}
}

func (h *Handler) respondError(w http.ResponseWriter, status int) {
	h.respondJSON(w, map[string]string{"error": http.StatusText(status)}, status)
}
