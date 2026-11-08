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

package components

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// ClusterAutoscalerOptionsBuilder adds options for cluster autoscaler to the model
type ClusterAutoscalerOptionsBuilder struct {
	*OptionsContext
}

var _ loader.ClusterOptionsBuilder = &ClusterAutoscalerOptionsBuilder{}

func (b *ClusterAutoscalerOptionsBuilder) BuildOptions(o *kops.Cluster) error {
	clusterSpec := &o.Spec
	cas := clusterSpec.ClusterAutoscaler
	if cas == nil || !fi.ValueOf(cas.Enabled) {
		return nil
	}

	if cas.Image == nil {
		// cluster-autoscaler v1.35+ always watches DRA objects through resource.k8s.io/v1,
		// which the API server serves only from Kubernetes 1.34. On older control planes
		// those informers never sync and the autoscaler never starts its main loop.
		// v1.34 handles DRA only when --enable-dynamic-resource-allocation is set, which
		// kOps does not set.
		if b.ControlPlaneKubernetesVersion().IsLT("1.34.0") {
			cas.Image = new("registry.k8s.io/autoscaling/cluster-autoscaler:v1.34.6")
		} else {
			cas.Image = new("registry.k8s.io/autoscaling/cluster-autoscaler:v1.37.0")
		}
	}

	if cas.Expander == "" {
		cas.Expander = "random"
	}
	if cas.IgnoreDaemonSetsUtilization == nil {
		cas.IgnoreDaemonSetsUtilization = new(false)
	}
	if cas.ScaleDownUtilizationThreshold == nil {
		cas.ScaleDownUtilizationThreshold = new("0.5")
	}
	if cas.SkipNodesWithCustomControllerPods == nil {
		cas.SkipNodesWithCustomControllerPods = new(true)
	}
	if cas.SkipNodesWithLocalStorage == nil {
		cas.SkipNodesWithLocalStorage = new(true)
	}
	if cas.SkipNodesWithSystemPods == nil {
		cas.SkipNodesWithSystemPods = new(true)
	}
	if cas.BalanceSimilarNodeGroups == nil {
		cas.BalanceSimilarNodeGroups = new(false)
	}
	if cas.EmitPerNodegroupMetrics == nil {
		cas.EmitPerNodegroupMetrics = new(false)
	}
	if cas.AWSUseStaticInstanceList == nil {
		cas.AWSUseStaticInstanceList = new(false)
	}
	if cas.NewPodScaleUpDelay == nil {
		cas.NewPodScaleUpDelay = new("0s")
	}
	if cas.ScaleDownDelayAfterAdd == nil {
		cas.ScaleDownDelayAfterAdd = new("10m0s")
	}
	if cas.ScaleDownUnneededTime == nil {
		cas.ScaleDownUnneededTime = new("10m0s")
	}
	if cas.ScaleDownUnreadyTime == nil {
		cas.ScaleDownUnreadyTime = new("20m0s")
	}
	if cas.MaxNodeProvisionTime == "" {
		cas.MaxNodeProvisionTime = "15m0s"
	}
	if cas.Expander == "priority" {
		cas.CreatePriorityExpenderConfig = new(true)
	}

	return nil
}
