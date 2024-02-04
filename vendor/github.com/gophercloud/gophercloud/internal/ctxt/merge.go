// package ctxt implements context merging.
package ctxt

import (
	"context"
	"time"
)

type mergeContext struct {
	context.Context
	ctx2 context.Context
}

// Merge returns a context that is cancelled when at least one of the parents
// is cancelled. The returned context also returns the values of ctx1, or ctx2
// if nil.
func Merge(ctx1, ctx2 context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := WithCancelCause(ctx1)
	stop := AfterFunc(ctx2, func() {
		cancel(Cause(ctx2))
	})

	return &mergeContext{
			Context: ctx,
			ctx2:    ctx2,
		}, func() {
			stop()
			cancel(context.Canceled)
		}
}

// Value returns ctx2's value if ctx's is nil.
func (ctx *mergeContext) Value(key interface{}) interface{} {
	if v := ctx.Context.Value(key); v != nil {
		return v
	}
	return ctx.ctx2.Value(key)
}

// Deadline returns the earlier deadline of the two parents of ctx.
func (ctx *mergeContext) Deadline() (time.Time, bool) {
	if d1, ok := ctx.Context.Deadline(); ok {
		if d2, ok := ctx.ctx2.Deadline(); ok {
			if d1.Before(d2) {
				return d1, true
			}
			return d2, true
		}
		return d1, ok
	}
	return ctx.ctx2.Deadline()
}
