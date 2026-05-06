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

package nodeup

import (
	"reflect"
	"testing"
)

func TestNodeupCgroupMemoryCandidatesFromProcCgroupV2(t *testing.T) {
	candidates := nodeupCgroupMemoryCandidatesFromProcCgroup("0::/system.slice/kops-configuration.service\n")

	expected := []nodeupCgroupMemoryCandidate{
		{version: "v2", path: "/sys/fs/cgroup/system.slice/kops-configuration.service"},
		{version: "v2", path: "/sys/fs/cgroup"},
		{version: "v1", path: "/sys/fs/cgroup/memory"},
	}
	if !reflect.DeepEqual(candidates, expected) {
		t.Fatalf("unexpected candidates:\nexpected: %#v\nactual:   %#v", expected, candidates)
	}
}

func TestNodeupCgroupMemoryCandidatesFromProcCgroupV1(t *testing.T) {
	candidates := nodeupCgroupMemoryCandidatesFromProcCgroup("9:cpu,cpuacct:/ignored\n8:memory:/system.slice/kops-configuration.service\n")

	expected := []nodeupCgroupMemoryCandidate{
		{version: "v1", path: "/sys/fs/cgroup/memory/system.slice/kops-configuration.service"},
		{version: "v2", path: "/sys/fs/cgroup"},
		{version: "v1", path: "/sys/fs/cgroup/memory"},
	}
	if !reflect.DeepEqual(candidates, expected) {
		t.Fatalf("unexpected candidates:\nexpected: %#v\nactual:   %#v", expected, candidates)
	}
}

func TestParseNodeupCgroupMemoryStat(t *testing.T) {
	stats := parseNodeupCgroupMemoryStat(`anon 123
file 456
malformed
kernel_stack not-a-number
slab 789
`)

	expected := map[string]int64{
		"anon": 123,
		"file": 456,
		"slab": 789,
	}
	if !reflect.DeepEqual(stats, expected) {
		t.Fatalf("unexpected stats:\nexpected: %#v\nactual:   %#v", expected, stats)
	}
}

func TestFormatNodeupCgroupMemoryStats(t *testing.T) {
	stats := map[string]int64{
		"file": 456,
		"anon": 123,
		"slab": 789,
	}

	expected := "anon=123,file=456,slab=789"
	if actual := formatNodeupCgroupMemoryStats(stats); actual != expected {
		t.Fatalf("unexpected formatted stats: expected %q, got %q", expected, actual)
	}
}
