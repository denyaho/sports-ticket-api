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

func classifyErrorOnRepository(code pq.ErrorCode) error {
	switch code {
	case "23505":
		return apperror.ErrConflict
	case "23503", "23514":
		return apperror.ErrValidation
	case "40001", "40P01", "55P03":
		return apperror.ErrRetryable
	case "57014":
		return apperror.ErrTimeout
	case "53300", "08006", "08003":
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
		return fmt.Errorf("%s: %w", statement, apperror.ErrUnavailable)
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
