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

// TODO(jba): link in google.cloud.audit.AuditLog, to support activity logs (after it is published)

// These features are missing now, but will likely be added:
// - There is no way to specify CallOptions.

// Package logadmin contains a Stackdriver Logging client that can be used
// for reading logs and working with sinks, metrics and monitored resources.
// For a client that can write logs, see package cloud.google.com/go/logging.
//
// The client uses Logging API v2.
// See https://cloud.google.com/logging/docs/api/v2/ for an introduction to the API.
//
// This package is experimental and subject to API changes.
package logadmin // import "cloud.google.com/go/preview/logging/logadmin"

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"

	vkit "cloud.google.com/go/logging/apiv2"
	"cloud.google.com/go/preview/logging"
	"cloud.google.com/go/preview/logging/internal"
	"github.com/golang/protobuf/ptypes"
	gax "github.com/googleapis/gax-go"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	logtypepb "google.golang.org/genproto/googleapis/logging/type"
	logpb "google.golang.org/genproto/googleapis/logging/v2"
)

// Client is a Logging client. A Client is associated with a single Cloud project.
type Client struct {
	lClient   *vkit.Client        // logging client
	sClient   *vkit.ConfigClient  // sink client
	mClient   *vkit.MetricsClient // metric client
	projectID string
	closed    bool
}

// NewClient returns a new logging client associated with the provided project ID.
//
// By default NewClient uses AdminScope. To use a different scope, call
// NewClient using a WithScopes option (see https://godoc.org/google.golang.org/api/option#WithScopes).
func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error) {
	// Check for '/' in project ID to reserve the ability to support various owning resources,
	// in the form "{Collection}/{Name}", for instance "organizations/my-org".
	if strings.ContainsRune(projectID, '/') {
		return nil, errors.New("logging: project ID contains '/'")
	}
	opts = append([]option.ClientOption{
		option.WithEndpoint(internal.ProdAddr),
		option.WithScopes(logging.AdminScope),
	}, opts...)
	lc, err := vkit.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	// TODO(jba): pass along any client options that should be provided to all clients.
	sc, err := vkit.NewConfigClient(ctx, option.WithGRPCConn(lc.Connection()))
	if err != nil {
		return nil, err
	}
	mc, err := vkit.NewMetricsClient(ctx, option.WithGRPCConn(lc.Connection()))
	if err != nil {
		return nil, err
	}
	lc.SetGoogleClientInfo("logging", internal.Version)
	sc.SetGoogleClientInfo("logging", internal.Version)
	mc.SetGoogleClientInfo("logging", internal.Version)
	client := &Client{
		lClient:   lc,
		sClient:   sc,
		mClient:   mc,
		projectID: projectID,
	}
	return client, nil
}

// parent returns the string used in many RPCs to denote the parent resource of the log.
func (c *Client) parent() string {
	return "projects/" + c.projectID
}

// Close closes the client.
func (c *Client) Close() error {
	if c.closed {
		return nil
	}
	// Return only the first error. Since all clients share an underlying connection,
	// Closes after the first always report a "connection is closing" error.
	err := c.lClient.Close()
	_ = c.sClient.Close()
	_ = c.mClient.Close()
	c.closed = true
	return err
}

// DeleteLog deletes a log and all its log entries. The log will reappear if it receives new entries.
// logID identifies the log within the project. An example log ID is "syslog". Requires AdminScope.
func (c *Client) DeleteLog(ctx context.Context, logID string) error {
	return c.lClient.DeleteLog(ctx, &logpb.DeleteLogRequest{
		LogName: internal.LogPath(c.parent(), logID),
	})
}

func toHTTPRequest(p *logtypepb.HttpRequest) (*logging.HTTPRequest, error) {
	if p == nil {
		return nil, nil
	}
	u, err := url.Parse(p.RequestUrl)
	if err != nil {
		return nil, err
	}
	hr := &http.Request{
		Method: p.RequestMethod,
		URL:    u,
		Header: map[string][]string{},
	}
	if p.UserAgent != "" {
		hr.Header.Set("User-Agent", p.UserAgent)
	}
	if p.Referer != "" {
		hr.Header.Set("Referer", p.Referer)
	}
	return &logging.HTTPRequest{
		Request:                        hr,
		RequestSize:                    p.RequestSize,
		Status:                         int(p.Status),
		ResponseSize:                   p.ResponseSize,
		RemoteIP:                       p.RemoteIp,
		CacheHit:                       p.CacheHit,
		CacheValidatedWithOriginServer: p.CacheValidatedWithOriginServer,
	}, nil
}

// An EntriesOption is an option for listing log entries.
type EntriesOption interface {
	set(*logpb.ListLogEntriesRequest)
}

// ProjectIDs sets the project IDs or project numbers from which to retrieve
// log entries. Examples of a project ID: "my-project-1A", "1234567890".
func ProjectIDs(pids []string) EntriesOption { return projectIDs(pids) }

type projectIDs []string

func (p projectIDs) set(r *logpb.ListLogEntriesRequest) { r.ProjectIds = []string(p) }

// Filter sets an advanced logs filter for listing log entries (see
// https://cloud.google.com/logging/docs/view/advanced_filters). The filter is
// compared against all log entries in the projects specified by ProjectIDs.
// Only entries that match the filter are retrieved. An empty filter (the
// default) matches all log entries.
//
// In the filter string, log names must be written in their full form, as
// "projects/PROJECT-ID/logs/LOG-ID". Forward slashes in LOG-ID must be
// replaced by %2F before calling Filter.
//
// Timestamps in the filter string must be written in RFC 3339 format. See the
// timestamp example.
func Filter(f string) EntriesOption { return filter(f) }

type filter string

func (f filter) set(r *logpb.ListLogEntriesRequest) { r.Filter = string(f) }

// NewestFirst causes log entries to be listed from most recent (newest) to
// least recent (oldest). By default, they are listed from oldest to newest.
func NewestFirst() EntriesOption { return newestFirst{} }

type newestFirst struct{}

func (newestFirst) set(r *logpb.ListLogEntriesRequest) { r.OrderBy = "timestamp desc" }

// OrderBy determines how a listing of log entries should be sorted. Presently,
// the only permitted values are "timestamp asc" (default) and "timestamp
// desc". The first option returns entries in order of increasing values of
// timestamp (oldest first), and the second option returns entries in order of
// decreasing timestamps (newest first). Entries with equal timestamps are
// returned in order of InsertID.
func OrderBy(ob string) EntriesOption { return orderBy(ob) }

type orderBy string

func (o orderBy) set(r *logpb.ListLogEntriesRequest) { r.OrderBy = string(o) }

// Entries returns an EntryIterator for iterating over log entries. By default,
// the log entries will be restricted to those from the project passed to
// NewClient. This may be overridden by passing a ProjectIDs option. Requires ReadScope or AdminScope.
func (c *Client) Entries(ctx context.Context, opts ...EntriesOption) *EntryIterator {
	it := &EntryIterator{
		ctx:    ctx,
		client: c.lClient,
		req:    listLogEntriesRequest(c.projectID, opts),
	}
	it.pageInfo, it.nextFunc = iterator.NewPageInfo(
		it.fetch,
		func() int { return len(it.items) },
		func() interface{} { b := it.items; it.items = nil; return b })
	return it
}

func listLogEntriesRequest(projectID string, opts []EntriesOption) *logpb.ListLogEntriesRequest {
	req := &logpb.ListLogEntriesRequest{
		ProjectIds: []string{projectID},
	}
	for _, opt := range opts {
		opt.set(req)
	}
	return req
}

// An EntryIterator iterates over log entries.
type EntryIterator struct {
	ctx      context.Context
	client   *vkit.Client
	pageInfo *iterator.PageInfo
	nextFunc func() error
	req      *logpb.ListLogEntriesRequest
	items    []*logging.Entry
}

// PageInfo supports pagination. See https://godoc.org/google.golang.org/api/iterator package for details.
func (it *EntryIterator) PageInfo() *iterator.PageInfo { return it.pageInfo }

// Next returns the next result. Its second return value is iterator.Done
// (https://godoc.org/google.golang.org/api/iterator) if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *EntryIterator) Next() (*logging.Entry, error) {
	if err := it.nextFunc(); err != nil {
		return nil, err
	}
	item := it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *EntryIterator) fetch(pageSize int, pageToken string) (string, error) {
	// TODO(jba): Do this a nicer way if the generated code supports one.
	// TODO(jba): If the above TODO can't be done, find a way to pass metadata in the call.
	client := logpb.NewLoggingServiceV2Client(it.client.Connection())
	var res *logpb.ListLogEntriesResponse
	err := gax.Invoke(it.ctx, func(ctx context.Context) error {
		it.req.PageSize = trunc32(pageSize)
		it.req.PageToken = pageToken
		var err error
		res, err = client.ListLogEntries(ctx, it.req)
		return err
	}, it.client.CallOptions.ListLogEntries...)
	if err != nil {
		return "", err
	}
	for _, ep := range res.Entries {
		e, err := fromLogEntry(ep)
		if err != nil {
			return "", err
		}
		it.items = append(it.items, e)
	}
	return res.NextPageToken, nil
}

func trunc32(i int) int32 {
	if i > math.MaxInt32 {
		i = math.MaxInt32
	}
	return int32(i)
}

var slashUnescaper = strings.NewReplacer("%2F", "/", "%2f", "/")

func fromLogEntry(le *logpb.LogEntry) (*logging.Entry, error) {
	time, err := ptypes.Timestamp(le.Timestamp)
	if err != nil {
		return nil, err
	}
	var payload interface{}
	switch x := le.Payload.(type) {
	case *logpb.LogEntry_TextPayload:
		payload = x.TextPayload

	case *logpb.LogEntry_ProtoPayload:
		var d ptypes.DynamicAny
		if err := ptypes.UnmarshalAny(x.ProtoPayload, &d); err != nil {
			return nil, fmt.Errorf("logging: unmarshalling proto payload: %v", err)
		}
		payload = d.Message

	case *logpb.LogEntry_JsonPayload:
		// Leave this as a Struct.
		// TODO(jba): convert to map[string]interface{}?
		payload = x.JsonPayload

	default:
		return nil, fmt.Errorf("logging: unknown payload type: %T", le.Payload)
	}
	hr, err := toHTTPRequest(le.HttpRequest)
	if err != nil {
		return nil, err
	}
	return &logging.Entry{
		Timestamp:   time,
		Severity:    logging.Severity(le.Severity),
		Payload:     payload,
		Labels:      le.Labels,
		InsertID:    le.InsertId,
		HTTPRequest: hr,
		Operation:   le.Operation,
		LogName:     slashUnescaper.Replace(le.LogName),
		Resource:    le.Resource,
	}, nil
}
