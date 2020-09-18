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

// NodeTerminationHandlerOptionsBuilder adds options for the node termination handler to the model.
type NodeTerminationHandlerOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &NodeTerminationHandlerOptionsBuilder{}

func (b *NodeTerminationHandlerOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)
	if clusterSpec.NodeTerminationHandler == nil {
		return nil
	}
	nth := clusterSpec.NodeTerminationHandler
	if nth.Enabled == nil {
		nth.Enabled = fi.Bool(true)
	}
	if nth.EnableSpotInterruptionDraining == nil {
		nth.EnableSpotInterruptionDraining = fi.Bool(true)
	}
	if nth.EnableScheduledEventDraining == nil {
		nth.EnableScheduledEventDraining = fi.Bool(false)
	}

	if nth.EnablePrometheusMetrics == nil {
		nth.EnablePrometheusMetrics = fi.Bool(false)
	}
	return nil
}
