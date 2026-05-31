package auth

import (
	"context"

	"github.com/google/uuid"
)

type userKeyType struct{}

var (
	userKey = userKeyType{}
)

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userKey, userID)
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	value, ok := ctx.Value(userKey).(uuid.UUID)
	return value, ok
}
