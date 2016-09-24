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
	"runtime"

	gax "github.com/googleapis/gax-go"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	agentTranslationProjectPathTemplate = gax.MustCompilePathTemplate("projects/{project}")
)

// AgentTranslationCallOptions contains the retry settings for each method of this client.
type AgentTranslationCallOptions struct {
	CreateCollectdTimeSeries []gax.CallOption
}

func defaultAgentTranslationClientOptions() []option.ClientOption {
	return []option.ClientOption{
		option.WithEndpoint("monitoring.googleapis.com:443"),
		option.WithScopes(),
	}
}

func defaultAgentTranslationCallOptions() *AgentTranslationCallOptions {
	retry := map[[2]string][]gax.CallOption{}

	return &AgentTranslationCallOptions{
		CreateCollectdTimeSeries: retry[[2]string{"default", "non_idempotent"}],
	}
}

// AgentTranslationClient is a client for interacting with AgentTranslationService.
type AgentTranslationClient struct {
	// The connection to the service.
	conn *grpc.ClientConn

	// The gRPC API client.
	client monitoringpb.AgentTranslationServiceClient

	// The call options for this service.
	CallOptions *AgentTranslationCallOptions

	// The metadata to be sent with each request.
	metadata map[string][]string
}

// NewAgentTranslationClient creates a new agent_translation service client.
//
// The AgentTranslation API allows `collectd`-based agents to
// write time series data to Cloud Monitoring.
// See [google.monitoring.v3.MetricService.CreateTimeSeries] instead.
func NewAgentTranslationClient(ctx context.Context, opts ...option.ClientOption) (*AgentTranslationClient, error) {
	conn, err := transport.DialGRPC(ctx, append(defaultAgentTranslationClientOptions(), opts...)...)
	if err != nil {
		return nil, err
	}
	c := &AgentTranslationClient{
		conn:        conn,
		client:      monitoringpb.NewAgentTranslationServiceClient(conn),
		CallOptions: defaultAgentTranslationCallOptions(),
	}
	c.SetGoogleClientInfo("gax", gax.Version)
	return c, nil
}

// Connection returns the client's connection to the API service.
func (c *AgentTranslationClient) Connection() *grpc.ClientConn {
	return c.conn
}

// Close closes the connection to the API service. The user should invoke this when
// the client is no longer required.
func (c *AgentTranslationClient) Close() error {
	return c.conn.Close()
}

// SetGoogleClientInfo sets the name and version of the application in
// the `x-goog-api-client` header passed on each request. Intended for
// use by Google-written clients.
func (c *AgentTranslationClient) SetGoogleClientInfo(name, version string) {
	c.metadata = map[string][]string{
		"x-goog-api-client": {fmt.Sprintf("%s/%s %s gax/%s go/%s", name, version, gapicNameVersion, gax.Version, runtime.Version())},
	}
}

// ProjectPath returns the path for the project resource.
func AgentTranslationProjectPath(project string) string {
	path, err := agentTranslationProjectPathTemplate.Render(map[string]string{
		"project": project,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// **Stackdriver Monitoring Agent only:** Creates a new time series.
//
// <aside class="caution">This method is only for use by the Google Monitoring Agent.
// Use [projects.timeSeries.create][google.monitoring.v3.MetricService.CreateTimeSeries]
// instead.</aside>
func (c *AgentTranslationClient) CreateCollectdTimeSeries(ctx context.Context, req *monitoringpb.CreateCollectdTimeSeriesRequest) error {
	ctx = metadata.NewContext(ctx, c.metadata)
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		_, err = c.client.CreateCollectdTimeSeries(ctx, req)
		return err
	}, c.CallOptions.CreateCollectdTimeSeries...)
	return err
}
