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

func newSubnetworkMetricContext(request, region string) *metricContext {
	return newGenericMetricContext("subnetworks", request, region, unusedMetricLabel, computeV1Version)
}

// GetSubnetwork returns the GCE resource for the compute.Subnetwork if it exists.
func (g *Cloud) GetSubnetwork(region, subnetworkName string) (*compute.Subnetwork, error) {
	ctx, cancel := cloud.ContextWithCallTimeout()
	defer cancel()

	mc := newSubnetworkMetricContext("get", region)
	key := meta.RegionalKey(subnetworkName, region)
	subnetwork, err := g.Compute().Subnetworks().Get(ctx, key)
	return subnetwork, mc.Observe(err)
}
