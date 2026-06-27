package handler

import (
	"net/http"
	"encoding/json"
	"42tokyo-road-to-dena-server/authbundle"
	"42tokyo-road-to-dena-server/internal/domain"
	"42tokyo-road-to-dena-server/internal/apperror"
	"github.com/google/uuid"
)

func (h *Handler) HandleCancelReservation(w http.ResponseWriter, r *http.Request) {

	userID, ok := authbundle.GetUserIDFromContext(r.Context())
	if !ok {
		h.handleError(w, apperror.ErrUnauthorized)
		return
	}

	ctx := r.Context()

	id := r.PathValue("id")
	reservationID, err := uuid.Parse(id)
	if err != nil {
		h.handleError(w, apperror.ErrBadRequest)
		return
	}

	err = h.reservationService.CancelReservation(ctx, reservationID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, nil, http.StatusNoContent)
}

func (h *Handler) HandleCreateReservation(w http.ResponseWriter, r *http.Request) {

	userID, ok := authbundle.GetUserIDFromContext(r.Context())
	if !ok {
		h.handleError(w, apperror.ErrUnauthorized)
		return
	}

	ctx := r.Context()
	
	var reqBody domain.ReservationRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		h.handleError(w, apperror.ErrBadRequest)
		return
	}

	reservation_response, err := h.reservationService.CreateReservation(ctx, &reqBody, userID)

	if err != nil {
		h.handleError(w, err)
		return
	}

	h.respondJSON(w, reservation_response, http.StatusOK)	
}

func (h *Handler) HandleGetUserReservations(w http.ResponseWriter, r *http.Request) {
	userID, ok := authbundle.GetUserIDFromContext(r.Context())

	if !ok {
		h.handleError(w, apperror.ErrUnauthorized)
		return
	}
	
	ctx := r.Context()

	reservations, err := h.reservationService.GetUserReservations(ctx, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, reservations, http.StatusOK)
}

func (h *Handler) HandleGetReservationByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := authbundle.GetUserIDFromContext(r.Context())
	if !ok {
		h.handleError(w, apperror.ErrUnauthorized)
		return
	}
	id := r.PathValue("id")
	reservationID, err := uuid.Parse(id)
	if err != nil {
		h.handleError(w, apperror.ErrBadRequest)
		return
	}
	reservation, err := h.reservationService.GetReservationByID(r.Context(), reservationID, userID)

	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, reservation, http.StatusOK)
}

func (h *Handler) HandlePurchaseReservation(w http.ResponseWriter, r *http.Request) {

	userID, ok := authbundle.GetUserIDFromContext(r.Context())
	if !ok {
		h.handleError(w, apperror.ErrUnauthorized)
		return
	}
	id := r.PathValue("id")
	reservationID, err := uuid.Parse(id)
	if err != nil {
		h.handleError(w, apperror.ErrBadRequest)
		return
	}
	reservation, err := h.reservationService.PurchaseReservation(r.Context(), reservationID, userID)

	if err != nil {
		h.handleError(w, err)
		return
	}
	h.respondJSON(w, reservation, http.StatusOK)
}