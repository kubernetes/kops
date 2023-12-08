/*
Copyright 2023 Google LLC

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

package cloud

import (
	"context"
	"fmt"
)

// CallObserver is called between the start and end of the operation.
type CallObserver interface {
	// Start event for a call to the API. This happens before the
	// RateLimiter.Accept().
	Start(ctx context.Context, key *RateLimitKey)

	// End event for a call to the API. This happens after the
	// operation is finished.
	End(ctx context.Context, key *RateLimitKey, err error)
}

type contextKey string

var callObserverContextKey = contextKey("call observer")

// WithCallObserver adds a CallObserver that will be called on the
// operation being called.
//
//	type obs struct{}
//	func (o *obs) Start(...) { fmt.Println("start") }
//	func (o *obs) End(...) { fmt.Println("end") }
//
//	ctx := WithCallObserver(ctx, &obs{})
//	g.Addresses.Insert(ctx, ...)
func WithCallObserver(ctx context.Context, obs CallObserver) context.Context {
	return context.WithValue(ctx, callObserverContextKey, obs)
}

func callObserverStart(ctx context.Context, key *CallContextKey) {
	obj := ctx.Value(callObserverContextKey)
	if obj == nil {
		return
	}
	co, ok := obj.(CallObserver)
	if !ok {
		panic(fmt.Sprintf("expected CallObserver, got %T", obj))
	}
	co.Start(ctx, key)
}

func callObserverEnd(ctx context.Context, key *CallContextKey, err error) {
	obj := ctx.Value(callObserverContextKey)
	if obj == nil {
		return
	}
	co, ok := obj.(CallObserver)
	if !ok {
		panic(fmt.Sprintf("expected CallObserver, got %T", obj))
	}
	co.End(ctx, key, err)
}
