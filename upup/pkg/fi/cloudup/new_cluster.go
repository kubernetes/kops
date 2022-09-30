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

package cloudup

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/blang/semver/v4"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops"
	api "k8s.io/kops/pkg/apis/kops"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/zones"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/vfs"
)

const (
	AuthorizationFlagAlwaysAllow = "AlwaysAllow"
	AuthorizationFlagRBAC        = "RBAC"
)

type NewClusterOptions struct {
	// ClusterName is the name of the cluster to initialize.
	ClusterName string

	// Authorization is the authorization mode to use. The options are "RBAC" (default) and "AlwaysAllow".
	Authorization string
	// Channel is a channel location for initializing the cluster. It defaults to "stable".
	Channel string
	// ConfigBase is the location where we will store the configuration. It defaults to the state store.
	ConfigBase string
	// DiscoveryStore is the location where we will store public OIDC-compatible discovery documents, under a cluster-specific directory. It defaults to not publishing discovery documents.
	DiscoveryStore string
	// KubernetesVersion is the version of Kubernetes to deploy. It defaults to the version recommended by the channel.
	KubernetesVersion string
	// AdminAccess is the set of CIDR blocks permitted to connect to the Kubernetes API. It defaults to "0.0.0.0/0" and "::/0".
	AdminAccess []string
	// SSHAccess is the set of CIDR blocks permitted to connect to SSH on the nodes. It defaults to the value of AdminAccess.
	SSHAccess []string

	// CloudProvider is the name of the cloud provider. The default is to guess based on the Zones name.
	CloudProvider string
	// Zones are the availability zones in which to run the cluster.
	Zones []string
	// MasterZones are the availability zones in which to run the masters. Defaults to the list in the Zones field.
	MasterZones []string

	// Project is the cluster's GCE project.
	Project string
	// GCEServiceAccount specifies the service account with which the GCE VM runs.
	GCEServiceAccount string

	// Spotinst options
	SpotinstProduct     string
	SpotinstOrientation string

	// NetworkID is the ID of the shared network (VPC).
	// If empty, SubnetIDs are not empty, and on AWS or OpenStack, determines network ID from the first SubnetID.
	// If empty otherwise, creates a new network/VPC to be owned by the cluster.
	NetworkID string
	// SubnetIDs are the IDs of the shared subnets.
	// If empty, creates new subnets to be owned by the cluster.
	SubnetIDs []string
	// UtilitySubnetIDs are the IDs of the shared utility subnets. If empty and the topology is "private", creates new subnets to be owned by the cluster.
	UtilitySubnetIDs []string
	// Egress defines the method of traffic egress for subnets.
	Egress string
	// IPv6 adds IPv6 CIDRs to subnets
	IPv6 bool

	// OpenstackExternalNet is the name of the external network for the openstack router.
	OpenstackExternalNet     string
	OpenstackExternalSubnet  string
	OpenstackStorageIgnoreAZ bool
	OpenstackDNSServers      string
	OpenstackLBSubnet        string
	// OpenstackLBOctavia is whether to use use octavia instead of haproxy.
	OpenstackLBOctavia       bool
	OpenstackOctaviaProvider string

	AzureSubscriptionID    string
	AzureTenantID          string
	AzureResourceGroupName string
	AzureRouteTableName    string
	AzureAdminUser         string

	// MasterCount is the number of masters to create. Defaults to the length of MasterZones
	// if MasterZones is explicitly nonempty, otherwise defaults to 1.
	MasterCount int32
	// APIServerCount is the number of API servers to create. Defaults to 0.
	APIServerCount int32
	// EncryptEtcdStorage is whether to encrypt the etcd volumes.
	EncryptEtcdStorage *bool
	// EtcdStorageType is the underlying cloud storage class of the etcd volumes.
	EtcdStorageType string

	// NodeCount is the number of nodes to create. Defaults to leaving the count unspecified
	// on the InstanceGroup, which results in a count of 2.
	NodeCount int32
	// Bastion enables the creation of a Bastion instance.
	Bastion bool
	// BastionLoadBalancerType is the bastion loadbalancer type to use; "public" or "internal".
	// Defaults to "public".
	BastionLoadBalancerType string

	// Networking is the networking provider/node to use.
	Networking string
	// Topology is the network topology to use. Defaults to "public".
	Topology string
	// DNSType is the DNS type to use; "public" or "private". Defaults to "public".
	DNSType string

	// APILoadBalancerClass determines whether to use classic or network load balancers for the API
	APILoadBalancerClass string
	// APILoadBalancerType is the Kubernetes API loadbalancer type to use; "public" or "internal".
	// Defaults to using DNS instead of a load balancer if using public topology and not gossip, otherwise "public".
	APILoadBalancerType string
	// APISSLCertificate is the SSL certificate to use for the API loadbalancer.
	// Currently only supported in AWS.
	APISSLCertificate string

	// InstanceManager specifies which manager to use for managing instances.
	InstanceManager string

	Image       string
	NodeImage   string
	MasterImage string
	MasterSize  string
	NodeSize    string
}

func (o *NewClusterOptions) InitDefaults() {
	o.Channel = api.DefaultChannel
	o.Authorization = AuthorizationFlagRBAC
	o.AdminAccess = []string{"0.0.0.0/0", "::/0"}
	o.Networking = "kubenet"
	o.Topology = api.TopologyPublic
	o.DNSType = string(api.DNSTypePublic)
	o.InstanceManager = "cloudgroups"
}

type NewClusterResult struct {
	// Cluster is the initialized Cluster resource.
	Cluster *api.Cluster
	// InstanceGroups are the initialized InstanceGroup resources.
	InstanceGroups []*api.InstanceGroup
	// Channel is the loaded Channel object.
	Channel *api.Channel
}

// NewCluster initializes cluster and instance groups specifications as
// intended for newly created clusters.
// It is the responsibility of the caller to call cloudup.PerformAssignments() on
// the returned cluster spec.
func NewCluster(opt *NewClusterOptions, clientset simple.Clientset) (*NewClusterResult, error) {
	if opt.ClusterName == "" {
		return nil, fmt.Errorf("name is required")
	}

	if opt.Channel == "" {
		opt.Channel = api.DefaultChannel
	}
	channel, err := api.LoadChannel(opt.Channel)
	if err != nil {
		return nil, err
	}

	cluster := api.Cluster{
		ObjectMeta: v1.ObjectMeta{
			Name: opt.ClusterName,
		},
	}

	if channel.Spec.Cluster != nil {
		cluster.Spec = *channel.Spec.Cluster

		kubernetesVersion := api.RecommendedKubernetesVersion(channel, kops.Version)
		if kubernetesVersion != nil {
			cluster.Spec.KubernetesVersion = kubernetesVersion.String()
		}
	}
	cluster.Spec.Channel = opt.Channel
	if opt.KubernetesVersion != "" {
		cluster.Spec.KubernetesVersion = opt.KubernetesVersion
	}

	cluster.Spec.ConfigBase = opt.ConfigBase
	configBase, err := clientset.ConfigBaseFor(&cluster)
	if err != nil {
		return nil, fmt.Errorf("error building ConfigBase for cluster: %v", err)
	}
	cluster.Spec.ConfigBase = configBase.Path()

	cluster.Spec.Authorization = &api.AuthorizationSpec{}
	if strings.EqualFold(opt.Authorization, AuthorizationFlagAlwaysAllow) {
		cluster.Spec.Authorization.AlwaysAllow = &api.AlwaysAllowAuthorizationSpec{}
	} else if opt.Authorization == "" || strings.EqualFold(opt.Authorization, AuthorizationFlagRBAC) {
		cluster.Spec.Authorization.RBAC = &api.RBACAuthorizationSpec{}
	} else {
		return nil, fmt.Errorf("unknown authorization mode %q", opt.Authorization)
	}

	cluster.Spec.IAM = &api.IAMSpec{
		AllowContainerRegistry: true,
	}
	cluster.Spec.Kubelet = &api.KubeletConfigSpec{
		AnonymousAuth: fi.Bool(false),
	}

	if len(opt.AdminAccess) == 0 {
		opt.AdminAccess = []string{"0.0.0.0/0", "::/0"}
	}
	cluster.Spec.KubernetesAPIAccess = opt.AdminAccess
	if len(opt.SSHAccess) != 0 {
		cluster.Spec.SSHAccess = opt.SSHAccess
	} else {
		cluster.Spec.SSHAccess = opt.AdminAccess
	}

	if len(opt.Zones) == 0 {
		return nil, fmt.Errorf("must specify at least one zone for the cluster (use --zones)")
	}
	allZones := sets.NewString()
	allZones.Insert(opt.Zones...)
	allZones.Insert(opt.MasterZones...)

	if opt.CloudProvider == "" {
		for _, zone := range allZones.List() {
			cloud, known := zones.GuessCloudForZone(zone)
			if known {
				klog.Infof("Inferred %q cloud provider from zone %q", cloud, zone)
				opt.CloudProvider = string(cloud)
				break
			}
		}
		if opt.CloudProvider == "" {
			if allZones.Len() == 0 {
				return nil, fmt.Errorf("must specify --zones or --cloud")
			}
			return nil, fmt.Errorf("unable to infer cloud provider from zones. pass in the cloud provider explicitly using --cloud")
		}
	}

	var cloud fi.Cloud

	switch api.CloudProviderID(opt.CloudProvider) {
	case api.CloudProviderAWS:
		cluster.Spec.CloudProvider.AWS = &api.AWSSpec{}
		cloudTags := map[string]string{}
		awsCloud, err := awsup.NewAWSCloud(opt.Zones[0][:len(opt.Zones[0])-1], cloudTags)
		if err != nil {
			return nil, err
		}
		cloud = awsCloud
	case api.CloudProviderAzure:
		cluster.Spec.CloudProvider.Azure = &api.AzureSpec{
			SubscriptionID:    opt.AzureSubscriptionID,
			TenantID:          opt.AzureTenantID,
			ResourceGroupName: opt.AzureResourceGroupName,
			RouteTableName:    opt.AzureRouteTableName,
			AdminUser:         opt.AzureAdminUser,
		}
	case api.CloudProviderDO:
		cluster.Spec.CloudProvider.DO = &api.DOSpec{}
	case api.CloudProviderGCE:
		cluster.Spec.CloudProvider.GCE = &api.GCESpec{}
	case api.CloudProviderHetzner:
		cluster.Spec.CloudProvider.Hetzner = &api.HetznerSpec{}
	case api.CloudProviderOpenstack:
		cluster.Spec.CloudProvider.Openstack = &api.OpenstackSpec{
			Router: &api.OpenstackRouter{
				ExternalNetwork: fi.String(opt.OpenstackExternalNet),
			},
			BlockStorage: &api.OpenstackBlockStorageConfig{
				Version:  fi.String("v3"),
				IgnoreAZ: fi.Bool(opt.OpenstackStorageIgnoreAZ),
			},
			Monitor: &api.OpenstackMonitor{
				Delay:      fi.String("15s"),
				Timeout:    fi.String("10s"),
				MaxRetries: fi.Int(3),
			},
		}
	case api.CloudProviderYandex:
		cluster.Spec.CloudProvider.Yandex = &api.YandexSpec{}
	default:
		return nil, fmt.Errorf("unsupported cloud provider %s", opt.CloudProvider)
	}

	if opt.DiscoveryStore != "" {
		discoveryPath, err := vfs.Context.BuildVfsPath(opt.DiscoveryStore)
		if err != nil {
			return nil, fmt.Errorf("error building DiscoveryStore for cluster: %v", err)
		}
		cluster.Spec.ServiceAccountIssuerDiscovery = &api.ServiceAccountIssuerDiscoveryConfig{
			DiscoveryStore: discoveryPath.Join(cluster.Name).Path(),
		}
		if cluster.Spec.GetCloudProvider() == api.CloudProviderAWS {
			cluster.Spec.ServiceAccountIssuerDiscovery.EnableAWSOIDCProvider = true
			cluster.Spec.IAM.UseServiceAccountExternalPermissions = fi.Bool(true)
		}
	}

	err = setupVPC(opt, &cluster, cloud)
	if err != nil {
		return nil, err
	}

	zoneToSubnetMap, err := setupZones(opt, &cluster, allZones)
	if err != nil {
		return nil, err
	}

	masters, err := setupMasters(opt, &cluster, zoneToSubnetMap)
	if err != nil {
		return nil, err
	}

	var nodes []*api.InstanceGroup

	switch opt.InstanceManager {
	case "karpenter":
		if opt.DiscoveryStore == "" {
			return nil, fmt.Errorf("karpenter requires --discovery-store")
		}
		cluster.Spec.Karpenter = &api.KarpenterConfig{
			Enabled: true,
		}
		nodes, err = setupKarpenterNodes(opt, &cluster, zoneToSubnetMap)
		if err != nil {
			return nil, err
		}
	case "cloudgroups":
		nodes, err = setupNodes(opt, &cluster, zoneToSubnetMap)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid value %q for --instance-manager", opt.InstanceManager)
	}

	apiservers, err := setupAPIServers(opt, &cluster, zoneToSubnetMap)
	if err != nil {
		return nil, err
	}

	err = setupNetworking(opt, &cluster)
	if err != nil {
		return nil, err
	}

	bastions, err := setupTopology(opt, &cluster, allZones)
	if err != nil {
		return nil, err
	}

	err = setupAPI(opt, &cluster)
	if err != nil {
		return nil, err
	}

	instanceGroups := append([]*api.InstanceGroup(nil), masters...)
	instanceGroups = append(instanceGroups, apiservers...)
	instanceGroups = append(instanceGroups, nodes...)
	instanceGroups = append(instanceGroups, bastions...)

	for _, instanceGroup := range instanceGroups {
		g := instanceGroup
		ig := g
		if instanceGroup.Spec.Image == "" {
			if opt.Image != "" {
				instanceGroup.Spec.Image = opt.Image
			} else {
				architecture, err := MachineArchitecture(cloud, instanceGroup.Spec.MachineType)
				if err != nil {
					return nil, err
				}
				instanceGroup.Spec.Image = defaultImage(&cluster, channel, architecture)
			}
		}

		// TODO: Clean up
		if g.IsMaster() {
			if g.Spec.MachineType == "" {
				g.Spec.MachineType, err = defaultMachineType(cloud, &cluster, ig)
				if err != nil {
					return nil, fmt.Errorf("error assigning default machine type for masters: %v", err)
				}

			}
		} else if g.Spec.Role == kopsapi.InstanceGroupRoleBastion {
			if g.Spec.MachineType == "" {
				g.Spec.MachineType, err = defaultMachineType(cloud, &cluster, g)
				if err != nil {
					return nil, fmt.Errorf("error assigning default machine type for bastions: %v", err)
				}
			}
		} else {
			if g.IsAPIServerOnly() && !featureflag.APIServerNodes.Enabled() {
				return nil, fmt.Errorf("apiserver nodes requires the APIServerNodes feature flag to be enabled")
			}
			if g.Spec.MachineType == "" {
				g.Spec.MachineType, err = defaultMachineType(cloud, &cluster, g)
				if err != nil {
					return nil, fmt.Errorf("error assigning default machine type for nodes: %v", err)
				}
			}

		}

		if ig.Spec.Tenancy != "" && ig.Spec.Tenancy != "default" {
			switch cluster.Spec.GetCloudProvider() {
			case kopsapi.CloudProviderAWS:
				if _, ok := awsDedicatedInstanceExceptions[g.Spec.MachineType]; ok {
					return nil, fmt.Errorf("invalid dedicated instance type: %s", g.Spec.MachineType)
				}
			default:
				klog.Warning("Trying to set tenancy on non-AWS environment")
			}
		}

		if ig.IsMaster() {
			if len(ig.Spec.Subnets) == 0 {
				return nil, fmt.Errorf("master InstanceGroup %s did not specify any Subnets", g.ObjectMeta.Name)
			}
		} else if ig.IsAPIServerOnly() && cluster.Spec.IsIPv6Only() {
			if len(ig.Spec.Subnets) == 0 {
				for _, subnet := range cluster.Spec.Subnets {
					if subnet.Type != kopsapi.SubnetTypePrivate && subnet.Type != kopsapi.SubnetTypeUtility {
						ig.Spec.Subnets = append(g.Spec.Subnets, subnet.Name)
					}
				}
			}
		} else {
			if len(ig.Spec.Subnets) == 0 {
				for _, subnet := range cluster.Spec.Subnets {
					if subnet.Type != kopsapi.SubnetTypeDualStack && subnet.Type != kopsapi.SubnetTypeUtility {
						g.Spec.Subnets = append(g.Spec.Subnets, subnet.Name)
					}
				}
			}

			if len(g.Spec.Subnets) == 0 {
				for _, subnet := range cluster.Spec.Subnets {
					if subnet.Type != kopsapi.SubnetTypeUtility {
						g.Spec.Subnets = append(g.Spec.Subnets, subnet.Name)
					}
				}
			}
		}

		if len(g.Spec.Subnets) == 0 {
			return nil, fmt.Errorf("unable to infer any Subnets for InstanceGroup %s ", g.ObjectMeta.Name)
		}
	}

	result := NewClusterResult{
		Cluster:        &cluster,
		InstanceGroups: instanceGroups,
		Channel:        channel,
	}
	return &result, nil
}

func setupVPC(opt *NewClusterOptions, cluster *api.Cluster, cloud fi.Cloud) error {
	cluster.Spec.NetworkID = opt.NetworkID

	switch cluster.Spec.GetCloudProvider() {
	case api.CloudProviderAWS:
		if cluster.Spec.NetworkID == "" && len(opt.SubnetIDs) > 0 {
			awsCloud := cloud.(awsup.AWSCloud)
			res, err := awsCloud.EC2().DescribeSubnets(&ec2.DescribeSubnetsInput{
				SubnetIds: []*string{aws.String(opt.SubnetIDs[0])},
			})
			if err != nil {
				return fmt.Errorf("error describing subnet %s: %v", opt.SubnetIDs[0], err)
			}
			if len(res.Subnets) == 0 || res.Subnets[0].VpcId == nil {
				return fmt.Errorf("failed to determine VPC id of subnet %s", opt.SubnetIDs[0])
			}
			cluster.Spec.NetworkID = *res.Subnets[0].VpcId
		}

	case api.CloudProviderGCE:
		if cluster.Spec.CloudConfig == nil {
			cluster.Spec.CloudConfig = &api.CloudConfiguration{}
		}
		cluster.Spec.Project = opt.Project
		if cluster.Spec.Project == "" {
			project, err := gce.DefaultProject()
			if err != nil {
				klog.Warningf("unable to get default google cloud project: %v", err)
			} else if project == "" {
				klog.Warningf("default google cloud project not set (try `gcloud config set project <name>`")
			} else {
				klog.Infof("using google cloud project: %s", project)
			}
			cluster.Spec.Project = project
		}
		if opt.GCEServiceAccount != "" {
			// TODO remove this logging?
			klog.Infof("VMs will be configured to use specified Service Account: %v", opt.GCEServiceAccount)
			cluster.Spec.CloudConfig.GCEServiceAccount = opt.GCEServiceAccount
		}

	case api.CloudProviderOpenstack:
		if cluster.Spec.CloudConfig == nil {
			cluster.Spec.CloudConfig = &api.CloudConfiguration{}
		}

		if cluster.Spec.NetworkID == "" && len(opt.SubnetIDs) > 0 {
			tags := make(map[string]string)
			tags[openstack.TagClusterName] = cluster.Name
			osCloud, err := openstack.NewOpenstackCloud(tags, &cluster.Spec, "new-cluster-setupvpc")
			if err != nil {
				return fmt.Errorf("error loading cloud: %v", err)
			}

			res, err := osCloud.FindNetworkBySubnetID(opt.SubnetIDs[0])
			if err != nil {
				return fmt.Errorf("error finding network: %v", err)
			}
			cluster.Spec.NetworkID = res.ID
		}

		if opt.OpenstackDNSServers != "" {
			cluster.Spec.CloudProvider.Openstack.Router.DNSServers = fi.String(opt.OpenstackDNSServers)
		}
		if opt.OpenstackExternalSubnet != "" {
			cluster.Spec.CloudProvider.Openstack.Router.ExternalSubnet = fi.String(opt.OpenstackExternalSubnet)
		}
	case api.CloudProviderAzure:
		// TODO(kenji): Find a right place for this.

		// Creating an empty CloudConfig so that --cloud-config is passed to kubelet, api-server, etc.
		if cluster.Spec.CloudConfig == nil {
			cluster.Spec.CloudConfig = &api.CloudConfiguration{}
		}
	}

	if featureflag.Spotinst.Enabled() {
		if cluster.Spec.CloudConfig == nil {
			cluster.Spec.CloudConfig = &api.CloudConfiguration{}
		}
		if opt.SpotinstProduct != "" {
			cluster.Spec.CloudConfig.SpotinstProduct = fi.String(opt.SpotinstProduct)
		}
		if opt.SpotinstOrientation != "" {
			cluster.Spec.CloudConfig.SpotinstOrientation = fi.String(opt.SpotinstOrientation)
		}
	}

	return nil
}

func setupZones(opt *NewClusterOptions, cluster *api.Cluster, allZones sets.String) (map[string]*api.ClusterSubnetSpec, error) {
	var err error
	zoneToSubnetMap := make(map[string]*api.ClusterSubnetSpec)

	var zoneToSubnetProviderID map[string]string

	switch cluster.Spec.GetCloudProvider() {
	case api.CloudProviderGCE:
		// On GCE, subnets are regional - we create one per region, not per zone
		for _, zoneName := range allZones.List() {
			region, err := gce.ZoneToRegion(zoneName)
			if err != nil {
				return nil, err
			}

			// We create default subnets named the same as the regions
			subnetName := region

			subnet := model.FindSubnet(cluster, subnetName)
			if subnet == nil {
				subnet = &api.ClusterSubnetSpec{
					Name:   subnetName,
					Region: region,
				}
				if len(opt.SubnetIDs) != 0 {
					// We don't support multi-region clusters, so we can't have more than one subnet
					if len(opt.SubnetIDs) != 1 {
						return nil, fmt.Errorf("expected exactly one subnet for GCE, got %d", len(opt.SubnetIDs))
					}
					providerID := opt.SubnetIDs[0]
					subnet.ProviderID = providerID
				}
				cluster.Spec.Subnets = append(cluster.Spec.Subnets, *subnet)
			}
			zoneToSubnetMap[zoneName] = subnet
		}

		return zoneToSubnetMap, nil

	case api.CloudProviderDO:
		if len(opt.Zones) > 1 {
			return nil, fmt.Errorf("digitalocean cloud provider currently supports one region only.")
		}

		// For DO we just pass in the region for --zones
		region := opt.Zones[0]
		subnet := model.FindSubnet(cluster, region)

		// for DO, subnets are just regions
		subnetName := region

		if subnet == nil {
			subnet = &api.ClusterSubnetSpec{
				Name: subnetName,
				// region and zone are the same for DO
				Region: region,
				Zone:   region,
			}
			cluster.Spec.Subnets = append(cluster.Spec.Subnets, *subnet)
		}
		zoneToSubnetMap[region] = subnet
		return zoneToSubnetMap, nil

	case api.CloudProviderHetzner:
		if len(opt.Zones) > 1 {
			return nil, fmt.Errorf("hetzner cloud provider currently supports only one zone (location)")
		}
		// TODO(hakman): Add customizations for Hetzner Cloud

	case api.CloudProviderAzure:
		// On Azure, subnets are regional - we create one per region, not per zone
		for _, zoneName := range allZones.List() {
			location, err := azure.ZoneToLocation(zoneName)
			if err != nil {
				return nil, err
			}

			// We create default subnets named the same as the regions
			subnetName := location

			subnet := model.FindSubnet(cluster, subnetName)
			if subnet == nil {
				subnet = &api.ClusterSubnetSpec{
					Name:   subnetName,
					Region: location,
				}
				cluster.Spec.Subnets = append(cluster.Spec.Subnets, *subnet)
			}
			zoneToSubnetMap[zoneName] = subnet
		}

	case api.CloudProviderAWS:
		if len(opt.Zones) > 0 && len(opt.SubnetIDs) > 0 {
			zoneToSubnetProviderID, err = getAWSZoneToSubnetProviderID(cluster.Spec.NetworkID, opt.Zones[0][:len(opt.Zones[0])-1], opt.SubnetIDs)
			if err != nil {
				return nil, err
			}
		}

	case api.CloudProviderOpenstack:
		if len(opt.Zones) > 0 && len(opt.SubnetIDs) > 0 {
			tags := make(map[string]string)
			tags[openstack.TagClusterName] = cluster.Name
			zoneToSubnetProviderID, err = getOpenstackZoneToSubnetProviderID(&cluster.Spec, allZones.List(), opt.SubnetIDs, tags)
			if err != nil {
				return nil, err
			}
		}

	case api.CloudProviderYandex:
		if len(opt.Zones) > 1 {
			klog.Warningf(strings.Join(opt.Zones[:], ","))
			return nil, fmt.Errorf("yandex cloud provider currently supports only one zone")
		}
		// TODO(YuraBeznos): yandex Add customizations for Yandex Cloud and more zones
	}

	for _, zoneName := range allZones.List() {
		// We create default subnets named the same as the zones
		subnetName := zoneName

		subnet := model.FindSubnet(cluster, subnetName)
		if subnet == nil {
			subnet = &api.ClusterSubnetSpec{
				Name:   subnetName,
				Zone:   subnetName,
				Egress: opt.Egress,
			}
			if subnetID, ok := zoneToSubnetProviderID[zoneName]; ok {
				subnet.ProviderID = subnetID
			}
			cluster.Spec.Subnets = append(cluster.Spec.Subnets, *subnet)
		}
		zoneToSubnetMap[zoneName] = subnet
	}

	return zoneToSubnetMap, nil
}

func getAWSZoneToSubnetProviderID(VPCID string, region string, subnetIDs []string) (map[string]string, error) {
	res := make(map[string]string)
	cloudTags := map[string]string{}
	awsCloud, err := awsup.NewAWSCloud(region, cloudTags)
	if err != nil {
		return res, fmt.Errorf("error loading cloud: %v", err)
	}
	vpcInfo, err := awsCloud.FindVPCInfo(VPCID)
	if err != nil {
		return res, fmt.Errorf("error describing VPC: %v", err)
	}
	if vpcInfo == nil {
		return res, fmt.Errorf("VPC %q not found", VPCID)
	}
	subnetByID := make(map[string]*fi.SubnetInfo)
	for _, subnetInfo := range vpcInfo.Subnets {
		subnetByID[subnetInfo.ID] = subnetInfo
	}
	for _, subnetID := range subnetIDs {
		subnet, ok := subnetByID[subnetID]
		if !ok {
			return res, fmt.Errorf("subnet %s not found in VPC %s", subnetID, VPCID)
		}
		if res[subnet.Zone] != "" {
			return res, fmt.Errorf("subnet %s and %s have the same zone", subnetID, res[subnet.Zone])
		}
		res[subnet.Zone] = subnetID
	}
	return res, nil
}

func getOpenstackZoneToSubnetProviderID(spec *api.ClusterSpec, zones []string, subnetIDs []string, tags map[string]string) (map[string]string, error) {
	res := make(map[string]string)
	osCloud, err := openstack.NewOpenstackCloud(tags, spec, "new-cluster-zone-to-subnet")
	if err != nil {
		return res, fmt.Errorf("error loading cloud: %v", err)
	}
	osCloud.UseZones(zones)

	networkInfo, err := osCloud.FindVPCInfo(spec.NetworkID)
	if err != nil {
		return res, fmt.Errorf("error describing Network: %v", err)
	}
	if networkInfo == nil {
		return res, fmt.Errorf("network %q not found", spec.NetworkID)
	}

	subnetByID := make(map[string]*fi.SubnetInfo)
	for _, subnetInfo := range networkInfo.Subnets {
		subnetByID[subnetInfo.ID] = subnetInfo
	}

	for _, subnetID := range subnetIDs {
		subnet, ok := subnetByID[subnetID]
		if !ok {
			return res, fmt.Errorf("subnet %s not found in network %s", subnetID, spec.NetworkID)
		}

		if res[subnet.Zone] != "" {
			return res, fmt.Errorf("subnet %s and %s have the same zone", subnetID, res[subnet.Zone])
		}
		res[subnet.Zone] = subnetID
	}
	return res, nil
}

func setupMasters(opt *NewClusterOptions, cluster *api.Cluster, zoneToSubnetMap map[string]*api.ClusterSubnetSpec) ([]*api.InstanceGroup, error) {
	cloudProvider := cluster.Spec.GetCloudProvider()

	var masters []*api.InstanceGroup

	// Build the master subnets
	// The master zones is the default set of zones unless explicitly set
	// The master count is the number of master zones unless explicitly set
	// We then round-robin around the zones
	{
		masterCount := opt.MasterCount
		masterZones := opt.MasterZones
		if len(masterZones) != 0 {
			if masterCount != 0 && masterCount < int32(len(masterZones)) {
				return nil, fmt.Errorf("specified %d master zones, but also requested %d masters.  If specifying both, the count should match.", len(masterZones), masterCount)
			}

			if masterCount == 0 {
				// If master count is not specified, default to the number of master zones
				masterCount = int32(len(masterZones))
			}
		} else {
			// masterZones not set; default to same as node Zones
			masterZones = opt.Zones

			if masterCount == 0 {
				// If master count is not specified, default to 1
				masterCount = 1
			}
		}

		if len(masterZones) == 0 {
			// Should be unreachable
			return nil, fmt.Errorf("cannot determine master zones")
		}

		for i := 0; i < int(masterCount); i++ {
			zone := masterZones[i%len(masterZones)]
			name := zone
			if cloudProvider == api.CloudProviderDO {
				if int(masterCount) >= len(masterZones) {
					name += "-" + strconv.Itoa(1+(i/len(masterZones)))
				}
			} else {
				if int(masterCount) > len(masterZones) {
					name += "-" + strconv.Itoa(1+(i/len(masterZones)))
				}
			}

			g := &api.InstanceGroup{}
			g.Spec.Role = api.InstanceGroupRoleMaster
			g.Spec.MinSize = fi.Int32(1)
			g.Spec.MaxSize = fi.Int32(1)
			g.ObjectMeta.Name = "master-" + name

			subnet := zoneToSubnetMap[zone]
			if subnet == nil {
				klog.Fatalf("subnet not found in zoneToSubnetMap")
			}

			g.Spec.Subnets = []string{subnet.Name}
			if opt.IPv6 && opt.Topology == api.TopologyPrivate {
				g.Spec.Subnets = []string{"dualstack-" + subnet.Name}
			}
			if cloudProvider == api.CloudProviderGCE || cloudProvider == api.CloudProviderAzure {
				g.Spec.Zones = []string{zone}
			}

			if cluster.IsKubernetesGTE("1.22") {
				if cloudProvider == api.CloudProviderAWS {
					g.Spec.InstanceMetadata = &api.InstanceMetadataOptions{
						HTTPPutResponseHopLimit: fi.Int64(3),
						HTTPTokens:              fi.String("required"),
					}
				}
			}

			g.Spec.MachineType = opt.MasterSize
			g.Spec.Image = opt.MasterImage

			masters = append(masters, g)
		}
	}

	// Build the Etcd clusters
	{
		masterAZs := sets.NewString()
		duplicateAZs := false
		for _, ig := range masters {
			zones, err := model.FindZonesForInstanceGroup(cluster, ig)
			if err != nil {
				return nil, err
			}
			for _, zone := range zones {
				if masterAZs.Has(zone) {
					duplicateAZs = true
				}

				masterAZs.Insert(zone)
			}
		}

		if duplicateAZs {
			klog.Warningf("Running with masters in the same AZs; redundancy will be reduced")
		}

		clusters := EtcdClusters

		if opt.Networking == "cilium-etcd" {
			clusters = append(clusters, "cilium")
		}

		encryptEtcdStorage := false
		if opt.EncryptEtcdStorage != nil {
			encryptEtcdStorage = fi.BoolValue(opt.EncryptEtcdStorage)
		} else if cloudProvider == api.CloudProviderAWS {
			encryptEtcdStorage = true
		}
		for _, etcdCluster := range clusters {
			etcd := createEtcdCluster(etcdCluster, masters, encryptEtcdStorage, opt.EtcdStorageType)
			cluster.Spec.EtcdClusters = append(cluster.Spec.EtcdClusters, etcd)
		}
	}

	return masters, nil
}

func trimCommonPrefix(names []string) []string {
	// Trim shared prefix to keep the lengths sane
	// (this only applies to new clusters...)
	for len(names) != 0 && len(names[0]) > 1 {
		prefix := names[0][:1]
		allMatch := true
		for _, name := range names {
			if !strings.HasPrefix(name, prefix) {
				allMatch = false
			}
		}

		if !allMatch {
			break
		}

		for i := range names {
			names[i] = strings.TrimPrefix(names[i], prefix)
		}
	}

	for i, name := range names {
		if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
			names[i] = "etcd-" + name
		}
	}

	return names
}

func setupNodes(opt *NewClusterOptions, cluster *api.Cluster, zoneToSubnetMap map[string]*api.ClusterSubnetSpec) ([]*api.InstanceGroup, error) {
	cloudProvider := cluster.Spec.GetCloudProvider()

	var nodes []*api.InstanceGroup

	// The node count is the number of zones unless explicitly set
	// We then divvy up amongst the zones
	numZones := len(opt.Zones)
	nodeCount := opt.NodeCount
	if nodeCount == 0 {
		// If node count is not specified, default to the number of zones
		nodeCount = int32(numZones)
	}

	countPerIG := nodeCount / int32(numZones)
	remainder := int(nodeCount) % numZones

	for i, zone := range opt.Zones {
		count := countPerIG
		if i < remainder {
			count++
		}

		g := &api.InstanceGroup{}
		g.Spec.Role = api.InstanceGroupRoleNode
		g.Spec.MinSize = fi.Int32(count)
		g.Spec.MaxSize = fi.Int32(count)
		g.ObjectMeta.Name = "nodes-" + zone

		subnet := zoneToSubnetMap[zone]
		if subnet == nil {
			klog.Fatalf("subnet not found in zoneToSubnetMap")
		}

		g.Spec.Subnets = []string{subnet.Name}
		if cloudProvider == api.CloudProviderGCE || cloudProvider == api.CloudProviderAzure {
			g.Spec.Zones = []string{zone}
		}

		if cluster.IsKubernetesGTE("1.22") {
			if cloudProvider == api.CloudProviderAWS {
				g.Spec.InstanceMetadata = &api.InstanceMetadataOptions{
					HTTPPutResponseHopLimit: fi.Int64(1),
					HTTPTokens:              fi.String("required"),
				}
			}
		}

		g.Spec.MachineType = opt.NodeSize
		g.Spec.Image = opt.NodeImage

		nodes = append(nodes, g)
	}

	return nodes, nil
}

func setupKarpenterNodes(opt *NewClusterOptions, cluster *api.Cluster, zoneToSubnetMap map[string]*api.ClusterSubnetSpec) ([]*api.InstanceGroup, error) {
	g := &api.InstanceGroup{}
	g.Spec.Role = api.InstanceGroupRoleNode
	g.Spec.Manager = api.InstanceManagerKarpenter
	g.ObjectMeta.Name = "nodes"

	g.Spec.InstanceMetadata = &api.InstanceMetadataOptions{
		HTTPPutResponseHopLimit: fi.Int64(1),
		HTTPTokens:              fi.String("required"),
	}

	return []*api.InstanceGroup{g}, nil
}

func setupAPIServers(opt *NewClusterOptions, cluster *api.Cluster, zoneToSubnetMap map[string]*api.ClusterSubnetSpec) ([]*api.InstanceGroup, error) {
	cloudProvider := cluster.Spec.GetCloudProvider()

	var nodes []*api.InstanceGroup

	numZones := len(opt.Zones)
	nodeCount := opt.APIServerCount

	if nodeCount == 0 {
		return nodes, nil
	}

	countPerIG := nodeCount / int32(numZones)
	remainder := int(nodeCount) % numZones

	for i, zone := range opt.Zones {
		count := countPerIG
		if i < remainder {
			count++
		}

		g := &api.InstanceGroup{}
		g.Spec.Role = api.InstanceGroupRoleAPIServer
		g.Spec.MinSize = fi.Int32(count)
		g.Spec.MaxSize = fi.Int32(count)
		g.ObjectMeta.Name = "apiserver-" + zone

		subnet := zoneToSubnetMap[zone]
		if subnet == nil {
			klog.Fatalf("subnet not found in zoneToSubnetMap")
		}

		g.Spec.Subnets = []string{subnet.Name}
		if cloudProvider == api.CloudProviderGCE || cloudProvider == api.CloudProviderAzure {
			g.Spec.Zones = []string{zone}
		}

		if cluster.IsKubernetesGTE("1.22") {
			if cloudProvider == api.CloudProviderAWS {
				g.Spec.InstanceMetadata = &api.InstanceMetadataOptions{
					HTTPPutResponseHopLimit: fi.Int64(1),
					HTTPTokens:              fi.String("required"),
				}
			}
		}

		nodes = append(nodes, g)
	}

	return nodes, nil
}

func setupNetworking(opt *NewClusterOptions, cluster *api.Cluster) error {
	cluster.Spec.Networking = &api.NetworkingSpec{}
	switch opt.Networking {
	case "kubenet", "":
		cluster.Spec.Networking.Kubenet = &api.KubenetNetworkingSpec{}
	case "external":
		cluster.Spec.Networking.External = &api.ExternalNetworkingSpec{}
	case "cni":
		cluster.Spec.Networking.CNI = &api.CNINetworkingSpec{}
	case "kopeio-vxlan", "kopeio":
		cluster.Spec.Networking.Kopeio = &api.KopeioNetworkingSpec{}
	case "weave":
		cluster.Spec.Networking.Weave = &api.WeaveNetworkingSpec{}

		if cluster.Spec.GetCloudProvider() == api.CloudProviderAWS {
			// AWS supports "jumbo frames" of 9001 bytes and weave adds up to 87 bytes overhead
			// sets the default to the largest number that leaves enough overhead and is divisible by 4
			jumboFrameMTUSize := int32(8912)
			cluster.Spec.Networking.Weave.MTU = &jumboFrameMTUSize
		}
	case "flannel", "flannel-vxlan":
		cluster.Spec.Networking.Flannel = &api.FlannelNetworkingSpec{
			Backend: "vxlan",
		}
	case "flannel-udp":
		klog.Warningf("flannel UDP mode is not recommended; consider flannel-vxlan instead")
		cluster.Spec.Networking.Flannel = &api.FlannelNetworkingSpec{
			Backend: "udp",
		}
	case "calico":
		cluster.Spec.Networking.Calico = &api.CalicoNetworkingSpec{}
	case "canal":
		cluster.Spec.Networking.Canal = &api.CanalNetworkingSpec{}
	case "kube-router":
		cluster.Spec.Networking.Kuberouter = &api.KuberouterNetworkingSpec{}
		if cluster.Spec.KubeProxy == nil {
			cluster.Spec.KubeProxy = &api.KubeProxyConfig{}
		}
		enabled := false
		cluster.Spec.KubeProxy.Enabled = &enabled
	case "amazonvpc", "amazon-vpc-routed-eni":
		cluster.Spec.Networking.AmazonVPC = &api.AmazonVPCNetworkingSpec{}
	case "cilium":
		addCiliumNetwork(cluster)
	case "cilium-etcd":
		addCiliumNetwork(cluster)
		cluster.Spec.Networking.Cilium.EtcdManaged = true
	case "gce":
		cluster.Spec.Networking.GCE = &api.GCENetworkingSpec{}
	default:
		return fmt.Errorf("unknown networking mode %q", opt.Networking)
	}

	klog.V(4).Infof("networking mode=%s => %s", opt.Networking, fi.DebugAsJsonString(cluster.Spec.Networking))

	return nil
}

func setupTopology(opt *NewClusterOptions, cluster *api.Cluster, allZones sets.String) ([]*api.InstanceGroup, error) {
	var bastions []*api.InstanceGroup

	switch opt.Topology {
	case api.TopologyPublic, "":
		cluster.Spec.Topology = &api.TopologySpec{
			Masters: api.TopologyPublic,
			Nodes:   api.TopologyPublic,
			// Bastion: &api.BastionSpec{Enable: c.Bastion},
		}

		if opt.Bastion {
			return nil, fmt.Errorf("bastion supports --topology='private' only")
		}

		for i := range cluster.Spec.Subnets {
			cluster.Spec.Subnets[i].Type = api.SubnetTypePublic
		}

	case api.TopologyPrivate:
		if cluster.Spec.Networking.Kubenet != nil {
			return nil, fmt.Errorf("invalid networking option %s. Kubenet does not support private topology", opt.Networking)
		}
		cluster.Spec.Topology = &api.TopologySpec{
			Masters: api.TopologyPrivate,
			Nodes:   api.TopologyPrivate,
		}

		for i := range cluster.Spec.Subnets {
			cluster.Spec.Subnets[i].Type = api.SubnetTypePrivate
		}

		var zoneToSubnetProviderID map[string]string
		var err error
		if len(opt.Zones) > 0 && len(opt.UtilitySubnetIDs) > 0 {
			switch cluster.Spec.GetCloudProvider() {
			case api.CloudProviderAWS:
				zoneToSubnetProviderID, err = getAWSZoneToSubnetProviderID(cluster.Spec.NetworkID, opt.Zones[0][:len(opt.Zones[0])-1], opt.UtilitySubnetIDs)
				if err != nil {
					return nil, err
				}
			case api.CloudProviderOpenstack:
				tags := make(map[string]string)
				tags[openstack.TagClusterName] = cluster.Name
				zoneToSubnetProviderID, err = getOpenstackZoneToSubnetProviderID(&cluster.Spec, allZones.List(), opt.UtilitySubnetIDs, tags)
				if err != nil {
					return nil, err
				}
			}
		}

		if opt.IPv6 {
			var dualStackSubnets []api.ClusterSubnetSpec

			for _, s := range cluster.Spec.Subnets {
				if s.Type != api.SubnetTypePrivate {
					continue
				}
				subnet := api.ClusterSubnetSpec{
					Name:   "dualstack-" + s.Name,
					Zone:   s.Zone,
					Type:   api.SubnetTypeDualStack,
					Region: s.Region,
				}
				if subnetID, ok := zoneToSubnetProviderID[s.Zone]; ok {
					subnet.ProviderID = subnetID
				}
				dualStackSubnets = append(dualStackSubnets, subnet)
			}
			cluster.Spec.Subnets = append(cluster.Spec.Subnets, dualStackSubnets...)
		}

		addUtilitySubnets := true
		switch cluster.Spec.GetCloudProvider() {
		case api.CloudProviderGCE:
			// GCE does not need utility subnets
			addUtilitySubnets = false
		}

		if addUtilitySubnets {
			var utilitySubnets []api.ClusterSubnetSpec

			for _, s := range cluster.Spec.Subnets {
				if s.Type != api.SubnetTypePrivate {
					continue
				}
				subnet := api.ClusterSubnetSpec{
					Name:   "utility-" + s.Name,
					Zone:   s.Zone,
					Type:   api.SubnetTypeUtility,
					Region: s.Region,
				}
				if subnetID, ok := zoneToSubnetProviderID[s.Zone]; ok {
					subnet.ProviderID = subnetID
				}
				utilitySubnets = append(utilitySubnets, subnet)
			}
			cluster.Spec.Subnets = append(cluster.Spec.Subnets, utilitySubnets...)
		}

		if opt.Bastion {
			bastionGroup := &api.InstanceGroup{}
			bastionGroup.Spec.Role = api.InstanceGroupRoleBastion
			bastionGroup.ObjectMeta.Name = "bastions"
			bastionGroup.Spec.MaxSize = fi.Int32(1)
			bastionGroup.Spec.MinSize = fi.Int32(1)
			bastions = append(bastions, bastionGroup)

			if !dns.IsGossipHostname(cluster.Name) {
				cluster.Spec.Topology.Bastion = &api.BastionSpec{
					PublicName: "bastion." + cluster.Name,
				}
			}
			if cluster.Spec.GetCloudProvider() == api.CloudProviderGCE {
				bastionGroup.Spec.Zones = allZones.List()
			}

			if cluster.IsKubernetesGTE("1.22") {
				bastionGroup.Spec.InstanceMetadata = &api.InstanceMetadataOptions{
					HTTPPutResponseHopLimit: fi.Int64(1),
					HTTPTokens:              fi.String("required"),
				}
			}

		}

	default:
		return nil, fmt.Errorf("invalid topology %s", opt.Topology)
	}

	if opt.IPv6 {
		cluster.Spec.NonMasqueradeCIDR = "::/0"
		cluster.Spec.ExternalCloudControllerManager = &api.CloudControllerManagerConfig{}
		if cluster.Spec.GetCloudProvider() == api.CloudProviderAWS {
			for i := range cluster.Spec.Subnets {
				cluster.Spec.Subnets[i].IPv6CIDR = fmt.Sprintf("/64#%x", i)
			}
		} else {
			klog.Errorf("IPv6 support is available only on AWS")
		}
	}

	cluster.Spec.Topology.DNS = &api.DNSSpec{}
	switch strings.ToLower(opt.DNSType) {
	case "public", "":
		cluster.Spec.Topology.DNS.Type = api.DNSTypePublic
	case "private":
		cluster.Spec.Topology.DNS.Type = api.DNSTypePrivate
	default:
		return nil, fmt.Errorf("unknown DNSType: %q", opt.DNSType)
	}

	return bastions, nil
}

func setupAPI(opt *NewClusterOptions, cluster *api.Cluster) error {
	// Populate the API access, so that it can be discoverable
	klog.Infof(" Cloud Provider ID = %s", cluster.Spec.GetCloudProvider())
	cluster.Spec.API = &api.AccessSpec{}
	if cluster.Spec.GetCloudProvider() == api.CloudProviderOpenstack {
		initializeOpenstackAPI(opt, cluster)
	} else if cluster.Spec.GetCloudProvider() == api.CloudProviderAzure {
		// Do nothing to disable the use of loadbalancer for the k8s API server.
		// TODO(kenji): Remove this condition once we support the loadbalancer
		// in pkg/model/azuremodel/api_loadbalancer.go.
		cluster.Spec.API = nil
		return nil
	} else if opt.APILoadBalancerType != "" || opt.APISSLCertificate != "" {
		cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}
	} else {
		switch cluster.Spec.Topology.Masters {
		case api.TopologyPublic:
			if dns.IsGossipHostname(cluster.Name) {
				// gossip DNS names don't work outside the cluster, so we use a LoadBalancer instead
				cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}
			} else {
				cluster.Spec.API.DNS = &api.DNSAccessSpec{}
			}

		case api.TopologyPrivate:
			cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}

		default:
			return fmt.Errorf("unknown master topology type: %q", cluster.Spec.Topology.Masters)
		}
	}

	if cluster.Spec.API.LoadBalancer != nil && cluster.Spec.API.LoadBalancer.Type == "" {
		switch opt.APILoadBalancerType {
		case "", "public":
			cluster.Spec.API.LoadBalancer.Type = api.LoadBalancerTypePublic
		case "internal":
			cluster.Spec.API.LoadBalancer.Type = api.LoadBalancerTypeInternal
		default:
			return fmt.Errorf("unknown api-loadbalancer-type: %q", opt.APILoadBalancerType)
		}
	}

	if cluster.Spec.API.LoadBalancer != nil && opt.APISSLCertificate != "" {
		cluster.Spec.API.LoadBalancer.SSLCertificate = opt.APISSLCertificate
	}

	if cluster.Spec.API.LoadBalancer != nil && cluster.Spec.API.LoadBalancer.Class == "" && cluster.Spec.GetCloudProvider() == api.CloudProviderAWS {
		switch opt.APILoadBalancerClass {
		case "", "classic":
			cluster.Spec.API.LoadBalancer.Class = api.LoadBalancerClassClassic
		case "network":
			cluster.Spec.API.LoadBalancer.Class = api.LoadBalancerClassNetwork
		default:
			return fmt.Errorf("unknown api-loadbalancer-class: %q", opt.APILoadBalancerClass)
		}
	}

	return nil
}

func initializeOpenstackAPI(opt *NewClusterOptions, cluster *api.Cluster) {
	if opt.APILoadBalancerType != "" {
		cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}
		provider := "haproxy"
		if opt.OpenstackLBOctavia {
			if opt.OpenstackOctaviaProvider != "" {
				provider = opt.OpenstackOctaviaProvider
			} else {
				provider = "octavia"
			}
		}

		LbMethod := "ROUND_ROBIN"
		if provider == "ovn" {
			LbMethod = "SOURCE_IP_PORT"
		}
		cluster.Spec.CloudProvider.Openstack.Loadbalancer = &api.OpenstackLoadbalancerConfig{
			FloatingNetwork: fi.String(opt.OpenstackExternalNet),
			Method:          fi.String(LbMethod),
			Provider:        fi.String(provider),
			UseOctavia:      fi.Bool(opt.OpenstackLBOctavia),
		}

		if opt.OpenstackLBSubnet != "" {
			cluster.Spec.CloudProvider.Openstack.Loadbalancer.FloatingSubnet = fi.String(opt.OpenstackLBSubnet)
		}
	}
}

func createEtcdCluster(etcdCluster string, masters []*api.InstanceGroup, encryptEtcdStorage bool, etcdStorageType string) api.EtcdClusterSpec {
	etcd := api.EtcdClusterSpec{}
	etcd.Name = etcdCluster

	// if this is the main cluster, we use 200 millicores by default.
	// otherwise we use 100 millicores by default.  100Mi is always default
	// for event and main clusters.  This is changeable in the kops cluster
	// configuration.
	if etcd.Name == "main" {
		cpuRequest := resource.MustParse("200m")
		etcd.CPURequest = &cpuRequest
	} else {
		cpuRequest := resource.MustParse("100m")
		etcd.CPURequest = &cpuRequest
	}
	memoryRequest := resource.MustParse("100Mi")
	etcd.MemoryRequest = &memoryRequest

	var names []string
	for _, ig := range masters {
		name := ig.ObjectMeta.Name
		// We expect the IG to have a `master-` prefix, but this is both superfluous
		// and not how we named things previously
		name = strings.TrimPrefix(name, "master-")
		names = append(names, name)
	}

	names = trimCommonPrefix(names)

	for i, ig := range masters {
		m := api.EtcdMemberSpec{}
		if encryptEtcdStorage {
			m.EncryptedVolume = &encryptEtcdStorage
		}
		if len(etcdStorageType) > 0 {
			m.VolumeType = fi.String(etcdStorageType)
		}
		m.Name = names[i]

		m.InstanceGroup = fi.String(ig.ObjectMeta.Name)
		etcd.Members = append(etcd.Members, m)
	}

	// Cilium etcd server is not compacted by the k8s API server.
	if etcd.Name == "cilium" {
		if etcd.Manager == nil {
			etcd.Manager = &api.EtcdManagerSpec{
				Env: []api.EnvVar{
					{Name: "ETCD_AUTO_COMPACTION_MODE", Value: "revision"},
					{Name: "ETCD_AUTO_COMPACTION_RETENTION", Value: "2500"},
				},
			}
		}
	}

	return etcd
}

func addCiliumNetwork(cluster *api.Cluster) {
	cilium := &api.CiliumNetworkingSpec{}
	cluster.Spec.Networking.Cilium = cilium
	cilium.EnableNodePort = true
	if cluster.Spec.KubeProxy == nil {
		cluster.Spec.KubeProxy = &api.KubeProxyConfig{}
	}
	enabled := false
	cluster.Spec.KubeProxy.Enabled = &enabled
}

// defaultImage returns the default Image, based on the cloudprovider
func defaultImage(cluster *kopsapi.Cluster, channel *kopsapi.Channel, architecture architectures.Architecture) string {
	if channel != nil {
		var kubernetesVersion *semver.Version
		if cluster.Spec.KubernetesVersion != "" {
			var err error
			kubernetesVersion, err = util.ParseKubernetesVersion(cluster.Spec.KubernetesVersion)
			if err != nil {
				klog.Warningf("cannot parse KubernetesVersion %q in cluster", cluster.Spec.KubernetesVersion)
			}
		}
		if kubernetesVersion != nil {
			image := channel.FindImage(cluster.Spec.GetCloudProvider(), *kubernetesVersion, architecture)
			if image != nil {
				return image.Name
			}
		}
	}

	switch cluster.Spec.GetCloudProvider() {
	case kopsapi.CloudProviderDO:
		return defaultDONodeImage
	}
	klog.Infof("Cannot set default Image for CloudProvider=%q", cluster.Spec.GetCloudProvider())
	return ""
}

func MachineArchitecture(cloud fi.Cloud, machineType string) (architectures.Architecture, error) {
	if machineType == "" {
		return architectures.ArchitectureAmd64, nil
	}

	// Some calls only have AWS initialised at this point and in other cases pass in nil as cloud.
	if cloud == nil {
		return architectures.ArchitectureAmd64, nil
	}

	switch cloud.ProviderID() {
	case kopsapi.CloudProviderAWS:
		info, err := cloud.(awsup.AWSCloud).DescribeInstanceType(machineType)
		if err != nil {
			return "", fmt.Errorf("error finding instance info for instance type %q: %w", machineType, err)
		}
		if info.ProcessorInfo == nil || len(info.ProcessorInfo.SupportedArchitectures) == 0 {
			return "", fmt.Errorf("error finding architecture info for instance type %q", machineType)
		}
		var unsupported []string
		for _, arch := range info.ProcessorInfo.SupportedArchitectures {
			// Return the first found supported architecture, in order of popularity
			switch fi.StringValue(arch) {
			case ec2.ArchitectureTypeX8664:
				return architectures.ArchitectureAmd64, nil
			case ec2.ArchitectureTypeArm64:
				return architectures.ArchitectureArm64, nil
			default:
				unsupported = append(unsupported, fi.StringValue(arch))
			}
		}
		return "", fmt.Errorf("unsupported architecture for instance type %q: %v", machineType, unsupported)
	default:
		// No other clouds are known to support any other architectures at this time
		return architectures.ArchitectureAmd64, nil
	}
}
