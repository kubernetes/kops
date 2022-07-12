package scw

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/internal/auth"
)

// RequestOption is a function that applies options to a ScalewayRequest.
type RequestOption func(*ScalewayRequest)

// WithContext request option sets the context of a ScalewayRequest
func WithContext(ctx context.Context) RequestOption {
	return func(s *ScalewayRequest) {
		s.ctx = ctx
	}
}

// WithAllPages aggregate all pages in the response of a List request.
// Will error when pagination is not supported on the request.
func WithAllPages() RequestOption {
	return func(s *ScalewayRequest) {
		s.allPages = true
	}
}

// WithAuthRequest overwrites the client access key and secret key used in the request.
func WithAuthRequest(accessKey, secretKey string) RequestOption {
	return func(s *ScalewayRequest) {
		s.auth = auth.NewToken(accessKey, secretKey)
	}
}
