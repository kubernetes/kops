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

package components

import (
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// KubeletOptionsBuilder adds options for kubelets
type KubeletOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &KubeletOptionsBuilder{}

// BuildOptions is responsible for filling the defaults for the kubelet
func (b *KubeletOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)

	kubernetesVersion, err := KubernetesVersion(clusterSpec)
	if err != nil {
		return err
	}

	if clusterSpec.Kubelet == nil {
		clusterSpec.Kubelet = &kops.KubeletConfigSpec{}
	}
	if clusterSpec.MasterKubelet == nil {
		clusterSpec.MasterKubelet = &kops.KubeletConfigSpec{}
	}

	if clusterSpec.KubeAPIServer != nil && clusterSpec.KubeAPIServer.EnableBootstrapAuthToken != nil {
		if *clusterSpec.KubeAPIServer.EnableBootstrapAuthToken {
			if clusterSpec.Kubelet.BootstrapKubeconfig == "" {
				clusterSpec.Kubelet.BootstrapKubeconfig = "/var/lib/kubelet/bootstrap-kubeconfig"
			}
		}
	}

	// Standard options
	clusterSpec.Kubelet.EnableDebuggingHandlers = fi.Bool(true)
	clusterSpec.Kubelet.PodManifestPath = "/etc/kubernetes/manifests"
	clusterSpec.Kubelet.LogLevel = fi.Int32(2)
	clusterSpec.Kubelet.ClusterDomain = clusterSpec.ClusterDNSDomain
	clusterSpec.Kubelet.NonMasqueradeCIDR = clusterSpec.NonMasqueradeCIDR

	// AllowPrivileged is deprecated and removed in v1.14.
	// See https://github.com/kubernetes/kubernetes/pull/71835
	if kubernetesVersion.Major == 1 && kubernetesVersion.Minor >= 14 {
		if clusterSpec.Kubelet.AllowPrivileged != nil {
			// If it is explicitly set to false, return an error, because this
			// behavior is no longer supported in v1.14 (the default was true, prior).
			if !*clusterSpec.Kubelet.AllowPrivileged {
				klog.Warningf("Kubelet's --allow-privileged flag is no longer supported in v1.14.")
			}
			// Explicitly set it to nil, so it won't be passed on the command line.
			clusterSpec.Kubelet.AllowPrivileged = nil
		}
	} else {
		clusterSpec.Kubelet.AllowPrivileged = fi.Bool(true)
	}

	if clusterSpec.Kubelet.ClusterDNS == "" {
		ip, err := WellKnownServiceIP(clusterSpec, 10)
		if err != nil {
			return err
		}
		clusterSpec.Kubelet.ClusterDNS = ip.String()
	}

	if b.Context.IsKubernetesLT("1.7") {
		// babysit-daemons removed in 1.7
		clusterSpec.Kubelet.BabysitDaemons = fi.Bool(true)
	}

	clusterSpec.MasterKubelet.RegisterSchedulable = fi.Bool(false)
	// Replace the CIDR with a CIDR allocated by KCM (the default, but included for clarity)
	// We _do_ allow debugging handlers, so we can do logs
	// This does allow more access than we would like though
	clusterSpec.MasterKubelet.EnableDebuggingHandlers = fi.Bool(true)

	// In 1.5 we fixed this, but in 1.4 we need to set the PodCIDR on the master
	// so that hostNetwork pods can come up
	if kubernetesVersion.Major == 1 && kubernetesVersion.Minor <= 4 {
		// We bootstrap with a fake CIDR, but then this will be replaced (unless we're running with _isolated_master)
		clusterSpec.MasterKubelet.PodCIDR = "10.123.45.0/28"
	}

	// 1.5 deprecates the reconcile cidr option (and 1.6 removes it)
	if kubernetesVersion.Major == 1 && kubernetesVersion.Minor <= 4 {
		clusterSpec.MasterKubelet.ReconcileCIDR = fi.Bool(true)

		if fi.BoolValue(clusterSpec.IsolateMasters) {
			clusterSpec.MasterKubelet.ReconcileCIDR = fi.Bool(false)
		}

		usesKubenet, err := UsesKubenet(clusterSpec)
		if err != nil {
			return err
		}
		if usesKubenet {
			clusterSpec.Kubelet.ReconcileCIDR = fi.Bool(true)
		}
	}

	if kubernetesVersion.Major == 1 && kubernetesVersion.Minor >= 4 {
		// For pod eviction in low memory or empty disk situations
		if clusterSpec.Kubelet.EvictionHard == nil {
			evictionHard := []string{
				// TODO: Some people recommend 250Mi, but this would hurt small machines
				"memory.available<100Mi",

				// Disk eviction (evict old images)
				// We don't need to specify both, but it seems harmless / safer
				"nodefs.available<10%",
				"nodefs.inodesFree<5%",
				"imagefs.available<10%",
				"imagefs.inodesFree<5%",
			}
			clusterSpec.Kubelet.EvictionHard = fi.String(strings.Join(evictionHard, ","))
		}
	}

	if b.Context.IsKubernetesGTE("1.6") {
		// for 1.6+ use kubeconfig instead of api-servers
		const kubeconfigPath = "/var/lib/kubelet/kubeconfig"
		clusterSpec.Kubelet.KubeconfigPath = kubeconfigPath
		clusterSpec.MasterKubelet.KubeconfigPath = kubeconfigPath

		// Only pass require-kubeconfig to versions prior to 1.9; deprecated & being removed
		if b.Context.IsKubernetesLT("1.9") {
			clusterSpec.Kubelet.RequireKubeconfig = fi.Bool(true)
			clusterSpec.MasterKubelet.RequireKubeconfig = fi.Bool(true)
		}
	} else {
		// Legacy behaviour for <= 1.5
		clusterSpec.Kubelet.APIServers = "https://" + clusterSpec.MasterInternalName
		clusterSpec.MasterKubelet.APIServers = "http://127.0.0.1:8080"
	}

	// IsolateMasters enables the legacy behaviour, where master pods on a separate network
	// In newer versions of kubernetes, most of that functionality has been removed though
	if fi.BoolValue(clusterSpec.IsolateMasters) {
		clusterSpec.MasterKubelet.EnableDebuggingHandlers = fi.Bool(false)
		clusterSpec.MasterKubelet.HairpinMode = "none"
	}

	cloudProvider := kops.CloudProviderID(clusterSpec.CloudProvider)

	clusterSpec.Kubelet.CgroupRoot = "/"

	klog.V(1).Infof("Cloud Provider: %s", cloudProvider)
	if cloudProvider == kops.CloudProviderAWS {
		clusterSpec.Kubelet.CloudProvider = "aws"

		// For 1.6 we're using much cleaner cgroup hierarchies
		// but we keep the settings we've tested for k8s 1.5 and lower
		// (see https://github.com/kubernetes/kubernetes/pull/41349)
		if kubernetesVersion.Major == 1 && kubernetesVersion.Minor <= 5 {
			clusterSpec.Kubelet.CgroupRoot = "docker"
		}

		// Use the hostname from the AWS metadata service
		// if hostnameOverride is not set.
		if clusterSpec.Kubelet.HostnameOverride == "" {
			clusterSpec.Kubelet.HostnameOverride = "@aws"
		}
	}

	if cloudProvider == kops.CloudProviderDO {
		clusterSpec.Kubelet.CloudProvider = "external"
		clusterSpec.Kubelet.HostnameOverride = "@digitalocean"
	}

	if cloudProvider == kops.CloudProviderGCE {
		clusterSpec.Kubelet.CloudProvider = "gce"
		clusterSpec.Kubelet.HairpinMode = "promiscuous-bridge"

		if clusterSpec.CloudConfig == nil {
			clusterSpec.CloudConfig = &kops.CloudConfiguration{}
		}
		clusterSpec.CloudConfig.Multizone = fi.Bool(true)
		clusterSpec.CloudConfig.NodeTags = fi.String(GCETagForRole(b.Context.ClusterName, kops.InstanceGroupRoleNode))
	}

	if cloudProvider == kops.CloudProviderVSphere {
		clusterSpec.Kubelet.CloudProvider = "vsphere"
		clusterSpec.Kubelet.HairpinMode = "promiscuous-bridge"
	}

	if cloudProvider == kops.CloudProviderOpenstack {
		clusterSpec.Kubelet.CloudProvider = "openstack"
	}

	if cloudProvider == kops.CloudProviderALI {
		clusterSpec.Kubelet.CloudProvider = "external"
		clusterSpec.Kubelet.HostnameOverride = "@alicloud"
		clusterSpec.Kubelet.ProviderID = "@alicloud"
	}

	if clusterSpec.ExternalCloudControllerManager != nil {
		clusterSpec.Kubelet.CloudProvider = "external"
	}

	usesKubenet, err := UsesKubenet(clusterSpec)
	if err != nil {
		return err
	}
	if usesKubenet {
		clusterSpec.Kubelet.NetworkPluginName = "kubenet"

		if kubernetesVersion.Major == 1 && kubernetesVersion.Minor >= 4 {
			// AWS MTU is 9001
			clusterSpec.Kubelet.NetworkPluginMTU = fi.Int32(9001)
		}
	}

	// Specify our pause image
	image := "k8s.gcr.io/pause-amd64:3.0"
	if image, err = b.Context.AssetBuilder.RemapImage(image); err != nil {
		return err
	}
	clusterSpec.Kubelet.PodInfraContainerImage = image

	if clusterSpec.Kubelet.FeatureGates == nil {
		clusterSpec.Kubelet.FeatureGates = make(map[string]string)
	}
	if _, found := clusterSpec.Kubelet.FeatureGates["ExperimentalCriticalPodAnnotation"]; !found {
		if b.Context.IsKubernetesGTE("1.5.2") && b.Context.IsKubernetesLT("1.16") {
			clusterSpec.Kubelet.FeatureGates["ExperimentalCriticalPodAnnotation"] = "true"
		}
	}

	return nil
}
