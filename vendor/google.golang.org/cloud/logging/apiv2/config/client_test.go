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

package config_test

import (
	gax "github.com/googleapis/gax-go"
	google_logging_v2 "github.com/googleapis/proto-client-go/logging/v2"
	"golang.org/x/net/context"
	"google.golang.org/cloud/logging/apiv2/config"
)

func ExampleNewClient() {
	ctx := context.Background()
	opts := []gax.ClientOption{ /* Optional client parameters. */ }
	c, err := config.NewClient(ctx, opts...)
	_, _ = c, err // Handle error.
}

func ExampleClient_ListSinks() {
	ctx := context.Background()
	c, err := config.NewClient(ctx)
	_ = err // Handle error.

	req := &google_logging_v2.ListSinksRequest{ /* Data... */ }
	it := c.ListSinks(ctx, req)
	var resp *google_logging_v2.LogSink
	for {
		resp, err = it.Next()
		if err != nil {
			break
		}
	}
	_ = resp
}

func ExampleClient_GetSink() {
	ctx := context.Background()
	c, err := config.NewClient(ctx)
	_ = err // Handle error.

	req := &google_logging_v2.GetSinkRequest{ /* Data... */ }
	var resp *google_logging_v2.LogSink
	resp, err = c.GetSink(ctx, req)
	_, _ = resp, err // Handle error.
}

func ExampleClient_CreateSink() {
	ctx := context.Background()
	c, err := config.NewClient(ctx)
	_ = err // Handle error.

	req := &google_logging_v2.CreateSinkRequest{ /* Data... */ }
	var resp *google_logging_v2.LogSink
	resp, err = c.CreateSink(ctx, req)
	_, _ = resp, err // Handle error.
}

func ExampleClient_UpdateSink() {
	ctx := context.Background()
	c, err := config.NewClient(ctx)
	_ = err // Handle error.

	req := &google_logging_v2.UpdateSinkRequest{ /* Data... */ }
	var resp *google_logging_v2.LogSink
	resp, err = c.UpdateSink(ctx, req)
	_, _ = resp, err // Handle error.
}

func ExampleClient_DeleteSink() {
	ctx := context.Background()
	c, err := config.NewClient(ctx)
	_ = err // Handle error.

	req := &google_logging_v2.DeleteSinkRequest{ /* Data... */ }
	err = c.DeleteSink(ctx, req)
	_ = err // Handle error.
}
