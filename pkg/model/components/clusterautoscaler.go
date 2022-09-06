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
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// ClusterAutoscalerOptionsBuilder adds options for cluster autoscaler to the model
type ClusterAutoscalerOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &ClusterAutoscalerOptionsBuilder{}

func (b *ClusterAutoscalerOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)
	cas := clusterSpec.ClusterAutoscaler
	if cas == nil || !fi.BoolValue(cas.Enabled) {
		return nil
	}

	if cas.Image == nil {

		image := ""
		v, err := util.ParseKubernetesVersion(clusterSpec.KubernetesVersion)
		if err == nil {
			switch v.Minor {
			case 25:
				image = "registry.k8s.io/autoscaling/cluster-autoscaler:v1.25.0"
			case 24:
				image = "registry.k8s.io/autoscaling/cluster-autoscaler:v1.24.0"
			case 23:
				image = "registry.k8s.io/autoscaling/cluster-autoscaler:v1.23.1"
			case 22:
				image = "registry.k8s.io/autoscaling/cluster-autoscaler:v1.22.3"
			case 21:
				image = "registry.k8s.io/autoscaling/cluster-autoscaler:v1.21.3"
			case 20:
				image = "registry.k8s.io/autoscaling/cluster-autoscaler:v1.20.3"
			default:
				image = "registry.k8s.io/autoscaling/cluster-autoscaler:v1.25.0"
			}
		}
		cas.Image = fi.String(image)
	}

	if cas.Expander == nil {
		cas.Expander = fi.String("random")
	}
	if cas.ScaleDownUtilizationThreshold == nil {
		cas.ScaleDownUtilizationThreshold = fi.String("0.5")
	}
	if cas.SkipNodesWithLocalStorage == nil {
		cas.SkipNodesWithLocalStorage = fi.Bool(true)
	}
	if cas.SkipNodesWithSystemPods == nil {
		cas.SkipNodesWithSystemPods = fi.Bool(true)
	}
	if cas.BalanceSimilarNodeGroups == nil {
		cas.BalanceSimilarNodeGroups = fi.Bool(false)
	}
	if cas.AWSUseStaticInstanceList == nil {
		cas.AWSUseStaticInstanceList = fi.Bool(false)
	}
	if cas.NewPodScaleUpDelay == nil {
		cas.NewPodScaleUpDelay = fi.String("0s")
	}
	if cas.ScaleDownDelayAfterAdd == nil {
		cas.ScaleDownDelayAfterAdd = fi.String("10m0s")
	}
	if cas.MaxNodeProvisionTime == "" {
		cas.MaxNodeProvisionTime = "15m0s"
	}

	return nil
}
