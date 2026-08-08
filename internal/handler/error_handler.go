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
	h.respondError(w, toStatus(r.Context(), err))
}

func toStatus(ctx context.Context, err error) int {
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
		if ctx.Err() != nil {
			return 499
		}
		return http.StatusInternalServerError
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout
	case errors.Is(err, apperror.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, apperror.ErrValidation):
		return http.StatusBadRequest
	case errors.Is(err, apperror.ErrRetryable):
		return http.StatusServiceUnavailable
	case errors.Is(err, apperror.ErrTimeout):
		return http.StatusGatewayTimeout
	case errors.Is(err, apperror.ErrUnavailable):
		return http.StatusServiceUnavailable
	case errors.Is(err, apperror.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, apperror.ErrInsufficientTickets):
		return http.StatusConflict
	case errors.Is(err, apperror.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, apperror.ErrInternal):
		return http.StatusInternalServerError
	case errors.Is(err, apperror.ErrBadRequest):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
