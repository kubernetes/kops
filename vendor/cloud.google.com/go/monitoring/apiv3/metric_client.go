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

package monitoring

import (
	"fmt"
	"math"
	"runtime"
	"time"

	gax "github.com/googleapis/gax-go"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

var (
	metricProjectPathTemplate                     = gax.MustCompilePathTemplate("projects/{project}")
	metricMetricDescriptorPathPathTemplate        = gax.MustCompilePathTemplate("projects/{project}/metricDescriptors/{metric_descriptor_path=**}")
	metricMonitoredResourceDescriptorPathTemplate = gax.MustCompilePathTemplate("projects/{project}/monitoredResourceDescriptors/{monitored_resource_descriptor}")
)

// MetricCallOptions contains the retry settings for each method of this client.
type MetricCallOptions struct {
	ListMonitoredResourceDescriptors []gax.CallOption
	GetMonitoredResourceDescriptor   []gax.CallOption
	ListMetricDescriptors            []gax.CallOption
	GetMetricDescriptor              []gax.CallOption
	CreateMetricDescriptor           []gax.CallOption
	DeleteMetricDescriptor           []gax.CallOption
	ListTimeSeries                   []gax.CallOption
	CreateTimeSeries                 []gax.CallOption
}

func defaultMetricClientOptions() []option.ClientOption {
	return []option.ClientOption{
		option.WithEndpoint("monitoring.googleapis.com:443"),
		option.WithScopes(),
	}
}

func defaultMetricCallOptions() *MetricCallOptions {
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

	return &MetricCallOptions{
		ListMonitoredResourceDescriptors: retry[[2]string{"default", "idempotent"}],
		GetMonitoredResourceDescriptor:   retry[[2]string{"default", "idempotent"}],
		ListMetricDescriptors:            retry[[2]string{"default", "idempotent"}],
		GetMetricDescriptor:              retry[[2]string{"default", "idempotent"}],
		CreateMetricDescriptor:           retry[[2]string{"default", "non_idempotent"}],
		DeleteMetricDescriptor:           retry[[2]string{"default", "idempotent"}],
		ListTimeSeries:                   retry[[2]string{"default", "idempotent"}],
		CreateTimeSeries:                 retry[[2]string{"default", "non_idempotent"}],
	}
}

// MetricClient is a client for interacting with MetricService.
type MetricClient struct {
	// The connection to the service.
	conn *grpc.ClientConn

	// The gRPC API client.
	client monitoringpb.MetricServiceClient

	// The call options for this service.
	CallOptions *MetricCallOptions

	// The metadata to be sent with each request.
	metadata map[string][]string
}

// NewMetricClient creates a new metric service client.
//
// Manages metric descriptors, monitored resource descriptors, and
// time series data.
func NewMetricClient(ctx context.Context, opts ...option.ClientOption) (*MetricClient, error) {
	conn, err := transport.DialGRPC(ctx, append(defaultMetricClientOptions(), opts...)...)
	if err != nil {
		return nil, err
	}
	c := &MetricClient{
		conn:        conn,
		client:      monitoringpb.NewMetricServiceClient(conn),
		CallOptions: defaultMetricCallOptions(),
	}
	c.SetGoogleClientInfo("gax", gax.Version)
	return c, nil
}

// Connection returns the client's connection to the API service.
func (c *MetricClient) Connection() *grpc.ClientConn {
	return c.conn
}

// Close closes the connection to the API service. The user should invoke this when
// the client is no longer required.
func (c *MetricClient) Close() error {
	return c.conn.Close()
}

// SetGoogleClientInfo sets the name and version of the application in
// the `x-goog-api-client` header passed on each request. Intended for
// use by Google-written clients.
func (c *MetricClient) SetGoogleClientInfo(name, version string) {
	c.metadata = map[string][]string{
		"x-goog-api-client": {fmt.Sprintf("%s/%s %s gax/%s go/%s", name, version, gapicNameVersion, gax.Version, runtime.Version())},
	}
}

// ProjectPath returns the path for the project resource.
func MetricProjectPath(project string) string {
	path, err := metricProjectPathTemplate.Render(map[string]string{
		"project": project,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// MetricDescriptorPathPath returns the path for the metric descriptor path resource.
func MetricMetricDescriptorPathPath(project string, metricDescriptorPath string) string {
	path, err := metricMetricDescriptorPathPathTemplate.Render(map[string]string{
		"project":                project,
		"metric_descriptor_path": metricDescriptorPath,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// MonitoredResourceDescriptorPath returns the path for the monitored resource descriptor resource.
func MetricMonitoredResourceDescriptorPath(project string, monitoredResourceDescriptor string) string {
	path, err := metricMonitoredResourceDescriptorPathTemplate.Render(map[string]string{
		"project":                       project,
		"monitored_resource_descriptor": monitoredResourceDescriptor,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// ListMonitoredResourceDescriptors lists monitored resource descriptors that match a filter. This method does not require a Stackdriver account.
func (c *MetricClient) ListMonitoredResourceDescriptors(ctx context.Context, req *monitoringpb.ListMonitoredResourceDescriptorsRequest) *MonitoredResourceDescriptorIterator {
	ctx = metadata.NewContext(ctx, c.metadata)
	it := &MonitoredResourceDescriptorIterator{}
	it.apiCall = func() error {
		var resp *monitoringpb.ListMonitoredResourceDescriptorsResponse
		err := gax.Invoke(ctx, func(ctx context.Context) error {
			var err error
			req.PageToken = it.nextPageToken
			req.PageSize = it.pageSize
			resp, err = c.client.ListMonitoredResourceDescriptors(ctx, req)
			return err
		}, c.CallOptions.ListMonitoredResourceDescriptors...)
		if err != nil {
			return err
		}
		if resp.NextPageToken == "" {
			it.atLastPage = true
		}
		it.nextPageToken = resp.NextPageToken
		it.items = resp.ResourceDescriptors
		return nil
	}
	return it
}

// GetMonitoredResourceDescriptor gets a single monitored resource descriptor. This method does not require a Stackdriver account.
func (c *MetricClient) GetMonitoredResourceDescriptor(ctx context.Context, req *monitoringpb.GetMonitoredResourceDescriptorRequest) (*monitoredrespb.MonitoredResourceDescriptor, error) {
	ctx = metadata.NewContext(ctx, c.metadata)
	var resp *monitoredrespb.MonitoredResourceDescriptor
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.GetMonitoredResourceDescriptor(ctx, req)
		return err
	}, c.CallOptions.GetMonitoredResourceDescriptor...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ListMetricDescriptors lists metric descriptors that match a filter. This method does not require a Stackdriver account.
func (c *MetricClient) ListMetricDescriptors(ctx context.Context, req *monitoringpb.ListMetricDescriptorsRequest) *MetricDescriptorIterator {
	ctx = metadata.NewContext(ctx, c.metadata)
	it := &MetricDescriptorIterator{}
	it.apiCall = func() error {
		var resp *monitoringpb.ListMetricDescriptorsResponse
		err := gax.Invoke(ctx, func(ctx context.Context) error {
			var err error
			req.PageToken = it.nextPageToken
			req.PageSize = it.pageSize
			resp, err = c.client.ListMetricDescriptors(ctx, req)
			return err
		}, c.CallOptions.ListMetricDescriptors...)
		if err != nil {
			return err
		}
		if resp.NextPageToken == "" {
			it.atLastPage = true
		}
		it.nextPageToken = resp.NextPageToken
		it.items = resp.MetricDescriptors
		return nil
	}
	return it
}

// GetMetricDescriptor gets a single metric descriptor. This method does not require a Stackdriver account.
func (c *MetricClient) GetMetricDescriptor(ctx context.Context, req *monitoringpb.GetMetricDescriptorRequest) (*metricpb.MetricDescriptor, error) {
	ctx = metadata.NewContext(ctx, c.metadata)
	var resp *metricpb.MetricDescriptor
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.GetMetricDescriptor(ctx, req)
		return err
	}, c.CallOptions.GetMetricDescriptor...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateMetricDescriptor creates a new metric descriptor.
// User-created metric descriptors define
// [custom metrics](/monitoring/custom-metrics).
func (c *MetricClient) CreateMetricDescriptor(ctx context.Context, req *monitoringpb.CreateMetricDescriptorRequest) (*metricpb.MetricDescriptor, error) {
	ctx = metadata.NewContext(ctx, c.metadata)
	var resp *metricpb.MetricDescriptor
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.CreateMetricDescriptor(ctx, req)
		return err
	}, c.CallOptions.CreateMetricDescriptor...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteMetricDescriptor deletes a metric descriptor. Only user-created
// [custom metrics](/monitoring/custom-metrics) can be deleted.
func (c *MetricClient) DeleteMetricDescriptor(ctx context.Context, req *monitoringpb.DeleteMetricDescriptorRequest) error {
	ctx = metadata.NewContext(ctx, c.metadata)
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		_, err = c.client.DeleteMetricDescriptor(ctx, req)
		return err
	}, c.CallOptions.DeleteMetricDescriptor...)
	return err
}

// ListTimeSeries lists time series that match a filter. This method does not require a Stackdriver account.
func (c *MetricClient) ListTimeSeries(ctx context.Context, req *monitoringpb.ListTimeSeriesRequest) *TimeSeriesIterator {
	ctx = metadata.NewContext(ctx, c.metadata)
	it := &TimeSeriesIterator{}
	it.apiCall = func() error {
		var resp *monitoringpb.ListTimeSeriesResponse
		err := gax.Invoke(ctx, func(ctx context.Context) error {
			var err error
			req.PageToken = it.nextPageToken
			req.PageSize = it.pageSize
			resp, err = c.client.ListTimeSeries(ctx, req)
			return err
		}, c.CallOptions.ListTimeSeries...)
		if err != nil {
			return err
		}
		if resp.NextPageToken == "" {
			it.atLastPage = true
		}
		it.nextPageToken = resp.NextPageToken
		it.items = resp.TimeSeries
		return nil
	}
	return it
}

// CreateTimeSeries creates or adds data to one or more time series.
// The response is empty if all time series in the request were written.
// If any time series could not be written, a corresponding failure message is
// included in the error response.
func (c *MetricClient) CreateTimeSeries(ctx context.Context, req *monitoringpb.CreateTimeSeriesRequest) error {
	ctx = metadata.NewContext(ctx, c.metadata)
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		_, err = c.client.CreateTimeSeries(ctx, req)
		return err
	}, c.CallOptions.CreateTimeSeries...)
	return err
}

// MonitoredResourceDescriptorIterator manages a stream of *monitoredrespb.MonitoredResourceDescriptor.
type MonitoredResourceDescriptorIterator struct {
	// The current page data.
	items         []*monitoredrespb.MonitoredResourceDescriptor
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
func (it *MonitoredResourceDescriptorIterator) NextPage() ([]*monitoredrespb.MonitoredResourceDescriptor, error) {
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
func (it *MonitoredResourceDescriptorIterator) Next() (*monitoredrespb.MonitoredResourceDescriptor, error) {
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
func (it *MonitoredResourceDescriptorIterator) PageSize() int {
	return int(it.pageSize)
}

// SetPageSize sets the page size for all subsequent calls to NextPage.
func (it *MonitoredResourceDescriptorIterator) SetPageSize(pageSize int) {
	if pageSize > math.MaxInt32 {
		pageSize = math.MaxInt32
	}
	it.pageSize = int32(pageSize)
}

// SetPageToken sets the page token for the next call to NextPage, to resume the iteration from
// a previous point.
func (it *MonitoredResourceDescriptorIterator) SetPageToken(token string) {
	it.nextPageToken = token
}

// NextPageToken returns a page token that can be used with SetPageToken to resume
// iteration from the next page. It returns the empty string if there are no more pages.
func (it *MonitoredResourceDescriptorIterator) NextPageToken() string {
	return it.nextPageToken
}

// MetricDescriptorIterator manages a stream of *metricpb.MetricDescriptor.
type MetricDescriptorIterator struct {
	// The current page data.
	items         []*metricpb.MetricDescriptor
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
func (it *MetricDescriptorIterator) NextPage() ([]*metricpb.MetricDescriptor, error) {
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
func (it *MetricDescriptorIterator) Next() (*metricpb.MetricDescriptor, error) {
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
func (it *MetricDescriptorIterator) PageSize() int {
	return int(it.pageSize)
}

// SetPageSize sets the page size for all subsequent calls to NextPage.
func (it *MetricDescriptorIterator) SetPageSize(pageSize int) {
	if pageSize > math.MaxInt32 {
		pageSize = math.MaxInt32
	}
	it.pageSize = int32(pageSize)
}

// SetPageToken sets the page token for the next call to NextPage, to resume the iteration from
// a previous point.
func (it *MetricDescriptorIterator) SetPageToken(token string) {
	it.nextPageToken = token
}

// NextPageToken returns a page token that can be used with SetPageToken to resume
// iteration from the next page. It returns the empty string if there are no more pages.
func (it *MetricDescriptorIterator) NextPageToken() string {
	return it.nextPageToken
}

// TimeSeriesIterator manages a stream of *monitoringpb.TimeSeries.
type TimeSeriesIterator struct {
	// The current page data.
	items         []*monitoringpb.TimeSeries
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
func (it *TimeSeriesIterator) NextPage() ([]*monitoringpb.TimeSeries, error) {
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
func (it *TimeSeriesIterator) Next() (*monitoringpb.TimeSeries, error) {
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
func (it *TimeSeriesIterator) PageSize() int {
	return int(it.pageSize)
}

// SetPageSize sets the page size for all subsequent calls to NextPage.
func (it *TimeSeriesIterator) SetPageSize(pageSize int) {
	if pageSize > math.MaxInt32 {
		pageSize = math.MaxInt32
	}
	it.pageSize = int32(pageSize)
}

// SetPageToken sets the page token for the next call to NextPage, to resume the iteration from
// a previous point.
func (it *TimeSeriesIterator) SetPageToken(token string) {
	it.nextPageToken = token
}

// NextPageToken returns a page token that can be used with SetPageToken to resume
// iteration from the next page. It returns the empty string if there are no more pages.
func (it *TimeSeriesIterator) NextPageToken() string {
	return it.nextPageToken
}
