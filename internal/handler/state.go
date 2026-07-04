package handler

import (
	"context"
	"github.com/google/uuid"
)


func SetUserIDInRequestState(ctx context.Context, userID uuid.UUID) {
	if state, ok := ctx.Value(requestStateKey).(*RequestState); ok {
		state.UserID = userID
	}
}

func SetErrorInRequestState(ctx context.Context, err error) {
	if state, ok := ctx.Value(requestStateKey).(*RequestState); ok {
		state.Err = err
	}
}