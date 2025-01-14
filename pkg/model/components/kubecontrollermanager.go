/*
Copyright 2017 The Kubernetes Authors.

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
	"net"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/loader"
	"k8s.io/kops/upup/pkg/fi/utils"
)

const (
	defaultAttachDetachReconcileSyncPeriod = time.Minute
)

// KubeControllerManagerOptionsBuilder adds options for the kubernetes controller manager to the model.
type KubeControllerManagerOptionsBuilder struct {
	*OptionsContext
}

var _ loader.ClusterOptionsBuilder = &KubeControllerManagerOptionsBuilder{}

// BuildOptions generates the configurations used to create kubernetes controller manager manifest
func (b *KubeControllerManagerOptionsBuilder) BuildOptions(o *kops.Cluster) error {
	clusterSpec := &o.Spec
	if clusterSpec.KubeControllerManager == nil {
		clusterSpec.KubeControllerManager = &kops.KubeControllerManagerConfig{}
	}
	kcm := clusterSpec.KubeControllerManager

	// Tune the duration upon which the volume attach detach component is called.
	// See https://github.com/kubernetes/kubernetes/pull/39551
	// TLDR; set this too low, and have a few EBS Volumes, and you will spam AWS api

	{
		klog.V(4).Infof("Kubernetes version %q supports AttachDetachReconcileSyncPeriod; will configure", b.ControlPlaneKubernetesVersion().String())
		// If not set ... or set to 0s ... which is stupid
		if kcm.AttachDetachReconcileSyncPeriod == nil ||
			kcm.AttachDetachReconcileSyncPeriod.Duration.String() == "0s" {

			klog.V(8).Infof("AttachDetachReconcileSyncPeriod is not set; will set to default %v", defaultAttachDetachReconcileSyncPeriod)
			kcm.AttachDetachReconcileSyncPeriod = &metav1.Duration{Duration: defaultAttachDetachReconcileSyncPeriod}

			// If less than 1 min and greater than 1 sec ... you get a warning
		} else if kcm.AttachDetachReconcileSyncPeriod.Duration < defaultAttachDetachReconcileSyncPeriod &&
			kcm.AttachDetachReconcileSyncPeriod.Duration > time.Second {

			klog.Infof("KubeControllerManager AttachDetachReconcileSyncPeriod is set lower than recommended: %s", defaultAttachDetachReconcileSyncPeriod)

			// If less than 1sec you get an error.  Controller is coded to not allow configuration
			// less than one second.
		} else if kcm.AttachDetachReconcileSyncPeriod.Duration < time.Second {
			return fmt.Errorf("AttachDetachReconcileSyncPeriod cannot be set to less than 1 second")
		}
	}

	kcm.ClusterName = b.ClusterName
	kcm.CloudProvider = "external"

	if kcm.LogLevel == 0 {
		kcm.LogLevel = 2
	}

	image, err := Image("kube-controller-manager", clusterSpec, b.AssetBuilder)
	if err != nil {
		return err
	}
	kcm.Image = image

	// Doesn't seem to be any real downside to always doing a leader election
	kcm.LeaderElection = &kops.LeaderElectionConfiguration{LeaderElect: fi.PtrTo(true)}

	kcm.AllocateNodeCIDRs = fi.PtrTo(!clusterSpec.IsKopsControllerIPAM())

	if kcm.ClusterCIDR == "" && !clusterSpec.IsKopsControllerIPAM() {
		kcm.ClusterCIDR = clusterSpec.Networking.PodCIDR
	}

	if utils.IsIPv6CIDR(kcm.ClusterCIDR) {
		_, clusterNet, _ := net.ParseCIDR(kcm.ClusterCIDR)
		clusterSize, _ := clusterNet.Mask.Size()
		nodeSize := (128 - clusterSize) / 2
		if nodeSize > 16 {
			// Kubernetes limitation
			nodeSize = 16
		}
		kcm.NodeCIDRMaskSize = fi.PtrTo(int32(clusterSize + nodeSize))
	}

	networking := &clusterSpec.Networking
	if networking.Kubenet != nil {
		kcm.ConfigureCloudRoutes = fi.PtrTo(true)
	} else if gce.UsesIPAliases(o) {
		kcm.ConfigureCloudRoutes = fi.PtrTo(false)
		if kcm.CloudProvider == "external" {
			// kcm should not allocate node cidrs with the CloudAllocator if we're using the external CCM
			kcm.AllocateNodeCIDRs = fi.PtrTo(false)
		} else {
			kcm.CIDRAllocatorType = fi.PtrTo("CloudAllocator")
		}
	} else if networking.External != nil {
		kcm.ConfigureCloudRoutes = fi.PtrTo(false)
	} else if UsesCNI(networking) {
		kcm.ConfigureCloudRoutes = fi.PtrTo(false)
	} else if networking.Kopeio != nil {
		// Kopeio is based on kubenet / external
		kcm.ConfigureCloudRoutes = fi.PtrTo(false)
	} else {
		return fmt.Errorf("no networking mode set")
	}

	if kcm.UseServiceAccountCredentials == nil {
		kcm.UseServiceAccountCredentials = fi.PtrTo(true)
	}

	if len(kcm.Controllers) == 0 {
		var changes []string
		if clusterSpec.IsKopsControllerIPAM() {
			changes = append(changes, "-nodeipam")
		}
		if len(changes) != 0 {
			kcm.Controllers = append([]string{"*"}, changes...)
		}
	}

	if clusterSpec.CloudProvider.AWS != nil {

		if kcm.FeatureGates == nil {
			kcm.FeatureGates = make(map[string]string)
		}

		if _, found := kcm.FeatureGates["InTreePluginAWSUnregister"]; !found && b.ControlPlaneKubernetesVersion().IsLT("1.31") {
			kcm.FeatureGates["InTreePluginAWSUnregister"] = "true"
		}

		if _, found := kcm.FeatureGates["CSIMigrationAWS"]; !found && b.ControlPlaneKubernetesVersion().IsLT("1.27") {
			kcm.FeatureGates["CSIMigrationAWS"] = "true"
		}
	}

	return nil
}
