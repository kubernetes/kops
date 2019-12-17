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

package aliup

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func TestFindRegion(t *testing.T) {
	testZones := []string{"cn-qingdao-b", "cn-hangzhou-b", "ap-southeast-1a", "ap-southeast-2a"}
	expectedRegionds := []string{"cn-qingdao", "cn-hangzhou", "ap-southeast-1", "ap-southeast-2"}

	for i, zone := range testZones {
		c := &kops.Cluster{}
		c.Spec.Subnets = append(c.Spec.Subnets, kops.ClusterSubnetSpec{Name: "subnet-" + zone, Zone: zone})

		region, err := FindRegion(c)
		if err != nil {
			t.Fatalf("unexpected error finding region for %q: %v", zone, err)
		}

		expected := expectedRegionds[i]
		if region != expected {
			t.Fatalf("unexpected region for zone: %q vs %q", expected, region)
		}
	}

}
