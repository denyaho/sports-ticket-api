package handler

import (
	"net/http"

	"42tokyo-road-to-dena-server/internal/apperror"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (h *Handler) HandleGetSeatsByGameID(w http.ResponseWriter, r *http.Request) error {
	ctx, span := tracer.Start(r.Context(), "GET /api/games/{id}/seats", trace.WithAttributes(
		attribute.String("http.request.method", r.Method),
		attribute.String("url.path", r.URL.Path),
		attribute.String("url.scheme", r.URL.Scheme),
	))
	defer span.End()
	id := r.PathValue("id")

	gameID, err := uuid.Parse(id)
	if err != nil {
		span.RecordError(apperror.ErrBadRequest)
		span.SetStatus(codes.Error, apperror.ErrBadRequest.Error())
		return apperror.ErrBadRequest
	}
	seats, err := h.seatsService.GetSeatsByGameID(ctx, gameID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetStatus(codes.Ok, "Seats retrieval successful")
	h.respondJSON(w, seats, http.StatusOK)
	return nil
}
