package message

import (
	"context"
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

func WithAfterParam(ctx context.Context, after int64) context.Context {
	return context.WithValue(ctx, afterKey, after)
}

func AfterParam(ctx context.Context) (int64, bool) {
	value, ok := ctx.Value(afterKey).(int64)
	return value, ok
}

func WithBeforeParam(ctx context.Context, before int64) context.Context {
	return context.WithValue(ctx, beforeKey, before)
}

func BeforeParam(ctx context.Context) (int64, bool) {
	value, ok := ctx.Value(beforeKey).(int64)
	return value, ok
}

func WithLimitParam(ctx context.Context, limit int) context.Context {
	return context.WithValue(ctx, limitKey, limit)
}

func LimitParam(ctx context.Context) (int, bool) {
	value, ok := ctx.Value(limitKey).(int)
	return value, ok
}
