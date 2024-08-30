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
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/blang/semver/v4"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/clouds"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/util/pkg/architectures"
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
	// KubernetesFeatureGates is the list of Kubernetes feature gates to enable/disable.
	KubernetesFeatureGates []string
	// AdminAccess is the set of CIDR blocks permitted to connect to the Kubernetes API. It defaults to "0.0.0.0/0" and "::/0".
	AdminAccess []string
	// SSHAccess is the set of CIDR blocks permitted to connect to SSH on the nodes. It defaults to the value of AdminAccess.
	SSHAccess []string
	// NetworkCIDRs is the set of CIDR blocks of the cluster network.
	NetworkCIDRs []string

	// CloudProvider is the name of the cloud provider. The default is to guess based on the Zones name.
	CloudProvider string
	// Zones are the availability zones in which to run the cluster.
	Zones []string
	// ControlPlaneZones are the availability zones in which to run the control-plane nodes. Defaults to the list in the Zones field.
	ControlPlaneZones []string

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

	// ControlPlaneCount is the number of control-plane nodes to create. Defaults to the length of ControlPlaneZones.
	// if ControlPlaneZones is explicitly nonempty, otherwise defaults to 1.
	ControlPlaneCount int32
	// APIServerCount is the number of API servers to create. Defaults to 0.
	APIServerCount int32
	// EncryptEtcdStorage is whether to encrypt the etcd volumes.
	EncryptEtcdStorage *bool

	// EtcdClusters contains the names of the etcd clusters.
	EtcdClusters []string
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
	// Topology is the network topology to use. Defaults to "public" for IPv4 clusters and "private" for IPv6 clusters.
	Topology string
	// DNSType is the DNS type to use; "public" or "private". Defaults to "public".
	DNSType string
	// DNSZone is the DNS zone to use.
	DNSZone string

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

	Image             string
	NodeImage         string
	ControlPlaneImage string
	BastionImage      string
	ControlPlaneSizes []string
	NodeSizes         []string
}

func (o *NewClusterOptions) InitDefaults() {
	o.Channel = api.DefaultChannel
	o.Authorization = AuthorizationFlagRBAC
	o.AdminAccess = []string{"0.0.0.0/0", "::/0"}
	o.EtcdClusters = []string{"main", "events"}
	o.Networking = "cilium"
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
	channel, err := api.LoadChannel(clientset.VFSContext(), opt.Channel)
	if err != nil {
		return nil, err
	}

	cluster := &api.Cluster{
		ObjectMeta: v1.ObjectMeta{
			Name: opt.ClusterName,
		},
	}

	cluster.Spec.Channel = opt.Channel
	if opt.KubernetesVersion == "" {
		kubernetesVersion := api.RecommendedKubernetesVersion(channel, kops.Version)
		if kubernetesVersion != nil {
			cluster.Spec.KubernetesVersion = kubernetesVersion.String()
		}
	} else {
		cluster.Spec.KubernetesVersion = opt.KubernetesVersion
	}

	// Verify the kubernetes version early (while we can still handle the version nicely)
	if cluster.Spec.KubernetesVersion != "" {
		if _, err := util.ParseKubernetesVersion(cluster.Spec.KubernetesVersion); err != nil {
			return nil, fmt.Errorf("cannot parse KubernetesVersion %q: %w", cluster.Spec.KubernetesVersion, err)
		}
	}

	cluster.Spec.ConfigStore = api.ConfigStoreSpec{
		Base: opt.ConfigBase,
	}
	configBase, err := clientset.ConfigBaseFor(cluster)
	if err != nil {
		return nil, fmt.Errorf("error building ConfigBase for cluster: %v", err)
	}
	cluster.Spec.ConfigStore.Base = configBase.Path()

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
		AnonymousAuth: fi.PtrTo(false),
	}

	if len(opt.KubernetesFeatureGates) > 0 {
		cluster.Spec.Kubelet.FeatureGates = make(map[string]string)
		cluster.Spec.KubeAPIServer = &api.KubeAPIServerConfig{
			FeatureGates: make(map[string]string),
		}
		cluster.Spec.KubeControllerManager = &api.KubeControllerManagerConfig{
			FeatureGates: make(map[string]string),
		}
		cluster.Spec.KubeProxy = &api.KubeProxyConfig{
			FeatureGates: make(map[string]string),
		}
		cluster.Spec.KubeScheduler = &api.KubeSchedulerConfig{
			FeatureGates: make(map[string]string),
		}

		for _, featureGate := range opt.KubernetesFeatureGates {
			enabled := true
			if featureGate[0] == '+' {
				featureGate = featureGate[1:]
			}
			if featureGate[0] == '-' {
				enabled = false
				featureGate = featureGate[1:]
			}
			cluster.Spec.Kubelet.FeatureGates[featureGate] = strconv.FormatBool(enabled)
			cluster.Spec.KubeAPIServer.FeatureGates[featureGate] = strconv.FormatBool(enabled)
			cluster.Spec.KubeControllerManager.FeatureGates[featureGate] = strconv.FormatBool(enabled)
			cluster.Spec.KubeProxy.FeatureGates[featureGate] = strconv.FormatBool(enabled)
			cluster.Spec.KubeScheduler.FeatureGates[featureGate] = strconv.FormatBool(enabled)
		}
	}

	if len(opt.AdminAccess) == 0 {
		opt.AdminAccess = []string{"0.0.0.0/0", "::/0"}
	}
	cluster.Spec.API.Access = opt.AdminAccess
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
	allZones.Insert(opt.ControlPlaneZones...)

	if opt.CloudProvider == "" {
		cloud, err := clouds.GuessCloudForPath(cluster.Spec.ConfigStore.Base)
		if err != nil {
			return nil, fmt.Errorf("pass in the cloud provider explicitly using --cloud: %w", err)
		}
		klog.V(2).Infof("Inferred %q cloud provider from state store %q", cloud, cluster.Spec.ConfigStore.Base)
		opt.CloudProvider = string(cloud)
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
				ExternalNetwork: fi.PtrTo(opt.OpenstackExternalNet),
			},
			BlockStorage: &api.OpenstackBlockStorageConfig{
				Version:     fi.PtrTo("v3"),
				IgnoreAZ:    fi.PtrTo(opt.OpenstackStorageIgnoreAZ),
				ClusterName: opt.ClusterName,
			},
			Monitor: &api.OpenstackMonitor{
				Delay:      fi.PtrTo("15s"),
				Timeout:    fi.PtrTo("10s"),
				MaxRetries: fi.PtrTo(3),
			},
		}
		initializeOpenstack(opt, cluster)
		osCloud, err := openstack.NewOpenstackCloud(cluster, "openstackmodel")
		if err != nil {
			return nil, err
		}
		cloud = osCloud
	case api.CloudProviderScaleway:
		cluster.Spec.CloudProvider.Scaleway = &api.ScalewaySpec{}
	case api.CloudProviderMetal:
		if !featureflag.Metal.Enabled() {
			return nil, fmt.Errorf("bare-metal support requires the Metal feature flag to be enabled")
		}
		if cluster.Labels == nil {
			cluster.Labels = make(map[string]string)
		}
		cluster.Labels[api.AlphaLabelCloudProvider] = string(api.CloudProviderMetal)
	default:
		return nil, fmt.Errorf("unsupported cloud provider %s", opt.CloudProvider)
	}

	if opt.DiscoveryStore != "" {
		discoveryPath, err := clientset.VFSContext().BuildVfsPath(opt.DiscoveryStore)
		if err != nil {
			return nil, fmt.Errorf("error building DiscoveryStore for cluster: %v", err)
		}
		cluster.Spec.ServiceAccountIssuerDiscovery = &api.ServiceAccountIssuerDiscoveryConfig{
			DiscoveryStore: discoveryPath.Join(cluster.Name).Path(),
		}
		if cluster.GetCloudProvider() == api.CloudProviderAWS {
			cluster.Spec.ServiceAccountIssuerDiscovery.EnableAWSOIDCProvider = true
			cluster.Spec.IAM.UseServiceAccountExternalPermissions = fi.PtrTo(true)
		}
	}

	err = setupVPC(opt, cluster, cloud)
	if err != nil {
		return nil, err
	}

	zoneToSubnetsMap, err := setupZones(opt, cluster, allZones)
	if err != nil {
		return nil, err
	}

	err = setupNetworking(opt, cluster)
	if err != nil {
		return nil, err
	}

	bastions, err := setupTopology(opt, cluster, allZones)
	if err != nil {
		return nil, err
	}

	controlPlanes, err := setupControlPlane(opt, cluster, zoneToSubnetsMap)
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
		nodes, err = setupKarpenterNodes(cluster)
		if err != nil {
			return nil, err
		}
	case "cloudgroups":
		nodes, err = setupNodes(opt, cluster, zoneToSubnetsMap)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid value %q for --instance-manager", opt.InstanceManager)
	}

	apiservers, err := setupAPIServers(opt, cluster, zoneToSubnetsMap)
	if err != nil {
		return nil, err
	}

	err = setupAPI(opt, cluster)
	if err != nil {
		return nil, err
	}

	instanceGroups := append([]*api.InstanceGroup(nil), controlPlanes...)
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
				instanceGroup.Spec.Image, err = defaultImage(cluster, channel, architecture)
				if err != nil {
					return nil, err
				}
			}
		}

		// TODO: Clean up
		if g.IsControlPlane() {
			if g.Spec.MachineType == "" {
				g.Spec.MachineType, err = defaultMachineType(cloud, cluster, ig)
				if err != nil {
					return nil, fmt.Errorf("error assigning default machine type for control plane: %v", err)
				}

			}
		} else if g.Spec.Role == api.InstanceGroupRoleBastion {
			if g.Spec.MachineType == "" {
				g.Spec.MachineType, err = defaultMachineType(cloud, cluster, g)
				if err != nil {
					return nil, fmt.Errorf("error assigning default machine type for bastions: %v", err)
				}
			}
		} else {
			if g.IsAPIServerOnly() && !featureflag.APIServerNodes.Enabled() {
				return nil, fmt.Errorf("apiserver nodes requires the APIServerNodes feature flag to be enabled")
			}
			if g.Spec.MachineType == "" {
				g.Spec.MachineType, err = defaultMachineType(cloud, cluster, g)
				if err != nil {
					return nil, fmt.Errorf("error assigning default machine type for nodes: %v", err)
				}
			}

		}

		if ig.Spec.Tenancy != "" && ig.Spec.Tenancy != "default" {
			switch cluster.GetCloudProvider() {
			case api.CloudProviderAWS:
				if _, ok := awsDedicatedInstanceExceptions[g.Spec.MachineType]; ok {
					return nil, fmt.Errorf("invalid dedicated instance type: %s", g.Spec.MachineType)
				}
			default:
				klog.Warning("Trying to set tenancy on non-AWS environment")
			}
		}

		if ig.IsControlPlane() {
			if len(ig.Spec.Subnets) == 0 {
				return nil, fmt.Errorf("control-plane InstanceGroup %s did not specify any Subnets", g.ObjectMeta.Name)
			}
		} else if ig.IsAPIServerOnly() && cluster.Spec.IsIPv6Only() {
			if len(ig.Spec.Subnets) == 0 {
				for _, subnet := range cluster.Spec.Networking.Subnets {
					if subnet.Type != api.SubnetTypePrivate && subnet.Type != api.SubnetTypeUtility {
						ig.Spec.Subnets = append(g.Spec.Subnets, subnet.Name)
					}
				}
			}
		} else {
			if len(ig.Spec.Subnets) == 0 {
				for _, subnet := range cluster.Spec.Networking.Subnets {
					if subnet.Type != api.SubnetTypeDualStack && subnet.Type != api.SubnetTypeUtility {
						g.Spec.Subnets = append(g.Spec.Subnets, subnet.Name)
					}
				}
			}

			if len(g.Spec.Subnets) == 0 {
				for _, subnet := range cluster.Spec.Networking.Subnets {
					if subnet.Type != api.SubnetTypeUtility {
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
		Cluster:        cluster,
		InstanceGroups: instanceGroups,
		Channel:        channel,
	}
	return &result, nil
}

func setupVPC(opt *NewClusterOptions, cluster *api.Cluster, cloud fi.Cloud) error {
	ctx := context.TODO()
	cluster.Spec.Networking.NetworkID = opt.NetworkID

	switch cluster.GetCloudProvider() {
	case api.CloudProviderAWS:
		if cluster.Spec.Networking.NetworkID == "" && len(opt.SubnetIDs) > 0 {
			awsCloud := cloud.(awsup.AWSCloud)
			res, err := awsCloud.EC2().DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
				SubnetIds: []string{opt.SubnetIDs[0]},
			})
			if err != nil {
				return fmt.Errorf("error describing subnet %s: %v", opt.SubnetIDs[0], err)
			}
			if len(res.Subnets) == 0 || res.Subnets[0].VpcId == nil {
				return fmt.Errorf("failed to determine VPC id of subnet %s", opt.SubnetIDs[0])
			}
			cluster.Spec.Networking.NetworkID = *res.Subnets[0].VpcId
		}

		if featureflag.Spotinst.Enabled() {
			if opt.SpotinstProduct != "" {
				cluster.Spec.CloudProvider.AWS.SpotinstProduct = fi.PtrTo(opt.SpotinstProduct)
			}
			if opt.SpotinstOrientation != "" {
				cluster.Spec.CloudProvider.AWS.SpotinstOrientation = fi.PtrTo(opt.SpotinstOrientation)
			}
		}

	case api.CloudProviderGCE:
		if cluster.Spec.CloudConfig == nil {
			cluster.Spec.CloudConfig = &api.CloudConfiguration{}
		}
		cluster.Spec.CloudProvider.GCE.Project = opt.Project
		if cluster.Spec.CloudProvider.GCE.Project == "" {
			project, err := gce.DefaultProject()
			if err != nil {
				klog.Warningf("unable to get default google cloud project: %v", err)
			} else if project == "" {
				klog.Warningf("default google cloud project not set (try `gcloud config set project <name>`")
			} else {
				klog.Infof("using google cloud project: %s", project)
			}
			cluster.Spec.CloudProvider.GCE.Project = project
		}
		if opt.GCEServiceAccount != "" {
			// TODO remove this logging?
			klog.Infof("VMs will be configured to use specified Service Account: %v", opt.GCEServiceAccount)
			cluster.Spec.CloudProvider.GCE.ServiceAccount = opt.GCEServiceAccount
		}

	case api.CloudProviderOpenstack:
		if cluster.Spec.CloudConfig == nil {
			cluster.Spec.CloudConfig = &api.CloudConfiguration{}
		}

		if cluster.Spec.Networking.NetworkID == "" && len(opt.SubnetIDs) > 0 {
			osCloud, err := openstack.NewOpenstackCloud(cluster, "new-cluster-setupvpc")
			if err != nil {
				return fmt.Errorf("error loading cloud: %v", err)
			}

			res, err := osCloud.FindNetworkBySubnetID(opt.SubnetIDs[0])
			if err != nil {
				return fmt.Errorf("error finding network: %v", err)
			}
			cluster.Spec.Networking.NetworkID = res.ID
		}

		if opt.OpenstackDNSServers != "" {
			cluster.Spec.CloudProvider.Openstack.Router.DNSServers = fi.PtrTo(opt.OpenstackDNSServers)
		}
		if opt.OpenstackExternalSubnet != "" {
			cluster.Spec.CloudProvider.Openstack.Router.ExternalSubnet = fi.PtrTo(opt.OpenstackExternalSubnet)
		}
	case api.CloudProviderAzure:
		// TODO(kenji): Find a right place for this.

		// Creating an empty CloudConfig so that --cloud-config is passed to kubelet, api-server, etc.
		if cluster.Spec.CloudConfig == nil {
			cluster.Spec.CloudConfig = &api.CloudConfiguration{}
		}
	}

	return nil
}

func setupZones(opt *NewClusterOptions, cluster *api.Cluster, allZones sets.String) (map[string][]*api.ClusterSubnetSpec, error) {
	var err error
	zoneToSubnetsMap := make(map[string][]*api.ClusterSubnetSpec)

	var zoneToSubnetProviderID map[string]string

	switch cluster.GetCloudProvider() {
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
					subnet.ID = opt.SubnetIDs[0]
				}
				cluster.Spec.Networking.Subnets = append(cluster.Spec.Networking.Subnets, *subnet)
			}
			zoneToSubnetsMap[zoneName] = append(zoneToSubnetsMap[zoneName], subnet)
		}

		return zoneToSubnetsMap, nil

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
			cluster.Spec.Networking.Subnets = append(cluster.Spec.Networking.Subnets, *subnet)
		}
		zoneToSubnetsMap[region] = append(zoneToSubnetsMap[region], subnet)
		return zoneToSubnetsMap, nil

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

			// We create default subnets named the same as the cluster
			subnetName := cluster.Name

			subnet := model.FindSubnet(cluster, subnetName)
			if subnet == nil {
				subnet = &api.ClusterSubnetSpec{
					Name:   subnetName,
					Region: location,
				}
				cluster.Spec.Networking.Subnets = append(cluster.Spec.Networking.Subnets, *subnet)
			}
			zoneToSubnetsMap[zoneName] = append(zoneToSubnetsMap[zoneName], subnet)
		}
		return zoneToSubnetsMap, nil

	case api.CloudProviderAWS:
		if len(opt.Zones) > 0 && len(opt.SubnetIDs) > 0 {
			zoneToSubnetProviderID, err = getAWSZoneToSubnetProviderID(cluster.Spec.Networking.NetworkID, opt.Zones[0][:len(opt.Zones[0])-1], opt.SubnetIDs)
			if err != nil {
				return nil, err
			}
		}

	case api.CloudProviderOpenstack:
		if len(opt.Zones) > 0 && len(opt.SubnetIDs) > 0 {
			zoneToSubnetProviderID, err = getOpenstackZoneToSubnetProviderID(cluster, allZones.List(), opt.SubnetIDs)
			if err != nil {
				return nil, err
			}
		}

	case api.CloudProviderScaleway:
		if len(opt.Zones) > 1 {
			return nil, fmt.Errorf("scaleway cloud provider currently supports only one availability zone")
		}
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
				subnet.ID = subnetID
			}
			cluster.Spec.Networking.Subnets = append(cluster.Spec.Networking.Subnets, *subnet)
		}
		zoneToSubnetsMap[zoneName] = append(zoneToSubnetsMap[zoneName], subnet)
	}

	if len(opt.NetworkCIDRs) > 1 {
		for i, cidr := range opt.NetworkCIDRs[1:] {
			zoneName := opt.Zones[i%len(opt.Zones)]
			subnetName := fmt.Sprintf("%s-%d", zoneName, i+1)

			subnet := model.FindSubnet(cluster, subnetName)
			if subnet == nil {
				subnet = &api.ClusterSubnetSpec{
					Name:   subnetName,
					Zone:   zoneName,
					CIDR:   cidr,
					Egress: opt.Egress,
				}
				if subnetID, ok := zoneToSubnetProviderID[zoneName]; ok {
					subnet.ID = subnetID
				}
				cluster.Spec.Networking.Subnets = append(cluster.Spec.Networking.Subnets, *subnet)
			}
			zoneToSubnetsMap[zoneName] = append(zoneToSubnetsMap[zoneName], subnet)
		}
	}

	return zoneToSubnetsMap, nil
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

func getOpenstackZoneToSubnetProviderID(cluster *api.Cluster, zones []string, subnetIDs []string) (map[string]string, error) {
	res := make(map[string]string)
	osCloud, err := openstack.NewOpenstackCloud(cluster, "new-cluster-zone-to-subnet")
	if err != nil {
		return res, fmt.Errorf("error loading cloud: %v", err)
	}
	osCloud.UseZones(zones)

	networkInfo, err := osCloud.FindVPCInfo(cluster.Spec.Networking.NetworkID)
	if err != nil {
		return res, fmt.Errorf("error describing Network: %v", err)
	}
	if networkInfo == nil {
		return res, fmt.Errorf("network %q not found", cluster.Spec.Networking.NetworkID)
	}

	subnetByID := make(map[string]*fi.SubnetInfo)
	for _, subnetInfo := range networkInfo.Subnets {
		subnetByID[subnetInfo.ID] = subnetInfo
	}

	for _, subnetID := range subnetIDs {
		subnet, ok := subnetByID[subnetID]
		if !ok {
			return res, fmt.Errorf("subnet %s not found in network %s", subnetID, cluster.Spec.Networking.NetworkID)
		}

		if res[subnet.Zone] != "" {
			return res, fmt.Errorf("subnet %s and %s have the same zone", subnetID, res[subnet.Zone])
		}
		res[subnet.Zone] = subnetID
	}
	return res, nil
}

func setupControlPlane(opt *NewClusterOptions, cluster *api.Cluster, zoneToSubnetsMap map[string][]*api.ClusterSubnetSpec) ([]*api.InstanceGroup, error) {
	cloudProvider := cluster.GetCloudProvider()

	var controlPlanes []*api.InstanceGroup

	// Build the control-plane subnets.
	// The control-plane zones is the default set of zones unless explicitly set.
	// The control-plane count is the number of control-plane zones unless explicitly set.
	// We then round-robin around the zones.
	{
		controlPlaneCount := opt.ControlPlaneCount
		controlPlaneZones := opt.ControlPlaneZones
		if len(controlPlaneZones) != 0 {
			if controlPlaneCount != 0 && controlPlaneCount < int32(len(controlPlaneZones)) {
				return nil, fmt.Errorf("specified %d control-plane zones, but also requested %d control-plane nodes.  If specifying both, the count should match.", len(controlPlaneZones), controlPlaneCount)
			}

			if controlPlaneCount == 0 {
				// If control-plane count is not specified, default to the number of control-plane zones
				controlPlaneCount = int32(len(controlPlaneZones))
			}
		} else {
			// controlPlaneZones not set; default to same as node Zones
			controlPlaneZones = opt.Zones

			if controlPlaneCount == 0 {
				// If control-plane count is not specified, default to 1
				controlPlaneCount = 1
			}
		}

		if len(controlPlaneZones) == 0 {
			// Should be unreachable
			return nil, fmt.Errorf("cannot determine control-plane zones")
		}

		for i := 0; i < int(controlPlaneCount); i++ {
			zone := controlPlaneZones[i%len(controlPlaneZones)]
			name := zone
			if cloudProvider == api.CloudProviderDO {
				if int(controlPlaneCount) >= len(controlPlaneZones) {
					name += "-" + strconv.Itoa(1+(i/len(controlPlaneZones)))
				}
			} else {
				if int(controlPlaneCount) > len(controlPlaneZones) {
					name += "-" + strconv.Itoa(1+(i/len(controlPlaneZones)))
				}
			}

			g := &api.InstanceGroup{}
			g.Spec.Role = api.InstanceGroupRoleControlPlane
			g.Spec.MinSize = fi.PtrTo(int32(1))
			g.Spec.MaxSize = fi.PtrTo(int32(1))
			g.ObjectMeta.Name = "control-plane-" + name

			subnets := zoneToSubnetsMap[zone]
			switch len(subnets) {
			case 0:
				klog.Fatalf("subnet not found in zoneToSubnetsMap")
			default:
				// Use only the main subnet for control-plane nodes
				subnet := subnets[0]
				if opt.IPv6 && opt.Topology == api.TopologyPrivate {
					g.Spec.Subnets = append(g.Spec.Subnets, "dualstack-"+subnet.Name)
				} else {
					g.Spec.Subnets = append(g.Spec.Subnets, subnet.Name)
				}
			}

			if cloudProvider == api.CloudProviderGCE || cloudProvider == api.CloudProviderAzure {
				g.Spec.Zones = []string{zone}
			}

			if cluster.IsKubernetesLT("1.27") && cloudProvider == api.CloudProviderAWS {
				g.Spec.InstanceMetadata = &api.InstanceMetadataOptions{
					HTTPTokens: fi.PtrTo("required"),
				}
			}

			for i, size := range opt.ControlPlaneSizes {
				if i == 0 {
					g.Spec.MachineType = size
				}
				if cloudProvider == api.CloudProviderAWS && len(opt.ControlPlaneSizes) > 1 {
					if g.Spec.MixedInstancesPolicy == nil {
						g.Spec.MixedInstancesPolicy = &api.MixedInstancesPolicySpec{}

					}
					g.Spec.MixedInstancesPolicy.Instances = append(g.Spec.MixedInstancesPolicy.Instances, size)
				}
			}
			g.Spec.Image = opt.ControlPlaneImage

			controlPlanes = append(controlPlanes, g)
		}
	}

	// Build the Etcd clusters
	{
		controlPlaneAZs := sets.NewString()
		duplicateAZs := false
		for _, ig := range controlPlanes {
			zones, err := model.FindZonesForInstanceGroup(cluster, ig)
			if err != nil {
				return nil, err
			}
			for _, zone := range zones {
				if controlPlaneAZs.Has(zone) {
					duplicateAZs = true
				}

				controlPlaneAZs.Insert(zone)
			}
		}

		if duplicateAZs {
			klog.Warningf("Running with control-plane nodes in the same AZs; redundancy will be reduced")
		}

		clusters := opt.EtcdClusters

		if opt.Networking == "cilium-etcd" {
			clusters = append(clusters, "cilium")
		}

		encryptEtcdStorage := false
		if opt.EncryptEtcdStorage != nil {
			encryptEtcdStorage = fi.ValueOf(opt.EncryptEtcdStorage)
		} else if cloudProvider == api.CloudProviderAWS {
			encryptEtcdStorage = true
		}
		for _, etcdCluster := range clusters {
			etcd := createEtcdCluster(etcdCluster, controlPlanes, encryptEtcdStorage, opt.EtcdStorageType)
			cluster.Spec.EtcdClusters = append(cluster.Spec.EtcdClusters, etcd)
		}
	}

	return controlPlanes, nil
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

func setupNodes(opt *NewClusterOptions, cluster *api.Cluster, zoneToSubnetsMap map[string][]*api.ClusterSubnetSpec) ([]*api.InstanceGroup, error) {
	cloudProvider := cluster.GetCloudProvider()

	var nodes []*api.InstanceGroup

	if featureflag.AWSSingleNodesInstanceGroup.Enabled() && cloudProvider == api.CloudProviderAWS && len(opt.SubnetIDs) == 0 {
		nodeCount := opt.NodeCount
		if nodeCount == 0 {
			nodeCount = 1
		}

		g := &api.InstanceGroup{}
		g.Spec.Role = api.InstanceGroupRoleNode
		g.Spec.MinSize = fi.PtrTo(nodeCount)
		g.Spec.MaxSize = fi.PtrTo(nodeCount)
		g.ObjectMeta.Name = "nodes"

		for _, zone := range opt.Zones {
			subnets := zoneToSubnetsMap[zone]
			for _, subnet := range subnets[1:] {
				g.Spec.Subnets = append(g.Spec.Subnets, subnet.Name)
			}
		}

		for i, size := range opt.NodeSizes {
			if i == 0 {
				g.Spec.MachineType = size
			}
			if len(opt.NodeSizes) > 1 {
				if g.Spec.MixedInstancesPolicy == nil {
					g.Spec.MixedInstancesPolicy = &api.MixedInstancesPolicySpec{}

				}
				g.Spec.MixedInstancesPolicy.Instances = append(g.Spec.MixedInstancesPolicy.Instances, size)
			}
		}
		g.Spec.Image = opt.NodeImage

		nodes = append(nodes, g)
		return nodes, nil
	}

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
		g.Spec.MinSize = fi.PtrTo(count)
		g.Spec.MaxSize = fi.PtrTo(count)
		g.ObjectMeta.Name = "nodes-" + zone

		subnets := zoneToSubnetsMap[zone]
		switch len(subnets) {
		case 0:
			klog.Fatalf("subnet not found in zoneToSubnetsMap")
		case 1:
			g.Spec.Subnets = append(g.Spec.Subnets, subnets[0].Name)
		default:
			// The main subnet may be smaller and should not be used for worker nodes
			for _, subnet := range subnets[1:] {
				g.Spec.Subnets = append(g.Spec.Subnets, subnet.Name)
			}
		}

		if cloudProvider == api.CloudProviderGCE || cloudProvider == api.CloudProviderAzure {
			g.Spec.Zones = []string{zone}
		}

		if cluster.IsKubernetesLT("1.27") {
			if cloudProvider == api.CloudProviderAWS {
				g.Spec.InstanceMetadata = &api.InstanceMetadataOptions{
					HTTPPutResponseHopLimit: fi.PtrTo(int64(1)),
					HTTPTokens:              fi.PtrTo("required"),
				}
			}
		}

		if cloudProvider == api.CloudProviderGCE {
			if g.Spec.NodeLabels == nil {
				g.Spec.NodeLabels = make(map[string]string)
			}
			if cluster.IsKubernetesLT("1.29") {
				g.Spec.NodeLabels["cloud.google.com/metadata-proxy-ready"] = "true"
			}
		}

		for i, size := range opt.NodeSizes {
			if i == 0 {
				g.Spec.MachineType = size
			}
			if cloudProvider == api.CloudProviderAWS && len(opt.NodeSizes) > 1 {
				if g.Spec.MixedInstancesPolicy == nil {
					g.Spec.MixedInstancesPolicy = &api.MixedInstancesPolicySpec{}

				}
				g.Spec.MixedInstancesPolicy.Instances = append(g.Spec.MixedInstancesPolicy.Instances, size)
			}
		}
		g.Spec.Image = opt.NodeImage

		nodes = append(nodes, g)
	}

	return nodes, nil
}

func setupKarpenterNodes(cluster *api.Cluster) ([]*api.InstanceGroup, error) {
	g := &api.InstanceGroup{}
	g.Spec.Role = api.InstanceGroupRoleNode
	g.Spec.Manager = api.InstanceManagerKarpenter
	g.ObjectMeta.Name = "nodes"

	if cluster.IsKubernetesLT("1.27") {
		g.Spec.InstanceMetadata = &api.InstanceMetadataOptions{
			HTTPPutResponseHopLimit: fi.PtrTo(int64(1)),
			HTTPTokens:              fi.PtrTo("required"),
		}
	}

	return []*api.InstanceGroup{g}, nil
}

func setupAPIServers(opt *NewClusterOptions, cluster *api.Cluster, zoneToSubnetsMap map[string][]*api.ClusterSubnetSpec) ([]*api.InstanceGroup, error) {
	cloudProvider := cluster.GetCloudProvider()

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
		g.Spec.MinSize = fi.PtrTo(count)
		g.Spec.MaxSize = fi.PtrTo(count)
		g.ObjectMeta.Name = "apiserver-" + zone

		subnets := zoneToSubnetsMap[zone]
		if len(subnets) == 0 {
			klog.Fatalf("subnet not found in zoneToSubnetsMap")
		}
		for _, subnet := range subnets {
			g.Spec.Subnets = append(g.Spec.Subnets, subnet.Name)
		}

		if cloudProvider == api.CloudProviderGCE || cloudProvider == api.CloudProviderAzure {
			g.Spec.Zones = []string{zone}
		}

		if cluster.IsKubernetesLT("1.27") {
			if cloudProvider == api.CloudProviderAWS {
				g.Spec.InstanceMetadata = &api.InstanceMetadataOptions{
					HTTPPutResponseHopLimit: fi.PtrTo(int64(1)),
					HTTPTokens:              fi.PtrTo("required"),
				}
			}
		}

		nodes = append(nodes, g)
	}

	return nodes, nil
}

func setupNetworking(opt *NewClusterOptions, cluster *api.Cluster) error {
	switch opt.Networking {
	case "kubenet":
		cluster.Spec.Networking.Kubenet = &api.KubenetNetworkingSpec{}
	case "external":
		cluster.Spec.Networking.External = &api.ExternalNetworkingSpec{}
	case "cni":
		cluster.Spec.Networking.CNI = &api.CNINetworkingSpec{}
	case "kopeio-vxlan", "kopeio":
		cluster.Spec.Networking.Kopeio = &api.KopeioNetworkingSpec{}
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
		cluster.Spec.Networking.KubeRouter = &api.KuberouterNetworkingSpec{}
		if cluster.Spec.KubeProxy == nil {
			cluster.Spec.KubeProxy = &api.KubeProxyConfig{}
		}
		enabled := false
		cluster.Spec.KubeProxy.Enabled = &enabled
	case "amazonvpc", "amazon-vpc-routed-eni":
		cluster.Spec.Networking.AmazonVPC = &api.AmazonVPCNetworkingSpec{}
	case "cilium", "":
		addCiliumNetwork(cluster)
	case "cilium-etcd":
		addCiliumNetwork(cluster)
		cluster.Spec.Networking.Cilium.EtcdManaged = true
	case "cilium-eni":
		addCiliumNetwork(cluster)
		cluster.Spec.Networking.Cilium.IPAM = "eni"
	case "gcp", "gce":
		cluster.Spec.Networking.GCP = &api.GCPNetworkingSpec{}
	default:
		return fmt.Errorf("unknown networking mode %q", opt.Networking)
	}

	klog.V(4).Infof("networking mode=%s => %s", opt.Networking, fi.DebugAsJsonString(cluster.Spec.Networking))

	return nil
}

func setupTopology(opt *NewClusterOptions, cluster *api.Cluster, allZones sets.String) ([]*api.InstanceGroup, error) {
	var bastions []*api.InstanceGroup

	if opt.Topology == "" {
		if opt.IPv6 {
			opt.Topology = api.TopologyPrivate
		} else {
			opt.Topology = api.TopologyPublic
		}
	}

	cluster.Spec.Networking.Topology = &api.TopologySpec{}

	err := setupDNSTopology(opt, cluster)
	if err != nil {
		return nil, err
	}

	switch opt.Topology {
	case api.TopologyPublic:

		if opt.Bastion {
			return nil, fmt.Errorf("bastion supports --topology='private' only")
		}

		for i := range cluster.Spec.Networking.Subnets {
			cluster.Spec.Networking.Subnets[i].Type = api.SubnetTypePublic
		}

	case api.TopologyPrivate:
		if cluster.Spec.Networking.Kubenet != nil {
			return nil, fmt.Errorf("invalid networking option %s. Kubenet does not support private topology", opt.Networking)
		}

		for i := range cluster.Spec.Networking.Subnets {
			cluster.Spec.Networking.Subnets[i].Type = api.SubnetTypePrivate
		}

		var zoneToSubnetProviderID map[string]string
		var err error
		if len(opt.Zones) > 0 && len(opt.UtilitySubnetIDs) > 0 {
			switch cluster.GetCloudProvider() {
			case api.CloudProviderAWS:
				zoneToSubnetProviderID, err = getAWSZoneToSubnetProviderID(cluster.Spec.Networking.NetworkID, opt.Zones[0][:len(opt.Zones[0])-1], opt.UtilitySubnetIDs)
				if err != nil {
					return nil, err
				}
			case api.CloudProviderOpenstack:
				zoneToSubnetProviderID, err = getOpenstackZoneToSubnetProviderID(cluster, allZones.List(), opt.UtilitySubnetIDs)
				if err != nil {
					return nil, err
				}
			}
		}

		if opt.IPv6 {
			var dualStackSubnets []api.ClusterSubnetSpec

			for _, s := range cluster.Spec.Networking.Subnets {
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
					subnet.ID = subnetID
				}
				dualStackSubnets = append(dualStackSubnets, subnet)
			}
			cluster.Spec.Networking.Subnets = append(cluster.Spec.Networking.Subnets, dualStackSubnets...)
		}

		addUtilitySubnets := true
		switch cluster.GetCloudProvider() {
		case api.CloudProviderGCE:
			// GCE does not need utility subnets
			addUtilitySubnets = false
		}

		if addUtilitySubnets {
			var utilitySubnets []api.ClusterSubnetSpec

			for _, s := range cluster.Spec.Networking.Subnets {
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
					subnet.ID = subnetID
				}
				utilitySubnets = append(utilitySubnets, subnet)
			}
			cluster.Spec.Networking.Subnets = append(cluster.Spec.Networking.Subnets, utilitySubnets...)
		}

		if opt.Bastion {
			bastionGroup := &api.InstanceGroup{}
			bastionGroup.Spec.Role = api.InstanceGroupRoleBastion
			bastionGroup.ObjectMeta.Name = "bastions"
			bastionGroup.Spec.MaxSize = fi.PtrTo(int32(1))
			bastionGroup.Spec.MinSize = fi.PtrTo(int32(1))
			bastions = append(bastions, bastionGroup)

			if cluster.PublishesDNSRecords() {
				cluster.Spec.Networking.Topology.Bastion = &api.BastionSpec{
					PublicName: "bastion." + cluster.Name,
				}
			}
			if opt.IPv6 {
				for _, s := range cluster.Spec.Networking.Subnets {
					if s.Type == api.SubnetTypeDualStack {
						bastionGroup.Spec.Subnets = append(bastionGroup.Spec.Subnets, s.Name)
					}
				}
			}
			if cluster.GetCloudProvider() == api.CloudProviderGCE {
				bastionGroup.Spec.Zones = allZones.List()
			}

			if cluster.IsKubernetesLT("1.27") {
				if cluster.GetCloudProvider() == api.CloudProviderAWS {
					bastionGroup.Spec.InstanceMetadata = &api.InstanceMetadataOptions{
						HTTPPutResponseHopLimit: fi.PtrTo(int64(1)),
						HTTPTokens:              fi.PtrTo("required"),
					}
				}
			}

			bastionGroup.Spec.Image = opt.BastionImage
		}

	default:
		return nil, fmt.Errorf("invalid topology %s", opt.Topology)
	}

	if opt.IPv6 {
		cluster.Spec.Networking.NonMasqueradeCIDR = "::/0"
		cluster.Spec.ExternalCloudControllerManager = &api.CloudControllerManagerConfig{}
		if cluster.GetCloudProvider() == api.CloudProviderAWS {
			for i := range cluster.Spec.Networking.Subnets {
				cluster.Spec.Networking.Subnets[i].IPv6CIDR = fmt.Sprintf("/64#%x", i)
			}
		} else {
			klog.Errorf("IPv6 support is available only on AWS")
		}
	}

	return bastions, nil
}

func setupDNSTopology(opt *NewClusterOptions, cluster *api.Cluster) error {
	switch strings.ToLower(opt.DNSType) {
	case "":
		if opt.DNSZone != "" {
			// Use dns=public if zone is specified
			cluster.Spec.Networking.Topology.DNS = api.DNSTypePublic
		} else {
			if cluster.UsesLegacyGossip() {
				// Warn about using dns=none instead of Gossip
				klog.Warningf("Gossip is deprecated, using None DNS instead")
			}
			// Default to dns=none instead of dns=public for all cloud providers
			cluster.Spec.Networking.Topology.DNS = api.DNSTypeNone
		}
	case "public":
		cluster.Spec.Networking.Topology.DNS = api.DNSTypePublic
	case "private":
		cluster.Spec.Networking.Topology.DNS = api.DNSTypePrivate
	case "none":
		cluster.Spec.Networking.Topology.DNS = api.DNSTypeNone
	default:
		return fmt.Errorf("unknown DNSType: %q", opt.DNSType)
	}
	return nil
}

func setupAPI(opt *NewClusterOptions, cluster *api.Cluster) error {
	// Populate the API access, so that it can be discoverable
	klog.Infof("Cloud Provider ID: %q", cluster.GetCloudProvider())
	if opt.APILoadBalancerType != "" || opt.APISSLCertificate != "" {
		cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}
	} else {
		switch opt.Topology {
		case api.TopologyPublic:
			if cluster.UsesLegacyGossip() || cluster.UsesNoneDNS() {
				// gossip DNS names don't work outside the cluster, so we use a LoadBalancer instead
				cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}
			} else {
				cluster.Spec.API.DNS = &api.DNSAccessSpec{}
			}

		case api.TopologyPrivate:
			cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}

		default:
			return fmt.Errorf("unknown topology type: %q", opt.Topology)
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

	if cluster.Spec.API.LoadBalancer != nil && cluster.Spec.API.LoadBalancer.Class == "" && cluster.GetCloudProvider() == api.CloudProviderAWS {
		switch opt.APILoadBalancerClass {
		case "classic":
			klog.Warning("AWS Classic Load Balancer support for API is deprecated and should not be used for newly created clusters")
			cluster.Spec.API.LoadBalancer.Class = api.LoadBalancerClassClassic
		case "", "network":
			cluster.Spec.API.LoadBalancer.Class = api.LoadBalancerClassNetwork
		default:
			return fmt.Errorf("unknown api-loadbalancer-class: %q", opt.APILoadBalancerClass)
		}
	}

	return nil
}

func initializeOpenstack(opt *NewClusterOptions, cluster *api.Cluster) {
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
			FloatingNetwork: fi.PtrTo(opt.OpenstackExternalNet),
			Method:          fi.PtrTo(LbMethod),
			Provider:        fi.PtrTo(provider),
			UseOctavia:      fi.PtrTo(opt.OpenstackLBOctavia),
		}

		if opt.OpenstackLBSubnet != "" {
			cluster.Spec.CloudProvider.Openstack.Loadbalancer.FloatingSubnet = fi.PtrTo(opt.OpenstackLBSubnet)
		}
	}

	// this is needed in new clusters, otherwise openstack clients will automatically try to use openstack designate
	if strings.ToLower(opt.DNSType) == "none" {
		if cluster.Spec.Networking.Topology == nil {
			cluster.Spec.Networking.Topology = &api.TopologySpec{
				DNS: api.DNSTypeNone,
			}
		} else {
			cluster.Spec.Networking.Topology.DNS = api.DNSTypeNone
		}
	}

	if cluster.Spec.ExternalCloudControllerManager == nil {
		cluster.Spec.ExternalCloudControllerManager = &api.CloudControllerManagerConfig{}
	}
	cluster.Spec.ExternalCloudControllerManager.ClusterName = opt.ClusterName
}

func createEtcdCluster(etcdCluster string, controlPlanes []*api.InstanceGroup, encryptEtcdStorage bool, etcdStorageType string) api.EtcdClusterSpec {
	etcd := api.EtcdClusterSpec{
		Name: etcdCluster,
		Manager: &api.EtcdManagerSpec{
			BackupRetentionDays: fi.PtrTo[uint32](90),
		},
	}

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
	for _, ig := range controlPlanes {
		name := ig.ObjectMeta.Name
		// We expect the IG to have a `control-plane-` or `master-` prefix, but this is both superfluous
		// and not how we named things previously
		name = strings.TrimPrefix(name, "control-plane-")
		name = strings.TrimPrefix(name, "master-")
		names = append(names, name)
	}

	names = trimCommonPrefix(names)

	for i, ig := range controlPlanes {
		m := api.EtcdMemberSpec{}
		if encryptEtcdStorage {
			m.EncryptedVolume = &encryptEtcdStorage
		}
		if len(etcdStorageType) > 0 {
			m.VolumeType = fi.PtrTo(etcdStorageType)
		}
		m.Name = names[i]

		m.InstanceGroup = fi.PtrTo(ig.ObjectMeta.Name)
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
func defaultImage(cluster *api.Cluster, channel *api.Channel, architecture architectures.Architecture) (string, error) {
	kubernetesVersion, err := util.ParseKubernetesVersion(cluster.Spec.KubernetesVersion)
	if err != nil {
		return "", fmt.Errorf("unable to parse kubernetes version %q", cluster.Spec.KubernetesVersion)
	}

	if channel != nil {
		image := channel.FindImage(cluster.GetCloudProvider(), *kubernetesVersion, architecture)
		if image != nil {
			return image.Name, nil
		}
	}

	if kubernetesVersion.LT(semver.MustParse("1.27.0")) {
		switch cluster.GetCloudProvider() {
		case api.CloudProviderDO:
			return defaultDOImageFocal, nil
		case api.CloudProviderHetzner:
			return defaultHetznerImageFocal, nil
		case api.CloudProviderScaleway:
			return defaultScalewayImageFocal, nil
		}
	} else {
		switch cluster.GetCloudProvider() {
		case api.CloudProviderDO:
			return defaultDOImageJammy, nil
		case api.CloudProviderHetzner:
			return defaultHetznerImageJammy, nil
		case api.CloudProviderScaleway:
			return defaultScalewayImageJammy, nil
		case api.CloudProviderMetal:
			return "dummy-metal-image", nil
		}
	}

	return "", fmt.Errorf("unable to determine default image for cloud provider %q and architecture %q", cluster.GetCloudProvider(), architecture)
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
	case api.CloudProviderAWS:
		info, err := cloud.(awsup.AWSCloud).DescribeInstanceType(machineType)
		if err != nil {
			return "", fmt.Errorf("error finding instance info for instance type %q: %w", machineType, err)
		}
		if info.ProcessorInfo == nil || len(info.ProcessorInfo.SupportedArchitectures) == 0 {
			return "", fmt.Errorf("unable to determine architecture info for instance type %q", machineType)
		}
		var unsupported []ec2types.ArchitectureType
		for _, arch := range info.ProcessorInfo.SupportedArchitectures {
			// Return the first found supported architecture, in order of popularity
			switch arch {
			case ec2types.ArchitectureTypeX8664:
				return architectures.ArchitectureAmd64, nil
			case ec2types.ArchitectureTypeArm64:
				return architectures.ArchitectureArm64, nil
			default:
				unsupported = append(unsupported, arch)
			}
		}
		return "", fmt.Errorf("unsupported architecture for instance type %q: %v", machineType, unsupported)
	default:
		// No other clouds are known to support any other architectures at this time
		return architectures.ArchitectureAmd64, nil
	}
}
