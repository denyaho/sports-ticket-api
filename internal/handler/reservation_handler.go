package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"42tokyo-road-to-dena-server/authbundle"
	"42tokyo-road-to-dena-server/internal/apperror"
	"42tokyo-road-to-dena-server/internal/domain"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (h *Handler) HandleCancelReservation(w http.ResponseWriter, r *http.Request) error {
	ctx, span := tracer.Start(r.Context(), "DELETE /api/reservations/{id}", trace.WithAttributes(
		attribute.String("http.request.method", r.Method),
		attribute.String("url.path", r.URL.Path),
		attribute.String("url.scheme", r.URL.Scheme),
	))
	defer span.End()

	userID, ok := authbundle.GetUserIDFromContext(r.Context())
	if !ok {
		span.RecordError(apperror.ErrUnauthorized)
		span.SetStatus(codes.Error, apperror.ErrUnauthorized.Error())
		return apperror.ErrUnauthorized
	}

	id := r.PathValue("id")
	reservationID, err := uuid.Parse(id)
	if err != nil {
		span.RecordError(apperror.ErrBadRequest)
		span.SetStatus(codes.Error, apperror.ErrBadRequest.Error())
		return apperror.ErrBadRequest
	}

	err = h.reservationService.CancelReservation(ctx, reservationID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetStatus(codes.Ok, "Reservation cancellation successful")
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) HandleCreateReservation(w http.ResponseWriter, r *http.Request) error {
	ctx, span := tracer.Start(r.Context(), "POST /api/reservations", trace.WithAttributes(
		attribute.String("http.request.method", r.Method),
		attribute.String("url.path", r.URL.Path),
		attribute.String("url.scheme", r.URL.Scheme),
	))
	defer span.End()

	userID, ok := authbundle.GetUserIDFromContext(r.Context())
	if !ok {
		span.RecordError(apperror.ErrUnauthorized)
		span.SetStatus(codes.Error, apperror.ErrUnauthorized.Error())
		return apperror.ErrUnauthorized
	}

	var reqBody domain.ReservationRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		span.RecordError(apperror.ErrBadRequest)
		span.SetStatus(codes.Error, apperror.ErrBadRequest.Error())
		return errors.Join(apperror.ErrBadRequest, err)
	}

	reservationResponse, err := h.reservationService.CreateReservation(ctx, &reqBody, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetStatus(codes.Ok, "Reservation creation successful")
	h.respondJSON(w, reservationResponse, http.StatusOK)
	return nil
}

func (h *Handler) HandleGetUserReservations(w http.ResponseWriter, r *http.Request) error {
	ctx, span := tracer.Start(r.Context(), "GET /api/reservations", trace.WithAttributes(
		attribute.String("http.request.method", r.Method),
		attribute.String("url.path", r.URL.Path),
		attribute.String("url.scheme", r.URL.Scheme),
	))
	defer span.End()

	userID, ok := authbundle.GetUserIDFromContext(r.Context())

	if !ok {
		span.RecordError(apperror.ErrUnauthorized)
		span.SetStatus(codes.Error, apperror.ErrUnauthorized.Error())
		return apperror.ErrUnauthorized
	}

	reservations, err := h.reservationService.GetUserReservations(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetStatus(codes.Ok, "User reservations retrieval successful")
	h.respondJSON(w, reservations, http.StatusOK)
	return nil
}

//nolint:dupl // this is why similar code exists for other reservation handlers
func (h *Handler) HandleGetReservationByID(w http.ResponseWriter, r *http.Request) error {
	ctx, span := tracer.Start(r.Context(), "GET /api/reservations/{id}", trace.WithAttributes(
		attribute.String("http.request.method", r.Method),
		attribute.String("url.path", r.URL.Path),
		attribute.String("url.scheme", r.URL.Scheme),
	))
	defer span.End()

	userID, ok := authbundle.GetUserIDFromContext(r.Context())
	if !ok {
		span.RecordError(apperror.ErrUnauthorized)
		span.SetStatus(codes.Error, apperror.ErrUnauthorized.Error())
		return apperror.ErrUnauthorized
	}
	id := r.PathValue("id")
	reservationID, err := uuid.Parse(id)
	if err != nil {
		span.RecordError(apperror.ErrBadRequest)
		span.SetStatus(codes.Error, apperror.ErrBadRequest.Error())
		return apperror.ErrBadRequest
	}
	reservation, err := h.reservationService.GetReservationByID(ctx, reservationID, userID)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetStatus(codes.Ok, "Reservation retrieval successful")
	h.respondJSON(w, reservation, http.StatusOK)
	return nil
}

//nolint:dupl // this is why similar code exists for other reservation handlers
func (h *Handler) HandlePurchaseReservation(w http.ResponseWriter, r *http.Request) error {
	ctx, span := tracer.Start(r.Context(), "POST /api/reservations/{id}/purchase", trace.WithAttributes(
		attribute.String("http.request.method", r.Method),
		attribute.String("url.path", r.URL.Path),
		attribute.String("url.scheme", r.URL.Scheme),
	))
	defer span.End()

	userID, ok := authbundle.GetUserIDFromContext(r.Context())
	if !ok {
		span.RecordError(apperror.ErrUnauthorized)
		span.SetStatus(codes.Error, apperror.ErrUnauthorized.Error())
		return apperror.ErrUnauthorized
	}
	id := r.PathValue("id")
	reservationID, err := uuid.Parse(id)
	if err != nil {
		span.RecordError(apperror.ErrBadRequest)
		span.SetStatus(codes.Error, apperror.ErrBadRequest.Error())
		return apperror.ErrBadRequest
	}
	reservation, err := h.reservationService.PurchaseReservation(ctx, reservationID, userID)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetStatus(codes.Ok, "Reservation purchase successful")
	h.respondJSON(w, reservation, http.StatusOK)
	return nil
}
