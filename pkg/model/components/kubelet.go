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
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// KubeletOptionsBuilder adds options for kubelets
type KubeletOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &KubeletOptionsBuilder{}

// BuildOptions is responsible for filling the defaults for the kubelet
func (b *KubeletOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)

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

	// AllowPrivileged is deprecated and removed in v1.14.
	// See https://github.com/kubernetes/kubernetes/pull/71835
	if clusterSpec.Kubelet.AllowPrivileged != nil {
		// If it is explicitly set to false, return an error, because this
		// behavior is no longer supported in v1.14 (the default was true, prior).
		if !*clusterSpec.Kubelet.AllowPrivileged {
			klog.Warningf("Kubelet's --allow-privileged flag is no longer supported in v1.14.")
		}
		// Explicitly set it to nil, so it won't be passed on the command line.
		clusterSpec.Kubelet.AllowPrivileged = nil
	}

	if clusterSpec.Kubelet.ClusterDNS == "" {
		if clusterSpec.KubeDNS != nil && clusterSpec.KubeDNS.NodeLocalDNS != nil && fi.BoolValue(clusterSpec.KubeDNS.NodeLocalDNS.Enabled) {
			clusterSpec.Kubelet.ClusterDNS = clusterSpec.KubeDNS.NodeLocalDNS.LocalIP
		} else {
			ip, err := WellKnownServiceIP(clusterSpec, 10)
			if err != nil {
				return err
			}
			clusterSpec.Kubelet.ClusterDNS = ip.String()
		}
	}

	clusterSpec.MasterKubelet.RegisterSchedulable = fi.Bool(false)
	// Replace the CIDR with a CIDR allocated by KCM (the default, but included for clarity)
	// We _do_ allow debugging handlers, so we can do logs
	// This does allow more access than we would like though
	clusterSpec.MasterKubelet.EnableDebuggingHandlers = fi.Bool(true)

	{
		// For pod eviction in low memory or empty disk situations
		if clusterSpec.Kubelet.EvictionHard == nil {
			evictionHard := []string{
				// TODO: Some people recommend 250Mi, but this would hurt small machines
				"memory.available<100Mi",

				// Disk based eviction (evict old images)
				// We don't need to specify both, but it seems harmless / safer
				"nodefs.available<10%",
				"nodefs.inodesFree<5%",
				"imagefs.available<10%",
				"imagefs.inodesFree<5%",
			}
			clusterSpec.Kubelet.EvictionHard = fi.String(strings.Join(evictionHard, ","))
		}
	}

	// use kubeconfig instead of api-servers
	const kubeconfigPath = "/var/lib/kubelet/kubeconfig"
	clusterSpec.Kubelet.KubeconfigPath = kubeconfigPath
	clusterSpec.MasterKubelet.KubeconfigPath = kubeconfigPath

	// IsolateMasters enables the legacy behaviour, where master pods on a separate network
	// In newer versions of kubernetes, most of that functionality has been removed though
	if fi.BoolValue(clusterSpec.IsolateMasters) {
		clusterSpec.MasterKubelet.EnableDebuggingHandlers = fi.Bool(false)
		clusterSpec.MasterKubelet.HairpinMode = "none"
	}

	cloudProvider := clusterSpec.GetCloudProvider()

	clusterSpec.Kubelet.CgroupRoot = "/"

	klog.V(1).Infof("Cloud Provider: %s", cloudProvider)
	if cloudProvider == kops.CloudProviderAWS {
		clusterSpec.Kubelet.CloudProvider = "aws"
	}

	if cloudProvider == kops.CloudProviderDO {
		clusterSpec.Kubelet.CloudProvider = "external"
	}

	if cloudProvider == kops.CloudProviderGCE {
		clusterSpec.Kubelet.CloudProvider = "gce"
		clusterSpec.Kubelet.HairpinMode = "promiscuous-bridge"

		if clusterSpec.CloudConfig == nil {
			clusterSpec.CloudConfig = &kops.CloudConfiguration{}
		}
		clusterSpec.CloudConfig.Multizone = fi.Bool(true)
		clusterSpec.CloudConfig.NodeTags = fi.String(gce.TagForRole(b.ClusterName, kops.InstanceGroupRoleNode))

	}

	if cloudProvider == kops.CloudProviderHetzner {
		clusterSpec.Kubelet.CloudProvider = "external"
	}

	if cloudProvider == kops.CloudProviderOpenstack {
		clusterSpec.Kubelet.CloudProvider = "openstack"
	}

	if cloudProvider == kops.CloudProviderAzure {
		clusterSpec.Kubelet.CloudProvider = "azure"
	}

	if cloudProvider == kops.CloudProviderYandex {
		clusterSpec.Kubelet.CloudProvider = "external"
	}

	if clusterSpec.ExternalCloudControllerManager != nil {
		clusterSpec.Kubelet.CloudProvider = "external"
	}

	if clusterSpec.ContainerRuntime == "docker" || clusterSpec.ContainerRuntime == "" {
		networking := clusterSpec.Networking
		if networking == nil {
			return fmt.Errorf("no networking mode set")
		}
		if UsesKubenet(networking) && b.IsKubernetesLT("1.24") {
			clusterSpec.Kubelet.NetworkPluginName = fi.String("kubenet")
			clusterSpec.Kubelet.NetworkPluginMTU = fi.Int32(9001)
			clusterSpec.Kubelet.NonMasqueradeCIDR = fi.String(clusterSpec.NonMasqueradeCIDR)
		}
	}

	// Prevent image GC from pruning the pause image
	// https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/2040-kubelet-cri#pinned-images
	image := "registry.k8s.io/pause:3.6"
	var err error
	if image, err = b.AssetBuilder.RemapImage(image); err != nil {
		return err
	}
	clusterSpec.Kubelet.PodInfraContainerImage = image

	if clusterSpec.Kubelet.FeatureGates == nil {
		clusterSpec.Kubelet.FeatureGates = make(map[string]string)
	}

	if clusterSpec.CloudConfig != nil && clusterSpec.CloudConfig.AWSEBSCSIDriver != nil && fi.BoolValue(clusterSpec.CloudConfig.AWSEBSCSIDriver.Enabled) {
		if _, found := clusterSpec.Kubelet.FeatureGates["CSIMigrationAWS"]; !found {
			clusterSpec.Kubelet.FeatureGates["CSIMigrationAWS"] = "true"
		}

		if _, found := clusterSpec.Kubelet.FeatureGates["InTreePluginAWSUnregister"]; !found {
			clusterSpec.Kubelet.FeatureGates["InTreePluginAWSUnregister"] = "true"
		}
	}

	// Set systemd as the default cgroup driver for kubelet
	if clusterSpec.Kubelet.CgroupDriver == "" {
		clusterSpec.Kubelet.CgroupDriver = "systemd"
	}

	if b.IsKubernetesGTE("1.22") && clusterSpec.Kubelet.ProtectKernelDefaults == nil {
		clusterSpec.Kubelet.ProtectKernelDefaults = fi.Bool(true)
	}

	// We do not enable graceful shutdown when using amazonaws due to leaking ENIs.
	// Graceful shutdown is also not available by default on k8s < 1.21
	if clusterSpec.Kubelet.ShutdownGracePeriod == nil && clusterSpec.Networking.AmazonVPC == nil {
		clusterSpec.Kubelet.ShutdownGracePeriod = &metav1.Duration{Duration: time.Duration(30 * time.Second)}
		clusterSpec.Kubelet.ShutdownGracePeriodCriticalPods = &metav1.Duration{Duration: time.Duration(10 * time.Second)}
	} else if clusterSpec.Networking.AmazonVPC != nil {
		clusterSpec.Kubelet.ShutdownGracePeriod = &metav1.Duration{Duration: 0}
		clusterSpec.Kubelet.ShutdownGracePeriodCriticalPods = &metav1.Duration{Duration: 0}
	}

	return nil
}
