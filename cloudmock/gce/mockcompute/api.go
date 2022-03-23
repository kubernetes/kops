/*
Copyright 2021 The Kubernetes Authors.

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

package mockcompute

import (
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

// MockClient represents a mocked compute client.
type MockClient struct {
	projectClient *projectClient
	zoneClient    *zoneClient

	networkClient          *networkClient
	subnetworkClient       *subnetworkClient
	backendServiceClient   *backendServiceClient
	routeClient            *routeClient
	forwardingRuleClient   *forwardingRuleClient
	httpHealthChecksClient *httpHealthChecksClient
	healthCheckClient      *healthCheckClient
	addressClient          *addressClient
	firewallClient         *firewallClient
	routerClient           *routerClient

	instanceTemplateClient     *instanceTemplateClient
	instanceGroupManagerClient *instanceGroupManagerClient
	targetPoolClient           *targetPoolClient

	diskClient *diskClient
}

var _ gce.ComputeClient = &MockClient{}

// NewMockClient creates a new mock client.
func NewMockClient(project string) *MockClient {
	return &MockClient{
		projectClient: newProjectClient(project),
		zoneClient:    newZoneClient(project),

		networkClient:          newNetworkClient(),
		subnetworkClient:       newSubnetworkClient(),
		backendServiceClient:   newBackendServiceClient(),
		routeClient:            newRouteClient(),
		forwardingRuleClient:   newForwardingRuleClient(),
		httpHealthChecksClient: newHttpHealthChecksClient(),
		healthCheckClient:      newHealthCheckClient(),
		addressClient:          newAddressClient(),
		firewallClient:         newFirewallClient(),
		routerClient:           newRouterClient(),

		instanceTemplateClient:     newInstanceTemplateClient(),
		instanceGroupManagerClient: newInstanceGroupManagerClient(),
		targetPoolClient:           newTargetPoolClient(),

		diskClient: newDiskClient(),
	}
}

func (c *MockClient) AllResources() map[string]interface{} {
	all := map[string]interface{}{}
	fs := []func() map[string]interface{}{
		c.projectClient.All,
		c.zoneClient.All,
		// Do not call c.networkClient.All() or c.subnetworkClient.All,
		// as currently pkg/resources/gce/gce.go
		// does not delete a network or subnetwork.
		// TODO(kenji): Fix this.
		c.routeClient.All,
		c.forwardingRuleClient.All,
		c.httpHealthChecksClient.All,
		c.healthCheckClient.All,
		c.addressClient.All,
		c.firewallClient.All,
		c.routerClient.All,
		c.instanceTemplateClient.All,
		c.instanceGroupManagerClient.All,
		c.targetPoolClient.All,
		c.diskClient.All,
		c.backendServiceClient.All,
	}
	for _, f := range fs {
		m := f()
		for k, v := range m {
			all[k] = v
		}
	}

	return all
}

func (c *MockClient) Projects() gce.ProjectClient {
	return c.projectClient
}

func (c *MockClient) Regions() gce.RegionClient {
	// Not implemented.
	return nil
}

func (c *MockClient) Zones() gce.ZoneClient {
	return c.zoneClient
}

func (c *MockClient) Networks() gce.NetworkClient {
	return c.networkClient
}

func (c *MockClient) Subnetworks() gce.SubnetworkClient {
	return c.subnetworkClient
}

func (c *MockClient) Routes() gce.RouteClient {
	return c.routeClient
}

func (c *MockClient) ForwardingRules() gce.ForwardingRuleClient {
	return c.forwardingRuleClient
}

func (c *MockClient) HTTPHealthChecks() gce.HttpHealthChecksClient {
	return c.httpHealthChecksClient
}

func (c *MockClient) RegionHealthChecks() gce.RegionHealthChecksClient {
	return c.healthCheckClient
}

func (c *MockClient) RegionBackendServices() gce.RegionBackendServiceClient {
	return c.backendServiceClient
}

func (c *MockClient) Addresses() gce.AddressClient {
	return c.addressClient
}

func (c *MockClient) Firewalls() gce.FirewallClient {
	return c.firewallClient
}

func (c *MockClient) Routers() gce.RouterClient {
	return c.routerClient
}

func (c *MockClient) Instances() gce.InstanceClient {
	// Not implemented.
	return nil
}

func (c *MockClient) InstanceTemplates() gce.InstanceTemplateClient {
	return c.instanceTemplateClient
}

func (c *MockClient) InstanceGroupManagers() gce.InstanceGroupManagerClient {
	return c.instanceGroupManagerClient
}

func (c *MockClient) TargetPools() gce.TargetPoolClient {
	return c.targetPoolClient
}

func (c *MockClient) Disks() gce.DiskClient {
	return c.diskClient
}

func notFoundError() error {
	return &googleapi.Error{
		Code: 404,
	}
}

func doneOperation() *compute.Operation {
	return &compute.Operation{
		Status: "DONE",
	}
}
