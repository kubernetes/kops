/*
Copyright 2019 The Kubernetes Authors.

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

package main // import "k8s.io/kops/cmd/kops"

import (
	"context"
	"fmt"
	"os"

	"k8s.io/kops"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Set up OpenTelemetry.
	serviceName := "kops"
	serviceVersion := kops.Version
	if kops.GitVersion != "" {
		serviceVersion += ".git-" + kops.GitVersion
	}

	otelShutdown, err := setupOTelSDK(ctx, serviceName, serviceVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		// We use a background context because the main context has probably been shut down.
		if err := otelShutdown(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down otel: %v\n", err)
		}
	}()

	if err := Execute(ctx); err != nil {
		return err
	}

	return nil
}
