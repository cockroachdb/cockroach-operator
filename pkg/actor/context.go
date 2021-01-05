/*
Copyright 2021 The Cockroach Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package actor

import (
	"context"
	"errors"
)

type cancelFuncKey struct{}

//ContextWithCancelFn func
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

//CancelLoop func
func CancelLoop(ctx context.Context) {
	getCancelFn(ctx)()
}
