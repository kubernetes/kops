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

package config

import (
	"errors"
	"fmt"
	"runtime"
	"time"

	gax "github.com/googleapis/gax-go"
	google_logging_v2 "github.com/googleapis/proto-client-go/logging/v2"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

const (
	gapicNameVersion = "gapic/0.1.0"
)

var (
	// Done is returned by iterators on successful completion.
	Done = errors.New("iterator done")

	projectPathTemplate = gax.MustCompilePathTemplate("projects/{project}")
	sinkPathTemplate    = gax.MustCompilePathTemplate("projects/{project}/sinks/{sink}")
)

func defaultClientSettings() gax.ClientSettings {
	return gax.ClientSettings{
		AppName:    "gax",
		AppVersion: gax.Version,
		Endpoint:   "logging.googleapis.com:443",
		Scopes: []string{
			"https://www.googleapis.com/auth/logging.write",
			"https://www.googleapis.com/auth/logging.admin",
			"https://www.googleapis.com/auth/logging.read",
			"https://www.googleapis.com/auth/cloud-platform.read-only",
			"https://www.googleapis.com/auth/cloud-platform",
		},
		CallOptions: map[string][]gax.CallOption{
			"ListSinks":  append([]gax.CallOption{withIdempotentRetryCodes()}, defaultRetryOptions()...),
			"GetSink":    append([]gax.CallOption{withIdempotentRetryCodes()}, defaultRetryOptions()...),
			"CreateSink": append([]gax.CallOption{withNonIdempotentRetryCodes()}, defaultRetryOptions()...),
			"UpdateSink": append([]gax.CallOption{withNonIdempotentRetryCodes()}, defaultRetryOptions()...),
			"DeleteSink": append([]gax.CallOption{withIdempotentRetryCodes()}, defaultRetryOptions()...),
		},
	}
}

func withIdempotentRetryCodes() gax.CallOption {
	return gax.WithRetryCodes([]codes.Code{
		codes.DeadlineExceeded,
		codes.Unavailable,
	})
}

func withNonIdempotentRetryCodes() gax.CallOption {
	return gax.WithRetryCodes([]codes.Code{})
}

func defaultRetryOptions() []gax.CallOption {
	return []gax.CallOption{
		gax.WithTimeout(45000 * time.Millisecond),
		gax.WithDelayTimeoutSettings(100*time.Millisecond, 1000*time.Millisecond, 1.2),
		gax.WithRPCTimeoutSettings(2000*time.Millisecond, 30000*time.Millisecond, 1.5),
	}
}

// Client is a client for interacting with ConfigServiceV2.
type Client struct {
	// The connection to the service.
	conn *grpc.ClientConn

	// The gRPC API client.
	client google_logging_v2.ConfigServiceV2Client

	// The map from the method name to the default call options for the method of this service.
	callOptions map[string][]gax.CallOption

	// The metadata to be sent with each request.
	metadata map[string][]string
}

// NewClient creates a new API service client.
func NewClient(ctx context.Context, opts ...gax.ClientOption) (*Client, error) {
	s := defaultClientSettings()
	for _, opt := range opts {
		opt.Resolve(&s)
	}
	conn, err := gax.DialGRPC(ctx, s)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:        conn,
		client:      google_logging_v2.NewConfigServiceV2Client(conn),
		callOptions: s.CallOptions,
		metadata: map[string][]string{
			"x-goog-api-client": []string{fmt.Sprintf("%s/%s %s gax/%s go/%s", s.AppName, s.AppVersion, gapicNameVersion, gax.Version, runtime.Version())},
		},
	}, nil
}

// Close closes the connection to the API service. The user should invoke this when
// the client is no longer required.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Path templates.

// ProjectPath returns the path for the project resource.
func ProjectPath(project string) string {
	path, err := projectPathTemplate.Render(map[string]string{
		"project": project,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// SinkPath returns the path for the sink resource.
func SinkPath(project string, sink string) string {
	path, err := sinkPathTemplate.Render(map[string]string{
		"project": project,
		"sink":    sink,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// AUTO-GENERATED DOCUMENTATION AND METHOD -- see instructions at the top of the file for editing.

// ListSinks lists sinks.
func (c *Client) ListSinks(ctx context.Context, req *google_logging_v2.ListSinksRequest) *LogSinkIterator {
	ctx = metadata.NewContext(ctx, c.metadata)
	it := &LogSinkIterator{}
	it.apiCall = func() error {
		if it.atLastPage {
			return Done
		}
		var resp *google_logging_v2.ListSinksResponse
		err := gax.Invoke(ctx, func(ctx context.Context) error {
			var err error
			req.PageToken = it.nextPageToken
			req.PageSize = it.pageSize
			resp, err = c.client.ListSinks(ctx, req)
			return err
		}, c.callOptions["ListSinks"]...)
		if err != nil {
			return err
		}
		if resp.NextPageToken == "" {
			it.atLastPage = true
		} else {
			it.nextPageToken = resp.NextPageToken
		}
		it.items = resp.Sinks
		return nil
	}
	return it
}

// AUTO-GENERATED DOCUMENTATION AND METHOD -- see instructions at the top of the file for editing.

// GetSink gets a sink.
func (c *Client) GetSink(ctx context.Context, req *google_logging_v2.GetSinkRequest) (*google_logging_v2.LogSink, error) {
	ctx = metadata.NewContext(ctx, c.metadata)
	var resp *google_logging_v2.LogSink
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.GetSink(ctx, req)
		return err
	}, c.callOptions["GetSink"]...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// AUTO-GENERATED DOCUMENTATION AND METHOD -- see instructions at the top of the file for editing.

// CreateSink creates a sink.
func (c *Client) CreateSink(ctx context.Context, req *google_logging_v2.CreateSinkRequest) (*google_logging_v2.LogSink, error) {
	ctx = metadata.NewContext(ctx, c.metadata)
	var resp *google_logging_v2.LogSink
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.CreateSink(ctx, req)
		return err
	}, c.callOptions["CreateSink"]...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// AUTO-GENERATED DOCUMENTATION AND METHOD -- see instructions at the top of the file for editing.

// UpdateSink creates or updates a sink.
func (c *Client) UpdateSink(ctx context.Context, req *google_logging_v2.UpdateSinkRequest) (*google_logging_v2.LogSink, error) {
	ctx = metadata.NewContext(ctx, c.metadata)
	var resp *google_logging_v2.LogSink
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.UpdateSink(ctx, req)
		return err
	}, c.callOptions["UpdateSink"]...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// AUTO-GENERATED DOCUMENTATION AND METHOD -- see instructions at the top of the file for editing.

// DeleteSink deletes a sink.
func (c *Client) DeleteSink(ctx context.Context, req *google_logging_v2.DeleteSinkRequest) error {
	ctx = metadata.NewContext(ctx, c.metadata)
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		_, err = c.client.DeleteSink(ctx, req)
		return err
	}, c.callOptions["DeleteSink"]...)
	return err
}

// Iterators.
//

// LogSinkIterator manages a stream of *google_logging_v2.LogSink.
type LogSinkIterator struct {
	// The current page data.
	items         []*google_logging_v2.LogSink
	atLastPage    bool
	currentIndex  int
	pageSize      int32
	nextPageToken string
	apiCall       func() error
}

// NextPage moves to the next page and updates its internal data.
// It returns Done if no more pages exist.
func (it *LogSinkIterator) NextPage() ([]*google_logging_v2.LogSink, error) {
	err := it.apiCall()
	if err != nil {
		return nil, err
	}
	return it.items, err
}

// Next returns the next element in the stream. It returns Done at
// the end of the stream.
func (it *LogSinkIterator) Next() (*google_logging_v2.LogSink, error) {
	for it.currentIndex >= len(it.items) {
		_, err := it.NextPage()
		if err != nil {
			return nil, err
		}
		it.currentIndex = 0
	}
	result := it.items[it.currentIndex]
	it.currentIndex++
	return result, nil
}

// SetPageSize sets the maximum size of the next page to be
// retrieved.
func (it *LogSinkIterator) SetPageSize(pageSize int32) {
	it.pageSize = pageSize
}

// SetPageToken sets the next page token to be retrieved. Note, it
// does not retrieve the next page, or modify the cached page. If
// Next is called, there is no guarantee that the result returned
// will be from the next page until NextPage is called.
func (it *LogSinkIterator) SetPageToken(token string) {
	it.nextPageToken = token
}

// NextPageToken returns the next page token.
func (it *LogSinkIterator) NextPageToken() string {
	return it.nextPageToken
}
