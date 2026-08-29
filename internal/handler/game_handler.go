package handler

import (
	"fmt"
	"net/http"

	"42tokyo-road-to-dena-server/internal/apperror"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel"
)

func (h *Handler) HandleGetAllGames(w http.ResponseWriter, r *http.Request) error {

	ctx, span := otel.Tracer("handler").Start(r.Context(), "handler HandleGetAllGames", trace.WithAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.url", r.URL.String()),
	))
	defer span.End()

	games, err := h.gameService.GetAllGames(ctx)
	if err != nil {
		span.SetAttributes(
			attribute.String("error", err.Error()),
		)
		return err
	}

	span.SetAttributes(
		attribute.Int("http.status_code", http.StatusOK),
	)

	h.respondJSON(w, games, http.StatusOK)
	return nil
}

func (h *Handler) HandleGetGameByID(w http.ResponseWriter, r *http.Request) error {
	ctx, span := otel.Tracer("handler").Start(r.Context(), "handler HandleGetGameByID")
	defer span.End()

	id := r.PathValue("id")
	gameID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", apperror.ErrBadRequest)
	}
	game, err := h.gameService.GetGameByID(ctx, gameID)
	if err != nil {
		return err
	}
	h.respondJSON(w, game, http.StatusOK)
	return nil
}
