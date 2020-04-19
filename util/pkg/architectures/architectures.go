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

package architectures

import (
	"fmt"
	"runtime"

	"k8s.io/klog"
)

type Architecture string

var (
	ArchitectureAmd64 Architecture = "amd64"
	ArchitectureArm64 Architecture = "arm64"
)

func FindArchitecture() (Architecture, error) {
	switch runtime.GOARCH {
	case "amd64":
		return ArchitectureAmd64, nil
	case "arm64":
		return ArchitectureArm64, nil
	default:
		return "", fmt.Errorf("unsupported arch: %q", runtime.GOARCH)
	}
}

func GetArchitectures() []Architecture {
	return []Architecture{
		ArchitectureAmd64,
		ArchitectureArm64,
	}
}

func (a Architecture) BuildTags() []string {
	var t []string

	switch a {
	case ArchitectureAmd64:
		t = []string{"_amd64"}
	case ArchitectureArm64:
		t = []string{"_arm64"}
	default:
		klog.Fatalf("unknown architecture: %s", a)
		return nil
	}

	return t
}

func (a Architecture) IsAmd64() bool {
	switch a {
	case ArchitectureAmd64:
		return true
	case ArchitectureArm64:
		return false
	default:
		klog.Fatalf("unknown architecture: %s", a)
		return false
	}
}

func (a Architecture) IsArm64() bool {
	switch a {
	case ArchitectureAmd64:
		return false
	case ArchitectureArm64:
		return true
	default:
		klog.Fatalf("unknown architecture: %s", a)
		return false
	}
}
