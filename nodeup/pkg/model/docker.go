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

package model

import (
	"github.com/golang/glog"
	"k8s.io/kops/nodeup/pkg/model/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// DockerBuilder install docker (just the packages at the moment)
type DockerBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &DockerBuilder{}

type dockerVersion struct {
	Name    string
	Version string
	Source  string
	Hash    string

	DockerVersion string
	Distros       []Distribution
	Dependencies  []string
	Architectures []Architecture
}

const DefaultDockerVersion = "1.12.3"

var dockerVersions = []dockerVersion{
	// 1.11.2 - Jessie
	{
		DockerVersion: "1.11.2",
		Name:          "docker-engine",
		Distros:       []Distribution{DistributionJessie},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.11.2-0~jessie",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.11.2-0~jessie_amd64.deb",
		Hash:          "c312f1f6fa0b34df4589bb812e4f7af8e28fd51d",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 1.11.2 - Xenial
	{
		DockerVersion: "1.11.2",
		Name:          "docker-engine",
		Distros:       []Distribution{DistributionXenial},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.11.2-0~xenial",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.11.2-0~xenial_amd64.deb",
		Hash:          "194bfa864f0424d1bbdc7d499ccfa0445ce09b9f",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 1.11.2 - Centos / Rhel7 (two packages)
	{
		DockerVersion: "1.11.2",
		Name:          "docker-engine",
		Distros:       []Distribution{DistributionRhel7, DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.11.2",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.11.2-1.el7.centos.x86_64.rpm",
		Hash:          "432e6d7948df9e05f4190fce2f423eedbfd673d5",
		Dependencies:  []string{"libtool-ltdl"},
	},
	{
		DockerVersion: "1.11.2",
		Name:          "docker-engine-selinux",
		Distros:       []Distribution{DistributionRhel7, DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.11.2",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.11.2-1.el7.centos.noarch.rpm",
		Hash:          "f6da608fa8eeb2be8071489086ed9ff035f6daba",
	},

	// 1.12.1 - Jessie
	{
		DockerVersion: "1.12.1",
		Name:          "docker-engine",
		Distros:       []Distribution{DistributionJessie},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.1-0~jessie",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.1-0~jessie_amd64.deb",
		Hash:          "0401866231749abaabe8e09ee24432132839fe53",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 1.12.1 - Xenial
	{
		DockerVersion: "1.12.1",
		Name:          "docker-engine",
		Distros:       []Distribution{DistributionXenial},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.1-0~xenial",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.1-0~xenial_amd64.deb",
		Hash:          "30f7840704361673db2b62f25b6038628184b056",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 1.12.1 - Centos / Rhel7 (two packages)
	{
		DockerVersion: "1.12.1",
		Name:          "docker-engine",
		Distros:       []Distribution{DistributionRhel7, DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.1",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.12.1-1.el7.centos.x86_64.rpm",
		Hash:          "636471665665546224444052c3b48001397036be",
		Dependencies:  []string{"libtool-ltdl"},
	},
	{
		DockerVersion: "1.12.1",
		Name:          "docker-engine-selinux",
		Distros:       []Distribution{DistributionRhel7, DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.1",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.12.1-1.el7.centos.noarch.rpm",
		Hash:          "52ec22128e70acc2f76b3a8e87ff96785995116a",
	},

	// 1.12.3 - Jessie
	{
		DockerVersion: "1.12.3",
		Name:          "docker-engine",
		Distros:       []Distribution{DistributionJessie},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.3-0~jessie",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.3-0~jessie_amd64.deb",
		Hash:          "7c7eb45542b67a9cfb33c292ba245710efb5d773",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
		//Depends: iptables, init-system-helpers (>= 1.18~), libapparmor1 (>= 2.6~devel), libc6 (>= 2.17), libdevmapper1.02.1 (>= 2:1.02.90), libltdl7 (>= 2.4.2), libsystemd0
		//Recommends: aufs-tools, ca-certificates, cgroupfs-mount | cgroup-lite, git, xz-utils
	},

	// 1.12.3 - Jessie on ARM
	{
		DockerVersion: "1.12.3",
		Name:          "docker-engine",
		Distros:       []Distribution{DistributionJessie},
		Architectures: []Architecture{ArchitectureArm},
		Version:       "1.12.3-0~jessie",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.3-0~jessie_armhf.deb",
		Hash:          "aa2f2f710360268dc5fd3eb066868c5883d95698",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 1.12.3 - Xenial
	{
		DockerVersion: "1.12.3",
		Name:          "docker-engine",
		Distros:       []Distribution{DistributionXenial},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.3-0~xenial",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.3-0~xenial_amd64.deb",
		Hash:          "b758fc88346a1e5eebf7408b0d0c99f4f134166c",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 1.12.3 - Centos / Rhel7 (two packages)
	{
		DockerVersion: "1.12.3",
		Name:          "docker-engine",
		Distros:       []Distribution{DistributionRhel7, DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.3",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.12.3-1.el7.centos.x86_64.rpm",
		Hash:          "67fbb78cfb9526aaf8142c067c10384df199d8f9",
		Dependencies:  []string{"libtool-ltdl"},
	},
	{
		DockerVersion: "1.12.3",
		Name:          "docker-engine-selinux",
		Distros:       []Distribution{DistributionRhel7, DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.3",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.12.3-1.el7.centos.noarch.rpm",
		Hash:          "a6b0243af348140236ed96f2e902b259c590eefa",
	},
}

func (d *dockerVersion) matches(arch Architecture, dockerVersion string, distro Distribution) bool {
	if d.DockerVersion != dockerVersion {
		return false
	}
	foundDistro := false
	for _, d := range d.Distros {
		if d == distro {
			foundDistro = true
		}
	}
	if !foundDistro {
		return false
	}

	foundArch := false
	for _, a := range d.Architectures {
		if a == arch {
			foundArch = true
		}
	}
	if !foundArch {
		return false
	}

	return true
}

func (b *DockerBuilder) Build(c *fi.ModelBuilderContext) error {
	// Add Apache2 license
	{
		t := &nodetasks.File{
			Path:     "/usr/share/doc/docker/apache.txt",
			Contents: fi.NewStringResource(resources.DockerApache2License),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	dockerVersion := ""
	if b.Cluster.Spec.Docker != nil {
		dockerVersion = fi.StringValue(b.Cluster.Spec.Docker.Version)
	}
	if dockerVersion == "" {
		dockerVersion = DefaultDockerVersion
		glog.Warningf("DockerVersion not specified; using default %q", dockerVersion)
	}

	// Add packages
	{
		for i := range dockerVersions {
			dv := &dockerVersions[i]
			if !dv.matches(b.Architecture, dockerVersion, b.Distribution) {
				continue
			}

			c.AddTask(&nodetasks.Package{
				Name:    dv.Name,
				Version: s(dv.Version),
				Source:  s(dv.Source),
				Hash:    s(dv.Hash),

				// TODO: PreventStart is now unused?
				PreventStart: fi.Bool(true),
			})

			for _, dep := range dv.Dependencies {
				c.AddTask(&nodetasks.Package{Name: dep})
			}

			// Note we do _not_ stop looping... centos/rhel comprises multiple packages
		}
	}

	return nil
}
