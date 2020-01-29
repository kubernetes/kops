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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/loader"
)

const (
	defaultAttachDetachReconcileSyncPeriod = time.Minute
)

// KubeControllerManagerOptionsBuilder adds options for the kubernetes controller manager to the model.
type KubeControllerManagerOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &KubeControllerManagerOptionsBuilder{}

// BuildOptions generates the configurations used to create kubernetes controller manager manifest
func (b *KubeControllerManagerOptionsBuilder) BuildOptions(o interface{}) error {

	clusterSpec := o.(*kops.ClusterSpec)
	if clusterSpec.KubeControllerManager == nil {
		clusterSpec.KubeControllerManager = &kops.KubeControllerManagerConfig{}
	}
	kcm := clusterSpec.KubeControllerManager

	kubernetesVersion, err := KubernetesVersion(clusterSpec)
	if err != nil {
		return fmt.Errorf("unable to parse kubernetesVersion %s", err)
	}

	// Tune the duration upon which the volume attach detach component is called.
	// See https://github.com/kubernetes/kubernetes/pull/39551
	// TLDR; set this too low, and have a few EBS Volumes, and you will spam AWS api

	{
		klog.V(4).Infof("Kubernetes version %q supports AttachDetachReconcileSyncPeriod; will configure", kubernetesVersion)
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

	kcm.ClusterName = b.Context.ClusterName
	switch kops.CloudProviderID(clusterSpec.CloudProvider) {
	case kops.CloudProviderAWS:
		kcm.CloudProvider = "aws"

	case kops.CloudProviderGCE:
		kcm.CloudProvider = "gce"
		kcm.ClusterName = gce.SafeClusterName(b.Context.ClusterName)

	case kops.CloudProviderDO:
		kcm.CloudProvider = "external"

	case kops.CloudProviderVSphere:
		kcm.CloudProvider = "vsphere"

	case kops.CloudProviderBareMetal:
		// No cloudprovider

	case kops.CloudProviderOpenstack:
		kcm.CloudProvider = "openstack"

	case kops.CloudProviderALI:
		kcm.CloudProvider = "alicloud"

	default:
		return fmt.Errorf("unknown cloudprovider %q", clusterSpec.CloudProvider)
	}

	if featureflag.EnableExternalCloudController.Enabled() && clusterSpec.ExternalCloudControllerManager != nil {
		kcm.CloudProvider = "external"
	}

	kcm.LogLevel = 2

	image, err := Image("kube-controller-manager", b.Context.Architecture(), clusterSpec, b.Context.AssetBuilder)
	if err != nil {
		return err
	}
	kcm.Image = image

	// Doesn't seem to be any real downside to always doing a leader election
	kcm.LeaderElection = &kops.LeaderElectionConfiguration{LeaderElect: fi.Bool(true)}

	kcm.AllocateNodeCIDRs = fi.Bool(true)
	kcm.ConfigureCloudRoutes = fi.Bool(false)

	networking := clusterSpec.Networking
	if networking == nil || networking.Classic != nil {
		kcm.ConfigureCloudRoutes = fi.Bool(true)
	} else if networking.Kubenet != nil {
		kcm.ConfigureCloudRoutes = fi.Bool(true)
	} else if networking.GCE != nil {
		kcm.ConfigureCloudRoutes = fi.Bool(false)
		kcm.CIDRAllocatorType = fi.String("CloudAllocator")

		if kcm.ClusterCIDR == "" {
			kcm.ClusterCIDR = clusterSpec.PodCIDR
		}
	} else if networking.External != nil {
		kcm.ConfigureCloudRoutes = fi.Bool(false)
	} else if networking.CNI != nil || networking.Weave != nil || networking.Flannel != nil || networking.Calico != nil || networking.Canal != nil || networking.Kuberouter != nil || networking.Romana != nil || networking.AmazonVPC != nil || networking.Cilium != nil || networking.LyftVPC != nil {
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

	// @check if the node authorization is enabled and if so enable the tokencleaner controller (disabled by default)
	// This is responsible for cleaning up bootstrap tokens which have expired
	if b.Context.IsKubernetesGTE("1.10") {
		if fi.BoolValue(clusterSpec.KubeAPIServer.EnableBootstrapAuthToken) && len(kcm.Controllers) <= 0 {
			kcm.Controllers = []string{"*", "tokencleaner"}
		}
	}

	return nil
}
