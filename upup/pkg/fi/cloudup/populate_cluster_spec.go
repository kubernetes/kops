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

package cloudup

import (
	"context"
	"fmt"
	"net"
	"strings"

	"k8s.io/klog/v2"

	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/model/components/etcdmanager"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
	"k8s.io/kops/util/pkg/reflectutils"
	"k8s.io/kops/util/pkg/vfs"
)

type populateClusterSpec struct {
	cloud fi.Cloud

	// InputCluster is the api object representing the whole cluster, as input by the user
	// We build it up into a complete config, but we write the values as input
	InputCluster *kopsapi.Cluster

	// InputInstanceGroups are the instance groups in the cluster
	InputInstanceGroups []*kopsapi.InstanceGroup

	// fullCluster holds the built completed cluster spec
	fullCluster *kopsapi.Cluster

	// assetBuilder holds the AssetBuilder, used to store assets we discover / remap
	assetBuilder *assets.AssetBuilder
}

// PopulateClusterSpec takes a user-specified cluster spec, and computes the full specification that should be set on the cluster.
// We do this so that we don't need any real "brains" on the node side.
func PopulateClusterSpec(ctx context.Context, clientset simple.Clientset, cluster *kopsapi.Cluster, instanceGroups []*kopsapi.InstanceGroup, cloud fi.Cloud, assetBuilder *assets.AssetBuilder) (*kopsapi.Cluster, error) {
	c := &populateClusterSpec{
		cloud:               cloud,
		InputCluster:        cluster,
		InputInstanceGroups: instanceGroups,
		assetBuilder:        assetBuilder,
	}
	err := c.run(ctx, clientset)
	if err != nil {
		return nil, err
	}
	return c.fullCluster, nil
}

// Here be dragons
//
// This function has some `interesting` things going on.
// In an effort to let the cluster.Spec fall through I am
// hard coding topology in two places.. It seems and feels
// very wrong.. but at least now my new cluster.Spec.Topology
// struct is falling through..
// @kris-nova
func (c *populateClusterSpec) run(ctx context.Context, clientset simple.Clientset) error {
	if errs := validation.ValidateCluster(c.InputCluster, false, clientset.VFSContext()); len(errs) != 0 {
		return errs.ToAggregate()
	}

	cloud := c.cloud

	// Copy cluster & instance groups, so we can modify them freely
	cluster := &kopsapi.Cluster{}

	reflectutils.JSONMergeStruct(cluster, c.InputCluster)

	err := c.assignSubnets(cluster)
	if err != nil {
		return err
	}

	err = cluster.FillDefaults()
	if err != nil {
		return err
	}

	err = PerformAssignments(cluster, clientset.VFSContext(), cloud)
	if err != nil {
		return err
	}

	// TODO: Move to validate?
	// Check that instance groups are defined in valid zones
	{
		// TODO: Check that instance groups referenced here exist
		//clusterSubnets := make(map[string]*kopsapi.ClusterSubnetSpec)
		//for _, subnet := range cluster.Spec.Subnets {
		//	if clusterSubnets[subnet.Name] != nil {
		//		return fmt.Errorf("Subnets contained a duplicate value: %v", subnet.Name)
		//	}
		//	clusterSubnets[subnet.Name] = subnet
		//}

		// Check etcd configuration
		{
			for i, etcd := range cluster.Spec.EtcdClusters {
				if etcd.Name == "" {
					return fmt.Errorf("EtcdClusters #%d did not specify a Name", i)
				}

				for i, m := range etcd.Members {
					if m.Name == "" {
						return fmt.Errorf("EtcdMember #%d of etcd-cluster %s did not specify a Name", i, etcd.Name)
					}

					if fi.ValueOf(m.InstanceGroup) == "" {
						return fmt.Errorf("EtcdMember %s:%s did not specify a InstanceGroup", etcd.Name, m.Name)
					}
				}

				etcdInstanceGroups := make(map[string]kopsapi.EtcdMemberSpec)
				etcdNames := make(map[string]kopsapi.EtcdMemberSpec)

				for _, m := range etcd.Members {
					if _, ok := etcdNames[m.Name]; ok {
						return fmt.Errorf("EtcdMembers found with same name %q in etcd-cluster %q", m.Name, etcd.Name)
					}

					instanceGroupName := fi.ValueOf(m.InstanceGroup)

					if _, ok := etcdInstanceGroups[instanceGroupName]; ok {
						klog.Warningf("EtcdMembers are in the same InstanceGroup %q in etcd-cluster %q (fault-tolerance may be reduced)", instanceGroupName, etcd.Name)
					}

					//if clusterSubnets[zone] == nil {
					//	return fmt.Errorf("EtcdMembers for %q is configured in zone %q, but that is not configured at the k8s-cluster level", etcd.Name, m.Zone)
					//}
					etcdNames[m.Name] = m
					etcdInstanceGroups[instanceGroupName] = m
				}

				if (len(etcdNames) % 2) == 0 {
					// Not technically a requirement, but doesn't really make sense to allow
					return fmt.Errorf("there should be an odd number of control-plane-zones, for etcd's quorum.  Hint: Use --zones and --control-plane-zones to declare worker and control plane node zones separately")
				}
			}
		}
	}

	configBase, err := clientset.VFSContext().BuildVfsPath(cluster.Spec.ConfigStore.Base)
	if err != nil {
		return fmt.Errorf("error parsing ConfigStore.Base %q: %v", cluster.Spec.ConfigStore.Base, err)
	}
	if !vfs.IsClusterReadable(configBase) {
		// We could implement this approach, but it seems better to get all clouds using cluster-readable storage
		return fmt.Errorf("ConfigStore.Base path is not cluster readable: %v", cluster.Spec.ConfigStore.Base)
	}

	keyStore, err := clientset.KeyStore(cluster)
	if err != nil {
		return err
	}

	if cluster.Spec.ConfigStore.Keypairs == "" {
		hasVFSPath, ok := keyStore.(fi.HasVFSPath)
		if !ok {
			// We will mirror to ConfigBase
			basedir := configBase.Join("pki")
			cluster.Spec.ConfigStore.Keypairs = basedir.Path()
		} else if vfs.IsClusterReadable(hasVFSPath.VFSPath()) {
			vfsPath := hasVFSPath.VFSPath()
			cluster.Spec.ConfigStore.Keypairs = vfsPath.Path()
		} else {
			// We could implement this approach, but it seems better to get all clouds using cluster-readable storage
			return fmt.Errorf("keyStore path is not cluster readable: %v", hasVFSPath.VFSPath())
		}
	}

	secretStore, err := clientset.SecretStore(cluster)
	if err != nil {
		return err
	}

	if cluster.Spec.ConfigStore.Secrets == "" {
		hasVFSPath, ok := secretStore.(fi.HasVFSPath)
		if !ok {
			// We will mirror to ConfigBase
			basedir := configBase.Join("secrets")
			cluster.Spec.ConfigStore.Secrets = basedir.Path()
		} else if vfs.IsClusterReadable(hasVFSPath.VFSPath()) {
			vfsPath := hasVFSPath.VFSPath()
			cluster.Spec.ConfigStore.Secrets = vfsPath.Path()
		} else {
			// We could implement this approach, but it seems better to get all clouds using cluster-readable storage
			return fmt.Errorf("secrets path is not cluster readable: %v", hasVFSPath.VFSPath())
		}
	}

	// Normalize k8s version
	versionWithoutV := strings.TrimSpace(cluster.Spec.KubernetesVersion)
	versionWithoutV = strings.TrimPrefix(versionWithoutV, "v")
	if cluster.Spec.KubernetesVersion != versionWithoutV {
		klog.V(2).Infof("Normalizing kubernetes version: %q -> %q", cluster.Spec.KubernetesVersion, versionWithoutV)
		cluster.Spec.KubernetesVersion = versionWithoutV
	}

	if cluster.Spec.CloudProvider.Openstack == nil {
		if cluster.Spec.API.DNS == nil && cluster.Spec.API.LoadBalancer == nil {
			subnetTypesByName := map[string]kopsapi.SubnetType{}
			for _, subnet := range cluster.Spec.Networking.Subnets {
				subnetTypesByName[subnet.Name] = subnet.Type
			}

			haveAPIServerNodes := false
			for _, ig := range c.InputInstanceGroups {
				if ig.Spec.Role == kopsapi.InstanceGroupRoleAPIServer {
					haveAPIServerNodes = true
				}
			}
			for _, ig := range c.InputInstanceGroups {
				if ig.Spec.Role == kopsapi.InstanceGroupRoleAPIServer || (!haveAPIServerNodes && ig.Spec.Role == kopsapi.InstanceGroupRoleControlPlane) {
					for _, subnet := range ig.Spec.Subnets {
						switch subnetTypesByName[subnet] {
						case kopsapi.SubnetTypePrivate:
							cluster.Spec.API.LoadBalancer = &kopsapi.LoadBalancerAccessSpec{}
						case kopsapi.SubnetTypePublic:
							cluster.Spec.API.DNS = &kopsapi.DNSAccessSpec{}
						}
					}
				}
			}
			// If both public and private, go with the load balancer.
			if cluster.Spec.API.LoadBalancer != nil {
				cluster.Spec.API.DNS = nil
			}
		}

		if cluster.Spec.API.LoadBalancer != nil && cluster.Spec.API.LoadBalancer.Type == "" {
			cluster.Spec.API.LoadBalancer.Type = kopsapi.LoadBalancerTypePublic
		}
	}

	if cluster.Spec.API.LoadBalancer != nil && cluster.Spec.API.LoadBalancer.Class == "" && cluster.Spec.CloudProvider.AWS != nil {
		cluster.Spec.API.LoadBalancer.Class = kopsapi.LoadBalancerClassClassic
	}

	if cluster.Spec.DNSZone == "" && cluster.PublishesDNSRecords() {
		dns, err := cloud.DNS()
		if err != nil {
			return err
		}
		if dns != nil {
			dnsType := kopsapi.DNSTypePublic
			if cluster.Spec.Networking.Topology != nil && cluster.Spec.Networking.Topology.DNS != "" {
				dnsType = cluster.Spec.Networking.Topology.DNS
			}

			dnsZone, err := FindDNSHostedZone(dns, cluster.ObjectMeta.Name, dnsType)
			if err != nil {
				return fmt.Errorf("error determining default DNS zone: %v", err)
			}

			klog.V(2).Infof("Defaulting DNS zone to: %s", dnsZone)
			cluster.Spec.DNSZone = dnsZone
		}
	}

	if !cluster.UsesNoneDNS() {
		if cluster.Spec.DNSZone != "" && cluster.Spec.API.PublicName == "" {
			cluster.Spec.API.PublicName = "api." + cluster.Name
		}
		if cluster.Spec.ExternalDNS == nil {
			cluster.Spec.ExternalDNS = &kopsapi.ExternalDNSConfig{
				Provider: kopsapi.ExternalDNSProviderDNSController,
			}
		}
	}

	if cluster.Spec.KubernetesVersion == "" {
		return fmt.Errorf("KubernetesVersion is required")
	}

	optionsContext, err := components.NewOptionsContext(cluster, c.assetBuilder, c.assetBuilder.KubeletSupportedVersion)
	if err != nil {
		return err
	}

	var codeModels []loader.ClusterOptionsBuilder
	{
		{
			// Note: DefaultOptionsBuilder comes first
			codeModels = append(codeModels, &components.DefaultsOptionsBuilder{Context: optionsContext})
			codeModels = append(codeModels, &components.EtcdOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &etcdmanager.EtcdManagerOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &components.KubeAPIServerOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &components.ContainerdOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &components.NetworkingOptionsBuilder{Context: optionsContext})
			codeModels = append(codeModels, &components.KubeDnsOptionsBuilder{Context: optionsContext})
			codeModels = append(codeModels, &components.KubeletOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &components.KubeControllerManagerOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &components.KubeSchedulerOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &components.KubeProxyOptionsBuilder{Context: optionsContext})
			codeModels = append(codeModels, &components.CloudConfigurationOptionsBuilder{Context: optionsContext})
			codeModels = append(codeModels, &components.CalicoOptionsBuilder{Context: optionsContext})
			codeModels = append(codeModels, &components.CiliumOptionsBuilder{Context: optionsContext})
			codeModels = append(codeModels, &components.OpenStackOptionsBuilder{Context: optionsContext})
			codeModels = append(codeModels, &components.DiscoveryOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &components.ClusterAutoscalerOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &components.NodeTerminationHandlerOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &components.NodeProblemDetectorOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &components.AWSOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &components.AWSEBSCSIDriverOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &components.AWSCloudControllerManagerOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &components.GCPCloudControllerManagerOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &components.GCPPDCSIDriverOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &components.HetznerCloudControllerManagerOptionsBuilder{OptionsContext: optionsContext})
			codeModels = append(codeModels, &components.KarpenterOptionsBuilder{Context: optionsContext})
		}
	}

	specBuilder := &SpecBuilder{
		OptionsLoader: loader.NewClusterOptionsLoader(codeModels),
	}

	completed, err := specBuilder.BuildCompleteSpec(cluster)
	if err != nil {
		return fmt.Errorf("error building complete spec: %v", err)
	}

	// TODO: This should not be needed...
	completed.Spec.Networking.Topology = c.InputCluster.Spec.Networking.Topology
	// completed.Topology.Bastion = c.InputCluster.Spec.Topology.Bastion

	if errs := validation.ValidateCluster(completed, true, clientset.VFSContext()); len(errs) != 0 {
		return fmt.Errorf("completed cluster failed validation: %v", errs.ToAggregate())
	}

	c.fullCluster = completed
	return nil
}

func (c *populateClusterSpec) assignSubnets(cluster *kopsapi.Cluster) error {
	if cluster.Spec.Networking.NonMasqueradeCIDR == "" {
		klog.Warningf("NonMasqueradeCIDR not set; can't auto-assign dependent subnets")
		return nil
	}

	_, nonMasqueradeCIDR, err := net.ParseCIDR(cluster.Spec.Networking.NonMasqueradeCIDR)
	if err != nil {
		return fmt.Errorf("error parsing NonMasqueradeCIDR %q: %v", cluster.Spec.Networking.NonMasqueradeCIDR, err)
	}
	nmOnes, nmBits := nonMasqueradeCIDR.Mask.Size()

	if cluster.Spec.KubeControllerManager == nil {
		cluster.Spec.KubeControllerManager = &kopsapi.KubeControllerManagerConfig{}
	}

	if cluster.Spec.Networking.PodCIDR == "" && nmBits == 32 {
		// Allocate as big a range as possible: the NonMasqueradeCIDR mask + 1, with a '1' in the extra bit
		ip := nonMasqueradeCIDR.IP.Mask(nonMasqueradeCIDR.Mask)
		ip[nmOnes/8] |= 128 >> (nmOnes % 8)
		cidr := net.IPNet{IP: ip, Mask: net.CIDRMask(nmOnes+1, nmBits)}
		cluster.Spec.Networking.PodCIDR = cidr.String()
		klog.V(2).Infof("Defaulted PodCIDR to %v", cluster.Spec.Networking.PodCIDR)
	}

	if cluster.Spec.Networking.ServiceClusterIPRange == "" {
		if nmBits > 32 {
			cluster.Spec.Networking.ServiceClusterIPRange = "fd00:5e4f:ce::/108"
		} else {
			cluster.Spec.Networking.ServiceClusterIPRange = "100.64.0.0/13"
		}
		klog.V(2).Infof("Defaulted ServiceClusterIPRange to %v", cluster.Spec.Networking.ServiceClusterIPRange)
	}

	return nil
}
