/*
Copyright 2021 The Kubernetes Authors.

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

package model

import (
	"fmt"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// NvidiaBuilder installs the Nvidia driver and runtime.
type NvidiaBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &NvidiaBuilder{}

// Build is responsible for installing packages.
func (b *NvidiaBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.InstallNvidiaRuntime() && b.Distribution.IsUbuntu() {
		version := ""
		if b.Distribution.Version() >= 22.04 {
			version = "22.04"
		} else if b.Distribution.Version() >= 20.04 {
			version = "20.04"
		} else {
			version = "18.04"
		}
		c.AddTask(&nodetasks.AptSource{
			Name:    "nvidia-container-runtime",
			Keyring: "https://nvidia.github.io/nvidia-container-runtime/gpgkey",
			Sources: []string{
				fmt.Sprintf("deb https://nvidia.github.io/nvidia-container-runtime/stable/ubuntu%s/$(ARCH) /", version),
				fmt.Sprintf("deb https://nvidia.github.io/libnvidia-container/stable/ubuntu%s/$(ARCH) /", version),
				fmt.Sprintf("deb https://nvidia.github.io/nvidia-docker/ubuntu%s/$(ARCH) /", version),
			},
		})
		c.AddTask(&nodetasks.Package{Name: "nvidia-container-runtime"})
		c.AddTask(&nodetasks.Package{Name: b.NodeupConfig.NvidiaGPU.DriverPackage})
	}
	return nil
}
