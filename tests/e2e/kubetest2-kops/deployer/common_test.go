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

package deployer

import (
	"strings"
	"testing"
)

func TestDefaultClusterNamePresubmitIncludesBuildID(t *testing.T) {
	t.Setenv("JOB_NAME", "pull-kops-e2e-cni-calico-ipv6")
	t.Setenv("JOB_TYPE", "presubmit")
	t.Setenv("BUILD_ID", "2050895904046583808")
	t.Setenv("PULL_NUMBER", "18266")
	t.Setenv("KOPS_DNS_DOMAIN", "tests-kops-aws.k8s.io")

	d := &deployer{CloudProvider: "aws"}
	got, err := d.defaultClusterName()
	if err != nil {
		t.Fatalf("defaultClusterName returned error: %v", err)
	}

	const want = "e2e-pr18266-20508959040465.pull-kops-e2e-cni-calico-ipv6.tests-kops-aws.k8s.io"
	if got != want {
		t.Fatalf("defaultClusterName mismatch: got %q, want %q", got, want)
	}
}

func TestDefaultClusterNamePeriodicIncludesBuildID(t *testing.T) {
	t.Setenv("JOB_NAME", "periodic-kops-e2e-foo")
	t.Setenv("JOB_TYPE", "periodic")
	t.Setenv("BUILD_ID", "2050895904046583808")
	t.Setenv("KOPS_DNS_DOMAIN", "tests-kops-aws.k8s.io")

	d := &deployer{CloudProvider: "aws"}
	got, err := d.defaultClusterName()
	if err != nil {
		t.Fatalf("defaultClusterName returned error: %v", err)
	}

	const want = "e2e-20508959040465-periodic-kops-e2e-foo.tests-kops-aws.k8s.io"
	if got != want {
		t.Fatalf("defaultClusterName mismatch: got %q, want %q", got, want)
	}
}

func TestDefaultClusterNameGCETruncationKeepsBuildIDPrefix(t *testing.T) {
	t.Setenv("JOB_NAME", "periodic-kops-e2e-this-name-is-way-too-long-for-gce-and-must-be-truncated")
	t.Setenv("JOB_TYPE", "periodic")
	t.Setenv("BUILD_ID", "2050895904046583808")

	d := &deployer{CloudProvider: "gce"}
	got, err := d.defaultClusterName()
	if err != nil {
		t.Fatalf("defaultClusterName returned error: %v", err)
	}

	if !strings.HasPrefix(got, "e2e-20508959040465-") {
		t.Fatalf("defaultClusterName should keep build-id prefix, got %q", got)
	}
	if !strings.HasSuffix(got, ".k8s.local") {
		t.Fatalf("defaultClusterName should use k8s.local suffix, got %q", got)
	}
	if len(got) > 63 {
		t.Fatalf("defaultClusterName for gce should be <= 63 chars, got len=%d value=%q", len(got), got)
	}
}
