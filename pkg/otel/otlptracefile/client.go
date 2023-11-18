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

package otlptracefile

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

type client struct {
	cfg Config

	writerMutex sync.RWMutex
	writer      *writer
}

var _ otlptrace.Client = (*client)(nil)

// newClient constructs a client.
func newClient(opts ...Option) *client {
	var cfg Config
	for _, option := range opts {
		option(&cfg)
	}

	c := &client{
		cfg: cfg,
	}

	return c
}

// Start implements otlptrace.Client.
func (c *client) Start(ctx context.Context) error {
	c.writerMutex.Lock()
	defer c.writerMutex.Unlock()

	if c.writer != nil {
		return fmt.Errorf("already started")
	}

	w, err := newWriter(c.cfg)
	if err != nil {
		return err
	}
	c.writer = w

	return nil
}

// Stop implements otlptrace.Client.
func (c *client) Stop(ctx context.Context) error {
	c.writerMutex.Lock()
	defer c.writerMutex.Unlock()

	if c.writer != nil {
		err := c.writer.Close()
		if err != nil {
			return err
		}
		c.writer = nil
	}

	return nil
}

var errShutdown = errors.New("the client is shutdown")

// UploadTraces implements otlptrace.Client.
func (c *client) UploadTraces(ctx context.Context, protoSpans []*tracepb.ResourceSpans) error {
	c.writerMutex.RLock()
	defer c.writerMutex.RUnlock()

	if c.writer == nil {
		return errShutdown
	}

	return c.writer.writeTraces(ctx, &coltracepb.ExportTraceServiceRequest{
		ResourceSpans: protoSpans,
	})
}

// MarshalLog is the marshaling function used by the logging system to represent this Client.
func (c *client) MarshalLog() interface{} {
	return struct {
		Type string
	}{
		Type: "otlptracefile",
	}
}
