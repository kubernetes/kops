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
	"os"
	"runtime"
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

func GetSupported() []Architecture {
	// Kubernetes PR builds only generate AMD64 binaries at the moment
	// Force support only for AMD64 or ARM64
	arch := os.Getenv("KOPS_ARCH")
	switch arch {
	case "amd64":
		return []Architecture{ArchitectureAmd64}
	case "arm64":
		return []Architecture{ArchitectureArm64}
	}

	return []Architecture{
		ArchitectureAmd64,
		ArchitectureArm64,
	}
}
