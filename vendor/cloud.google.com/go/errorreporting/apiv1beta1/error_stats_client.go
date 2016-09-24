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

// AUTO-GENERATED CODE. DO NOT EDIT.

package errorreporting

import (
	"fmt"
	"math"
	"runtime"
	"time"

	gax "github.com/googleapis/gax-go"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	clouderrorreportingpb "google.golang.org/genproto/googleapis/devtools/clouderrorreporting/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

var (
	errorStatsProjectPathTemplate = gax.MustCompilePathTemplate("projects/{project}")
)

// ErrorStatsCallOptions contains the retry settings for each method of this client.
type ErrorStatsCallOptions struct {
	ListGroupStats []gax.CallOption
	ListEvents     []gax.CallOption
	DeleteEvents   []gax.CallOption
}

func defaultErrorStatsClientOptions() []option.ClientOption {
	return []option.ClientOption{
		option.WithEndpoint("clouderrorreporting.googleapis.com:443"),
		option.WithScopes(
			"https://www.googleapis.com/auth/cloud-platform",
		),
	}
}

func defaultErrorStatsCallOptions() *ErrorStatsCallOptions {
	retry := map[[2]string][]gax.CallOption{
		{"default", "idempotent"}: {
			gax.WithRetry(func() gax.Retryer {
				return gax.OnCodes([]codes.Code{
					codes.DeadlineExceeded,
					codes.Unavailable,
				}, gax.Backoff{
					Initial:    100 * time.Millisecond,
					Max:        60000 * time.Millisecond,
					Multiplier: 1.3,
				})
			}),
		},
	}

	return &ErrorStatsCallOptions{
		ListGroupStats: retry[[2]string{"default", "idempotent"}],
		ListEvents:     retry[[2]string{"default", "idempotent"}],
		DeleteEvents:   retry[[2]string{"default", "idempotent"}],
	}
}

// ErrorStatsClient is a client for interacting with ErrorStatsService.
type ErrorStatsClient struct {
	// The connection to the service.
	conn *grpc.ClientConn

	// The gRPC API client.
	client clouderrorreportingpb.ErrorStatsServiceClient

	// The call options for this service.
	CallOptions *ErrorStatsCallOptions

	// The metadata to be sent with each request.
	metadata map[string][]string
}

// NewErrorStatsClient creates a new error_stats service client.
//
// An API for retrieving and managing error statistics as well as data for
// individual events.
func NewErrorStatsClient(ctx context.Context, opts ...option.ClientOption) (*ErrorStatsClient, error) {
	conn, err := transport.DialGRPC(ctx, append(defaultErrorStatsClientOptions(), opts...)...)
	if err != nil {
		return nil, err
	}
	c := &ErrorStatsClient{
		conn:        conn,
		client:      clouderrorreportingpb.NewErrorStatsServiceClient(conn),
		CallOptions: defaultErrorStatsCallOptions(),
	}
	c.SetGoogleClientInfo("gax", gax.Version)
	return c, nil
}

// Connection returns the client's connection to the API service.
func (c *ErrorStatsClient) Connection() *grpc.ClientConn {
	return c.conn
}

// Close closes the connection to the API service. The user should invoke this when
// the client is no longer required.
func (c *ErrorStatsClient) Close() error {
	return c.conn.Close()
}

// SetGoogleClientInfo sets the name and version of the application in
// the `x-goog-api-client` header passed on each request. Intended for
// use by Google-written clients.
func (c *ErrorStatsClient) SetGoogleClientInfo(name, version string) {
	c.metadata = map[string][]string{
		"x-goog-api-client": {fmt.Sprintf("%s/%s %s gax/%s go/%s", name, version, gapicNameVersion, gax.Version, runtime.Version())},
	}
}

// ProjectPath returns the path for the project resource.
func ErrorStatsProjectPath(project string) string {
	path, err := errorStatsProjectPathTemplate.Render(map[string]string{
		"project": project,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// ListGroupStats lists the specified groups.
func (c *ErrorStatsClient) ListGroupStats(ctx context.Context, req *clouderrorreportingpb.ListGroupStatsRequest) *ErrorGroupStatsIterator {
	ctx = metadata.NewContext(ctx, c.metadata)
	it := &ErrorGroupStatsIterator{}
	it.apiCall = func() error {
		var resp *clouderrorreportingpb.ListGroupStatsResponse
		err := gax.Invoke(ctx, func(ctx context.Context) error {
			var err error
			req.PageToken = it.nextPageToken
			req.PageSize = it.pageSize
			resp, err = c.client.ListGroupStats(ctx, req)
			return err
		}, c.CallOptions.ListGroupStats...)
		if err != nil {
			return err
		}
		if resp.NextPageToken == "" {
			it.atLastPage = true
		}
		it.nextPageToken = resp.NextPageToken
		it.items = resp.ErrorGroupStats
		return nil
	}
	return it
}

// ListEvents lists the specified events.
func (c *ErrorStatsClient) ListEvents(ctx context.Context, req *clouderrorreportingpb.ListEventsRequest) *ErrorEventIterator {
	ctx = metadata.NewContext(ctx, c.metadata)
	it := &ErrorEventIterator{}
	it.apiCall = func() error {
		var resp *clouderrorreportingpb.ListEventsResponse
		err := gax.Invoke(ctx, func(ctx context.Context) error {
			var err error
			req.PageToken = it.nextPageToken
			req.PageSize = it.pageSize
			resp, err = c.client.ListEvents(ctx, req)
			return err
		}, c.CallOptions.ListEvents...)
		if err != nil {
			return err
		}
		if resp.NextPageToken == "" {
			it.atLastPage = true
		}
		it.nextPageToken = resp.NextPageToken
		it.items = resp.ErrorEvents
		return nil
	}
	return it
}

// DeleteEvents deletes all error events of a given project.
func (c *ErrorStatsClient) DeleteEvents(ctx context.Context, req *clouderrorreportingpb.DeleteEventsRequest) (*clouderrorreportingpb.DeleteEventsResponse, error) {
	ctx = metadata.NewContext(ctx, c.metadata)
	var resp *clouderrorreportingpb.DeleteEventsResponse
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.DeleteEvents(ctx, req)
		return err
	}, c.CallOptions.DeleteEvents...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ErrorGroupStatsIterator manages a stream of *clouderrorreportingpb.ErrorGroupStats.
type ErrorGroupStatsIterator struct {
	// The current page data.
	items         []*clouderrorreportingpb.ErrorGroupStats
	atLastPage    bool
	currentIndex  int
	pageSize      int32
	nextPageToken string
	apiCall       func() error
}

// NextPage returns the next page of results.
// It will return at most the number of results specified by the last call to SetPageSize.
// If SetPageSize was never called or was called with a value less than 1,
// the page size is determined by the underlying service.
//
// NextPage may return a second return value of Done along with the last page of results. After
// NextPage returns Done, all subsequent calls to NextPage will return (nil, Done).
//
// Next and NextPage should not be used with the same iterator.
func (it *ErrorGroupStatsIterator) NextPage() ([]*clouderrorreportingpb.ErrorGroupStats, error) {
	if it.atLastPage {
		// We already returned Done with the last page of items. Continue to
		// return Done, but with no items.
		return nil, Done
	}
	if err := it.apiCall(); err != nil {
		return nil, err
	}
	if it.atLastPage {
		return it.items, Done
	}
	return it.items, nil
}

// Next returns the next result. Its second return value is Done if there are no more results.
// Once next returns Done, all subsequent calls will return Done.
//
// Internally, Next retrieves results in bulk. You can call SetPageSize as a performance hint to
// affect how many results are retrieved in a single RPC.
//
// SetPageToken should not be called when using Next.
//
// Next and NextPage should not be used with the same iterator.
func (it *ErrorGroupStatsIterator) Next() (*clouderrorreportingpb.ErrorGroupStats, error) {
	for it.currentIndex >= len(it.items) {
		if it.atLastPage {
			return nil, Done
		}
		if err := it.apiCall(); err != nil {
			return nil, err
		}
		it.currentIndex = 0
	}
	result := it.items[it.currentIndex]
	it.currentIndex++
	return result, nil
}

// PageSize returns the page size for all subsequent calls to NextPage.
func (it *ErrorGroupStatsIterator) PageSize() int {
	return int(it.pageSize)
}

// SetPageSize sets the page size for all subsequent calls to NextPage.
func (it *ErrorGroupStatsIterator) SetPageSize(pageSize int) {
	if pageSize > math.MaxInt32 {
		pageSize = math.MaxInt32
	}
	it.pageSize = int32(pageSize)
}

// SetPageToken sets the page token for the next call to NextPage, to resume the iteration from
// a previous point.
func (it *ErrorGroupStatsIterator) SetPageToken(token string) {
	it.nextPageToken = token
}

// NextPageToken returns a page token that can be used with SetPageToken to resume
// iteration from the next page. It returns the empty string if there are no more pages.
func (it *ErrorGroupStatsIterator) NextPageToken() string {
	return it.nextPageToken
}

// ErrorEventIterator manages a stream of *clouderrorreportingpb.ErrorEvent.
type ErrorEventIterator struct {
	// The current page data.
	items         []*clouderrorreportingpb.ErrorEvent
	atLastPage    bool
	currentIndex  int
	pageSize      int32
	nextPageToken string
	apiCall       func() error
}

// NextPage returns the next page of results.
// It will return at most the number of results specified by the last call to SetPageSize.
// If SetPageSize was never called or was called with a value less than 1,
// the page size is determined by the underlying service.
//
// NextPage may return a second return value of Done along with the last page of results. After
// NextPage returns Done, all subsequent calls to NextPage will return (nil, Done).
//
// Next and NextPage should not be used with the same iterator.
func (it *ErrorEventIterator) NextPage() ([]*clouderrorreportingpb.ErrorEvent, error) {
	if it.atLastPage {
		// We already returned Done with the last page of items. Continue to
		// return Done, but with no items.
		return nil, Done
	}
	if err := it.apiCall(); err != nil {
		return nil, err
	}
	if it.atLastPage {
		return it.items, Done
	}
	return it.items, nil
}

// Next returns the next result. Its second return value is Done if there are no more results.
// Once next returns Done, all subsequent calls will return Done.
//
// Internally, Next retrieves results in bulk. You can call SetPageSize as a performance hint to
// affect how many results are retrieved in a single RPC.
//
// SetPageToken should not be called when using Next.
//
// Next and NextPage should not be used with the same iterator.
func (it *ErrorEventIterator) Next() (*clouderrorreportingpb.ErrorEvent, error) {
	for it.currentIndex >= len(it.items) {
		if it.atLastPage {
			return nil, Done
		}
		if err := it.apiCall(); err != nil {
			return nil, err
		}
		it.currentIndex = 0
	}
	result := it.items[it.currentIndex]
	it.currentIndex++
	return result, nil
}

// PageSize returns the page size for all subsequent calls to NextPage.
func (it *ErrorEventIterator) PageSize() int {
	return int(it.pageSize)
}

// SetPageSize sets the page size for all subsequent calls to NextPage.
func (it *ErrorEventIterator) SetPageSize(pageSize int) {
	if pageSize > math.MaxInt32 {
		pageSize = math.MaxInt32
	}
	it.pageSize = int32(pageSize)
}

// SetPageToken sets the page token for the next call to NextPage, to resume the iteration from
// a previous point.
func (it *ErrorEventIterator) SetPageToken(token string) {
	it.nextPageToken = token
}

// NextPageToken returns a page token that can be used with SetPageToken to resume
// iteration from the next page. It returns the empty string if there are no more pages.
func (it *ErrorEventIterator) NextPageToken() string {
	return it.nextPageToken
}
