/*
Copyright 2024 The Kubernetes Authors.

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

package awslog

import (
	"context"

	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go/middleware"
	"k8s.io/klog/v2"
)

type awsLogger struct{}

var _ middleware.FinalizeMiddleware = (*awsLogger)(nil)

func (*awsLogger) ID() string {
	return "kops/logger"
}

func (*awsLogger) HandleFinalize(ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler) (
	out middleware.FinalizeOutput, metadata middleware.Metadata, err error,
) {
	service := awsmiddleware.GetServiceID(ctx)
	name := "?"
	if n := awsmiddleware.GetOperationName(ctx); n != "" {
		name = n
	}
	klog.V(4).Infof("AWS request: %s %s", service, name)

	out, metadata, err = next.HandleFinalize(ctx, in)
	// the entire operation was invoked above - the deserialized response is
	// available opaquely in out.Result, run post-op actions here...
	return out, metadata, err
}

// WithAWSLogger adds middleware to aws-sdk-go-v2/config that logs AWS requests
func WithAWSLogger() func(*config.LoadOptions) error {
	return func(lo *config.LoadOptions) error {
		lo.APIOptions = append(lo.APIOptions, func(s *middleware.Stack) error {
			return s.Finalize.Add(&awsLogger{}, middleware.After)
		})
		return nil
	}
}
