/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package aws

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/middleware"
	"github.com/aws/smithy-go/transport/http"
	"k8s.io/klog/v2"
)

// Middleware for AWS SDK Go V2 clients. Logs requests at the Finalize stage.
func awsHandlerLoggerMiddleware() middleware.FinalizeMiddleware {
	return middleware.FinalizeMiddlewareFunc(
		"k8s/logger",
		func(ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler) (
			out middleware.FinalizeOutput, metadata middleware.Metadata, err error,
		) {
			service, name := awsServiceAndName(ctx)

			klog.V(4).Infof("AWS request: %s %s", service, name)
			return next.HandleFinalize(ctx, in)
		},
	)
}

// Logs details about the response at the Deserialization stage
func awsValidateResponseHandlerLoggerMiddleware() middleware.DeserializeMiddleware {
	return middleware.DeserializeMiddlewareFunc(
		"k8s/api-validate-response",
		func(ctx context.Context, in middleware.DeserializeInput, next middleware.DeserializeHandler) (
			out middleware.DeserializeOutput, metadata middleware.Metadata, err error,
		) {
			out, metadata, err = next.HandleDeserialize(ctx, in)
			response, ok := out.RawResponse.(*http.Response)
			if !ok {
				return out, metadata, &smithy.DeserializationError{Err: fmt.Errorf("unknown transport type %T", out.RawResponse)}
			}
			service, name := awsServiceAndName(ctx)
			klog.V(4).Infof("AWS API ValidateResponse: %s %s %d", service, name, response.StatusCode)
			return out, metadata, err
		},
	)
}

// Logs details about the request at the Serialize stage
func awsSendHandlerLoggerMiddleware() middleware.SerializeMiddleware {
	return middleware.SerializeMiddlewareFunc(
		"k8s/api-request",
		func(ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler) (
			out middleware.SerializeOutput, metadata middleware.Metadata, err error,
		) {
			service, name := awsServiceAndName(ctx)
			klog.V(4).Infof("AWS API Send: %s %s %v", service, name, in.Parameters)
			return next.HandleSerialize(ctx, in)
		},
	)
}

// Gets the service and operation name from AWS SDK Go V2 client requests.
func awsServiceAndName(ctx context.Context) (string, string) {
	service := middleware.GetServiceID(ctx)

	name := "?"
	if opName := middleware.GetOperationName(ctx); opName != "" {
		name = opName
	}
	return service, name
}
