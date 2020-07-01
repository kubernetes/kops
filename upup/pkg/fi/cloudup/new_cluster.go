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
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	"k8s.io/kops"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
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

	// CloudProvider is the name of the cloud provider. The default is to guess based on the Zones name.
	CloudProvider string
	// Zones are the availability zones in which to run the cluster.
	Zones []string
	// MasterZones are the availability zones in which to run the masters. Defaults to the list in the Zones field.
	MasterZones []string

	// NetworkID is the ID of the shared network (VPC).
	// If empty, SubnetIDs are not empty, and on AWS or OpenStack, determines network ID from the first SubnetID.
	// If empty otherwise, creates a new network/VPC to be owned by the cluster.
	NetworkID string
	// SubnetIDs are the IDs of the shared subnets. If empty, creates new subnets to be owned by the cluster.
	SubnetIDs []string
	// Egress defines the method of traffic egress for subnets.
	Egress string

	// OpenstackExternalNet is the name of the external network for the openstack router
	OpenstackExternalNet     string
	OpenstackExternalSubnet  string
	OpenstackStorageIgnoreAZ bool
	OpenstackDNSServers      string
	OpenstackLbSubnet        string
	// OpenstackLBOctavia is boolean value should we use octavia or old loadbalancer api
	OpenstackLBOctavia bool

	// MasterCount is the number of masters to create. Defaults to the length of MasterZones
	// if MasterZones is explicitly nonempty, otherwise defaults to 1.
	MasterCount int32
	// EncryptEtcdStorage is whether to encrypt the etcd volumes.
	EncryptEtcdStorage bool
	// EtcdStorageType is the underlying cloud storage class of the etcd volumes.
	EtcdStorageType string
}

func (o *NewClusterOptions) InitDefaults() {
	o.Channel = api.DefaultChannel
	o.Authorization = AuthorizationFlagRBAC
}

type NewClusterResult struct {
	// Cluster is the initialized Cluster resource.
	Cluster *api.Cluster
	// InstanceGroups are the initialized InstanceGroup resources.
	InstanceGroups []*api.InstanceGroup

	// TODO remove after more create_cluster logic refactored in
	Channel         *api.Channel
	AllZones        sets.String
	ZoneToSubnetMap map[string]*api.ClusterSubnetSpec
	Masters         []*api.InstanceGroup
}

// NewCluster initializes cluster and instance groups specifications as
// intended for newly created clusters.
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

	allZones := sets.NewString()
	allZones.Insert(opt.Zones...)
	allZones.Insert(opt.MasterZones...)

	cluster.Spec.CloudProvider = opt.CloudProvider
	if cluster.Spec.CloudProvider == "" {
		for _, zone := range allZones.List() {
			cloud, known := fi.GuessCloudForZone(zone)
			if known {
				klog.Infof("Inferred %q cloud provider from zone %q", cloud, zone)
				cluster.Spec.CloudProvider = string(cloud)
				break
			}
		}
		if cluster.Spec.CloudProvider == "" {
			if allZones.Len() == 0 {
				return nil, fmt.Errorf("must specify --zones or --cloud")
			}
			return nil, fmt.Errorf("unable to infer cloud provider from zones (is there a typo in --zones?)")
		}
	}

	err = setupVPC(opt, &cluster)
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

	result := NewClusterResult{
		Cluster:         &cluster,
		InstanceGroups:  masters,
		Channel:         channel,
		AllZones:        allZones,
		ZoneToSubnetMap: zoneToSubnetMap,
		Masters:         masters,
	}
	return &result, nil
}

func setupVPC(opt *NewClusterOptions, cluster *api.Cluster) error {
	cluster.Spec.NetworkID = opt.NetworkID

	switch api.CloudProviderID(cluster.Spec.CloudProvider) {
	case api.CloudProviderAWS:
		if cluster.Spec.NetworkID == "" && len(opt.SubnetIDs) > 0 {
			cloudTags := map[string]string{}
			awsCloud, err := awsup.NewAWSCloud(opt.Zones[0][:len(opt.Zones[0])-1], cloudTags)
			if err != nil {
				return fmt.Errorf("error loading cloud: %v", err)
			}
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

	case api.CloudProviderOpenstack:
		if cluster.Spec.CloudConfig == nil {
			cluster.Spec.CloudConfig = &api.CloudConfiguration{}
		}
		cluster.Spec.CloudConfig.Openstack = &api.OpenstackConfiguration{
			Router: &api.OpenstackRouter{
				ExternalNetwork: fi.String(opt.OpenstackExternalNet),
			},
			BlockStorage: &api.OpenstackBlockStorageConfig{
				Version:  fi.String("v3"),
				IgnoreAZ: fi.Bool(opt.OpenstackStorageIgnoreAZ),
			},
			Monitor: &api.OpenstackMonitor{
				Delay:      fi.String("1m"),
				Timeout:    fi.String("30s"),
				MaxRetries: fi.Int(3),
			},
		}

		if cluster.Spec.NetworkID == "" && len(opt.SubnetIDs) > 0 {
			tags := make(map[string]string)
			tags[openstack.TagClusterName] = opt.ClusterName
			osCloud, err := openstack.NewOpenstackCloud(tags, &cluster.Spec)
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
			cluster.Spec.CloudConfig.Openstack.Router.DNSServers = fi.String(opt.OpenstackDNSServers)
		}
		if opt.OpenstackExternalSubnet != "" {
			cluster.Spec.CloudConfig.Openstack.Router.ExternalSubnet = fi.String(opt.OpenstackExternalSubnet)
		}
	}

	return nil
}

func setupZones(opt *NewClusterOptions, cluster *api.Cluster, allZones sets.String) (map[string]*api.ClusterSubnetSpec, error) {
	var err error
	zoneToSubnetMap := make(map[string]*api.ClusterSubnetSpec)

	if len(opt.Zones) == 0 {
		return nil, fmt.Errorf("must specify at least one zone for the cluster (use --zones)")
	}

	var zoneToSubnetProviderID map[string]string

	switch api.CloudProviderID(cluster.Spec.CloudProvider) {
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
				cluster.Spec.Subnets = append(cluster.Spec.Subnets, *subnet)
			}
			zoneToSubnetMap[zoneName] = subnet
		}
		return zoneToSubnetMap, nil

	case api.CloudProviderDO:
		if len(opt.Zones) > 1 {
			return nil, fmt.Errorf("digitalocean cloud provider currently only supports 1 region, expect multi-region support when digitalocean support is in beta")
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

	case api.CloudProviderALI:
		if len(opt.Zones) > 0 && len(opt.SubnetIDs) > 0 {
			zoneToSubnetProviderID, err = aliup.ZoneToVSwitchID(cluster.Spec.NetworkID, opt.Zones, opt.SubnetIDs)
			if err != nil {
				return nil, err
			}
		}

	case api.CloudProviderAWS:
		if len(opt.Zones) > 0 && len(opt.SubnetIDs) > 0 {
			zoneToSubnetProviderID, err = getZoneToSubnetProviderID(cluster.Spec.NetworkID, opt.Zones[0][:len(opt.Zones[0])-1], opt.SubnetIDs)
			if err != nil {
				return nil, err
			}
		}

	case api.CloudProviderOpenstack:
		if len(opt.Zones) > 0 && len(opt.SubnetIDs) > 0 {
			tags := make(map[string]string)
			tags[openstack.TagClusterName] = opt.ClusterName
			zoneToSubnetProviderID, err = getSubnetProviderID(&cluster.Spec, allZones.List(), opt.SubnetIDs, tags)
			if err != nil {
				return nil, err
			}
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
				subnet.ProviderID = subnetID
			}
			cluster.Spec.Subnets = append(cluster.Spec.Subnets, *subnet)
		}
		zoneToSubnetMap[zoneName] = subnet
	}

	return zoneToSubnetMap, nil
}

// TODO rename to getAWSZoneToSubnetProviderID
func getZoneToSubnetProviderID(VPCID string, region string, subnetIDs []string) (map[string]string, error) {
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

// TODO rename to getOpenstackZoneToSubnetProviderID
func getSubnetProviderID(spec *api.ClusterSpec, zones []string, subnetIDs []string, tags map[string]string) (map[string]string, error) {
	res := make(map[string]string)
	osCloud, err := openstack.NewOpenstackCloud(tags, spec)
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
			if api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderDO {
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
			if api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderGCE {
				g.Spec.Zones = []string{zone}
			}

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

		for _, etcdCluster := range EtcdClusters {
			etcd := &api.EtcdClusterSpec{}
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
				m := &api.EtcdMemberSpec{}
				if opt.EncryptEtcdStorage {
					m.EncryptedVolume = fi.Bool(opt.EncryptEtcdStorage)
				}
				if opt.EtcdStorageType != "" {
					m.VolumeType = fi.String(opt.EtcdStorageType)
				}
				m.Name = names[i]

				m.InstanceGroup = fi.String(ig.ObjectMeta.Name)
				etcd.Members = append(etcd.Members, m)
			}
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

	return names
}
