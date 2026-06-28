package apperror

import (
	"errors"
)

var (
	//reservation related errors
	ErrInsufficientTickets = errors.New("insufficient tickets available")
	ErrReservationExpired = errors.New("reservation has expired")
	ErrForbidden = errors.New("forbidden")
	ErrNotFound = errors.New("not found")

	//user related errors
	ErrUnauthorized = errors.New("unauthorized")

	ErrDuplicateEmail = errors.New("email already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrDatabase = errors.New("database error")
	ErrUserNotCreated = errors.New("failed to create user")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrInvalidInput = errors.New("invalid input")

	ErrReservationConflict = errors.New("reservation conflict")
	ErrBadRequest = errors.New("bad request")
	ErrReservationNotPending = errors.New("reservation is not pending")
	ErrInternal = errors.New("internal server error")	
)
