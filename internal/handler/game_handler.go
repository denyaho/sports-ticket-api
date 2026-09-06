package handler

import (
	"fmt"
	"net/http"

	"42tokyo-road-to-dena-server/internal/apperror"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (h *Handler) HandleGetAllGames(w http.ResponseWriter, r *http.Request) error {
	ctx, span := tracer.Start(r.Context(), "GET /api/games", trace.WithAttributes(
		attribute.String("http.request.method", r.Method),
		attribute.String("url.path", r.URL.Path),
		attribute.String("url.scheme", r.URL.Scheme),
	))
	defer span.End()

	games, err := h.gameService.GetAllGames(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "Games retrieval successful")

	h.respondJSON(w, games, http.StatusOK)
	return nil
}

func (h *Handler) HandleGetGameByID(w http.ResponseWriter, r *http.Request) error {
	ctx, span := tracer.Start(r.Context(), "GET /api/games/{id}", trace.WithAttributes(
		attribute.String("http.request.method", r.Method),
		attribute.String("url.path", r.URL.Path),
		attribute.String("url.scheme", r.URL.Scheme),
	))
	defer span.End()

	id := r.PathValue("id")
	gameID, err := uuid.Parse(id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("invalid UUID: %w", apperror.ErrBadRequest)
	}
	game, err := h.gameService.GetGameByID(ctx, gameID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetStatus(codes.Ok, "Game retrieval successful")
	h.respondJSON(w, game, http.StatusOK)
	return nil
}
