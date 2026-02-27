/*
 /*
Copyright The Kubernetes Authors.

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

package dra_support

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kops/tests/e2e/scenarios/ai-conformance/validators"
)

// TestAcceleratorsDRASupport corresponds to the accelerators/dra_support scenario
func TestAcceleratorsDRASupport(t *testing.T) {
	// Description:
	//   Support Dynamic Resource Allocation (DRA) APIs to enable more flexible and fine-grained resource requests beyond simple counts.
	h := validators.NewValidatorHarness(t)

	h.Logf("# DRA API Availability")

	gv := schema.GroupVersion{Group: "resource.k8s.io", Version: "v1"}

	// Check resource.k8s.io/v1 is registered
	h.Logf("## Checking for DRA API version v1")
	{
		result := h.ShellExec(fmt.Sprintf("kubectl api-versions | grep %s", gv.Group+"/"+gv.Version))
		if !strings.Contains(result.Stdout(), "resource.k8s.io/v1\n") {
			h.Fatalf("Expected DRA API version group %s version %s, but it was not found", gv.Group, gv.Version)
		}
		h.Success("DRA API version %s is available.", gv.String())
	}

	// Check all expected DRA API resources are registered
	for _, resource := range []string{"deviceclasses", "resourceclaims", "resourceclaimtemplates", "resourceslices"} {
		h.Logf("## Checking for %s", resource)
		result := h.ShellExec(fmt.Sprintf("kubectl api-resources --api-group=%s | grep %s", gv.Group, resource))
		if !strings.Contains(result.Stdout(), "resource.k8s.io/v1") {
			h.Fatalf("Expected DRA API resource %s to be available in group %s version %s, but it was not found", resource, gv.Group, gv.Version)
		}
		h.Success("DRA API resource %s is available.", resource)
	}
}
