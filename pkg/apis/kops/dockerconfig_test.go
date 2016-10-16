/*
Copyright 2016 The Kubernetes Authors.

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

package kops

import (
	"k8s.io/kubernetes/pkg/util/validation/field"
	"testing"
)

func Test_Validate_DockerConfig_Storage(t *testing.T) {
	for _, name := range []string{"aufs", "zfs", "overlay"} {
		config := &DockerConfig{Storage: &name}
		errs := ValidateDockerConfig(config, field.NewPath("docker"))
		if len(errs) != 0 {
			t.Fatalf("Unexpected errors validating DockerConfig %q", errs)
		}
	}

	for _, name := range []string{"overlayfs", "", "au"} {
		config := &DockerConfig{Storage: &name}
		errs := ValidateDockerConfig(config, field.NewPath("docker"))
		if len(errs) != 1 {
			t.Fatalf("Expected errors validating DockerConfig %q", config)
		}
		if errs[0].Field != "docker.storage" || errs[0].Type != field.ErrorTypeNotSupported {
			t.Fatalf("Not the expected error validating DockerConfig %q", errs)
		}
	}
}
