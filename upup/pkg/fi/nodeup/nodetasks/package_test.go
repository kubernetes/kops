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

package nodetasks

import (
	"slices"
	"testing"

	"k8s.io/kops/upup/pkg/fi"
)

func TestPackageRunExpectsHealthy(t *testing.T) {
	// The zero-value target skips Find, so the test does not read the host package database.
	target := &fi.NodeupDryRunTarget{}
	p := &Package{Name: "runsc"}
	if err := p.Run(&fi.NodeupContext{Target: target}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// The default must be in place before the delta run compares the task.
	creates, _ := target.Changes()
	if created, ok := creates["Package"].(*Package); !ok || !fi.ValueOf(created.Healthy) {
		t.Fatalf("Run did not record a Package change with Healthy=true")
	}

	actual := &Package{Name: p.Name, Version: new("1.0"), Healthy: new(false)}
	changes := &Package{}
	fi.BuildChanges(actual, p, changes)
	if !fi.ValueOf(changes.Healthy) || changes.Version != nil {
		t.Errorf("unconfigured package: want a change in Healthy only")
	}

	actual.Healthy = new(true)
	if fi.BuildChanges(actual, p, &Package{}) {
		t.Errorf("configured package: want no changes")
	}
}

func TestAptGetArgs(t *testing.T) {
	gotUpdate := aptGetArgs("update")
	wantUpdate := []string{
		"apt-get",
		"-o", "Acquire::http::Timeout=30",
		"-o", "Acquire::https::Timeout=30",
		"-o", "DPkg::Lock::Timeout=60",
		"update",
	}
	if !slices.Equal(gotUpdate, wantUpdate) {
		t.Errorf("aptGetArgs(update) = %v, want %v", gotUpdate, wantUpdate)
	}

	gotInstall := aptGetArgs("install", "--yes", "--no-install-recommends")
	wantInstall := []string{
		"apt-get",
		"-o", "Acquire::http::Timeout=30",
		"-o", "Acquire::https::Timeout=30",
		"-o", "DPkg::Lock::Timeout=60",
		"install", "--yes", "--no-install-recommends",
	}
	if !slices.Equal(gotInstall, wantInstall) {
		t.Errorf("aptGetArgs(install) = %v, want %v", gotInstall, wantInstall)
	}
}
