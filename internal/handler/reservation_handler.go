package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"42tokyo-road-to-dena-server/authbundle"
	"42tokyo-road-to-dena-server/internal/apperror"
	"42tokyo-road-to-dena-server/internal/domain"

	"github.com/google/uuid"
)

func (h *Handler) HandleCancelReservation(w http.ResponseWriter, r *http.Request) error {
	userID, ok := authbundle.GetUserIDFromContext(r.Context())
	if !ok {
		return apperror.ErrUnauthorized
	}

	ctx := r.Context()

	id := r.PathValue("id")
	reservationID, err := uuid.Parse(id)
	if err != nil {
		return apperror.ErrBadRequest
	}

	err = h.reservationService.CancelReservation(ctx, reservationID, userID)
	if err != nil {
		return err
	}
	h.respondJSON(w, nil, http.StatusNoContent)
	return nil
}

func (h *Handler) HandleCreateReservation(w http.ResponseWriter, r *http.Request) error {
	userID, ok := authbundle.GetUserIDFromContext(r.Context())
	if !ok {
		return apperror.ErrUnauthorized
	}

	ctx := r.Context()

	var reqBody domain.ReservationRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		return errors.Join(apperror.ErrBadRequest, err)
	}

	reservationResponse, err := h.reservationService.CreateReservation(ctx, &reqBody, userID)

	if err != nil {
		return err
	}

	h.respondJSON(w, reservationResponse, http.StatusOK)
	return nil
}

func (h *Handler) HandleGetUserReservations(w http.ResponseWriter, r *http.Request) error {
	userID, ok := authbundle.GetUserIDFromContext(r.Context())

	if !ok {
		return apperror.ErrUnauthorized
	}

	ctx := r.Context()

	reservations, err := h.reservationService.GetUserReservations(ctx, userID)
	if err != nil {
		return err
	}
	h.respondJSON(w, reservations, http.StatusOK)
	return nil
}

func (h *Handler) HandleGetReservationByID(w http.ResponseWriter, r *http.Request) error {
	userID, ok := authbundle.GetUserIDFromContext(r.Context())
	if !ok {
		return apperror.ErrUnauthorized
	}
	id := r.PathValue("id")
	reservationID, err := uuid.Parse(id)
	if err != nil {
		return apperror.ErrBadRequest
	}
	reservation, err := h.reservationService.GetReservationByID(r.Context(), reservationID, userID)

	if err != nil {
		return err
	}
	h.respondJSON(w, reservation, http.StatusOK)
	return nil
}

func (h *Handler) HandlePurchaseReservation(w http.ResponseWriter, r *http.Request) error {
	userID, ok := authbundle.GetUserIDFromContext(r.Context())
	if !ok {
		return apperror.ErrUnauthorized
	}
	id := r.PathValue("id")
	reservationID, err := uuid.Parse(id)
	if err != nil {
		return apperror.ErrBadRequest
	}
	reservation, err := h.reservationService.PurchaseReservation(r.Context(), reservationID, userID)

	if err != nil {
		return err
	}
	h.respondJSON(w, reservation, http.StatusOK)
	return nil
}
