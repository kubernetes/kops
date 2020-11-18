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
	"sync"

	"k8s.io/klog/v2"
)

type Architecture string

const (
	ArchitectureAmd64 Architecture = "amd64"
	ArchitectureArm64 Architecture = "arm64"
)

var supportedArchitecturesMutex sync.Mutex
var supportedArchitectures []Architecture

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
	supportedArchitecturesMutex.Lock()
	defer supportedArchitecturesMutex.Unlock()

	if len(supportedArchitectures) > 0 {
		return supportedArchitectures
	}

	klog.Warningf("could not find any supported CPU architecture, falling back to AMD64")
	return []Architecture{ArchitectureAmd64}
}

func AddSupported(arch Architecture) {
	supportedArchitecturesMutex.Lock()
	defer supportedArchitecturesMutex.Unlock()

	for _, supported := range supportedArchitectures {
		if supported == arch {
			return
		}
	}

	supportedArchitectures = append(supportedArchitectures, arch)
}

func FromString(arch string) (Architecture, error) {
	switch arch {
	case "amd64":
		return ArchitectureAmd64, nil
	case "arm64":
		return ArchitectureArm64, nil
	default:
		return "", fmt.Errorf("unsupported arch: %q", arch)
	}
}
