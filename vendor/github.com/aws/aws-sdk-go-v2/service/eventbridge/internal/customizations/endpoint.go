package customizations

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/internal/v4a"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/internal/endpoints"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// EndpointResolver interface for resolving service endpoints.
type EndpointResolver interface {
	ResolveEndpoint(region string, options endpoints.Options) (aws.Endpoint, error)
}

// UpdateEndpointOptions provides configuration options for the UpdateEndpoint middleware.
type UpdateEndpointOptions struct {
	GetEndpointIDFromInput  func(interface{}) (*string, bool)
	EndpointResolver        EndpointResolver
	EndpointResolverOptions endpoints.Options
}

// UpdateEndpoint is a middleware that handles routing an EventBridge operation to a multi-region endpoint.
func UpdateEndpoint(stack *middleware.Stack, options UpdateEndpointOptions) error {
	const serializerID = "OperationSerializer"

	return stack.Serialize.Insert(&updateEndpoint{
		getEndpointIDFromInput:  options.GetEndpointIDFromInput,
		endpointResolver:        options.EndpointResolver,
		endpointResolverOptions: options.EndpointResolverOptions,
	}, serializerID, middleware.Before)
}

type updateEndpoint struct {
	getEndpointIDFromInput  func(interface{}) (*string, bool)
	endpointResolver        EndpointResolver
	endpointResolverOptions endpoints.Options
}

func (u *updateEndpoint) ID() string {
	return "EventBridge:UpdateEndpoint"
}

func (u *updateEndpoint) HandleSerialize(ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler) (out middleware.SerializeOutput, metadata middleware.Metadata, err error) {
	if !awsmiddleware.GetRequiresLegacyEndpoints(ctx) {
		return next.HandleSerialize(ctx, in)
	}

	// If getEndpointIDFromInput is nil but the middleware got attached just skip to the next handler
	if u.getEndpointIDFromInput == nil {
		return next.HandleSerialize(ctx, in)
	}

	value, ok := u.getEndpointIDFromInput(in.Parameters)
	if !ok || value == nil {
		return next.HandleSerialize(ctx, in)
	}

	req, ok := in.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, fmt.Errorf("unknown transport type %T", req)
	}

	endpointID := aws.ToString(value)

	if len(endpointID) == 0 {
		return out, metadata, &smithy.SerializationError{
			Err: fmt.Errorf("EndpointId must not be a zero-length string"),
		}
	}

	if u.endpointResolverOptions.UseFIPSEndpoint == aws.FIPSEndpointStateEnabled {
		return out, metadata, &smithy.SerializationError{
			Err: fmt.Errorf("EventBridge multi-region endpoints do not support FIPS endpoint configuration"),
		}
	}

	labels := strings.Split(endpointID, ".")

	for _, label := range labels {
		if !smithyhttp.ValidHostLabel(label) {
			return out, metadata, &smithy.SerializationError{
				Err: fmt.Errorf("EndpointId is not a valid host label, %s", endpointID),
			}
		}
	}

	region := awsmiddleware.GetRegion(ctx)

	endpoint, err := u.endpointResolver.ResolveEndpoint(region, u.endpointResolverOptions)
	if err != nil {
		return out, metadata, &smithy.SerializationError{
			Err: err,
		}
	}

	if len(endpoint.SigningRegion) > 0 {
		region = endpoint.SigningRegion
	}

	// set signing region and version for MRAP
	endpoint.SigningRegion = "*"
	ctx = awsmiddleware.SetSigningRegion(ctx, endpoint.SigningRegion)
	ctx = SetSignerVersion(ctx, v4a.Version)

	if len(endpoint.SigningName) != 0 {
		ctx = awsmiddleware.SetSigningName(ctx, endpoint.SigningName)
	}

	if endpoint.Source == aws.EndpointSourceCustom {
		return next.HandleSerialize(ctx, in)
	}

	dnsSuffix, err := endpoints.GetDNSSuffixFromRegion(region, u.endpointResolverOptions)
	if err != nil {
		return out, metadata, &smithy.SerializationError{
			Err: err,
		}
	}

	// modify endpoint host to use s3-global host prefix
	scheme := strings.SplitN(endpoint.URL, "://", 2)

	// set url as per partition
	endpoint.URL = scheme[0] + "://" + endpointID + ".endpoint.events." + dnsSuffix

	// assign resolved endpoint url to request url
	req.URL, err = url.Parse(endpoint.URL)
	if err != nil {
		return out, metadata, &smithy.SerializationError{
			Err: fmt.Errorf("failed to parse endpoint URL: %w", err),
		}
	}

	return next.HandleSerialize(ctx, in)
}
