package handler

import (
	"errors"
	"net/http"
	"context"
	"database/sql"
	"42tokyo-road-to-dena-server/internal/apperror"
)

func (h *Handler) HandleError(w http.ResponseWriter, err error) {

	switch {
	case errors.Is(err, context.Canceled):
		h.respondError(w, err, http.StatusRequestTimeout) // 408 Request Timeout
	case errors.Is(err, context.DeadlineExceeded):
		h.respondError(w, err, http.StatusGatewayTimeout)// 504 Gateway Timeout
	case errors.Is(err, sql.ErrConnDone):
		h.respondError(w, err, http.StatusInternalServerError)
	case errors.Is(err, sql.ErrNoRows):
		h.respondError(w, err, http.StatusNotFound)
	case errors.Is(err, apperror.ErrNotFound):
		h.respondError(w, err, http.StatusNotFound)
	case errors.Is(err, apperror.ErrInsufficientTickets):
		h.respondError(w, err, http.StatusConflict) // 409 Conflict
	case errors.Is(err, apperror.ErrUserNotFound):
		h.respondError(w, err, http.StatusNotFound)
	case errors.Is(err, apperror.ErrDatabase), errors.Is(err, apperror.ErrInternal):
		h.respondError(w, err, http.StatusInternalServerError)
	case errors.Is(err, apperror.ErrDuplicateEmail):
		h.respondError(w, err, http.StatusConflict)
	case errors.Is(err, apperror.ErrUnauthorized):
		h.respondError(w, err, http.StatusUnauthorized)
	case errors.Is(err, apperror.ErrReservationExpired):
		h.respondError(w, err, http.StatusGone) // 410 Gone
	case errors.Is(err, apperror.ErrReservationConflict):
		h.respondError(w, err, http.StatusConflict) // 409 Conflict
	case errors.Is(err, apperror.ErrReservationNotPending):
		h.respondError(w, err, http.StatusBadRequest) // 400 Bad Request
	case errors.Is(err, apperror.ErrBadRequest):
		h.respondError(w, err, http.StatusBadRequest) // 400 Bad Request
	case errors.Is(err, apperror.ErrInvalidInput):
		h.respondError(w, err, http.StatusBadRequest) // 400 Bad Request
	case errors.Is(err, apperror.ErrAuthenticationFailed):
		h.respondError(w, err, http.StatusUnauthorized) // 401 Unauthorized
	case errors.Is(err, apperror.ErrForbidden):
		h.respondError(w, err, http.StatusForbidden) // 403 Forbidden
	default:
		h.respondError(w, err, http.StatusInternalServerError)
	}
}