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

package metrics

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
	metricPathTemplate  = gax.MustCompilePathTemplate("projects/{project}/metrics/{metric}")
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
			"ListLogMetrics":  append([]gax.CallOption{withIdempotentRetryCodes()}, defaultRetryOptions()...),
			"GetLogMetric":    append([]gax.CallOption{withIdempotentRetryCodes()}, defaultRetryOptions()...),
			"CreateLogMetric": append([]gax.CallOption{withNonIdempotentRetryCodes()}, defaultRetryOptions()...),
			"UpdateLogMetric": append([]gax.CallOption{withNonIdempotentRetryCodes()}, defaultRetryOptions()...),
			"DeleteLogMetric": append([]gax.CallOption{withIdempotentRetryCodes()}, defaultRetryOptions()...),
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

// Client is a client for interacting with MetricsServiceV2.
type Client struct {
	// The connection to the service.
	conn *grpc.ClientConn

	// The gRPC API client.
	client google_logging_v2.MetricsServiceV2Client

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
		client:      google_logging_v2.NewMetricsServiceV2Client(conn),
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

// MetricPath returns the path for the metric resource.
func MetricPath(project string, metric string) string {
	path, err := metricPathTemplate.Render(map[string]string{
		"project": project,
		"metric":  metric,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// AUTO-GENERATED DOCUMENTATION AND METHOD -- see instructions at the top of the file for editing.

// ListLogMetrics lists logs-based metrics.
func (c *Client) ListLogMetrics(ctx context.Context, req *google_logging_v2.ListLogMetricsRequest) *LogMetricIterator {
	ctx = metadata.NewContext(ctx, c.metadata)
	it := &LogMetricIterator{}
	it.apiCall = func() error {
		if it.atLastPage {
			return Done
		}
		var resp *google_logging_v2.ListLogMetricsResponse
		err := gax.Invoke(ctx, func(ctx context.Context) error {
			var err error
			req.PageToken = it.nextPageToken
			req.PageSize = it.pageSize
			resp, err = c.client.ListLogMetrics(ctx, req)
			return err
		}, c.callOptions["ListLogMetrics"]...)
		if err != nil {
			return err
		}
		if resp.NextPageToken == "" {
			it.atLastPage = true
		} else {
			it.nextPageToken = resp.NextPageToken
		}
		it.items = resp.Metrics
		return nil
	}
	return it
}

// AUTO-GENERATED DOCUMENTATION AND METHOD -- see instructions at the top of the file for editing.

// GetLogMetric gets a logs-based metric.
func (c *Client) GetLogMetric(ctx context.Context, req *google_logging_v2.GetLogMetricRequest) (*google_logging_v2.LogMetric, error) {
	ctx = metadata.NewContext(ctx, c.metadata)
	var resp *google_logging_v2.LogMetric
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.GetLogMetric(ctx, req)
		return err
	}, c.callOptions["GetLogMetric"]...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// AUTO-GENERATED DOCUMENTATION AND METHOD -- see instructions at the top of the file for editing.

// CreateLogMetric creates a logs-based metric.
func (c *Client) CreateLogMetric(ctx context.Context, req *google_logging_v2.CreateLogMetricRequest) (*google_logging_v2.LogMetric, error) {
	ctx = metadata.NewContext(ctx, c.metadata)
	var resp *google_logging_v2.LogMetric
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.CreateLogMetric(ctx, req)
		return err
	}, c.callOptions["CreateLogMetric"]...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// AUTO-GENERATED DOCUMENTATION AND METHOD -- see instructions at the top of the file for editing.

// UpdateLogMetric creates or updates a logs-based metric.
func (c *Client) UpdateLogMetric(ctx context.Context, req *google_logging_v2.UpdateLogMetricRequest) (*google_logging_v2.LogMetric, error) {
	ctx = metadata.NewContext(ctx, c.metadata)
	var resp *google_logging_v2.LogMetric
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.UpdateLogMetric(ctx, req)
		return err
	}, c.callOptions["UpdateLogMetric"]...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// AUTO-GENERATED DOCUMENTATION AND METHOD -- see instructions at the top of the file for editing.

// DeleteLogMetric deletes a logs-based metric.
func (c *Client) DeleteLogMetric(ctx context.Context, req *google_logging_v2.DeleteLogMetricRequest) error {
	ctx = metadata.NewContext(ctx, c.metadata)
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		_, err = c.client.DeleteLogMetric(ctx, req)
		return err
	}, c.callOptions["DeleteLogMetric"]...)
	return err
}

// Iterators.
//

// LogMetricIterator manages a stream of *google_logging_v2.LogMetric.
type LogMetricIterator struct {
	// The current page data.
	items         []*google_logging_v2.LogMetric
	atLastPage    bool
	currentIndex  int
	pageSize      int32
	nextPageToken string
	apiCall       func() error
}

// NextPage moves to the next page and updates its internal data.
// It returns Done if no more pages exist.
func (it *LogMetricIterator) NextPage() ([]*google_logging_v2.LogMetric, error) {
	err := it.apiCall()
	if err != nil {
		return nil, err
	}
	return it.items, err
}

// Next returns the next element in the stream. It returns Done at
// the end of the stream.
func (it *LogMetricIterator) Next() (*google_logging_v2.LogMetric, error) {
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
func (it *LogMetricIterator) SetPageSize(pageSize int32) {
	it.pageSize = pageSize
}

// SetPageToken sets the next page token to be retrieved. Note, it
// does not retrieve the next page, or modify the cached page. If
// Next is called, there is no guarantee that the result returned
// will be from the next page until NextPage is called.
func (it *LogMetricIterator) SetPageToken(token string) {
	it.nextPageToken = token
}

// NextPageToken returns the next page token.
func (it *LogMetricIterator) NextPageToken() string {
	return it.nextPageToken
}
