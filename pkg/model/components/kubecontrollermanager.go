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

var _ loader.OptionsBuilder = &KubeControllerManagerOptionsBuilder{}

// BuildOptions generates the configurations used to create kubernetes controller manager manifest
func (b *KubeControllerManagerOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)
	if clusterSpec.KubeControllerManager == nil {
		clusterSpec.KubeControllerManager = &kops.KubeControllerManagerConfig{}
	}
	kcm := clusterSpec.KubeControllerManager

	// Tune the duration upon which the volume attach detach component is called.
	// See https://github.com/kubernetes/kubernetes/pull/39551
	// TLDR; set this too low, and have a few EBS Volumes, and you will spam AWS api

	{
		klog.V(4).Infof("Kubernetes version %q supports AttachDetachReconcileSyncPeriod; will configure", b.KubernetesVersion)
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
	if b.IsKubernetesGTE("1.24") {
		kcm.CloudProvider = "external"
	} else {
		switch kops.CloudProviderID(clusterSpec.GetCloudProvider()) {
		case kops.CloudProviderAWS:
			kcm.CloudProvider = "aws"

		case kops.CloudProviderGCE:
			kcm.CloudProvider = "gce"
			kcm.ClusterName = gce.SafeClusterName(b.ClusterName)

		case kops.CloudProviderDO:
			kcm.CloudProvider = "external"

		case kops.CloudProviderHetzner:
			kcm.CloudProvider = "external"

		case kops.CloudProviderOpenstack:
			kcm.CloudProvider = "openstack"

		case kops.CloudProviderAzure:
			kcm.CloudProvider = "azure"

		case kops.CloudProviderYandex:
			kcm.CloudProvider = "external"

		default:
			return fmt.Errorf("unknown cloudprovider %q", clusterSpec.GetCloudProvider())
		}
	}

	if clusterSpec.ExternalCloudControllerManager == nil {
		if b.IsKubernetesGTE("1.23") && (kcm.CloudProvider == "aws" || kcm.CloudProvider == "gce") {
			kcm.EnableLeaderMigration = fi.Bool(true)
		}
	} else {
		kcm.CloudProvider = "external"
	}

	if kcm.LogLevel == 0 {
		kcm.LogLevel = 2
	}

	image, err := Image("kube-controller-manager", clusterSpec, b.AssetBuilder)
	if err != nil {
		return err
	}
	kcm.Image = image

	// Doesn't seem to be any real downside to always doing a leader election
	kcm.LeaderElection = &kops.LeaderElectionConfiguration{LeaderElect: fi.Bool(true)}

	kcm.AllocateNodeCIDRs = fi.Bool(!clusterSpec.IsKopsControllerIPAM())

	if kcm.ClusterCIDR == "" && !clusterSpec.IsKopsControllerIPAM() {
		kcm.ClusterCIDR = clusterSpec.PodCIDR
	}

	if utils.IsIPv6CIDR(kcm.ClusterCIDR) {
		_, clusterNet, _ := net.ParseCIDR(kcm.ClusterCIDR)
		clusterSize, _ := clusterNet.Mask.Size()
		nodeSize := (128 - clusterSize) / 2
		if nodeSize > 16 {
			// Kubernetes limitation
			nodeSize = 16
		}
		kcm.NodeCIDRMaskSize = fi.Int32(int32(clusterSize + nodeSize))
	}

	networking := clusterSpec.Networking
	if networking == nil {
		kcm.ConfigureCloudRoutes = fi.Bool(true)
	} else if networking.Kubenet != nil {
		kcm.ConfigureCloudRoutes = fi.Bool(true)
	} else if networking.GCE != nil {
		kcm.ConfigureCloudRoutes = fi.Bool(false)
		kcm.CIDRAllocatorType = fi.String("CloudAllocator")
	} else if networking.External != nil {
		kcm.ConfigureCloudRoutes = fi.Bool(false)
	} else if UsesCNI(networking) {
		kcm.ConfigureCloudRoutes = fi.Bool(false)
	} else if networking.Kopeio != nil {
		// Kopeio is based on kubenet / external
		kcm.ConfigureCloudRoutes = fi.Bool(false)
	} else {
		return fmt.Errorf("no networking mode set")
	}

	if kcm.UseServiceAccountCredentials == nil {
		kcm.UseServiceAccountCredentials = fi.Bool(true)
	}

	if len(kcm.Controllers) == 0 {
		var changes []string
		// @check if the node authorization is enabled and if so enable the tokencleaner controller (disabled by default)
		// This is responsible for cleaning up bootstrap tokens which have expired
		if fi.BoolValue(clusterSpec.KubeAPIServer.EnableBootstrapAuthToken) {
			changes = append(changes, "tokencleaner")
		}
		if clusterSpec.IsKopsControllerIPAM() {
			changes = append(changes, "-nodeipam")
		}
		if len(changes) != 0 {
			kcm.Controllers = append([]string{"*"}, changes...)
		}
	}

	if clusterSpec.CloudConfig != nil && clusterSpec.CloudConfig.AWSEBSCSIDriver != nil && fi.BoolValue(clusterSpec.CloudConfig.AWSEBSCSIDriver.Enabled) {

		if kcm.FeatureGates == nil {
			kcm.FeatureGates = make(map[string]string)
		}

		if _, found := kcm.FeatureGates["InTreePluginAWSUnregister"]; !found {
			kcm.FeatureGates["InTreePluginAWSUnregister"] = "true"
		}

		if _, found := kcm.FeatureGates["CSIMigrationAWS"]; !found {
			kcm.FeatureGates["CSIMigrationAWS"] = "true"
		}
	}

	return nil
}
