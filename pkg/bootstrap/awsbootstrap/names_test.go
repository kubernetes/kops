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

package awsbootstrap

import "testing"

func TestPrivateDNSName(t *testing.T) {
	grid := []struct {
		privateIPv4 string
		region      string
		expected    string
	}{
		{"10.24.34.0", "us-east-1", "ip-10-24-34-0.ec2.internal"},
		{"10.24.34.0", "us-west-2", "ip-10-24-34-0.us-west-2.compute.internal"},
		{"10.99.1.236", "eu-central-1", "ip-10-99-1-236.eu-central-1.compute.internal"},
		{"10.0.0.1", "us-gov-west-1", "ip-10-0-0-1.us-gov-west-1.compute.internal"},
		{"172.20.1.2", "cn-north-1", "ip-172-20-1-2.cn-north-1.compute.internal"},
	}
	for _, g := range grid {
		actual := PrivateDNSName(g.privateIPv4, g.region)
		if actual != g.expected {
			t.Errorf("PrivateDNSName(%q, %q) = %q, expected %q", g.privateIPv4, g.region, actual, g.expected)
		}
	}
}
