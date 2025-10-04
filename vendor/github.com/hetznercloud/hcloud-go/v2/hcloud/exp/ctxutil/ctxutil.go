package ctxutil

import (
	"context"
	"strings"
)

// key is an unexported type to prevents collisions with keys defined in other packages.
type key struct{}

// opPathKey is the key for operation path in Contexts.
var opPathKey = key{}

// SetOpPath processes the operation path and save it in the context before returning it.
//
// Experimental: `exp` package is experimental, breaking changes may occur within minor releases.
func SetOpPath(ctx context.Context, path string) context.Context {
	path, _, _ = strings.Cut(path, "?")
	path = strings.ReplaceAll(path, "%d", "-")
	path = strings.ReplaceAll(path, "%s", "-")

	return context.WithValue(ctx, opPathKey, path)
}

// OpPath returns the operation path from the context.
//
// Experimental: `exp` package is experimental, breaking changes may occur within minor releases.
func OpPath(ctx context.Context) string {
	result, ok := ctx.Value(opPathKey).(string)
	if !ok {
		return ""
	}
	return result
}
