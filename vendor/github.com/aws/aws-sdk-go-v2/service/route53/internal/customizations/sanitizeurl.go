package customizations

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// AddSanitizeURLMiddlewareOptions provides the options for Route53SanitizeURL middleware setup
type AddSanitizeURLMiddlewareOptions struct {
	// functional pointer to sanitize hosted zone id member
	// The function is intended to take an input value,
	// look for hosted zone id input member and sanitize the value
	// to strip out an excess `/hostedzone/` prefix that can be present in
	// the hosted zone id input member.
	//
	// returns an error if any.
	SanitizeURLInput func(interface{}) error
}

// AddSanitizeURLMiddleware add the middleware necessary to modify Route53 input before op serialization.
func AddSanitizeURLMiddleware(stack *middleware.Stack, options AddSanitizeURLMiddlewareOptions) error {
	return stack.Serialize.Insert(&sanitizeURL{
		sanitizeURLInput: options.SanitizeURLInput,
	}, "OperationSerializer", middleware.Before)
}

// sanitizeURL cleans up potential formatting issues in the Route53 path.
//
// Notably it will strip out an excess `/hostedzone/` prefix that can be present in
// the hosted zone id input member. That excess prefix is there because some route53 apis return
// the id in that format, so this middleware enables round-tripping those values.
type sanitizeURL struct {
	sanitizeURLInput func(interface{}) error
}

// ID returns the id for the middleware.
func (*sanitizeURL) ID() string {
	return "Route53:SanitizeURL"
}

// HandleSerialize implements the SerializeMiddleware interface.
func (m *sanitizeURL) HandleSerialize(
	ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler,
) (
	out middleware.SerializeOutput, metadata middleware.Metadata, err error,
) {
	_, ok := in.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, &smithy.SerializationError{
			Err: fmt.Errorf("unknown request type %T", in.Request),
		}
	}

	if err := m.sanitizeURLInput(in.Parameters); err != nil {
		return out, metadata, err
	}

	return next.HandleSerialize(ctx, in)
}
