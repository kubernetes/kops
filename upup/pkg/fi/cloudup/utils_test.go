/*
Copyright 2026 The Kubernetes Authors.

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

package cloudup

import (
	"strings"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func newLinodeTestCluster(region string) *kops.Cluster {
	return &kops.Cluster{
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{
				Linode: &kops.LinodeSpec{},
			},
			Networking: kops.NetworkingSpec{
				Subnets: []kops.ClusterSubnetSpec{
					{Region: region},
				},
			},
		},
	}
}

func TestBuildCloudLinodeRequiresSubnetRegion(t *testing.T) {
	t.Setenv("LINODE_TOKEN", "test-token")

	_, err := BuildCloud(newLinodeTestCluster(""))
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "subnets must include Regions") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildCloudLinodeRequiresToken(t *testing.T) {
	t.Setenv("LINODE_TOKEN", "")

	_, err := BuildCloud(newLinodeTestCluster("us-east"))
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "LINODE_TOKEN is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildCloudLinodeSuccess(t *testing.T) {
	t.Setenv("LINODE_TOKEN", "test-token")

	cloud, err := BuildCloud(newLinodeTestCluster("us-east"))
	if err != nil {
		t.Fatalf("BuildCloud returned error: %v", err)
	}
	if got, want := cloud.ProviderID(), kops.CloudProviderLinode; got != want {
		t.Fatalf("provider mismatch: got %q, want %q", got, want)
	}
	if got, want := cloud.Region(), "us-east"; got != want {
		t.Fatalf("region mismatch: got %q, want %q", got, want)
	}
}
