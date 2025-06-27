/*
Copyright 2023 The Kubernetes Authors.

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

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"

	"k8s.io/kops/pkg/otel/otlptracefile"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context, serviceName, serviceVersion string) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Setup resource.
	res, err := newResource(serviceName, serviceVersion)
	if err != nil {
		handleErr(err)
		return
	}

	// Setup trace provider.
	tracerProvider, err := newTraceProvider(ctx, res)
	if err != nil {
		handleErr(err)
		return
	}
	if tracerProvider != nil {
		shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
		otel.SetTracerProvider(tracerProvider)

		http.DefaultClient = &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		}
	}

	return
}

func newResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		))
}

func newTraceProvider(ctx context.Context, res *resource.Resource) (*trace.TracerProvider, error) {
	destIsDirectory := false

	dest := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_FILE")
	if dest == "" {
		dest = os.Getenv("OTEL_EXPORTER_OTLP_FILE")
	}
	if dest == "" {
		dest = os.Getenv("OTEL_EXPORTER_OTLP_TRACES_DIR")
		if dest != "" {
			destIsDirectory = true
		}
	}
	if dest == "" {
		dest = os.Getenv("OTEL_EXPORTER_OTLP_DIR")
		if dest != "" {
			destIsDirectory = true
		}
	}
	if dest == "" {
		return nil, nil
	}

	// If we are writing to a directory, construct a (likely) unique name
	if destIsDirectory {
		if err := os.MkdirAll(dest, 0755); err != nil {
			return nil, fmt.Errorf("creating directories %q: %w", dest, err)
		}
		processName, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("getting process name: %w", err)
		}
		processName = filepath.Base(processName)
		processName = strings.TrimSuffix(processName, ".exe")
		pid := os.Getpid()
		timestamp := time.Now().UTC().Format(time.RFC3339)
		filename := fmt.Sprintf("%s-%d-%s.otel", processName, pid, timestamp)
		dest = filepath.Join(dest, filename)
	}

	traceExporter, err := otlptracefile.New(ctx, otlptracefile.WithPath(dest))
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
	)
	return traceProvider, nil
}
