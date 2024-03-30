package customizations

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/internal/v4a"
	"github.com/aws/smithy-go/middleware"
)

type signerVersionKey struct{}

// GetSignerVersion retrieves the signer version to use for signing
//
// Scoped to stack values. Use github.com/aws/smithy-go/middleware#ClearStackValues
// to clear all stack values.
func GetSignerVersion(ctx context.Context) (v string) {
	v, _ = middleware.GetStackValue(ctx, signerVersionKey{}).(string)
	return v
}

// SetSignerVersion sets the signer version to be used for signing the request
//
// Scoped to stack values. Use github.com/aws/smithy-go/middleware#ClearStackValues
// to clear all stack values.
func SetSignerVersion(ctx context.Context, version string) context.Context {
	return middleware.WithStackValue(ctx, signerVersionKey{}, version)
}

// SignHTTPRequestMiddlewareOptions is the configuration options for the SignHTTPRequestMiddleware middleware.
type SignHTTPRequestMiddlewareOptions struct {
	// credential provider
	CredentialsProvider aws.CredentialsProvider

	// log signing
	LogSigning bool

	// v4 signer
	V4Signer v4.HTTPSigner

	//v4a signer
	V4aSigner v4a.HTTPSigner
}

// NewSignHTTPRequestMiddleware constructs a SignHTTPRequestMiddleware using the given Signer for signing requests
func NewSignHTTPRequestMiddleware(options SignHTTPRequestMiddlewareOptions) *SignHTTPRequestMiddleware {
	return &SignHTTPRequestMiddleware{
		credentialsProvider: options.CredentialsProvider,
		v4Signer:            options.V4Signer,
		v4aSigner:           options.V4aSigner,
		logSigning:          options.LogSigning,
	}
}

// SignHTTPRequestMiddleware is a `FinalizeMiddleware` implementation to select HTTP Signing method
type SignHTTPRequestMiddleware struct {

	// credential provider
	credentialsProvider aws.CredentialsProvider

	// log signing
	logSigning bool

	// v4 signer
	v4Signer v4.HTTPSigner

	//v4a signer
	v4aSigner v4a.HTTPSigner
}

// ID is the SignHTTPRequestMiddleware identifier
func (s *SignHTTPRequestMiddleware) ID() string {
	return "Signing"
}

// HandleFinalize will take the provided input and sign the request using the SigV4 authentication scheme
func (s *SignHTTPRequestMiddleware) HandleFinalize(ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler) (
	out middleware.FinalizeOutput, metadata middleware.Metadata, err error,
) {
	// fetch signer type from context
	signerVersion := GetSignerVersion(ctx)
	// SigV4a
	if strings.EqualFold(signerVersion, v4a.Version) {
		v4aCredentialProvider, ok := s.credentialsProvider.(v4a.CredentialsProvider)
		if !ok {
			return out, metadata, fmt.Errorf("invalid credential-provider provided for sigV4a Signer")
		}

		mw := v4a.NewSignHTTPRequestMiddleware(v4a.SignHTTPRequestMiddlewareOptions{
			Credentials: v4aCredentialProvider,
			Signer:      s.v4aSigner,
			LogSigning:  s.logSigning,
		})
		return mw.HandleFinalize(ctx, in, next)
	}
	// SigV4
	mw := v4.NewSignHTTPRequestMiddleware(v4.SignHTTPRequestMiddlewareOptions{
		CredentialsProvider: s.credentialsProvider,
		Signer:              s.v4Signer,
		LogSigning:          s.logSigning,
	})
	return mw.HandleFinalize(ctx, in, next)
}

// RegisterSigningMiddleware registers the wrapper signing middleware to the stack. If a signing middleware is already
// present, this provided middleware will be swapped. Otherwise the middleware will be added at the tail of the
// finalize step.
func RegisterSigningMiddleware(stack *middleware.Stack, signingMiddleware *SignHTTPRequestMiddleware) (err error) {
	const signedID = "Signing"
	_, present := stack.Finalize.Get(signedID)
	if present {
		_, err = stack.Finalize.Swap(signedID, signingMiddleware)
	} else {
		err = stack.Finalize.Add(signingMiddleware, middleware.After)
	}
	return err
}
