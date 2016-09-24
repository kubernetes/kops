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
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

var (
	groupProjectPathTemplate = gax.MustCompilePathTemplate("projects/{project}")
	groupGroupPathTemplate   = gax.MustCompilePathTemplate("projects/{project}/groups/{group}")
)

// GroupCallOptions contains the retry settings for each method of this client.
type GroupCallOptions struct {
	ListGroups       []gax.CallOption
	GetGroup         []gax.CallOption
	CreateGroup      []gax.CallOption
	UpdateGroup      []gax.CallOption
	DeleteGroup      []gax.CallOption
	ListGroupMembers []gax.CallOption
}

func defaultGroupClientOptions() []option.ClientOption {
	return []option.ClientOption{
		option.WithEndpoint("monitoring.googleapis.com:443"),
		option.WithScopes(),
	}
}

func defaultGroupCallOptions() *GroupCallOptions {
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

	return &GroupCallOptions{
		ListGroups:       retry[[2]string{"default", "idempotent"}],
		GetGroup:         retry[[2]string{"default", "idempotent"}],
		CreateGroup:      retry[[2]string{"default", "non_idempotent"}],
		UpdateGroup:      retry[[2]string{"default", "idempotent"}],
		DeleteGroup:      retry[[2]string{"default", "idempotent"}],
		ListGroupMembers: retry[[2]string{"default", "idempotent"}],
	}
}

// GroupClient is a client for interacting with GroupService.
type GroupClient struct {
	// The connection to the service.
	conn *grpc.ClientConn

	// The gRPC API client.
	client monitoringpb.GroupServiceClient

	// The call options for this service.
	CallOptions *GroupCallOptions

	// The metadata to be sent with each request.
	metadata map[string][]string
}

// NewGroupClient creates a new group service client.
//
// The Group API lets you inspect and manage your
// [groups](google.monitoring.v3.Group).
//
// A group is a named filter that is used to identify
// a collection of monitored resources. Groups are typically used to
// mirror the physical and/or logical topology of the environment.
// Because group membership is computed dynamically, monitored
// resources that are started in the future are automatically placed
// in matching groups. By using a group to name monitored resources in,
// for example, an alert policy, the target of that alert policy is
// updated automatically as monitored resources are added and removed
// from the infrastructure.
func NewGroupClient(ctx context.Context, opts ...option.ClientOption) (*GroupClient, error) {
	conn, err := transport.DialGRPC(ctx, append(defaultGroupClientOptions(), opts...)...)
	if err != nil {
		return nil, err
	}
	c := &GroupClient{
		conn:        conn,
		client:      monitoringpb.NewGroupServiceClient(conn),
		CallOptions: defaultGroupCallOptions(),
	}
	c.SetGoogleClientInfo("gax", gax.Version)
	return c, nil
}

// Connection returns the client's connection to the API service.
func (c *GroupClient) Connection() *grpc.ClientConn {
	return c.conn
}

// Close closes the connection to the API service. The user should invoke this when
// the client is no longer required.
func (c *GroupClient) Close() error {
	return c.conn.Close()
}

// SetGoogleClientInfo sets the name and version of the application in
// the `x-goog-api-client` header passed on each request. Intended for
// use by Google-written clients.
func (c *GroupClient) SetGoogleClientInfo(name, version string) {
	c.metadata = map[string][]string{
		"x-goog-api-client": {fmt.Sprintf("%s/%s %s gax/%s go/%s", name, version, gapicNameVersion, gax.Version, runtime.Version())},
	}
}

// ProjectPath returns the path for the project resource.
func GroupProjectPath(project string) string {
	path, err := groupProjectPathTemplate.Render(map[string]string{
		"project": project,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// GroupPath returns the path for the group resource.
func GroupGroupPath(project string, group string) string {
	path, err := groupGroupPathTemplate.Render(map[string]string{
		"project": project,
		"group":   group,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// ListGroups lists the existing groups. The project ID in the URL path must refer
// to a Stackdriver account.
func (c *GroupClient) ListGroups(ctx context.Context, req *monitoringpb.ListGroupsRequest) *GroupIterator {
	ctx = metadata.NewContext(ctx, c.metadata)
	it := &GroupIterator{}
	it.apiCall = func() error {
		var resp *monitoringpb.ListGroupsResponse
		err := gax.Invoke(ctx, func(ctx context.Context) error {
			var err error
			req.PageToken = it.nextPageToken
			req.PageSize = it.pageSize
			resp, err = c.client.ListGroups(ctx, req)
			return err
		}, c.CallOptions.ListGroups...)
		if err != nil {
			return err
		}
		if resp.NextPageToken == "" {
			it.atLastPage = true
		}
		it.nextPageToken = resp.NextPageToken
		it.items = resp.Group
		return nil
	}
	return it
}

// GetGroup gets a single group. The project ID in the URL path must refer to a
// Stackdriver account.
func (c *GroupClient) GetGroup(ctx context.Context, req *monitoringpb.GetGroupRequest) (*monitoringpb.Group, error) {
	ctx = metadata.NewContext(ctx, c.metadata)
	var resp *monitoringpb.Group
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.GetGroup(ctx, req)
		return err
	}, c.CallOptions.GetGroup...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateGroup creates a new group. The project ID in the URL path must refer to a
// Stackdriver account.
func (c *GroupClient) CreateGroup(ctx context.Context, req *monitoringpb.CreateGroupRequest) (*monitoringpb.Group, error) {
	ctx = metadata.NewContext(ctx, c.metadata)
	var resp *monitoringpb.Group
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.CreateGroup(ctx, req)
		return err
	}, c.CallOptions.CreateGroup...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateGroup updates an existing group.
// You can change any group attributes except `name`.
// The project ID in the URL path must refer to a Stackdriver account.
func (c *GroupClient) UpdateGroup(ctx context.Context, req *monitoringpb.UpdateGroupRequest) (*monitoringpb.Group, error) {
	ctx = metadata.NewContext(ctx, c.metadata)
	var resp *monitoringpb.Group
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.UpdateGroup(ctx, req)
		return err
	}, c.CallOptions.UpdateGroup...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteGroup deletes an existing group. The project ID in the URL path must refer to a
// Stackdriver account.
func (c *GroupClient) DeleteGroup(ctx context.Context, req *monitoringpb.DeleteGroupRequest) error {
	ctx = metadata.NewContext(ctx, c.metadata)
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		_, err = c.client.DeleteGroup(ctx, req)
		return err
	}, c.CallOptions.DeleteGroup...)
	return err
}

// ListGroupMembers lists the monitored resources that are members of a group. The project ID
// in the URL path must refer to a Stackdriver account.
func (c *GroupClient) ListGroupMembers(ctx context.Context, req *monitoringpb.ListGroupMembersRequest) *MonitoredResourceIterator {
	ctx = metadata.NewContext(ctx, c.metadata)
	it := &MonitoredResourceIterator{}
	it.apiCall = func() error {
		var resp *monitoringpb.ListGroupMembersResponse
		err := gax.Invoke(ctx, func(ctx context.Context) error {
			var err error
			req.PageToken = it.nextPageToken
			req.PageSize = it.pageSize
			resp, err = c.client.ListGroupMembers(ctx, req)
			return err
		}, c.CallOptions.ListGroupMembers...)
		if err != nil {
			return err
		}
		if resp.NextPageToken == "" {
			it.atLastPage = true
		}
		it.nextPageToken = resp.NextPageToken
		it.items = resp.Members
		return nil
	}
	return it
}

// GroupIterator manages a stream of *monitoringpb.Group.
type GroupIterator struct {
	// The current page data.
	items         []*monitoringpb.Group
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
func (it *GroupIterator) NextPage() ([]*monitoringpb.Group, error) {
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
func (it *GroupIterator) Next() (*monitoringpb.Group, error) {
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
func (it *GroupIterator) PageSize() int {
	return int(it.pageSize)
}

// SetPageSize sets the page size for all subsequent calls to NextPage.
func (it *GroupIterator) SetPageSize(pageSize int) {
	if pageSize > math.MaxInt32 {
		pageSize = math.MaxInt32
	}
	it.pageSize = int32(pageSize)
}

// SetPageToken sets the page token for the next call to NextPage, to resume the iteration from
// a previous point.
func (it *GroupIterator) SetPageToken(token string) {
	it.nextPageToken = token
}

// NextPageToken returns a page token that can be used with SetPageToken to resume
// iteration from the next page. It returns the empty string if there are no more pages.
func (it *GroupIterator) NextPageToken() string {
	return it.nextPageToken
}

// MonitoredResourceIterator manages a stream of *monitoredrespb.MonitoredResource.
type MonitoredResourceIterator struct {
	// The current page data.
	items         []*monitoredrespb.MonitoredResource
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
func (it *MonitoredResourceIterator) NextPage() ([]*monitoredrespb.MonitoredResource, error) {
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
func (it *MonitoredResourceIterator) Next() (*monitoredrespb.MonitoredResource, error) {
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
func (it *MonitoredResourceIterator) PageSize() int {
	return int(it.pageSize)
}

// SetPageSize sets the page size for all subsequent calls to NextPage.
func (it *MonitoredResourceIterator) SetPageSize(pageSize int) {
	if pageSize > math.MaxInt32 {
		pageSize = math.MaxInt32
	}
	it.pageSize = int32(pageSize)
}

// SetPageToken sets the page token for the next call to NextPage, to resume the iteration from
// a previous point.
func (it *MonitoredResourceIterator) SetPageToken(token string) {
	it.nextPageToken = token
}

// NextPageToken returns a page token that can be used with SetPageToken to resume
// iteration from the next page. It returns the empty string if there are no more pages.
func (it *MonitoredResourceIterator) NextPageToken() string {
	return it.nextPageToken
}
