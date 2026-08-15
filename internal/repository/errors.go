package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"42tokyo-road-to-dena-server/internal/apperror"
)

const (
	pqUniqueViolation        pq.ErrorCode = "23505"
	pqForeignKeyViolation    pq.ErrorCode = "23503"
	pqCheckViolation         pq.ErrorCode = "23514"
	pqSerializationFailure   pq.ErrorCode = "40001"
	pqDeadlockDetected       pq.ErrorCode = "40P01"
	pqLockNotAvailable       pq.ErrorCode = "55P03"
	pqQueryCanceled          pq.ErrorCode = "57014"
	pqTooManyConnections     pq.ErrorCode = "53300"
	pqConnectionFailure      pq.ErrorCode = "08006"
	pqConnectionDoesNotExist pq.ErrorCode = "08003"
)

func classifyErrorOnRepository(code pq.ErrorCode) error {
	switch code {
	case pqUniqueViolation:
		return apperror.ErrConflict
	case pqForeignKeyViolation, pqCheckViolation:
		return apperror.ErrValidation
	case pqSerializationFailure, pqDeadlockDetected, pqLockNotAvailable:
		return apperror.ErrRetryable
	case pqQueryCanceled:
		return apperror.ErrTimeout
	case pqTooManyConnections, pqConnectionFailure, pqConnectionDoesNotExist:
		return apperror.ErrUnavailable
	}
	return nil
}

func wrapDBError(statement string, err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return fmt.Errorf("%s: %w", statement, apperror.ErrNotFound)
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return fmt.Errorf("%s: %w", statement, err)
	case errors.Is(err, sql.ErrConnDone), errors.Is(err, sql.ErrTxDone):
		return fmt.Errorf("%s: %w", statement, apperror.ErrDatabase)
	case errors.Is(err, driver.ErrBadConn):
		return fmt.Errorf("%s: %w", statement, apperror.ErrUnavailable)
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if sentinelErr := classifyErrorOnRepository(pqErr.Code); sentinelErr != nil {
			return fmt.Errorf("%s: %w: %w", statement, sentinelErr, err)
		}
	}
	return fmt.Errorf("%s: %w: %w", statement, apperror.ErrDatabase, err)
}
