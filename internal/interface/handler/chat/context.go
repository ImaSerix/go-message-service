package chat

import (
	"context"

	"github.com/google/uuid"
)

type chatIDKeyType struct{}
type afterParamKeyType struct{}
type beforeParamKeyType struct{}
type limitParamKeyType struct{}

var (
	chatIDKey = chatIDKeyType{}
	afterKey  = afterParamKeyType{}
	beforeKey = beforeParamKeyType{}
	limitKey  = limitParamKeyType{}
)

func WithChatID(ctx context.Context, chatID uuid.UUID) context.Context {
	return context.WithValue(ctx, chatIDKey, chatID)
}

func ChatID(ctx context.Context) (uuid.UUID, bool) {
	value, ok := ctx.Value(chatIDKey).(uuid.UUID)
	return value, ok
}
