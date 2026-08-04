package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"42tokyo-road-to-dena-server/internal/apperror"
)

func (h *Handler) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	SetErrorInRequestState(r.Context(), err)
	h.respondError(w, toStatus(err))
}

func toStatus(err error) int {
	var maxErr *http.MaxBytesError
	var syntaxErr *json.SyntaxError
	var unmarshalTypeErr *json.UnmarshalTypeError
	switch {
	case errors.As(err, &maxErr):
		return http.StatusRequestEntityTooLarge
	case errors.As(err, &syntaxErr):
		return http.StatusBadRequest
	case errors.As(err, &unmarshalTypeErr):
		return http.StatusBadRequest
	case errors.Is(err, context.Canceled):
		return http.StatusRequestTimeout // 408 Request Timeout
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout // 504 Gateway Timeout
	case errors.Is(err, apperror.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, apperror.ErrInsufficientTickets):
		return http.StatusConflict // 409 Conflict
	case errors.Is(err, apperror.ErrUserNotFound):
		return http.StatusNotFound
	case errors.Is(err, apperror.ErrDatabase), errors.Is(err, apperror.ErrInternal):
		return http.StatusInternalServerError
	case errors.Is(err, apperror.ErrDuplicateEmail):
		return http.StatusConflict
	case errors.Is(err, apperror.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, apperror.ErrReservationExpired):
		return http.StatusGone // 410 Gone
	case errors.Is(err, apperror.ErrReservationConflict):
		return http.StatusConflict // 409 Conflict
	case errors.Is(err, apperror.ErrReservationNotPending):
		return http.StatusBadRequest // 400 Bad Request
	case errors.Is(err, apperror.ErrBadRequest):
		return http.StatusBadRequest // 400 Bad Request
	case errors.Is(err, apperror.ErrInvalidInput):
		return http.StatusBadRequest // 400 Bad Request
	case errors.Is(err, apperror.ErrAuthenticationFailed):
		return http.StatusUnauthorized // 401 Unauthorized
	case errors.Is(err, apperror.ErrForbidden):
		return http.StatusForbidden // 403 Forbidden
	default:
		return http.StatusInternalServerError
	}
}
