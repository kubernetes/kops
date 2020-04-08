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
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud"
	os "github.com/gophercloud/gophercloud/openstack"
	cinder "github.com/gophercloud/gophercloud/openstack/blockstorage/v2/volumes"
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
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/openstack/designate"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

const (
	TagNameEtcdClusterPrefix = "k8s.io/etcd/"
	TagNameRolePrefix        = "k8s.io/role/"
	TagClusterName           = "KubernetesCluster"
	TagRoleMaster            = "master"
	TagKopsNetwork           = "KopsNetwork"
	TagNameDetach            = "KopsDetach"
	TagKopsName              = "KopsName"
	ResourceTypePort         = "ports"
	ResourceTypeNetwork      = "networks"
	ResourceTypeSubnet       = "subnets"
	FloatingType             = "floating"
	FixedType                = "fixed"
)

// ErrNotFound is used to inform that the object is not found
var ErrNotFound = "Resource not found"

// readBackoff is the backoff strategy for openstack read retries.
var readBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   1.5,
	Jitter:   0.1,
	Steps:    4,
}

// writeBackoff is the backoff strategy for openstack write retries.
var writeBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   1.5,
	Jitter:   0.1,
	Steps:    5,
}

type OpenstackCloud interface {
	fi.Cloud

	ComputeClient() *gophercloud.ServiceClient
	BlockStorageClient() *gophercloud.ServiceClient
	NetworkingClient() *gophercloud.ServiceClient
	LoadBalancerClient() *gophercloud.ServiceClient
	DNSClient() *gophercloud.ServiceClient
	UseOctavia() bool
	UseZones([]string)

	// GetInstance will return a openstack server provided its ID
	GetInstance(id string) (*servers.Server, error)

	// ListInstances will return a slice of openstack servers provided list opts
	ListInstances(servers.ListOptsBuilder) ([]servers.Server, error)

	// CreateInstance will create an openstack server provided create opts
	CreateInstance(servers.CreateOptsBuilder) (*servers.Server, error)

	//DeleteInstanceWithID will delete instance
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

	//DeleteVolume will delete volume
	DeleteVolume(volumeID string) error

	//ListSecurityGroups will return the Neutron security groups which match the options
	ListSecurityGroups(opt sg.ListOpts) ([]sg.SecGroup, error)

	//CreateSecurityGroup will create a new Neutron security group
	CreateSecurityGroup(opt sg.CreateOptsBuilder) (*sg.SecGroup, error)

	//DeleteSecurityGroup will delete securitygroup
	DeleteSecurityGroup(sgID string) error

	//DeleteSecurityGroupRule will delete securitygrouprule
	DeleteSecurityGroupRule(ruleID string) error

	//ListSecurityGroupRules will return the Neutron security group rules which match the options
	ListSecurityGroupRules(opt sgr.ListOpts) ([]sgr.SecGroupRule, error)

	//CreateSecurityGroupRule will create a new Neutron security group rule
	CreateSecurityGroupRule(opt sgr.CreateOptsBuilder) (*sgr.SecGroupRule, error)

	//GetNetwork will return the Neutron network which match the id
	GetNetwork(networkID string) (*networks.Network, error)

	//FindNetworkBySubnetID will return network
	FindNetworkBySubnetID(subnetID string) (*networks.Network, error)

	//GetSubnet returns subnet using subnet id
	GetSubnet(subnetID string) (*subnets.Subnet, error)

	//ListNetworks will return the Neutron networks which match the options
	ListNetworks(opt networks.ListOptsBuilder) ([]networks.Network, error)

	//ListExternalNetworks will return the Neutron networks with the router:external property
	GetExternalNetwork() (*networks.Network, error)

	// GetExternalSubnet will return the subnet for floatingip which is used in external router
	GetExternalSubnet() (*subnets.Subnet, error)

	// GetLBFloatingSubnet will return the subnet for floatingip which is used in lb
	GetLBFloatingSubnet() (*subnets.Subnet, error)

	//CreateNetwork will create a new Neutron network
	CreateNetwork(opt networks.CreateOptsBuilder) (*networks.Network, error)

	//DeleteNetwork will delete neutron network
	DeleteNetwork(networkID string) error

	//AppendTag appends tag to resource
	AppendTag(resource string, id string, tag string) error

	//DeleteTag removes tag from resource
	DeleteTag(resource string, id string, tag string) error

	//ListRouters will return the Neutron routers which match the options
	ListRouters(opt routers.ListOpts) ([]routers.Router, error)

	//CreateRouter will create a new Neutron router
	CreateRouter(opt routers.CreateOptsBuilder) (*routers.Router, error)

	//DeleteRouter will delete neutron router
	DeleteRouter(routerID string) error

	//DeleteSubnet will delete neutron subnet
	DeleteSubnet(subnetID string) error

	//ListSubnets will return the Neutron subnets which match the options
	ListSubnets(opt subnets.ListOptsBuilder) ([]subnets.Subnet, error)

	//CreateSubnet will create a new Neutron subnet
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

	//GetPort will return a Neutron port by ID
	GetPort(id string) (*ports.Port, error)

	//ListPorts will return the Neutron ports which match the options
	ListPorts(opt ports.ListOptsBuilder) ([]ports.Port, error)

	// DeletePort will delete a neutron port
	DeletePort(portID string) error

	//CreateRouterInterface will create a new Neutron router interface
	CreateRouterInterface(routerID string, opt routers.AddInterfaceOptsBuilder) (*routers.InterfaceInfo, error)

	//DeleteRouterInterface will delete router interface from subnet
	DeleteRouterInterface(routerID string, opt routers.RemoveInterfaceOptsBuilder) error

	// CreateServerGroup will create a new server group.
	CreateServerGroup(opt servergroups.CreateOptsBuilder) (*servergroups.ServerGroup, error)

	// ListServerGroups will list available server groups
	ListServerGroups() ([]servergroups.ServerGroup, error)

	// DeleteServerGroup will delete a nova server group
	DeleteServerGroup(groupID string) error

	// ListDNSZones will list available DNS zones
	ListDNSZones(opt zones.ListOptsBuilder) ([]zones.Zone, error)

	// ListDNSRecordsets will list the DNS recordsets for the given zone id
	ListDNSRecordsets(zoneID string, opt recordsets.ListOptsBuilder) ([]recordsets.RecordSet, error)

	GetLB(loadbalancerID string) (*loadbalancers.LoadBalancer, error)

	CreateLB(opt loadbalancers.CreateOptsBuilder) (*loadbalancers.LoadBalancer, error)

	ListLBs(opt loadbalancers.ListOptsBuilder) ([]loadbalancers.LoadBalancer, error)

	// DeleteLB will delete loadbalancer
	DeleteLB(lbID string, opt loadbalancers.DeleteOpts) error

	GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error)

	FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error)

	// DefaultInstanceType determines a suitable instance type for the specified instance group
	DefaultInstanceType(cluster *kops.Cluster, ig *kops.InstanceGroup) (string, error)

	// Returns the availability zones for the service client passed (compute, volume, network)
	ListAvailabilityZones(serviceClient *gophercloud.ServiceClient) ([]az.AvailabilityZone, error)

	AssociateToPool(server *servers.Server, poolID string, opts v2pools.CreateMemberOpts) (*v2pools.Member, error)

	CreatePool(opts v2pools.CreateOpts) (*v2pools.Pool, error)

	GetPool(poolID string, memberID string) (*v2pools.Member, error)

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

	GetFloatingIP(id string) (fip *floatingips.FloatingIP, err error)

	GetImage(name string) (i *images.Image, err error)

	AssociateFloatingIPToInstance(serverID string, opts floatingips.AssociateOpts) (err error)

	ListServerIPs(id string, IPType string) ([]*string, error)

	ListFloatingIPs() (fips []floatingips.FloatingIP, err error)
	ListL3FloatingIPs(opts l3floatingip.ListOpts) (fips []l3floatingip.FloatingIP, err error)
	CreateFloatingIP(opts floatingips.CreateOpts) (*floatingips.FloatingIP, error)
	CreateL3FloatingIP(opts l3floatingip.CreateOpts) (fip *l3floatingip.FloatingIP, err error)
	DeleteFloatingIP(id string) error
	DeleteL3FloatingIP(id string) error
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
}

var _ fi.Cloud = &openstackCloud{}

func NewOpenstackCloud(tags map[string]string, spec *kops.ClusterSpec) (OpenstackCloud, error) {
	config := vfs.OpenstackConfig{}

	authOption, err := config.GetCredential()
	if err != nil {
		return nil, err
	}

	provider, err := os.NewClient(authOption.IdentityEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error building openstack provider client: %v", err)
	}

	region, err := config.GetRegion()
	if err != nil {
		return nil, fmt.Errorf("error finding openstack region: %v", err)
	}

	if spec != nil && spec.CloudConfig != nil && spec.CloudConfig.Openstack != nil && spec.CloudConfig.Openstack.InsecureSkipVerify != nil {
		tlsconfig := &tls.Config{}
		tlsconfig.InsecureSkipVerify = fi.BoolValue(spec.CloudConfig.Openstack.InsecureSkipVerify)
		transport := &http.Transport{TLSClientConfig: tlsconfig}
		provider.HTTPClient = http.Client{
			Transport: transport,
		}
	}

	klog.V(2).Info("authenticating to keystone")

	err = os.Authenticate(provider, authOption)
	if err != nil {
		return nil, fmt.Errorf("error building openstack authenticated client: %v", err)
	}

	//TODO: maybe try v2, and v3?
	cinderClient, err := os.NewBlockStorageV2(provider, gophercloud.EndpointOpts{
		Type:   "volumev2",
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("error building cinder client: %v", err)
	}

	neutronClient, err := os.NewNetworkV2(provider, gophercloud.EndpointOpts{
		Type:   "network",
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("error building neutron client: %v", err)
	}

	novaClient, err := os.NewComputeV2(provider, gophercloud.EndpointOpts{
		Type:   "compute",
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("error building nova client: %v", err)
	}

	glanceClient, err := os.NewImageServiceV2(provider, gophercloud.EndpointOpts{
		Type:   "image",
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("error building glance client: %v", err)
	}

	var dnsClient *gophercloud.ServiceClient
	if !dns.IsGossipHostname(tags[TagClusterName]) {
		//TODO: This should be replaced with the environment variable methods as done above
		endpointOpt, err := config.GetServiceConfig("Designate")
		if err != nil {
			return nil, err
		}

		dnsClient, err = os.NewDNSV2(provider, endpointOpt)
		if err != nil {
			return nil, fmt.Errorf("error building dns client: %v", err)
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

	octavia := false
	floatingEnabled := false
	if spec != nil &&
		spec.CloudConfig != nil &&
		spec.CloudConfig.Openstack != nil &&
		spec.CloudConfig.Openstack.Router != nil {

		floatingEnabled = true
		c.extNetworkName = spec.CloudConfig.Openstack.Router.ExternalNetwork

		if spec.CloudConfig.Openstack.Router.ExternalSubnet != nil {
			c.extSubnetName = spec.CloudConfig.Openstack.Router.ExternalSubnet
		}
		if spec.CloudConfig.Openstack.Loadbalancer != nil &&
			spec.CloudConfig.Openstack.Loadbalancer.FloatingNetworkID == nil &&
			spec.CloudConfig.Openstack.Loadbalancer.FloatingNetwork != nil {
			// This field is derived
			lbNet, err := c.ListNetworks(networks.ListOpts{
				Name: fi.StringValue(spec.CloudConfig.Openstack.Loadbalancer.FloatingNetwork),
			})
			if err != nil || len(lbNet) != 1 {
				return c, fmt.Errorf("could not establish floating network id")
			}
			spec.CloudConfig.Openstack.Loadbalancer.FloatingNetworkID = fi.String(lbNet[0].ID)
		}
		if spec.CloudConfig.Openstack.Loadbalancer != nil {
			if spec.CloudConfig.Openstack.Loadbalancer.UseOctavia != nil {
				octavia = fi.BoolValue(spec.CloudConfig.Openstack.Loadbalancer.UseOctavia)
			}
			if spec.CloudConfig.Openstack.Loadbalancer.FloatingSubnet != nil {
				c.floatingSubnet = spec.CloudConfig.Openstack.Loadbalancer.FloatingSubnet
			}
		}
	}
	c.floatingEnabled = floatingEnabled
	c.useOctavia = octavia
	var lbClient *gophercloud.ServiceClient
	if spec != nil && spec.CloudConfig != nil && spec.CloudConfig.Openstack != nil {
		if spec.CloudConfig.Openstack.Loadbalancer != nil && octavia {
			klog.V(2).Infof("Openstack using Octavia lbaasv2 api")
			lbClient, err = os.NewLoadBalancerV2(provider, gophercloud.EndpointOpts{
				Region: region,
			})
			if err != nil {
				return nil, fmt.Errorf("error building lb client: %v", err)
			}
		} else if spec.CloudConfig.Openstack.Loadbalancer != nil {
			klog.V(2).Infof("Openstack using deprecated lbaasv2 api")
			lbClient, err = os.NewNetworkV2(provider, gophercloud.EndpointOpts{
				Region: region,
			})
			if err != nil {
				return nil, fmt.Errorf("error building lb client: %v", err)
			}
		} else {
			klog.V(2).Infof("Openstack disabled loadbalancer support")
		}
	}
	c.lbClient = lbClient
	return c, nil
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
	vpcInfo := &fi.VPCInfo{}
	// Find subnets in the network
	{
		if len(c.zones) == 0 {
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
			zone := c.zones[int(index)%len(c.zones)]
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
	grp := g.Raw.(*servergroups.ServerGroup)

	for _, id := range grp.Members {
		err := c.DeleteInstanceWithID(id)
		if err != nil {
			return fmt.Errorf("could not delete instance %q: %v", id, err)
		}
	}

	ports, err := c.ListPorts(ports.ListOpts{})
	if err != nil {
		return fmt.Errorf("Could not list ports %v", err)
	}

	for _, port := range ports {
		if strings.Contains(port.Name, grp.Name) {
			err := c.DeletePort(port.ID)
			if err != nil {
				return fmt.Errorf("could not delete port %q: %v", port.ID, err)
			}
		}
	}

	err = c.DeleteServerGroup(grp.ID)
	if err != nil {
		return fmt.Errorf("could not server group %q: %v", grp.ID, err)
	}

	return nil
}

func (c *openstackCloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	nodeMap := cloudinstances.GetNodeMap(nodes, cluster)
	groups := make(map[string]*cloudinstances.CloudInstanceGroup)

	serverGrps, err := c.ListServerGroups()
	if err != nil {
		return nil, fmt.Errorf("unable to list servergroups: %v", err)
	}

	for _, grp := range serverGrps {
		name := grp.Name
		instancegroup, err := matchInstanceGroup(name, cluster.ObjectMeta.Name, instancegroups)
		if err != nil {
			return nil, fmt.Errorf("error getting instance group for servergroup %q", name)
		}
		if instancegroup == nil {
			if warnUnmatched {
				klog.Warningf("Found servergrp with no corresponding instance group %q", name)
			}
			continue
		}
		groups[instancegroup.ObjectMeta.Name], err = c.osBuildCloudInstanceGroup(cluster, instancegroup, &grp, nodeMap)
		if err != nil {
			return nil, fmt.Errorf("error getting cloud instance group %q: %v", instancegroup.ObjectMeta.Name, err)
		}
	}
	return groups, nil
}

func (c *openstackCloud) GetCloudTags() map[string]string {
	return c.tags
}

type Address struct {
	IPType string `mapstructure:"OS-EXT-IPS:type"`
	Addr   string
}

func (c *openstackCloud) GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error) {
	var ingresses []kops.ApiIngressStatus
	if cluster.Spec.CloudConfig.Openstack.Loadbalancer != nil {
		if cluster.Spec.MasterPublicName != "" {
			// Note that this must match OpenstackModel lb name
			klog.V(2).Infof("Querying Openstack to find Loadbalancers for API (%q)", cluster.Name)
			lbList, err := c.ListLBs(loadbalancers.ListOpts{
				Name: cluster.Spec.MasterPublicName,
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
						ingresses = append(ingresses, kops.ApiIngressStatus{
							IP: fip.FloatingIP,
						})
					}
				}
			}
		}
	} else {
		instances, err := c.ListInstances(servers.ListOpts{})
		if err != nil {
			return ingresses, fmt.Errorf("GetApiIngressStatus: Failed to list master nodes: %v", err)
		}
		for _, instance := range instances {
			val, ok := instance.Metadata["k8s"]
			val2, ok2 := instance.Metadata["KopsRole"]
			if ok && val == cluster.Name && ok2 {
				role, success := kops.ParseInstanceGroupRole(val2, false)
				if success && role == kops.InstanceGroupRoleMaster {
					ips, err := c.ListServerIPs(instance.ID, FloatingType)
					if err != nil {
						return ingresses, err
					}
					for _, ip := range ips {
						ingresses = append(ingresses, kops.ApiIngressStatus{
							IP: fi.StringValue(ip),
						})
					}
				}
			}
		}
	}

	return ingresses, nil
}

func isNotFound(err error) bool {
	if _, ok := err.(gophercloud.ErrDefault404); ok {
		return true
	}

	if errCode, ok := err.(gophercloud.ErrUnexpectedResponseCode); ok {
		if errCode.Actual == http.StatusNotFound {
			return true
		}
	}

	return false
}
