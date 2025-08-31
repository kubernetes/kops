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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// KubeletOptionsBuilder adds options for kubelets
type KubeletOptionsBuilder struct {
	*OptionsContext
}

var _ loader.ClusterOptionsBuilder = &KubeletOptionsBuilder{}

// BuildOptions is responsible for filling the defaults for the kubelet
func (b *KubeletOptionsBuilder) BuildOptions(cluster *kops.Cluster) error {
	if cluster.Spec.Kubelet == nil {
		cluster.Spec.Kubelet = &kops.KubeletConfigSpec{}
	}
	if cluster.Spec.ControlPlaneKubelet == nil {
		cluster.Spec.ControlPlaneKubelet = &kops.KubeletConfigSpec{}
	}

	if err := b.configureKubelet(cluster, cluster.Spec.Kubelet, b.NodeKubernetesVersion()); err != nil {
		return err
	}
	if err := b.configureKubelet(cluster, cluster.Spec.ControlPlaneKubelet, b.ControlPlaneKubernetesVersion()); err != nil {
		return err
	}

	// We _do_ allow debugging handlers, so we can do logs
	// This does allow more access than we would like though
	cluster.Spec.ControlPlaneKubelet.EnableDebuggingHandlers = fi.PtrTo(true)

	// IsolateControlPlane enables the legacy behaviour, where master pods on a separate network
	// In newer versions of kubernetes, most of that functionality has been removed though
	if fi.ValueOf(cluster.Spec.Networking.IsolateControlPlane) {
		cluster.Spec.ControlPlaneKubelet.EnableDebuggingHandlers = fi.PtrTo(false)
		cluster.Spec.ControlPlaneKubelet.HairpinMode = "none"
	}

	return nil
}

func (b *KubeletOptionsBuilder) configureKubelet(cluster *kops.Cluster, kubelet *kops.KubeletConfigSpec, kubernetesVersion model.KubernetesVersion) error {
	// Standard options
	kubelet.EnableDebuggingHandlers = fi.PtrTo(true)
	kubelet.PodManifestPath = "/etc/kubernetes/manifests"
	kubelet.LogLevel = fi.PtrTo(int32(2))
	kubelet.ClusterDomain = cluster.Spec.ClusterDNSDomain

	// AllowPrivileged is deprecated and removed in v1.14.
	// See https://github.com/kubernetes/kubernetes/pull/71835
	if kubelet.AllowPrivileged != nil {
		// If it is explicitly set to false, return an error, because this
		// behavior is no longer supported in v1.14 (the default was true, prior).
		if !*kubelet.AllowPrivileged {
			klog.Warningf("Kubelet's --allow-privileged flag is no longer supported in v1.14.")
		}
		// Explicitly set it to nil, so it won't be passed on the command line.
		kubelet.AllowPrivileged = nil
	}

	if kubelet.ClusterDNS == "" {
		if cluster.Spec.KubeDNS != nil && cluster.Spec.KubeDNS.NodeLocalDNS != nil && fi.ValueOf(cluster.Spec.KubeDNS.NodeLocalDNS.Enabled) {
			kubelet.ClusterDNS = cluster.Spec.KubeDNS.NodeLocalDNS.LocalIP
		} else {
			ip, err := WellKnownServiceIP(&cluster.Spec.Networking, 10)
			if err != nil {
				return err
			}
			kubelet.ClusterDNS = ip.String()
		}
	}

	{
		// For pod eviction in low memory or empty disk situations
		if kubelet.EvictionHard == nil {
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
			kubelet.EvictionHard = fi.PtrTo(strings.Join(evictionHard, ","))
		}
	}

	// use kubeconfig instead of api-servers
	const kubeconfigPath = "/var/lib/kubelet/kubeconfig"
	kubelet.KubeconfigPath = kubeconfigPath

	kubelet.CgroupRoot = "/"

	cloudProvider := cluster.GetCloudProvider()
	klog.V(1).Infof("Cloud Provider: %s", cloudProvider)
	if b.controlPlaneKubernetesVersion.IsLT("1.31") {
		switch cloudProvider {
		case kops.CloudProviderAWS:
			kubelet.CloudProvider = "aws"
		case kops.CloudProviderGCE:
			kubelet.CloudProvider = "gce"
		case kops.CloudProviderDO:
			kubelet.CloudProvider = "external"
		case kops.CloudProviderHetzner:
			kubelet.CloudProvider = "external"
		case kops.CloudProviderOpenstack:
			kubelet.CloudProvider = "openstack"
		case kops.CloudProviderAzure:
			kubelet.CloudProvider = "azure"
		case kops.CloudProviderScaleway:
			kubelet.CloudProvider = "external"
		case kops.CloudProviderMetal:
			kubelet.CloudProvider = ""
		default:
			kubelet.CloudProvider = "external"
		}

		if cluster.Spec.ExternalCloudControllerManager != nil {
			kubelet.CloudProvider = "external"
		}
	} else {
		if cloudProvider == kops.CloudProviderMetal {
			// metal does not (yet) have a cloud-controller-manager, so we don't need to set the cloud-provider flag
			// If we do set it to external, kubelet will taint the node with the node.kops.k8s.io/uninitialized taint
			// and there is no cloud-controller-manager to remove it
			kubelet.CloudProvider = ""
		} else {
			kubelet.CloudProvider = "external"
		}
	}

	if cloudProvider == kops.CloudProviderGCE {
		kubelet.HairpinMode = "promiscuous-bridge"

		if cluster.Spec.CloudConfig == nil {
			cluster.Spec.CloudConfig = &kops.CloudConfiguration{}
		}
		cluster.Spec.CloudProvider.GCE.Multizone = fi.PtrTo(true)
		cluster.Spec.CloudProvider.GCE.NodeTags = fi.PtrTo(gce.TagForRole(b.ClusterName, kops.InstanceGroupRoleNode))
	}

	// Prevent image GC from pruning the pause image
	// https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/2040-kubelet-cri#pinned-images
	image := "registry.k8s.io/pause:3.9"
	kubelet.PodInfraContainerImage = b.AssetBuilder.RemapImage(image)

	if kubelet.FeatureGates == nil {
		kubelet.FeatureGates = make(map[string]string)
	}

	if cluster.Spec.CloudProvider.AWS != nil {
		if _, found := kubelet.FeatureGates["InTreePluginAWSUnregister"]; !found && kubernetesVersion.IsLT("1.31") {
			kubelet.FeatureGates["InTreePluginAWSUnregister"] = "true"
		}
	}

	// Set systemd as the default cgroup driver for kubelet
	// In Kubernetes 1.34, with the KubeletCgroupDriverFromCRI feature gate enabled and a container runtime
	// that supports the RuntimeConfig CRI RPC, the kubelet automatically detects the appropriate cgroup driver
	// from the runtime, and ignores the cgroupDriver setting within the kubelet configuration.
	kubelet.CgroupDriver = "systemd"

	if kubelet.ProtectKernelDefaults == nil {
		kubelet.ProtectKernelDefaults = fi.PtrTo(true)
	}

	// We do not enable graceful shutdown when using amazonaws due to leaking ENIs.
	// Graceful shutdown is also not available by default on k8s < 1.21
	if kubelet.ShutdownGracePeriod == nil && cluster.Spec.Networking.AmazonVPC == nil {
		kubelet.ShutdownGracePeriod = &metav1.Duration{Duration: time.Duration(30 * time.Second)}
		kubelet.ShutdownGracePeriodCriticalPods = &metav1.Duration{Duration: time.Duration(10 * time.Second)}
	} else if cluster.Spec.Networking.AmazonVPC != nil {
		kubelet.ShutdownGracePeriod = &metav1.Duration{Duration: 0}
		kubelet.ShutdownGracePeriodCriticalPods = &metav1.Duration{Duration: 0}
	}

	if kubernetesVersion.IsLT("1.34") {
		kubelet.RegisterSchedulable = fi.PtrTo(true)
	}

	return nil
}
