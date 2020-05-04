package actor

import (
	"context"
	"errors"
)

type cancelFuncKey struct{}

func ContextWithCancelFn(ctx context.Context, fn context.CancelFunc) context.Context {
	return context.WithValue(ctx, cancelFuncKey{}, fn)
}

func getCancelFn(ctx context.Context) context.CancelFunc {
	f, ok := ctx.Value(cancelFuncKey{}).(context.CancelFunc)

	if f == nil || !ok {
		return func() {
			Log.Error(errors.New("missing parent cancel function in context"), "")
		}
	}

	return f
}

func CancelLoop(ctx context.Context) {
	getCancelFn(ctx)()
}
