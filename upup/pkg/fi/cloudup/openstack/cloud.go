/*
Copyright 2019 The Kubernetes Authors.

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
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	cinder "github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	az "github.com/gophercloud/gophercloud/v2/openstack/compute/v2/availabilityzones"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/keypairs"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servergroups"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/volumeattach"
	"github.com/gophercloud/gophercloud/v2/openstack/dns/v2/recordsets"
	"github.com/gophercloud/gophercloud/v2/openstack/dns/v2/zones"
	"github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
	"github.com/gophercloud/gophercloud/v2/openstack/loadbalancer/v2/apiversions"
	"github.com/gophercloud/gophercloud/v2/openstack/loadbalancer/v2/listeners"
	"github.com/gophercloud/gophercloud/v2/openstack/loadbalancer/v2/loadbalancers"
	"github.com/gophercloud/gophercloud/v2/openstack/loadbalancer/v2/monitors"
	v2pools "github.com/gophercloud/gophercloud/v2/openstack/loadbalancer/v2/pools"
	l3floatingip "github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/routers"
	sg "github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/security/groups"
	sgr "github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/security/rules"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/openstack/designate"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

const (
	TagNameEtcdClusterPrefix = "k8s.io/etcd/"
	TagNameRolePrefix        = "k8s.io/role/"
	TagClusterName           = "KubernetesCluster"
	TagRoleControlPlane      = "control-plane"
	TagRoleMaster            = "master"
	TagKopsInstanceGroup     = "KopsInstanceGroup"
	TagKopsNetwork           = "KopsNetwork"
	TagKopsName              = "KopsName"
	TagKopsRole              = "KopsRole"
	ResourceTypePort         = "ports"
	ResourceTypeNetwork      = "networks"
	ResourceTypeSubnet       = "subnets"
)

// readBackoff is the backoff strategy for openstack read retries.
var readBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   1.5,
	Jitter:   0.1,
	Steps:    10,
}

// writeBackoff is the backoff strategy for openstack write retries.
var writeBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   2,
	Jitter:   0.1,
	Steps:    5,
}

// deleteBackoff is the backoff strategy for openstack delete retries.
var deleteBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   5,
	Jitter:   0.1,
	Steps:    4,
}

type OpenstackCloud interface {
	fi.Cloud
	ComputeClient() *gophercloud.ServiceClient
	BlockStorageClient() *gophercloud.ServiceClient
	NetworkingClient() *gophercloud.ServiceClient
	LoadBalancerClient() *gophercloud.ServiceClient
	DNSClient() *gophercloud.ServiceClient
	ImageClient() *gophercloud.ServiceClient
	UseOctavia() bool
	UseZones([]string)

	// GetInstance will return a openstack server provided its ID
	GetInstance(id string) (*servers.Server, error)

	// ListInstances will return a slice of openstack servers provided list opts
	ListInstances(servers.ListOptsBuilder) ([]servers.Server, error)

	// CreateInstance will create an openstack server provided create opts
	CreateInstance(servers.CreateOptsBuilder, servers.SchedulerHintOptsBuilder, string) (*servers.Server, error)

	// DeleteInstanceWithID will delete instance
	DeleteInstanceWithID(instanceID string) error

	// SetVolumeTags will set the tags for the Cinder volume
	SetVolumeTags(id string, tags map[string]string) error

	// GetCloudTags will return the tags attached on cloud
	GetCloudTags() map[string]string

	// ListVolumes will return the Cinder volumes which match the options
	ListVolumes(opt cinder.ListOptsBuilder) ([]cinder.Volume, error)

	// CreateVolume will create a new Cinder Volume
	CreateVolume(opt cinder.CreateOptsBuilder) (*cinder.Volume, error)
	AttachVolume(serverID string, opt volumeattach.CreateOpts) (*volumeattach.VolumeAttachment, error)

	// DeleteVolume will delete volume
	DeleteVolume(volumeID string) error

	// ListSecurityGroups will return the Neutron security groups which match the options
	ListSecurityGroups(opt sg.ListOpts) ([]sg.SecGroup, error)

	// CreateSecurityGroup will create a new Neutron security group
	CreateSecurityGroup(opt sg.CreateOptsBuilder) (*sg.SecGroup, error)

	// DeleteSecurityGroup will delete securitygroup
	DeleteSecurityGroup(sgID string) error

	// DeleteSecurityGroupRule will delete securitygrouprule
	DeleteSecurityGroupRule(ruleID string) error

	// ListSecurityGroupRules will return the Neutron security group rules which match the options
	ListSecurityGroupRules(opt sgr.ListOpts) ([]sgr.SecGroupRule, error)

	// CreateSecurityGroupRule will create a new Neutron security group rule
	CreateSecurityGroupRule(opt sgr.CreateOptsBuilder) (*sgr.SecGroupRule, error)

	// GetNetwork will return the Neutron network which match the id
	GetNetwork(networkID string) (*networks.Network, error)

	// FindNetworkBySubnetID will return network
	FindNetworkBySubnetID(subnetID string) (*networks.Network, error)

	// GetSubnet returns subnet using subnet id
	GetSubnet(subnetID string) (*subnets.Subnet, error)

	// ListNetworks will return the Neutron networks which match the options
	ListNetworks(opt networks.ListOptsBuilder) ([]networks.Network, error)

	// GetExternalNetwork will return the Neutron networks with the router:external property
	GetExternalNetwork() (*networks.Network, error)

	// GetExternalSubnet will return the subnet for floatingip which is used in external router
	GetExternalSubnet() (*subnets.Subnet, error)

	// GetLBFloatingSubnet will return the subnet for floatingip which is used in lb
	GetLBFloatingSubnet() (*subnets.Subnet, error)

	// CreateNetwork will create a new Neutron network
	CreateNetwork(opt networks.CreateOptsBuilder) (*networks.Network, error)

	// DeleteNetwork will delete neutron network
	DeleteNetwork(networkID string) error

	// AppendTag appends tag to resource
	AppendTag(resource string, id string, tag string) error

	// DeleteTag removes tag from resource
	DeleteTag(resource string, id string, tag string) error

	// ListRouters will return the Neutron routers which match the options
	ListRouters(opt routers.ListOpts) ([]routers.Router, error)

	// CreateRouter will create a new Neutron router
	CreateRouter(opt routers.CreateOptsBuilder) (*routers.Router, error)

	// DeleteRouter will delete neutron router
	DeleteRouter(routerID string) error

	// DeleteSubnet will delete neutron subnet
	DeleteSubnet(subnetID string) error

	// ListSubnets will return the Neutron subnets which match the options
	ListSubnets(opt subnets.ListOptsBuilder) ([]subnets.Subnet, error)

	// CreateSubnet will create a new Neutron subnet
	CreateSubnet(opt subnets.CreateOptsBuilder) (*subnets.Subnet, error)

	// GetKeypair will return the Nova keypair
	GetKeypair(name string) (*keypairs.KeyPair, error)

	// ListKeypairs will return the all Nova keypairs
	ListKeypairs() ([]keypairs.KeyPair, error)

	// DeleteKeyPair will delete a Nova keypair
	DeleteKeyPair(name string) error

	// CreateKeypair will create a new Nova Keypair
	CreateKeypair(opt keypairs.CreateOptsBuilder) (*keypairs.KeyPair, error)
	CreatePort(opt ports.CreateOptsBuilder) (*ports.Port, error)

	// GetPort will return a Neutron port by ID
	GetPort(id string) (*ports.Port, error)

	// UpdatePort will update a Neutron port by ID and options
	UpdatePort(id string, opt ports.UpdateOptsBuilder) (*ports.Port, error)

	// ListPorts will return the Neutron ports which match the options
	ListPorts(opt ports.ListOptsBuilder) ([]ports.Port, error)

	// DeletePort will delete a neutron port
	DeletePort(portID string) error

	// CreateRouterInterface will create a new Neutron router interface
	CreateRouterInterface(routerID string, opt routers.AddInterfaceOptsBuilder) (*routers.InterfaceInfo, error)

	// DeleteRouterInterface will delete router interface from subnet
	DeleteRouterInterface(routerID string, opt routers.RemoveInterfaceOptsBuilder) error

	// CreateServerGroup will create a new server group.
	CreateServerGroup(opt servergroups.CreateOptsBuilder) (*servergroups.ServerGroup, error)

	// ListServerGroups will list available server groups
	ListServerGroups(opts servergroups.ListOptsBuilder) ([]servergroups.ServerGroup, error)

	// DeleteServerGroup will delete a nova server group
	DeleteServerGroup(groupID string) error

	// ListDNSZones will list available DNS zones
	ListDNSZones(opt zones.ListOptsBuilder) ([]zones.Zone, error)

	// ListDNSRecordsets will list the DNS recordsets for the given zone id
	ListDNSRecordsets(zoneID string, opt recordsets.ListOptsBuilder) ([]recordsets.RecordSet, error)
	DeleteDNSRecordset(zoneID string, rrsetID string) error

	GetLB(loadbalancerID string) (*loadbalancers.LoadBalancer, error)
	GetLBStats(loadbalancerID string) (*loadbalancers.Stats, error)
	CreateLB(opt loadbalancers.CreateOptsBuilder) (*loadbalancers.LoadBalancer, error)
	ListLBs(opt loadbalancers.ListOptsBuilder) ([]loadbalancers.LoadBalancer, error)
	UpdateMemberInPool(poolID string, memberID string, opts v2pools.UpdateMemberOptsBuilder) (*v2pools.Member, error)
	ListPoolMembers(poolID string, opts v2pools.ListMembersOpts) ([]v2pools.Member, error)

	// DeleteLB will delete loadbalancer
	DeleteLB(lbID string, opt loadbalancers.DeleteOpts) error

	// DefaultInstanceType determines a suitable instance type for the specified instance group
	DefaultInstanceType(cluster *kops.Cluster, ig *kops.InstanceGroup) (string, error)

	// Returns the availability zones for the service client passed (compute, volume, network)
	ListAvailabilityZones(serviceClient *gophercloud.ServiceClient) ([]az.AvailabilityZone, error)
	AssociateToPool(server *servers.Server, poolID string, opts v2pools.CreateMemberOpts) (*v2pools.Member, error)
	CreatePool(opts v2pools.CreateOpts) (*v2pools.Pool, error)
	CreatePoolMonitor(opts monitors.CreateOpts) (*monitors.Monitor, error)
	GetPool(poolID string) (*v2pools.Pool, error)
	GetPoolMember(poolID string, memberID string) (*v2pools.Member, error)
	ListPools(v2pools.ListOpts) ([]v2pools.Pool, error)

	// ListMonitors will list HealthMonitors matching the provided options
	ListMonitors(monitors.ListOpts) ([]monitors.Monitor, error)

	// DeleteMonitor will delete a Pool resources Health Monitor
	DeleteMonitor(monitorID string) error

	// DeletePool will delete loadbalancer pool
	DeletePool(poolID string) error
	ListListeners(opts listeners.ListOpts) ([]listeners.Listener, error)
	CreateListener(opts listeners.CreateOpts) (*listeners.Listener, error)

	// DeleteListener will delete loadbalancer listener
	DeleteListener(listenerID string) error
	GetStorageAZFromCompute(azName string) (*az.AvailabilityZone, error)
	GetL3FloatingIP(id string) (fip *l3floatingip.FloatingIP, err error)
	GetImage(name string) (i *images.Image, err error)
	GetFlavor(name string) (f *flavors.Flavor, err error)
	ListServerFloatingIPs(id string) ([]*string, error)
	ListL3FloatingIPs(opts l3floatingip.ListOpts) (fips []l3floatingip.FloatingIP, err error)
	CreateL3FloatingIP(opts l3floatingip.CreateOpts) (fip *l3floatingip.FloatingIP, err error)
	DeleteFloatingIP(id string) error
	DeleteL3FloatingIP(id string) error
	UseLoadBalancerVIPACL() (bool, error)
}

type openstackCloud struct {
	cinderClient    *gophercloud.ServiceClient
	neutronClient   *gophercloud.ServiceClient
	novaClient      *gophercloud.ServiceClient
	dnsClient       *gophercloud.ServiceClient
	lbClient        *gophercloud.ServiceClient
	glanceClient    *gophercloud.ServiceClient
	extNetworkName  *string
	extSubnetName   *string
	floatingSubnet  *string
	tags            map[string]string
	region          string
	useOctavia      bool
	zones           []string
	floatingEnabled bool
	useVIPACL       *bool
}

var _ fi.Cloud = &openstackCloud{}

var openstackCloudInstances = make(map[string]OpenstackCloud)

func NewOpenstackCloud(cluster *kops.Cluster, uagent string) (OpenstackCloud, error) {
	config := vfs.OpenstackConfig{}

	region, err := config.GetRegion()
	if err != nil {
		return nil, fmt.Errorf("error finding openstack region: %v", err)
	}

	raw := openstackCloudInstances[region]
	if raw != nil {
		return raw, nil
	}

	authOption, err := config.GetCredential()
	if err != nil {
		return nil, err
	}

	provider, err := openstack.NewClient(authOption.IdentityEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building openstack provider client: %v", err)
	}
	ua := gophercloud.UserAgent{}
	ua.Prepend(fmt.Sprintf("kops/%s", uagent))
	provider.UserAgent = ua
	klog.V(4).Infof("Using user-agent %s", ua.Join())

	if cluster != nil && cluster.Spec.CloudProvider.Openstack != nil && cluster.Spec.CloudProvider.Openstack.InsecureSkipVerify != nil {
		tlsconfig := &tls.Config{}
		tlsconfig.InsecureSkipVerify = fi.ValueOf(cluster.Spec.CloudProvider.Openstack.InsecureSkipVerify)
		transport := &http.Transport{TLSClientConfig: tlsconfig}
		provider.HTTPClient = http.Client{
			Transport: transport,
		}
	}

	klog.V(2).Info("authenticating to keystone")

	err = openstack.Authenticate(context.TODO(), provider, authOption)
	if err != nil {
		return nil, fmt.Errorf("error building openstack authenticated client: %v", err)
	}

	if cluster != nil {
		hasDNS := cluster.PublishesDNSRecords()
		tags := map[string]string{
			TagClusterName: cluster.Name,
		}
		return buildClients(provider, tags, cluster.Spec.CloudProvider.Openstack, config, region, hasDNS)
	}
	// used by protokube
	return buildClients(provider, nil, nil, config, region, false)
}

func buildClients(provider *gophercloud.ProviderClient, tags map[string]string, spec *kops.OpenstackSpec, config vfs.OpenstackConfig, region string, hasDNS bool) (OpenstackCloud, error) {
	cinderClient, err := openstack.NewBlockStorageV3(provider, gophercloud.EndpointOpts{
		Type:   "volumev3",
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("error building cinder client: %w", err)
	}

	neutronClient, err := openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{
		Type:   "network",
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("error building neutron client: %w", err)
	}

	novaClient, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Type:   "compute",
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("error building nova client: %w", err)
	}
	// 2.47 is the minimum version where the compute API /server/details returns flavor names
	novaClient.Microversion = "2.47"

	glanceClient, err := openstack.NewImageV2(provider, gophercloud.EndpointOpts{
		Type:   "image",
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("error building glance client: %w", err)
	}

	var dnsClient *gophercloud.ServiceClient
	if hasDNS {
		dnsClient, err = openstack.NewDNSV2(provider, gophercloud.EndpointOpts{
			Type:   "dns",
			Region: region,
		})
		if err != nil {
			return nil, fmt.Errorf("error building dns client: %w", err)
		}
	}

	c := &openstackCloud{
		cinderClient:  cinderClient,
		neutronClient: neutronClient,
		novaClient:    novaClient,
		dnsClient:     dnsClient,
		glanceClient:  glanceClient,
		tags:          tags,
		region:        region,
		useOctavia:    false,
	}

	setFloatingIPSupport(c, spec)
	err = buildLoadBalancerClient(c, spec, provider, region)
	if err != nil {
		return nil, fmt.Errorf("failed to build load balancer client: %w", err)
	}
	openstackCloudInstances[region] = c

	return c, nil

}

func setFloatingIPSupport(c *openstackCloud, spec *kops.OpenstackSpec) {
	if spec == nil || spec.Router == nil {
		c.floatingEnabled = false
		klog.V(2).Infof("Floating IP support for OpenStack disabled")
		return
	}

	c.floatingEnabled = true
	c.extNetworkName = spec.Router.ExternalNetwork

	if spec.Router.ExternalSubnet != nil {
		c.extSubnetName = spec.Router.ExternalSubnet
	}
}

func buildLoadBalancerClient(c *openstackCloud, spec *kops.OpenstackSpec, provider *gophercloud.ProviderClient, region string) error {
	if spec == nil || spec.Loadbalancer == nil {
		klog.V(2).Infof("Loadbalancer support for OpenStack disabled")
		return nil
	}

	octavia := false
	if spec.Router != nil {
		if spec.Loadbalancer.FloatingNetworkID == nil &&
			spec.Loadbalancer.FloatingNetwork != nil {
			// This field is derived
			lbNet, err := c.ListNetworks(networks.ListOpts{
				Name: fi.ValueOf(spec.Loadbalancer.FloatingNetwork),
			})
			if err != nil || len(lbNet) != 1 {
				return fmt.Errorf("could not establish floating network id")
			}
			spec.Loadbalancer.FloatingNetworkID = fi.PtrTo(lbNet[0].ID)
		}

		if spec.Loadbalancer.UseOctavia != nil {
			octavia = fi.ValueOf(spec.Loadbalancer.UseOctavia)
		}
		if spec.Loadbalancer.FloatingSubnet != nil {
			c.floatingSubnet = spec.Loadbalancer.FloatingSubnet
		}
	} else if fi.ValueOf(spec.Loadbalancer.UseOctavia) {
		return fmt.Errorf("cluster configured to use octavia, but router was not configured")
	}
	c.useOctavia = octavia

	var lbClient *gophercloud.ServiceClient
	if octavia {
		klog.V(2).Infof("Openstack using Octavia lbaasv2 api")
		client, err := openstack.NewLoadBalancerV2(provider, gophercloud.EndpointOpts{
			Region: region,
		})
		if err != nil {
			return fmt.Errorf("error building lb client: %w", err)
		}
		lbClient = client
	} else {
		klog.V(2).Infof("Openstack using deprecated lbaasv2 api")
		client, err := openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{
			Region: region,
		})
		if err != nil {
			return fmt.Errorf("error building lb client: %w", err)
		}
		lbClient = client
	}
	c.lbClient = lbClient
	return nil
}

// UseZones add unique zone names to openstackcloud
func (c *openstackCloud) UseZones(zones []string) {
	c.zones = zones
}

func (c *openstackCloud) UseOctavia() bool {
	return c.useOctavia
}

func (c *openstackCloud) ComputeClient() *gophercloud.ServiceClient {
	return c.novaClient
}

func (c *openstackCloud) BlockStorageClient() *gophercloud.ServiceClient {
	return c.cinderClient
}

func (c *openstackCloud) NetworkingClient() *gophercloud.ServiceClient {
	return c.neutronClient
}

func (c *openstackCloud) LoadBalancerClient() *gophercloud.ServiceClient {
	return c.lbClient
}

func (c *openstackCloud) DNSClient() *gophercloud.ServiceClient {
	return c.dnsClient
}

func (c *openstackCloud) ImageClient() *gophercloud.ServiceClient {
	return c.glanceClient
}

func (c *openstackCloud) Region() string {
	return c.region
}

func (c *openstackCloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderOpenstack
}

func (c *openstackCloud) DNS() (dnsprovider.Interface, error) {
	provider, err := dnsprovider.GetDnsProvider(designate.ProviderName, nil)
	if err != nil {
		return nil, fmt.Errorf("error building (Designate) DNS provider: %v", err)
	}
	return provider, nil
}

// FindVPCInfo list subnets in network
func (c *openstackCloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return findVPCInfo(c, id, c.zones)
}

func findVPCInfo(c OpenstackCloud, id string, zones []string) (*fi.VPCInfo, error) {
	vpcInfo := &fi.VPCInfo{}
	// Find subnets in the network
	{
		if len(zones) == 0 {
			return nil, fmt.Errorf("could not initialize zones")
		}
		klog.V(2).Infof("Calling ListSubnets for subnets in Network %q", id)
		opt := subnets.ListOpts{
			NetworkID: id,
		}
		subnets, err := c.ListSubnets(opt)
		if err != nil {
			return nil, fmt.Errorf("error listing subnets in network %q: %v", id, err)
		}

		for index, subnet := range subnets {
			zone := zones[int(index)%len(zones)]
			subnetInfo := &fi.SubnetInfo{
				ID:   subnet.ID,
				CIDR: subnet.CIDR,
				Zone: zone,
			}
			vpcInfo.Subnets = append(vpcInfo.Subnets, subnetInfo)
		}
	}
	return vpcInfo, nil
}

// DeleteGroup in openstack will delete servergroup, instances and ports
func (c *openstackCloud) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return deleteGroup(c, g)
}

// InstanceInClusterAndIG checks if instance is in current cluster and instancegroup
func InstanceInClusterAndIG(instance servers.Server, clusterName string, instanceGroupName string) bool {
	value, ok := instance.Metadata[TagKopsInstanceGroup]
	if !ok || value != instanceGroupName {
		return false
	}
	cName, clusterok := instance.Metadata["k8s"]
	if !clusterok || cName != clusterName {
		return false
	}
	return true
}

func deletePorts(c OpenstackCloud, instanceGroupName string, clusterName string) error {
	tags := []string{
		fmt.Sprintf("%s=%s", TagClusterName, clusterName),
		fmt.Sprintf("%s=%s", TagKopsInstanceGroup, instanceGroupName),
	}

	ports, err := c.ListPorts(ports.ListOpts{Tags: strings.Join(tags, ",")})
	if err != nil {
		return fmt.Errorf("could not list ports %v", err)
	}

	for _, port := range ports {
		// previous approach was problematic:
		//   for example in case there is a group called "worker" and "worker-2", it will delete ports of "worker" as well,
		//   because there might be port names like:
		//     * "port-worker-2-<clusterName>"
		//     * "port-worker-20-<clusterName>"
		klog.V(2).Infof("Delete port '%s' (%s)", port.Name, port.ID)
		err := c.DeletePort(port.ID)

		// TODO:
		//   really give up after trying to delete one port? other ports will be orphaned
		//   better to try all ports and collect errors?
		if err != nil {
			return fmt.Errorf("could not delete port %q: %v", port.ID, err)
		}
	}

	return nil
}

func deleteGroup(c OpenstackCloud, g *cloudinstances.CloudInstanceGroup) error {
	cluster := g.Raw.(*kops.Cluster)
	allInstances, err := c.ListInstances(servers.ListOpts{
		Name: fmt.Sprintf("^%s", g.InstanceGroup.Name),
	})
	if err != nil {
		return err
	}

	instances := []servers.Server{}
	for _, instance := range allInstances {
		if !InstanceInClusterAndIG(instance, cluster.Name, g.InstanceGroup.Name) {
			continue
		}
		instances = append(instances, instance)
	}
	for _, instance := range instances {
		err := c.DeleteInstanceWithID(instance.ID)
		if err != nil {
			return fmt.Errorf("could not delete instance %q: %v", instance.ID, err)
		}
	}

	err = deletePorts(c, g.InstanceGroup.Name, cluster.Name)
	if err != nil {
		return err
	}

	sgName := g.InstanceGroup.Name
	if name, ok := g.InstanceGroup.Annotations[OS_ANNOTATION+SERVER_GROUP_NAME]; ok {
		sgName = name
	}
	sgs, err := c.ListServerGroups(servergroups.ListOpts{})
	if err != nil {
		return fmt.Errorf("could not list server groups %v", err)
	}

	for _, sg := range sgs {
		if fmt.Sprintf("%s-%s", cluster.Name, sgName) == sg.Name {
			if len(sg.Members) == 0 {
				err = c.DeleteServerGroup(sg.ID)
				if err != nil {
					return fmt.Errorf("could not delete server group %q: %v", sg.ID, err)
				}
			} else {
				klog.Infof("Server group %q still has members (IDs: %s), delete not executed", sg.ID, strings.Join(sg.Members, ", "))
			}
			break
		}
	}
	return nil
}

func (c *openstackCloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return getCloudGroups(c, cluster, instancegroups, warnUnmatched, nodes)
}

func getCloudGroups(c OpenstackCloud, cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	nodeMap := cloudinstances.GetNodeMap(nodes, cluster)
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)
	for _, ig := range instancegroups {
		var err error
		groups[ig.ObjectMeta.Name], err = osBuildCloudInstanceGroup(c, cluster, ig, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("error getting cloud instance group %q: %v", ig.ObjectMeta.Name, err)
		}
	}
	return groups, nil
}

func (c *openstackCloud) GetCloudTags() map[string]string {
	return c.tags
}

func (c *openstackCloud) UseLoadBalancerVIPACL() (bool, error) {
	if c.useVIPACL != nil {
		return *c.useVIPACL, nil
	}
	use, err := useLoadBalancerVIPACL(c)
	if err != nil {
		return false, err
	}
	c.useVIPACL = &use
	return use, nil
}

func useLoadBalancerVIPACL(c OpenstackCloud) (bool, error) {
	if c.LoadBalancerClient() == nil {
		return false, nil
	}
	allPages, err := apiversions.List(c.LoadBalancerClient()).AllPages(context.TODO())
	if err != nil {
		return false, err
	}
	versions, err := apiversions.ExtractAPIVersions(allPages)
	if err != nil {
		return false, err
	}
	if len(versions) == 0 {
		return false, fmt.Errorf("loadbalancer API versions not found")
	}
	ver, err := semver.ParseTolerant(versions[len(versions)-1].ID)
	if err != nil {
		return false, err
	}
	// https://github.com/kubernetes/cloud-provider-openstack/blob/721615aa256bbddbd481cfb4a887c3ab180c5563/pkg/util/openstack/loadbalancer.go#L108
	return ver.Compare(semver.MustParse("2.12.0")) > 0, nil
}

type Address struct {
	IPType string `mapstructure:"OS-EXT-IPS:type"`
	Addr   string
}

func (c *openstackCloud) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	return getApiIngressStatus(c, cluster)
}

func getApiIngressStatus(c OpenstackCloud, cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	if cluster.Spec.CloudProvider.Openstack.Loadbalancer != nil {
		return getLoadBalancerIngressStatus(c, cluster)
	} else {
		return getIPIngressStatus(c, cluster)
	}
}

func getLoadBalancerIngressStatus(c OpenstackCloud, cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	var ingresses []fi.ApiIngressStatus
	lbName := "api." + cluster.Name
	if cluster.Spec.API.PublicName != "" {
		lbName = cluster.Spec.API.PublicName
	}
	// Note that this must match OpenstackModel lb name
	klog.V(2).Infof("Querying Openstack to find Loadbalancers for API (%q)", cluster.Name)
	lbList, err := c.ListLBs(loadbalancers.ListOpts{
		Name: lbName,
	})
	if err != nil {
		return ingresses, fmt.Errorf("GetApiIngressStatus: Failed to list openstack loadbalancers: %v", err)
	}
	for _, lb := range lbList {
		// Must Find Floating IP related to this lb
		fips, err := c.ListL3FloatingIPs(l3floatingip.ListOpts{
			PortID: lb.VipPortID,
		})
		if err != nil {
			return ingresses, fmt.Errorf("GetApiIngressStatus: Failed to list floating IP's: %v", err)
		}
		for _, fip := range fips {
			if fip.PortID == lb.VipPortID {
				ingresses = append(ingresses, fi.ApiIngressStatus{
					IP: fip.FloatingIP,
				})
			}
		}
	}

	return ingresses, nil
}

func getIPIngressStatus(c OpenstackCloud, cluster *kops.Cluster) (ingresses []fi.ApiIngressStatus, err error) {
	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		instances, err := c.ListInstances(servers.ListOpts{})
		if err != nil {
			return false, fmt.Errorf("GetApiIngressStatus: Failed to list control plane nodes: %v", err)
		}
		for _, instance := range instances {
			val, ok := instance.Metadata["k8s"]
			val2, ok2 := instance.Metadata["KopsRole"]
			if ok && val == cluster.Name && ok2 {
				role, success := kops.ParseInstanceGroupRole(val2, false)
				if success && role == kops.InstanceGroupRoleControlPlane {
					ifName := instance.Metadata[TagKopsNetwork]
					address, err := GetServerFixedIP(&instance, ifName)
					if err == nil {
						ingresses = append(ingresses, fi.ApiIngressStatus{
							IP: address,
						})
					} else {
						ips, err := c.ListServerFloatingIPs(instance.ID)
						if err != nil {
							return false, err
						}
						for _, ip := range ips {
							ingresses = append(ingresses, fi.ApiIngressStatus{
								IP: fi.ValueOf(ip),
							})
						}
					}
				}
			}
		}
		return true, nil
	})
	if done {
		return ingresses, nil
	} else {
		if err == nil {
			err = wait.ErrWaitTimeout
		}
		return ingresses, err
	}
}

func isNotFound(err error) bool {
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return true
	}

	if _, ok := err.(gophercloud.ErrResourceNotFound); ok {
		return true
	}

	if errCode, ok := err.(gophercloud.ErrUnexpectedResponseCode); ok {
		if errCode.Actual == http.StatusNotFound {
			return true
		}
	}

	return false
}

func MakeCloudConfig(osc *kops.OpenstackSpec) []string {
	var lines []string

	// Support mapping of older keystone API
	tenantName := os.Getenv("OS_TENANT_NAME")
	if tenantName == "" {
		tenantName = os.Getenv("OS_PROJECT_NAME")
	}
	tenantID := os.Getenv("OS_TENANT_ID")
	if tenantID == "" {
		tenantID = os.Getenv("OS_PROJECT_ID")
	}
	lines = append(lines,
		fmt.Sprintf("auth-url=\"%s\"", os.Getenv("OS_AUTH_URL")),
		fmt.Sprintf("username=\"%s\"", os.Getenv("OS_USERNAME")),
		fmt.Sprintf("password=\"%s\"", os.Getenv("OS_PASSWORD")),
		fmt.Sprintf("region=\"%s\"", os.Getenv("OS_REGION_NAME")),
		fmt.Sprintf("tenant-id=\"%s\"", tenantID),
		fmt.Sprintf("tenant-name=\"%s\"", tenantName),
		fmt.Sprintf("domain-name=\"%s\"", os.Getenv("OS_DOMAIN_NAME")),
		fmt.Sprintf("domain-id=\"%s\"", os.Getenv("OS_DOMAIN_ID")),
		fmt.Sprintf("application-credential-id=\"%s\"", os.Getenv("OS_APPLICATION_CREDENTIAL_ID")),
		fmt.Sprintf("application-credential-secret=\"%s\"", os.Getenv("OS_APPLICATION_CREDENTIAL_SECRET")),
		"",
	)

	if lb := osc.Loadbalancer; lb != nil {
		ingressHostnameSuffix := "nip.io"
		if fi.ValueOf(lb.IngressHostnameSuffix) != "" {
			ingressHostnameSuffix = fi.ValueOf(lb.IngressHostnameSuffix)
		}

		lines = append(lines,
			"[LoadBalancer]",
			fmt.Sprintf("floating-network-id=%s", fi.ValueOf(lb.FloatingNetworkID)),
			fmt.Sprintf("lb-method=%s", fi.ValueOf(lb.Method)),
			fmt.Sprintf("lb-provider=%s", fi.ValueOf(lb.Provider)),
			fmt.Sprintf("use-octavia=%t", fi.ValueOf(lb.UseOctavia)),
			fmt.Sprintf("manage-security-groups=%t", fi.ValueOf(lb.ManageSecGroups)),
			fmt.Sprintf("enable-ingress-hostname=%t", fi.ValueOf(lb.EnableIngressHostname)),
			fmt.Sprintf("ingress-hostname-suffix=%s", ingressHostnameSuffix),
			"",
		)

		if monitor := osc.Monitor; monitor != nil {
			lines = append(lines,
				"create-monitor=yes",
				fmt.Sprintf("monitor-delay=%s", fi.ValueOf(monitor.Delay)),
				fmt.Sprintf("monitor-timeout=%s", fi.ValueOf(monitor.Timeout)),
				fmt.Sprintf("monitor-max-retries=%d", fi.ValueOf(monitor.MaxRetries)),
				"",
			)
		}
	}

	if bs := osc.BlockStorage; bs != nil {
		// Block Storage Config
		lines = append(lines,
			"[BlockStorage]",
			fmt.Sprintf("bs-version=%s", fi.ValueOf(bs.Version)),
			fmt.Sprintf("ignore-volume-az=%t", fi.ValueOf(bs.IgnoreAZ)),
			fmt.Sprintf("ignore-volume-microversion=%t", fi.ValueOf(bs.IgnoreVolumeMicroVersion)),
			"")
	}

	if networking := osc.Network; networking != nil {
		// Networking Config
		// https://github.com/kubernetes/cloud-provider-openstack/blob/master/docs/openstack-cloud-controller-manager/using-openstack-cloud-controller-manager.md#networking
		var networkingLines []string

		if networking.IPv6SupportDisabled != nil {
			networkingLines = append(networkingLines, fmt.Sprintf("ipv6-support-disabled=%t", fi.ValueOf(networking.IPv6SupportDisabled)))
		}
		for _, name := range networking.PublicNetworkNames {
			networkingLines = append(networkingLines, fmt.Sprintf("public-network-name=%s", fi.ValueOf(name)))
		}
		for _, name := range networking.InternalNetworkNames {
			networkingLines = append(networkingLines, fmt.Sprintf("internal-network-name=%s", fi.ValueOf(name)))
		}
		if networking.AddressSortOrder != nil {
			networkingLines = append(networkingLines, fmt.Sprintf("address-sort-order=%s", fi.ValueOf(networking.AddressSortOrder)))
		}

		if len(networkingLines) > 0 {
			lines = append(lines, "[Networking]")
			lines = append(lines, networkingLines...)
			lines = append(lines, "")
		}
	}

	return lines
}
