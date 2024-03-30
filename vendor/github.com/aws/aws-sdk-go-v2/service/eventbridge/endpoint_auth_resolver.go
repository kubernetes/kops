package eventbridge

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	smithy "github.com/aws/smithy-go"
	smithyauth "github.com/aws/smithy-go/auth"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

type endpointAuthResolver struct {
	EndpointResolver EndpointResolverV2
}

var _ AuthSchemeResolver = (*endpointAuthResolver)(nil)

func (r *endpointAuthResolver) ResolveAuthSchemes(
	ctx context.Context, params *AuthResolverParameters,
) (
	[]*smithyauth.Option, error,
) {
	if params.endpointParams.Region == nil {
		// #2502: We're correcting the endpoint binding behavior to treat empty
		// Region as "unset" (nil), but auth resolution technically doesn't
		// care and someone could be using V1 or non-default V2 endpoint
		// resolution, both of which would bypass the required-region check.
		// They shouldn't be broken because the region is technically required
		// by this service's endpoint-based auth resolver, so we stub it here.
		params.endpointParams.Region = aws.String("")
	}

	opts, err := r.resolveAuthSchemes(ctx, params)
	if err != nil {
		return nil, err
	}

	// preserve pre-SRA behavior where everything technically had anonymous
	return append(opts, &smithyauth.Option{
		SchemeID: smithyauth.SchemeIDAnonymous,
	}), nil
}

func (r *endpointAuthResolver) resolveAuthSchemes(
	ctx context.Context, params *AuthResolverParameters,
) (
	[]*smithyauth.Option, error,
) {
	endpt, err := r.EndpointResolver.ResolveEndpoint(ctx, *params.endpointParams)
	if err != nil {
		return nil, fmt.Errorf("resolve endpoint: %w", err)
	}

	if opts, ok := smithyauth.GetAuthOptions(&endpt.Properties); ok {
		return opts, nil
	}

	// endpoint rules didn't specify, fallback to sigv4
	return []*smithyauth.Option{
		{
			SchemeID: smithyauth.SchemeIDSigV4,
			SignerProperties: func() smithy.Properties {
				var props smithy.Properties
				smithyhttp.SetSigV4SigningName(&props, "events")
				smithyhttp.SetSigV4SigningRegion(&props, params.Region)
				return props
			}(),
		},
		{
			SchemeID: smithyauth.SchemeIDSigV4A,
		},
	}, nil
}

func finalizeServiceEndpointAuthResolver(options *Options) {
	if _, ok := options.AuthSchemeResolver.(*defaultAuthSchemeResolver); !ok {
		return
	}

	options.AuthSchemeResolver = &endpointAuthResolver{
		EndpointResolver: options.EndpointResolverV2,
	}
}

func finalizeOperationEndpointAuthResolver(options *Options) {
	resolver, ok := options.AuthSchemeResolver.(*endpointAuthResolver)
	if !ok {
		return
	}

	if resolver.EndpointResolver == options.EndpointResolverV2 {
		return
	}

	options.AuthSchemeResolver = &endpointAuthResolver{
		EndpointResolver: options.EndpointResolverV2,
	}
}
