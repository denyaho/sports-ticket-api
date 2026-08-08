package apperror

import (
	"errors"
)

var (
	ErrConflict = errors.New("conflict")
	ErrValidation = errors.New("validation error")
	ErrRetryable = errors.New("retryable error")
	ErrTimeout              = errors.New("timeout")
	ErrUnavailable          = errors.New("service unavailable")
	ErrNotFound			 = errors.New("not found")
	ErrDeadlineExceeded      = errors.New("deadline exceeded")
	ErrDatabase              = errors.New("database error")
	ErrInsufficientTickets = errors.New("insufficient tickets")
)

