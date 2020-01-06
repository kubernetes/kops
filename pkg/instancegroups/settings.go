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

package instancegroups

import (
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kops/pkg/apis/kops"
)

func resolveSettings(cluster *kops.Cluster, group *kops.InstanceGroup, numInstances int) kops.RollingUpdate {
	rollingUpdate := kops.RollingUpdate{}
	if group.Spec.RollingUpdate != nil {
		rollingUpdate = *group.Spec.RollingUpdate
	}

	if def := cluster.Spec.RollingUpdate; def != nil {
		if rollingUpdate.MaxUnavailable == nil {
			rollingUpdate.MaxUnavailable = def.MaxUnavailable
		}
		if rollingUpdate.MaxSurge == nil {
			rollingUpdate.MaxSurge = def.MaxSurge
		}
	}

	if rollingUpdate.MaxSurge == nil {
		zero := intstr.FromInt(0)
		rollingUpdate.MaxSurge = &zero
	}

	if rollingUpdate.MaxSurge.Type == intstr.String {
		surge, _ := intstr.GetValueFromIntOrPercent(rollingUpdate.MaxSurge, numInstances, true)
		surgeInt := intstr.FromInt(surge)
		rollingUpdate.MaxSurge = &surgeInt
	}

	maxUnavailableDefault := intstr.FromInt(0)
	if rollingUpdate.MaxSurge.Type == intstr.Int && rollingUpdate.MaxSurge.IntVal == 0 {
		maxUnavailableDefault = intstr.FromInt(1)
	}
	if rollingUpdate.MaxUnavailable == nil || rollingUpdate.MaxUnavailable.IntVal < 0 {
		rollingUpdate.MaxUnavailable = &maxUnavailableDefault
	}

	if rollingUpdate.MaxUnavailable.Type == intstr.String {
		unavailable, err := intstr.GetValueFromIntOrPercent(rollingUpdate.MaxUnavailable, numInstances, false)
		if err != nil {
			// If unparseable use the default value
			unavailable = maxUnavailableDefault.IntValue()
		}
		if unavailable <= 0 {
			// While we round down, percentages should resolve to a minimum of 1
			unavailable = 1
		}
		unavailableInt := intstr.FromInt(unavailable)
		rollingUpdate.MaxUnavailable = &unavailableInt
	}

	return rollingUpdate
}
