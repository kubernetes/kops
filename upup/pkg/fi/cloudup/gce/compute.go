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

package gce

import (
	"context"
	"fmt"

	compute "google.golang.org/api/compute/v1"
)

type ComputeClient interface {
	Projects() ProjectClient
	Regions() RegionClient
	Zones() ZoneClient
	Networks() NetworkClient
	Subnetworks() SubnetworkClient
	Routes() RouteClient
	ForwardingRules() ForwardingRuleClient
	HTTPHealthChecks() HttpHealthChecksClient
	HealthChecks() HealthChecksClient
	Addresses() AddressClient
	Firewalls() FirewallClient
	Routers() RouterClient
	Instances() InstanceClient
	InstanceTemplates() InstanceTemplateClient
	InstanceGroupManagers() InstanceGroupManagerClient
	TargetPools() TargetPoolClient
	Disks() DiskClient
	BackendServices() BackendServiceClient
}

type computeClientImpl struct {
	srv *compute.Service
}

var _ ComputeClient = &computeClientImpl{}

func newComputeClientImpl(ctx context.Context) (*computeClientImpl, error) {
	srv, err := compute.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("error building compute API client: %v", err)
	}
	return &computeClientImpl{
		srv: srv,
	}, nil
}

func (c *computeClientImpl) Projects() ProjectClient {
	return &projectClientImpl{
		srv: c.srv.Projects,
	}
}

func (c *computeClientImpl) Regions() RegionClient {
	return &regionClientImpl{
		srv: c.srv.Regions,
	}
}

func (c *computeClientImpl) Zones() ZoneClient {
	return &zoneClientImpl{
		srv: c.srv.Zones,
	}
}

func (c *computeClientImpl) Networks() NetworkClient {
	return &networkClientImpl{
		srv: c.srv.Networks,
	}
}

func (c *computeClientImpl) Subnetworks() SubnetworkClient {
	return &subnetworkClientImpl{
		srv: c.srv.Subnetworks,
	}
}

func (c *computeClientImpl) Routes() RouteClient {
	return &routeClientImpl{
		srv: c.srv.Routes,
	}
}

func (c *computeClientImpl) ForwardingRules() ForwardingRuleClient {
	return &forwardingRuleClientImpl{
		srv:  c.srv.ForwardingRules,
		gsrv: c.srv.GlobalForwardingRules,
	}
}

func (c *computeClientImpl) BackendServices() BackendServiceClient {
	return &backendServiceClientImpl{
		srv:  c.srv.BackendServices,
		rsrv: c.srv.RegionBackendServices,
	}
}

func (c *computeClientImpl) HTTPHealthChecks() HttpHealthChecksClient {
	return &httpHealthCheckClientImpl{
		srv: c.srv.HttpHealthChecks,
	}
}

func (c *computeClientImpl) HealthChecks() HealthChecksClient {
	return &healthCheckClientImpl{
		srv:  c.srv.HealthChecks,
		rsrv: c.srv.RegionHealthChecks,
	}
}

func (c *computeClientImpl) Addresses() AddressClient {
	return &addressClientImpl{
		srv:  c.srv.Addresses,
		gsrv: c.srv.GlobalAddresses,
	}
}

func (c *computeClientImpl) Firewalls() FirewallClient {
	return &firewallClientImpl{
		srv: c.srv.Firewalls,
	}
}

func (c *computeClientImpl) Routers() RouterClient {
	return &routerClientImpl{
		srv: c.srv.Routers,
	}
}

func (c *computeClientImpl) Instances() InstanceClient {
	return &instanceClientImpl{
		srv: c.srv.Instances,
	}
}

func (c *computeClientImpl) InstanceTemplates() InstanceTemplateClient {
	return &instanceTemplateClientImpl{
		srv: c.srv.InstanceTemplates,
	}
}

func (c *computeClientImpl) InstanceGroupManagers() InstanceGroupManagerClient {
	return &instanceGroupManagerClientImpl{
		srv: c.srv.InstanceGroupManagers,
	}
}

func (c *computeClientImpl) TargetPools() TargetPoolClient {
	return &targetPoolClientImpl{
		srv: c.srv.TargetPools,
	}
}

func (c *computeClientImpl) Disks() DiskClient {
	return &diskClientImpl{
		srv: c.srv.Disks,
	}
}

type ProjectClient interface {
	Get(project string) (*compute.Project, error)
}

type projectClientImpl struct {
	srv *compute.ProjectsService
}

var _ ProjectClient = &projectClientImpl{}

func (c *projectClientImpl) Get(project string) (*compute.Project, error) {
	return c.srv.Get(project).Do()
}

type RegionClient interface {
	List(ctx context.Context, project string) ([]*compute.Region, error)
}

type regionClientImpl struct {
	srv *compute.RegionsService
}

var _ RegionClient = &regionClientImpl{}

func (c *regionClientImpl) List(ctx context.Context, project string) ([]*compute.Region, error) {
	var regions []*compute.Region
	err := c.srv.List(project).Pages(ctx, func(page *compute.RegionList) error {
		regions = append(regions, page.Items...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return regions, nil
}

type ZoneClient interface {
	List(ctx context.Context, project string) ([]*compute.Zone, error)
}

type zoneClientImpl struct {
	srv *compute.ZonesService
}

var _ ZoneClient = &zoneClientImpl{}

func (c *zoneClientImpl) List(ctx context.Context, project string) ([]*compute.Zone, error) {
	var zones []*compute.Zone
	err := c.srv.List(project).Pages(ctx, func(page *compute.ZoneList) error {
		zones = append(zones, page.Items...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return zones, nil
}

type NetworkClient interface {
	Insert(project string, nw *compute.Network) (*compute.Operation, error)
	Get(project, name string) (*compute.Network, error)
	Delete(project, name string) (*compute.Operation, error)
	List(project string) (*compute.NetworkList, error)
}

type networkClientImpl struct {
	srv *compute.NetworksService
}

var _ NetworkClient = &networkClientImpl{}

func (c *networkClientImpl) Insert(project string, nw *compute.Network) (*compute.Operation, error) {
	return c.srv.Insert(project, nw).Do()
}

func (c *networkClientImpl) Delete(project string, name string) (*compute.Operation, error) {
	return c.srv.Delete(project, name).Do()
}

func (c *networkClientImpl) Get(project, name string) (*compute.Network, error) {
	return c.srv.Get(project, name).Do()
}

func (c *networkClientImpl) List(project string) (*compute.NetworkList, error) {
	return c.srv.List(project).Do()
}

type SubnetworkClient interface {
	Insert(project, region string, subnet *compute.Subnetwork) (*compute.Operation, error)
	Patch(project, region, name string, subnet *compute.Subnetwork) (*compute.Operation, error)
	Delete(project, region, name string) (*compute.Operation, error)
	Get(project, region, name string) (*compute.Subnetwork, error)
	List(ctx context.Context, project, region string) ([]*compute.Subnetwork, error)
}

type subnetworkClientImpl struct {
	srv *compute.SubnetworksService
}

var _ SubnetworkClient = &subnetworkClientImpl{}

func (c *subnetworkClientImpl) Insert(project, region string, subnet *compute.Subnetwork) (*compute.Operation, error) {
	return c.srv.Insert(project, region, subnet).Do()
}

func (c *subnetworkClientImpl) Patch(project, region, name string, subnet *compute.Subnetwork) (*compute.Operation, error) {
	return c.srv.Patch(project, region, name, subnet).Do()
}

func (c *subnetworkClientImpl) Delete(project, region, name string) (*compute.Operation, error) {
	return c.srv.Delete(project, region, name).Do()
}

func (c *subnetworkClientImpl) Get(project, region, name string) (*compute.Subnetwork, error) {
	return c.srv.Get(project, region, name).Do()
}

func (c *subnetworkClientImpl) List(ctx context.Context, project, region string) ([]*compute.Subnetwork, error) {
	var subnetworks []*compute.Subnetwork
	if err := c.srv.List(project, region).Pages(ctx, func(p *compute.SubnetworkList) error {
		subnetworks = append(subnetworks, p.Items...)
		return nil
	}); err != nil {
		return nil, err
	}
	return subnetworks, nil
}

// ===
type BackendServiceClient interface {
	Insert(project, region string, fr *compute.BackendService) (*compute.Operation, error)
	Delete(project, region, name string) (*compute.Operation, error)
	Get(project, region, name string) (*compute.BackendService, error)
	List(ctx context.Context, project, region string) ([]*compute.BackendService, error)
}

type backendServiceClientImpl struct {
	srv  *compute.BackendServicesService
	rsrv *compute.RegionBackendServicesService
}

var _ BackendServiceClient = &backendServiceClientImpl{}

func (c *backendServiceClientImpl) Insert(project, region string, fr *compute.BackendService) (*compute.Operation, error) {
	if region == "" {
		return c.srv.Insert(project, fr).Do()
	}
	return c.rsrv.Insert(project, region, fr).Do()
}

func (c *backendServiceClientImpl) Delete(project, region, name string) (*compute.Operation, error) {
	if region == "" {
		return c.srv.Delete(project, name).Do()
	}
	return c.rsrv.Delete(project, region, name).Do()
}

func (c *backendServiceClientImpl) Get(project, region, name string) (*compute.BackendService, error) {
	if region == "" {
		return c.srv.Get(project, name).Do()
	}
	return c.rsrv.Get(project, region, name).Do()
}

func (c *backendServiceClientImpl) List(ctx context.Context, project, region string) ([]*compute.BackendService, error) {
	var bs []*compute.BackendService
	if region == "" {
		if err := c.srv.List(project).Pages(ctx, func(p *compute.BackendServiceList) error {
			bs = append(bs, p.Items...)
			return nil
		}); err != nil {
			return nil, err
		}
		return bs, nil
	}
	if err := c.rsrv.List(project, region).Pages(ctx, func(p *compute.BackendServiceList) error {
		bs = append(bs, p.Items...)
		return nil
	}); err != nil {
		return nil, err
	}
	return bs, nil
}

// ===

type RouteClient interface {
	Delete(project, name string) (*compute.Operation, error)
	List(ctx context.Context, project string) ([]*compute.Route, error)
}

type routeClientImpl struct {
	srv *compute.RoutesService
}

var _ RouteClient = &routeClientImpl{}

func (c *routeClientImpl) Delete(project, name string) (*compute.Operation, error) {
	return c.srv.Delete(project, name).Do()
}

func (c *routeClientImpl) List(ctx context.Context, project string) ([]*compute.Route, error) {
	var routes []*compute.Route
	if err := c.srv.List(project).Pages(ctx, func(p *compute.RouteList) error {
		routes = append(routes, p.Items...)
		return nil
	}); err != nil {
		return nil, err
	}
	return routes, nil
}

type ForwardingRuleClient interface {
	Insert(ctx context.Context, project, region string, fr *compute.ForwardingRule) (*compute.Operation, error)
	Delete(ctx context.Context, project, region, name string) (*compute.Operation, error)
	Get(ctx context.Context, project, region, name string) (*compute.ForwardingRule, error)
	List(ctx context.Context, project, region string) ([]*compute.ForwardingRule, error)
	SetLabels(ctx context.Context, project, region, resource string, request *compute.RegionSetLabelsRequest) (*compute.Operation, error)
}

type forwardingRuleClientImpl struct {
	srv  *compute.ForwardingRulesService
	gsrv *compute.GlobalForwardingRulesService
}

var _ ForwardingRuleClient = &forwardingRuleClientImpl{}

func (c *forwardingRuleClientImpl) Insert(ctx context.Context, project, region string, fr *compute.ForwardingRule) (*compute.Operation, error) {
	if region == "" {
		return c.gsrv.Insert(project, fr).Context(ctx).Do()
	}
	return c.srv.Insert(project, region, fr).Context(ctx).Do()
}

func (c *forwardingRuleClientImpl) Delete(ctx context.Context, project, region, name string) (*compute.Operation, error) {
	if region == "" {
		return c.gsrv.Delete(project, name).Context(ctx).Do()
	}
	return c.srv.Delete(project, region, name).Context(ctx).Do()
}

func (c *forwardingRuleClientImpl) Get(ctx context.Context, project, region, name string) (*compute.ForwardingRule, error) {
	if region == "" {
		return c.gsrv.Get(project, name).Do()
	}
	return c.srv.Get(project, region, name).Do()
}

func (c *forwardingRuleClientImpl) SetLabels(ctx context.Context, project string, region string, resource string, request *compute.RegionSetLabelsRequest) (*compute.Operation, error) {
	if region == "" {
		grequest := &compute.GlobalSetLabelsRequest{
			LabelFingerprint: request.LabelFingerprint,
			Labels:           request.Labels,
		}
		return c.gsrv.SetLabels(project, resource, grequest).Context(ctx).Do()
	}
	return c.srv.SetLabels(project, region, resource, request).Context(ctx).Do()
}

func (c *forwardingRuleClientImpl) List(ctx context.Context, project, region string) ([]*compute.ForwardingRule, error) {
	var frs []*compute.ForwardingRule
	if region == "" {
		if err := c.gsrv.List(project).Pages(ctx, func(p *compute.ForwardingRuleList) error {
			frs = append(frs, p.Items...)
			return nil
		}); err != nil {
			return nil, err
		}
		return frs, nil
	}
	if err := c.srv.List(project, region).Pages(ctx, func(p *compute.ForwardingRuleList) error {
		frs = append(frs, p.Items...)
		return nil
	}); err != nil {
		return nil, err
	}
	return frs, nil
}

type HealthChecksClient interface {
	Insert(project, region string, fr *compute.HealthCheck) (*compute.Operation, error)
	Delete(project, region, name string) (*compute.Operation, error)
	Get(project, region, name string) (*compute.HealthCheck, error)
	List(ctx context.Context, project, region string) ([]*compute.HealthCheck, error)
}

type healthCheckClientImpl struct {
	srv  *compute.HealthChecksService
	rsrv *compute.RegionHealthChecksService
}

var _ HealthChecksClient = &healthCheckClientImpl{}

func (c *healthCheckClientImpl) Insert(project, region string, hc *compute.HealthCheck) (*compute.Operation, error) {
	if region == "" {
		return c.srv.Insert(project, hc).Do()
	}
	return c.rsrv.Insert(project, region, hc).Do()
}

func (c *healthCheckClientImpl) Delete(project, region, name string) (*compute.Operation, error) {
	if region == "" {
		return c.srv.Delete(project, name).Do()
	}
	return c.rsrv.Delete(project, region, name).Do()
}

func (c *healthCheckClientImpl) Get(project, region, name string) (*compute.HealthCheck, error) {
	if region == "" {
		return c.srv.Get(project, name).Do()
	}
	return c.rsrv.Get(project, region, name).Do()
}

func (c *healthCheckClientImpl) List(ctx context.Context, project, region string) ([]*compute.HealthCheck, error) {
	var hcs []*compute.HealthCheck
	if region == "" {
		if err := c.srv.List(project).Pages(ctx, func(p *compute.HealthCheckList) error {
			hcs = append(hcs, p.Items...)
			return nil
		}); err != nil {
			return nil, err
		}
	}
	if err := c.rsrv.List(project, region).Pages(ctx, func(p *compute.HealthCheckList) error {
		hcs = append(hcs, p.Items...)
		return nil
	}); err != nil {
		return nil, err
	}
	return hcs, nil
}

type HttpHealthChecksClient interface {
	Insert(project string, fr *compute.HttpHealthCheck) (*compute.Operation, error)
	Delete(project, name string) (*compute.Operation, error)
	Get(project, name string) (*compute.HttpHealthCheck, error)
	List(ctx context.Context, project string) ([]*compute.HttpHealthCheck, error)
}

type httpHealthCheckClientImpl struct {
	srv *compute.HttpHealthChecksService
}

var _ HttpHealthChecksClient = &httpHealthCheckClientImpl{}

func (c *httpHealthCheckClientImpl) Insert(project string, fr *compute.HttpHealthCheck) (*compute.Operation, error) {
	return c.srv.Insert(project, fr).Do()
}

func (c *httpHealthCheckClientImpl) Delete(project, name string) (*compute.Operation, error) {
	return c.srv.Delete(project, name).Do()
}

func (c *httpHealthCheckClientImpl) Get(project, name string) (*compute.HttpHealthCheck, error) {
	return c.srv.Get(project, name).Do()
}

func (c *httpHealthCheckClientImpl) List(ctx context.Context, project string) ([]*compute.HttpHealthCheck, error) {
	var hcs []*compute.HttpHealthCheck
	if err := c.srv.List(project).Pages(ctx, func(p *compute.HttpHealthCheckList) error {
		hcs = append(hcs, p.Items...)
		return nil
	}); err != nil {
		return nil, err
	}
	return hcs, nil
}

type AddressClient interface {
	Insert(project, region string, addr *compute.Address) (*compute.Operation, error)
	Delete(project, region, name string) (*compute.Operation, error)
	Get(project, region, name string) (*compute.Address, error)
	List(ctx context.Context, project, region string) ([]*compute.Address, error)
	ListWithFilter(project, region, filter string) ([]*compute.Address, error)
}

type addressClientImpl struct {
	srv  *compute.AddressesService
	gsrv *compute.GlobalAddressesService
}

var _ AddressClient = &addressClientImpl{}

func (c *addressClientImpl) Insert(project, region string, addr *compute.Address) (*compute.Operation, error) {
	if region == "" {
		return c.gsrv.Insert(project, addr).Do()
	}
	return c.srv.Insert(project, region, addr).Do()
}

func (c *addressClientImpl) Delete(project, region, name string) (*compute.Operation, error) {
	if region == "" {
		return c.gsrv.Delete(project, name).Do()
	}
	return c.srv.Delete(project, region, name).Do()
}

func (c *addressClientImpl) Get(project, region, name string) (*compute.Address, error) {
	if region == "" {
		return c.gsrv.Get(project, name).Do()
	}
	return c.srv.Get(project, region, name).Do()
}

func (c *addressClientImpl) List(ctx context.Context, project, region string) ([]*compute.Address, error) {
	var addrs []*compute.Address
	if region == "" {
		if err := c.gsrv.List(project).Pages(ctx, func(p *compute.AddressList) error {
			addrs = append(addrs, p.Items...)
			return nil
		}); err != nil {
			return nil, err
		}
		return addrs, nil
	}
	if err := c.srv.List(project, region).Pages(ctx, func(p *compute.AddressList) error {
		addrs = append(addrs, p.Items...)
		return nil
	}); err != nil {
		return nil, err
	}
	return addrs, nil
}

func (c *addressClientImpl) ListWithFilter(project, region, filter string) ([]*compute.Address, error) {
	if region == "" {
		addrs, err := c.gsrv.List(project).Filter(filter).Do()
		if err != nil {
			return nil, err
		}
		return addrs.Items, nil
	}
	addrs, err := c.srv.List(project, region).Filter(filter).Do()
	if err != nil {
		return nil, err
	}
	return addrs.Items, nil
}

type FirewallClient interface {
	Insert(project string, fw *compute.Firewall) (*compute.Operation, error)
	Delete(project, name string) (*compute.Operation, error)
	Update(project, name string, fw *compute.Firewall) (*compute.Operation, error)
	Get(project, name string) (*compute.Firewall, error)
	List(ctx context.Context, project string) ([]*compute.Firewall, error)
}

type firewallClientImpl struct {
	srv *compute.FirewallsService
}

var _ FirewallClient = &firewallClientImpl{}

func (c *firewallClientImpl) Insert(project string, fw *compute.Firewall) (*compute.Operation, error) {
	return c.srv.Insert(project, fw).Do()
}

func (c *firewallClientImpl) Delete(project, name string) (*compute.Operation, error) {
	return c.srv.Delete(project, name).Do()
}

func (c *firewallClientImpl) Update(project, name string, fw *compute.Firewall) (*compute.Operation, error) {
	return c.srv.Update(project, name, fw).Do()
}

func (c *firewallClientImpl) Get(project, name string) (*compute.Firewall, error) {
	return c.srv.Get(project, name).Do()
}

func (c *firewallClientImpl) List(ctx context.Context, project string) ([]*compute.Firewall, error) {
	var fws []*compute.Firewall
	if err := c.srv.List(project).Pages(ctx, func(p *compute.FirewallList) error {
		fws = append(fws, p.Items...)
		return nil
	}); err != nil {
		return nil, err
	}
	return fws, nil
}

type RouterClient interface {
	Insert(project, region string, r *compute.Router) (*compute.Operation, error)
	Delete(project, region, name string) (*compute.Operation, error)
	Get(project, region, name string) (*compute.Router, error)
	List(ctx context.Context, project, region string) ([]*compute.Router, error)
}

type routerClientImpl struct {
	srv *compute.RoutersService
}

var _ RouterClient = &routerClientImpl{}

func (c *routerClientImpl) Insert(project, region string, r *compute.Router) (*compute.Operation, error) {
	return c.srv.Insert(project, region, r).Do()
}

func (c *routerClientImpl) Delete(project, region, name string) (*compute.Operation, error) {
	return c.srv.Delete(project, region, name).Do()
}

func (c *routerClientImpl) Get(project, region, name string) (*compute.Router, error) {
	return c.srv.Get(project, region, name).Do()
}

func (c *routerClientImpl) List(ctx context.Context, project, region string) ([]*compute.Router, error) {
	var rs []*compute.Router
	if err := c.srv.List(project, region).Pages(ctx, func(p *compute.RouterList) error {
		rs = append(rs, p.Items...)
		return nil
	}); err != nil {
		return nil, err
	}
	return rs, nil
}

type InstanceClient interface {
	Insert(project, zone string, i *compute.Instance) (*compute.Operation, error)
	Get(project, zone, name string) (*compute.Instance, error)
	List(ctx context.Context, project, zone string) ([]*compute.Instance, error)
	Delete(project, zone, name string) (*compute.Operation, error)
	SetMetadata(project, zone, name string, metadata *compute.Metadata) (*compute.Operation, error)
}

type instanceClientImpl struct {
	srv *compute.InstancesService
}

var _ InstanceClient = &instanceClientImpl{}

func (c *instanceClientImpl) Insert(project, zone string, i *compute.Instance) (*compute.Operation, error) {
	return c.srv.Insert(project, zone, i).Do()
}

func (c *instanceClientImpl) Get(project, zone, name string) (*compute.Instance, error) {
	return c.srv.Get(project, zone, name).Do()
}

func (c *instanceClientImpl) List(ctx context.Context, project, zone string) ([]*compute.Instance, error) {
	var insts []*compute.Instance
	if err := c.srv.List(project, zone).Pages(ctx, func(p *compute.InstanceList) error {
		insts = append(insts, p.Items...)
		return nil
	}); err != nil {
		return nil, err
	}
	return insts, nil
}

func (c *instanceClientImpl) Delete(project, zone, name string) (*compute.Operation, error) {
	return c.srv.Delete(project, zone, name).Do()
}

func (c *instanceClientImpl) SetMetadata(project, zone, name string, metadata *compute.Metadata) (*compute.Operation, error) {
	return c.srv.SetMetadata(project, zone, name, metadata).Do()
}

type InstanceTemplateClient interface {
	Insert(project string, template *compute.InstanceTemplate) (*compute.Operation, error)
	Delete(project, name string) (*compute.Operation, error)
	List(ctx context.Context, project string) ([]*compute.InstanceTemplate, error)
}

type instanceTemplateClientImpl struct {
	srv *compute.InstanceTemplatesService
}

var _ InstanceTemplateClient = &instanceTemplateClientImpl{}

func (c *instanceTemplateClientImpl) Insert(project string, template *compute.InstanceTemplate) (*compute.Operation, error) {
	return c.srv.Insert(project, template).Do()
}

func (c *instanceTemplateClientImpl) Delete(project, name string) (*compute.Operation, error) {
	return c.srv.Delete(project, name).Do()
}

func (c *instanceTemplateClientImpl) List(ctx context.Context, project string) ([]*compute.InstanceTemplate, error) {
	var its []*compute.InstanceTemplate
	if err := c.srv.List(project).Pages(ctx, func(page *compute.InstanceTemplateList) error {
		its = append(its, page.Items...)
		return nil
	}); err != nil {
		return nil, err
	}
	return its, nil
}

type InstanceGroupManagerClient interface {
	Insert(project, zone string, i *compute.InstanceGroupManager) (*compute.Operation, error)
	Delete(project, zone, name string) (*compute.Operation, error)
	Get(project, zone, name string) (*compute.InstanceGroupManager, error)
	List(ctx context.Context, project, zone string) ([]*compute.InstanceGroupManager, error)
	ListManagedInstances(ctx context.Context, project, zone, name string) ([]*compute.ManagedInstance, error)
	RecreateInstances(project, zone, name, id string) (*compute.Operation, error)
	SetTargetPools(project, zone, name string, targetPools []string) (*compute.Operation, error)
	SetInstanceTemplate(project, zone, name, instanceTemplateURL string) (*compute.Operation, error)
	Resize(project, zone, name string, newSize int64) (*compute.Operation, error)
}

type instanceGroupManagerClientImpl struct {
	srv *compute.InstanceGroupManagersService
}

var _ InstanceGroupManagerClient = &instanceGroupManagerClientImpl{}

func (c *instanceGroupManagerClientImpl) Insert(project, zone string, i *compute.InstanceGroupManager) (*compute.Operation, error) {
	return c.srv.Insert(project, zone, i).Do()
}

func (c *instanceGroupManagerClientImpl) Delete(project, zone, name string) (*compute.Operation, error) {
	return c.srv.Delete(project, zone, name).Do()
}

func (c *instanceGroupManagerClientImpl) Get(project, zone, name string) (*compute.InstanceGroupManager, error) {
	return c.srv.Get(project, zone, name).Do()
}

func (c *instanceGroupManagerClientImpl) List(ctx context.Context, project, zone string) ([]*compute.InstanceGroupManager, error) {
	var ms []*compute.InstanceGroupManager
	if err := c.srv.List(project, zone).Pages(ctx, func(page *compute.InstanceGroupManagerList) error {
		ms = append(ms, page.Items...)
		return nil
	}); err != nil {
		return nil, err
	}
	return ms, nil
}

func (c *instanceGroupManagerClientImpl) ListManagedInstances(ctx context.Context, project, zone, name string) ([]*compute.ManagedInstance, error) {
	var instances []*compute.ManagedInstance
	if err := c.srv.ListManagedInstances(project, zone, name).Pages(ctx, func(page *compute.InstanceGroupManagersListManagedInstancesResponse) error {
		instances = append(instances, page.ManagedInstances...)
		return nil
	}); err != nil {
		return nil, err
	}
	return instances, nil
}

func (c *instanceGroupManagerClientImpl) RecreateInstances(project, zone, name, id string) (*compute.Operation, error) {
	req := &compute.InstanceGroupManagersRecreateInstancesRequest{
		Instances: []string{
			id,
		},
	}
	return c.srv.RecreateInstances(project, zone, name, req).Do()
}

func (c *instanceGroupManagerClientImpl) SetTargetPools(project, zone, name string, targetPools []string) (*compute.Operation, error) {
	req := &compute.InstanceGroupManagersSetTargetPoolsRequest{
		TargetPools: targetPools,
	}
	return c.srv.SetTargetPools(project, zone, name, req).Do()
}

func (c *instanceGroupManagerClientImpl) SetInstanceTemplate(project, zone, name, instanceTemplateURL string) (*compute.Operation, error) {
	req := &compute.InstanceGroupManagersSetInstanceTemplateRequest{
		InstanceTemplate: instanceTemplateURL,
	}
	return c.srv.SetInstanceTemplate(project, zone, name, req).Do()
}

func (c *instanceGroupManagerClientImpl) Resize(project, zone, name string, newSize int64) (*compute.Operation, error) {
	return c.srv.Resize(project, zone, name, newSize).Do()
}

type TargetPoolClient interface {
	Insert(project, region string, tp *compute.TargetPool) (*compute.Operation, error)
	Delete(project, region, name string) (*compute.Operation, error)
	Get(project, region, name string) (*compute.TargetPool, error)
	List(ctx context.Context, project, region string) ([]*compute.TargetPool, error)
	AddHealthCheck(project, region, name string, req *compute.TargetPoolsAddHealthCheckRequest) (*compute.Operation, error)
}

type targetPoolClientImpl struct {
	srv *compute.TargetPoolsService
}

var _ TargetPoolClient = &targetPoolClientImpl{}

func (c *targetPoolClientImpl) Insert(project, region string, tp *compute.TargetPool) (*compute.Operation, error) {
	return c.srv.Insert(project, region, tp).Do()
}

func (c *targetPoolClientImpl) Delete(project, region, name string) (*compute.Operation, error) {
	return c.srv.Delete(project, region, name).Do()
}

func (c *targetPoolClientImpl) Get(project, region, name string) (*compute.TargetPool, error) {
	return c.srv.Get(project, region, name).Do()
}

func (c *targetPoolClientImpl) AddHealthCheck(project, region, name string, req *compute.TargetPoolsAddHealthCheckRequest) (*compute.Operation, error) {
	return c.srv.AddHealthCheck(project, region, name, req).Do()
}

func (c *targetPoolClientImpl) List(ctx context.Context, project, region string) ([]*compute.TargetPool, error) {
	var tps []*compute.TargetPool
	if err := c.srv.List(project, region).Pages(ctx, func(p *compute.TargetPoolList) error {
		tps = append(tps, p.Items...)
		return nil
	}); err != nil {
		return nil, err
	}
	return tps, nil
}

type DiskClient interface {
	Insert(project, zone string, disk *compute.Disk) (*compute.Operation, error)
	Delete(project, zone, name string) (*compute.Operation, error)
	Get(project, zone, name string) (*compute.Disk, error)
	List(ctx context.Context, project, zone string) ([]*compute.Disk, error)
	AggregatedList(ctx context.Context, project string) ([]compute.DisksScopedList, error)
	SetLabels(project, zone, name string, req *compute.ZoneSetLabelsRequest) error
}

type diskClientImpl struct {
	srv *compute.DisksService
}

var _ DiskClient = &diskClientImpl{}

func (c *diskClientImpl) Insert(project, zone string, disk *compute.Disk) (*compute.Operation, error) {
	return c.srv.Insert(project, zone, disk).Do()
}

func (c *diskClientImpl) Delete(project, zone, name string) (*compute.Operation, error) {
	return c.srv.Delete(project, zone, name).Do()
}

func (c *diskClientImpl) Get(project, zone, name string) (*compute.Disk, error) {
	return c.srv.Get(project, zone, name).Do()
}

func (c *diskClientImpl) List(ctx context.Context, project, zone string) ([]*compute.Disk, error) {
	var disks []*compute.Disk
	if err := c.srv.List(project, zone).Pages(ctx, func(page *compute.DiskList) error {
		disks = append(disks, page.Items...)
		return nil
	}); err != nil {
		return nil, err
	}
	return disks, nil
}

func (c *diskClientImpl) AggregatedList(ctx context.Context, project string) ([]compute.DisksScopedList, error) {
	var disks []compute.DisksScopedList
	if err := c.srv.AggregatedList(project).Pages(ctx, func(page *compute.DiskAggregatedList) error {
		for _, list := range page.Items {
			disks = append(disks, list)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return disks, nil
}

func (c *diskClientImpl) SetLabels(project, zone, name string, req *compute.ZoneSetLabelsRequest) error {
	_, err := c.srv.SetLabels(project, zone, name, req).Do()
	return err
}
