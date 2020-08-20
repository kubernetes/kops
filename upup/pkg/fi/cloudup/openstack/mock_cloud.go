/*
Copyright 2020 The Kubernetes Authors.

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

package openstack

import (
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"

	"github.com/gophercloud/gophercloud"
	cinder "github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	az "github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/availabilityzones"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/servergroups"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/volumeattach"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/dns/v2/recordsets"
	"github.com/gophercloud/gophercloud/openstack/dns/v2/zones"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/listeners"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/monitors"
	v2pools "github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/pools"
	l3floatingip "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/routers"
	sg "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	sgr "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kops/cloudmock/openstack/mockblockstorage"
	"k8s.io/kops/cloudmock/openstack/mockcompute"
	"k8s.io/kops/cloudmock/openstack/mockdns"
	"k8s.io/kops/cloudmock/openstack/mockimage"
	"k8s.io/kops/cloudmock/openstack/mockloadbalancer"
	"k8s.io/kops/cloudmock/openstack/mocknetworking"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	dnsproviderdesignate "k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/openstack/designate"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

type MockCloud struct {
	MockCinderClient  *mockblockstorage.MockClient
	MockNeutronClient *mocknetworking.MockClient
	MockNovaClient    *mockcompute.MockClient
	MockDNSClient     *mockdns.MockClient
	MockLBClient      *mockloadbalancer.MockClient
	MockImageClient   *mockimage.MockClient
	region            string
	tags              map[string]string
	useOctavia        bool
	zones             []string
	extNetworkName    *string
	extSubnetName     *string
	floatingSubnet    *string
}

func InstallMockOpenstackCloud(region string) *MockCloud {
	i := BuildMockOpenstackCloud(region)
	openstackCloudInstances[region] = i
	return i
}

func BuildMockOpenstackCloud(region string) *MockCloud {
	return &MockCloud{
		region: region,
	}
}

var _ fi.Cloud = (*MockCloud)(nil)

func (c *MockCloud) ComputeClient() *gophercloud.ServiceClient {
	client := c.MockNovaClient.ServiceClient()
	client.UserAgent.Prepend("compute")
	return client
}

func (c *MockCloud) BlockStorageClient() *gophercloud.ServiceClient {
	client := c.MockCinderClient.ServiceClient()
	client.UserAgent.Prepend("blockstorage")
	return client
}

func (c *MockCloud) NetworkingClient() *gophercloud.ServiceClient {
	client := c.MockNeutronClient.ServiceClient()
	client.UserAgent.Prepend("networking")
	return client
}

func (c *MockCloud) LoadBalancerClient() *gophercloud.ServiceClient {
	client := c.MockLBClient.ServiceClient()
	client.UserAgent.Prepend("loadbalancer")
	return client
}

func (c *MockCloud) DNSClient() *gophercloud.ServiceClient {
	client := c.MockDNSClient.ServiceClient()
	client.UserAgent.Prepend("dns")
	return client
}

func (c *MockCloud) ImageClient() *gophercloud.ServiceClient {
	client := c.MockImageClient.ServiceClient()
	client.UserAgent.Prepend("image")
	return client
}

func (c *MockCloud) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return deleteGroup(c, g)
}

func (c *MockCloud) DeleteInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	return deleteInstance(c, i)
}

func (c *MockCloud) DetachInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	return detachInstance(c, i)
}

func (c *MockCloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return getCloudGroups(c, cluster, instancegroups, warnUnmatched, nodes)
}

func (c *MockCloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderOpenstack
}

func (c *MockCloud) DNS() (dnsprovider.Interface, error) {
	if c.MockDNSClient == nil {
		return nil, fmt.Errorf("MockDNS not set")
	}
	return dnsproviderdesignate.New(c.DNSClient()), nil
}

func (c *MockCloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return findVPCInfo(c, id, c.zones)
}

func (c *MockCloud) Region() string {
	return c.region
}

func (c *MockCloud) AppendTag(resource string, id string, tag string) error {
	return appendTag(c, resource, id, tag)
}

func (c *MockCloud) AssociateToPool(server *servers.Server, poolID string, opts v2pools.CreateMemberOpts) (association *v2pools.Member, err error) {
	return associateToPool(c, server, poolID, opts)
}

func (c *MockCloud) AttachVolume(serverID string, opts volumeattach.CreateOpts) (attachment *volumeattach.VolumeAttachment, err error) {
	return attachVolume(c, serverID, opts)
}

func (c *MockCloud) CreateFloatingIP(opts floatingips.CreateOpts) (fip *floatingips.FloatingIP, err error) {
	return createFloatingIP(c, opts)
}

func (c *MockCloud) CreateInstance(opt servers.CreateOptsBuilder) (*servers.Server, error) {
	return createInstance(c, opt)
}

func (c *MockCloud) CreateKeypair(opt keypairs.CreateOptsBuilder) (*keypairs.KeyPair, error) {
	return createKeypair(c, opt)
}
func (c *MockCloud) CreateL3FloatingIP(opts l3floatingip.CreateOpts) (fip *l3floatingip.FloatingIP, err error) {
	return createL3FloatingIP(c, opts)
}
func (c *MockCloud) CreateLB(opt loadbalancers.CreateOptsBuilder) (*loadbalancers.LoadBalancer, error) {
	return createLB(c, opt)
}

func (c *MockCloud) CreateListener(opts listeners.CreateOpts) (listener *listeners.Listener, err error) {
	return createListener(c, opts)
}

func (c *MockCloud) CreateNetwork(opt networks.CreateOptsBuilder) (*networks.Network, error) {
	return createNetwork(c, opt)
}

func (c *MockCloud) CreatePool(opts v2pools.CreateOpts) (pool *v2pools.Pool, err error) {
	return createPool(c, opts)
}

func (c *MockCloud) CreatePort(opt ports.CreateOptsBuilder) (*ports.Port, error) {
	return createPort(c, opt)
}

func (c *MockCloud) CreateRouter(opt routers.CreateOptsBuilder) (*routers.Router, error) {
	return createRouter(c, opt)
}

func (c *MockCloud) CreateRouterInterface(routerID string, opt routers.AddInterfaceOptsBuilder) (*routers.InterfaceInfo, error) {
	return createRouterInterface(c, routerID, opt)
}

func (c *MockCloud) CreateSecurityGroup(opt sg.CreateOptsBuilder) (*sg.SecGroup, error) {
	return createSecurityGroup(c, opt)
}

func (c *MockCloud) CreateSecurityGroupRule(opt sgr.CreateOptsBuilder) (*sgr.SecGroupRule, error) {
	return createSecurityGroupRule(c, opt)
}

func (c *MockCloud) CreateServerGroup(opt servergroups.CreateOptsBuilder) (*servergroups.ServerGroup, error) {
	return createServerGroup(c, opt)
}

func (c *MockCloud) CreateSubnet(opt subnets.CreateOptsBuilder) (*subnets.Subnet, error) {
	return createSubnet(c, opt)
}

func (c *MockCloud) CreateVolume(opt cinder.CreateOptsBuilder) (*cinder.Volume, error) {
	return createVolume(c, opt)
}

func (c *MockCloud) DefaultInstanceType(cluster *kops.Cluster, ig *kops.InstanceGroup) (string, error) {
	return defaultInstanceType(c, cluster, ig)
}

func (c *MockCloud) DeleteFloatingIP(id string) (err error) {
	return deleteFloatingIP(c, id)
}
func (c *MockCloud) DeleteInstanceWithID(instanceID string) error {
	return deleteInstanceWithID(c, instanceID)
}
func (c *MockCloud) DeleteKeyPair(name string) error {
	return deleteKeyPair(c, name)
}

func (c *MockCloud) DeleteL3FloatingIP(id string) (err error) {
	return deleteL3FloatingIP(c, id)
}

func (c *MockCloud) DeleteLB(lbID string, opts loadbalancers.DeleteOpts) error {
	return deleteLB(c, lbID, opts)
}

func (c *MockCloud) DeleteListener(listenerID string) error {
	return deleteListener(c, listenerID)
}

func (c *MockCloud) DeleteMonitor(monitorID string) error {
	return deleteMonitor(c, monitorID)
}
func (c *MockCloud) DeleteNetwork(networkID string) error {
	return deleteNetwork(c, networkID)
}
func (c *MockCloud) DeletePool(poolID string) error {
	return deletePool(c, poolID)
}

func (c *MockCloud) DeletePort(portID string) error {
	return deletePort(c, portID)
}

func (c *MockCloud) DeleteRouter(routerID string) error {
	return deleteRouter(c, routerID)
}

func (c *MockCloud) DeleteSecurityGroup(sgID string) error {
	return deleteSecurityGroup(c, sgID)
}
func (c *MockCloud) DeleteSecurityGroupRule(ruleID string) error {
	return deleteSecurityGroupRule(c, ruleID)
}
func (c *MockCloud) DeleteRouterInterface(routerID string, opt routers.RemoveInterfaceOptsBuilder) error {
	return deleteRouterInterface(c, routerID, opt)
}

func (c *MockCloud) DeleteServerGroup(groupID string) error {
	return deleteServerGroup(c, groupID)
}

func (c *MockCloud) DeleteSubnet(subnetID string) error {
	return deleteSubnet(c, subnetID)
}

func (c *MockCloud) DeleteTag(resource string, id string, tag string) error {
	return deleteTag(c, resource, id, tag)
}
func (c *MockCloud) DeleteVolume(volumeID string) error {
	return deleteVolume(c, volumeID)
}

func (c *MockCloud) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	return findClusterStatus(c, cluster)
}

func (c *MockCloud) FindNetworkBySubnetID(subnetID string) (*networks.Network, error) {
	return findNetworkBySubnetID(c, subnetID)
}
func (c *MockCloud) GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error) {
	return getApiIngressStatus(c, cluster)
}
func (c *MockCloud) GetCloudTags() map[string]string {
	return c.tags
}
func (c *MockCloud) GetExternalNetwork() (net *networks.Network, err error) {
	return getExternalNetwork(c, *c.extNetworkName)
}
func (c *MockCloud) GetExternalSubnet() (subnet *subnets.Subnet, err error) {
	return getExternalSubnet(c, c.extSubnetName)
}

func (c *MockCloud) GetFloatingIP(id string) (fip *floatingips.FloatingIP, err error) {
	return getFloatingIP(c, id)
}

func (c *MockCloud) GetImage(name string) (*images.Image, error) {
	return getImage(c, name)
}

func (c *MockCloud) GetFlavor(name string) (*flavors.Flavor, error) {
	return getFlavor(c, name)
}

func (c *MockCloud) GetInstance(id string) (*servers.Server, error) {
	return getInstance(c, id)
}

func (c *MockCloud) GetKeypair(name string) (*keypairs.KeyPair, error) {
	return getKeypair(c, name)
}

func (c *MockCloud) GetLB(loadbalancerID string) (lb *loadbalancers.LoadBalancer, err error) {
	return getLB(c, loadbalancerID)
}
func (c *MockCloud) GetNetwork(id string) (*networks.Network, error) {
	return getNetwork(c, id)
}

func (c *MockCloud) GetLBFloatingSubnet() (subnet *subnets.Subnet, err error) {
	return getLBFloatingSubnet(c, c.floatingSubnet)
}

func (c *MockCloud) GetPool(poolID string, memberID string) (member *v2pools.Member, err error) {
	return getPool(c, poolID, memberID)
}

func (c *MockCloud) GetPort(id string) (*ports.Port, error) {
	return getPort(c, id)
}

func (c *MockCloud) GetStorageAZFromCompute(computeAZ string) (*az.AvailabilityZone, error) {
	return getStorageAZFromCompute(c, computeAZ)
}

func (c *MockCloud) GetSubnet(subnetID string) (*subnets.Subnet, error) {
	return getSubnet(c, subnetID)
}

func (c *MockCloud) ListAvailabilityZones(serviceClient *gophercloud.ServiceClient) (azList []az.AvailabilityZone, err error) {
	return listAvailabilityZones(c, serviceClient)
}
func (c *MockCloud) ListDNSZones(opt zones.ListOptsBuilder) ([]zones.Zone, error) {
	return listDNSZones(c, opt)
}
func (c *MockCloud) ListDNSRecordsets(zoneID string, opt recordsets.ListOptsBuilder) ([]recordsets.RecordSet, error) {
	return listDNSRecordsets(c, zoneID, opt)
}
func (c *MockCloud) ListFloatingIPs() (fips []floatingips.FloatingIP, err error) {
	return listFloatingIPs(c)
}

func (c *MockCloud) ListInstances(opt servers.ListOptsBuilder) ([]servers.Server, error) {
	return listInstances(c, opt)
}
func (c *MockCloud) ListKeypairs() ([]keypairs.KeyPair, error) {
	return listKeypairs(c)
}
func (c *MockCloud) ListL3FloatingIPs(opts l3floatingip.ListOpts) (fips []l3floatingip.FloatingIP, err error) {
	return listL3FloatingIPs(c, opts)
}

func (c *MockCloud) ListLBs(opt loadbalancers.ListOptsBuilder) (lbs []loadbalancers.LoadBalancer, err error) {
	return listLBs(c, opt)
}
func (c *MockCloud) ListListeners(opts listeners.ListOpts) (listenerList []listeners.Listener, err error) {
	return listListeners(c, opts)
}
func (c *MockCloud) ListMonitors(opts monitors.ListOpts) (monitorList []monitors.Monitor, err error) {
	return listMonitors(c, opts)
}

func (c *MockCloud) ListNetworks(opt networks.ListOptsBuilder) ([]networks.Network, error) {
	return listNetworks(c, opt)
}
func (c *MockCloud) ListPools(opts v2pools.ListOpts) (poolList []v2pools.Pool, err error) {
	return listPools(c, opts)
}

func (c *MockCloud) ListPorts(opt ports.ListOptsBuilder) ([]ports.Port, error) {
	return listPorts(c, opt)
}

func (c *MockCloud) ListRouters(opt routers.ListOpts) ([]routers.Router, error) {
	return listRouters(c, opt)
}

func (c *MockCloud) ListSecurityGroups(opt sg.ListOpts) ([]sg.SecGroup, error) {
	return listSecurityGroups(c, opt)
}

func (c *MockCloud) ListSecurityGroupRules(opt sgr.ListOpts) ([]sgr.SecGroupRule, error) {
	return listSecurityGroupRules(c, opt)
}

func (c *MockCloud) ListServerFloatingIPs(instanceID string) ([]*string, error) {
	return listServerFloatingIPs(c, instanceID, true)
}
func (c *MockCloud) ListServerGroups() ([]servergroups.ServerGroup, error) {
	return listServerGroups(c)
}
func (c *MockCloud) ListSubnets(opt subnets.ListOptsBuilder) ([]subnets.Subnet, error) {
	return listSubnets(c, opt)
}

func (c *MockCloud) ListVolumes(opt cinder.ListOptsBuilder) ([]cinder.Volume, error) {
	return listVolumes(c, opt)
}

func (c *MockCloud) SetExternalNetwork(name *string) {
	c.extNetworkName = name
}

func (c *MockCloud) SetExternalSubnet(name *string) {
	c.extSubnetName = name
}

func (c *MockCloud) SetLBFloatingSubnet(name *string) {
	c.floatingSubnet = name
}
func (c *MockCloud) SetVolumeTags(id string, tags map[string]string) error {
	return setVolumeTags(c, id, tags)
}

func (c *MockCloud) UseOctavia() bool {
	return c.useOctavia
}

func (c *MockCloud) UseZones(zones []string) {
	c.zones = zones
}
