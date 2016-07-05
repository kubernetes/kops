// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// AUTO-GENERATED DOCUMENTATION AND SERVICE

package logging_test

import (
	gax "github.com/googleapis/gax-go"
	google_api "github.com/googleapis/proto-client-go/api"
	google_logging_v2 "github.com/googleapis/proto-client-go/logging/v2"
	"golang.org/x/net/context"
	"google.golang.org/cloud/logging/apiv2/logging"
)

func ExampleNewClient() {
	ctx := context.Background()
	opts := []gax.ClientOption{ /* Optional client parameters. */ }
	c, err := logging.NewClient(ctx, opts...)
	_, _ = c, err // Handle error.
}

func ExampleClient_DeleteLog() {
	ctx := context.Background()
	c, err := logging.NewClient(ctx)
	_ = err // Handle error.

	req := &google_logging_v2.DeleteLogRequest{ /* Data... */ }
	err = c.DeleteLog(ctx, req)
	_ = err // Handle error.
}

func ExampleClient_WriteLogEntries() {
	ctx := context.Background()
	c, err := logging.NewClient(ctx)
	_ = err // Handle error.

	req := &google_logging_v2.WriteLogEntriesRequest{ /* Data... */ }
	var resp *google_logging_v2.WriteLogEntriesResponse
	resp, err = c.WriteLogEntries(ctx, req)
	_, _ = resp, err // Handle error.
}

func ExampleClient_ListLogEntries() {
	ctx := context.Background()
	c, err := logging.NewClient(ctx)
	_ = err // Handle error.

	req := &google_logging_v2.ListLogEntriesRequest{ /* Data... */ }
	it := c.ListLogEntries(ctx, req)
	var resp *google_logging_v2.LogEntry
	for {
		resp, err = it.Next()
		if err != nil {
			break
		}
	}
	_ = resp
}

func ExampleClient_ListMonitoredResourceDescriptors() {
	ctx := context.Background()
	c, err := logging.NewClient(ctx)
	_ = err // Handle error.

	req := &google_logging_v2.ListMonitoredResourceDescriptorsRequest{ /* Data... */ }
	it := c.ListMonitoredResourceDescriptors(ctx, req)
	var resp *google_api.MonitoredResourceDescriptor
	for {
		resp, err = it.Next()
		if err != nil {
			break
		}
	}
	_ = resp
}
