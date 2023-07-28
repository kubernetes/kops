/*
Copyright 2022 The Kubernetes Authors.

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

package gce

import (
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"

	compute "google.golang.org/api/compute/v1"
)

func newNetworkMetricContext(request string) *metricContext {
	return newGenericMetricContext("networks", request, unusedMetricLabel, unusedMetricLabel, computeV1Version)
}

// GetNetwork returns the GCE resource for the compute.Network if it exists.
func (g *Cloud) GetNetwork(networkName string) (*compute.Network, error) {
	ctx, cancel := cloud.ContextWithCallTimeout()
	defer cancel()

	mc := newNetworkMetricContext("get")
	key := meta.GlobalKey(networkName)
	network, err := g.Compute().Networks().Get(ctx, key)
	return network, mc.Observe(err)
}
