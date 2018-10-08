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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/nodeup/pkg/model/resources"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"

	"github.com/golang/glog"
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
	Distros       []distros.Distribution
	Dependencies  []string
	Architectures []Architecture

	// PlainBinary indicates that the Source is not an OS, but a "bare" tar.gz
	PlainBinary bool
}

// DefaultDockerVersion is the (legacy) docker version we use if one is not specified in the manifest.
// We don't change this with each version of kops, we expect newer versions of kops to populate the field.
const DefaultDockerVersion = "1.12.3"

var dockerVersions = []dockerVersion{
	// 1.11.2 - Jessie
	{
		DockerVersion: "1.11.2",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionJessie},
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
		Distros:       []distros.Distribution{distros.DistributionXenial},
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
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.11.2",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.11.2-1.el7.centos.x86_64.rpm",
		Hash:          "432e6d7948df9e05f4190fce2f423eedbfd673d5",
		Dependencies:  []string{"libtool-ltdl"},
	},
	{
		DockerVersion: "1.11.2",
		Name:          "docker-engine-selinux",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.11.2",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.11.2-1.el7.centos.noarch.rpm",
		Hash:          "f6da608fa8eeb2be8071489086ed9ff035f6daba",
	},

	// 1.12.1 - Jessie
	{
		DockerVersion: "1.12.1",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionJessie},
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
		Distros:       []distros.Distribution{distros.DistributionXenial},
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
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.1",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.12.1-1.el7.centos.x86_64.rpm",
		Hash:          "636471665665546224444052c3b48001397036be",
		Dependencies:  []string{"libtool-ltdl"},
	},
	{
		DockerVersion: "1.12.1",
		Name:          "docker-engine-selinux",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.1",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.12.1-1.el7.centos.noarch.rpm",
		Hash:          "52ec22128e70acc2f76b3a8e87ff96785995116a",
	},

	// 1.12.3 - k8s 1.5

	// 1.12.3 - Jessie
	{
		DockerVersion: "1.12.3",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionJessie},
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
		Distros:       []distros.Distribution{distros.DistributionJessie},
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
		Distros:       []distros.Distribution{distros.DistributionXenial},
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
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.3",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.12.3-1.el7.centos.x86_64.rpm",
		Hash:          "67fbb78cfb9526aaf8142c067c10384df199d8f9",
		Dependencies:  []string{"libtool-ltdl", "libseccomp"},
	},
	{
		DockerVersion: "1.12.3",
		Name:          "docker-engine-selinux",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.3",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.12.3-1.el7.centos.noarch.rpm",
		Hash:          "a6b0243af348140236ed96f2e902b259c590eefa",
	},

	// 1.12.6 - k8s 1.6

	// 1.12.6 - Jessie
	{
		DockerVersion: "1.12.6",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionJessie},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.6-0~debian-jessie",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.6-0~debian-jessie_amd64.deb",
		Hash:          "1a8b0c4e3386e12964676a126d284cebf599cc8e",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
		//Depends: iptables, init-system-helpers (>= 1.18~), libapparmor1 (>= 2.6~devel), libc6 (>= 2.17), libdevmapper1.02.1 (>= 2:1.02.90), libltdl7 (>= 2.4.2), libsystemd0
		//Recommends: aufs-tools, ca-certificates, cgroupfs-mount | cgroup-lite, git, xz-utils
	},

	// 1.12.6 - Debian9 (stretch)
	{
		DockerVersion: "1.12.6",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionDebian9},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.6-0~debian-stretch",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.6-0~debian-stretch_amd64.deb",
		Hash:          "18bb7d024658f27a1221eae4de78d792bf00611b",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl", "libseccomp2"},
		//Depends: iptables, init-system-helpers (>= 1.18~), libapparmor1 (>= 2.6~devel), libc6 (>= 2.17), libdevmapper1.02.1 (>= 2:1.02.97), libltdl7 (>= 2.4.6), libseccomp2 (>= 2.1.0), libsystemd0
		//Recommends: aufs-tools, ca-certificates, cgroupfs-mount | cgroup-lite, git, xz-utils
	},

	// 1.12.6 - Jessie on ARM
	{
		DockerVersion: "1.12.6",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionJessie},
		Architectures: []Architecture{ArchitectureArm},
		Version:       "1.12.6-0~debian-jessie",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.6-0~debian-jessie_armhf.deb",
		Hash:          "ac148e1f7381e4201e139584dd3c102372ad96fb",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 1.12.6 - Xenial
	{
		DockerVersion: "1.12.6",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionXenial},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.6-0~ubuntu-xenial",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.6-0~ubuntu-xenial_amd64.deb",
		Hash:          "fffc22da4ad5b20715bbb6c485b2d2bb7e84fd33",
		Dependencies:  []string{"bridge-utils", "iptables", "libapparmor1", "libltdl7", "perl"},
		// Depends: iptables, init-system-helpers (>= 1.18~), lsb-base (>= 4.1+Debian11ubuntu7), libapparmor1 (>= 2.6~devel), libc6 (>= 2.17), libdevmapper1.02.1 (>= 2:1.02.97), libltdl7 (>= 2.4.6), libseccomp2 (>= 2.1.0), libsystemd0
	},

	// 1.12.6 - Centos / Rhel7 (two packages)
	{
		DockerVersion: "1.12.6",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.6",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.12.6-1.el7.centos.x86_64.rpm",
		Hash:          "776dbefa9dc7733000e46049293555a9a422c50e",
		Dependencies:  []string{"libtool-ltdl", "libseccomp", "libcgroup"},
	},
	{
		DockerVersion: "1.12.6",
		Name:          "docker-engine-selinux",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.6",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.12.6-1.el7.centos.noarch.rpm",
		Hash:          "9a6ee0d631ca911b6927450a3c396e9a5be75047",
		Dependencies:  []string{"policycoreutils-python"},
	},

	// 1.13.1 - k8s 1.8

	// 1.13.1 - Debian9 (stretch)
	{
		DockerVersion: "1.13.1",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionDebian9},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.13.1-0~debian-stretch",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.13.1-0~debian-stretch_amd64.deb",
		Hash:          "19296514610aa2e5efddade5222cafae7894a689",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
		//Depends: iptables, init-system-helpers (>= 1.18~), libapparmor1 (>= 2.6~devel), libc6 (>= 2.17), libdevmapper1.02.1 (>= 2:1.02.90), libltdl7 (>= 2.4.2), libsystemd0
		//Recommends: aufs-tools, ca-certificates, cgroupfs-mount | cgroup-lite, git, xz-utils
	},

	// 1.13.1 - Jessie
	{
		DockerVersion: "1.13.1",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionJessie},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.13.1-0~debian-jessie",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.13.1-0~debian-jessie_amd64.deb",
		Hash:          "1d3370549e32ea13b2755b2db8dbc82b2b787ece",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
		//Depends: iptables, init-system-helpers (>= 1.18~), libapparmor1 (>= 2.6~devel), libc6 (>= 2.17), libdevmapper1.02.1 (>= 2:1.02.90), libltdl7 (>= 2.4.2), libsystemd0
		//Recommends: aufs-tools, ca-certificates, cgroupfs-mount | cgroup-lite, git, xz-utils
	},

	// 1.13.1 - Jessie on ARM
	{
		DockerVersion: "1.13.1",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionJessie},
		Architectures: []Architecture{ArchitectureArm},
		Version:       "1.13.1-0~debian-jessie",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.13.1-0~debian-jessie_armhf.deb",
		Hash:          "a3f252c5fbb2d3266be611bee50e1f331ff8d05f",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 1.13.1 - Xenial
	{
		DockerVersion: "1.13.1",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionXenial},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.13.1-0~ubuntu-xenial",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.13.1-0~ubuntu-xenial_amd64.deb",
		Hash:          "d12cbd686f44536c679a03cf0137df163f0bba5f",
		Dependencies:  []string{"bridge-utils", "iptables", "libapparmor1", "libltdl7", "perl"},
		// Depends: iptables, init-system-helpers (>= 1.18~), lsb-base (>= 4.1+Debian11ubuntu7), libapparmor1 (>= 2.6~devel), libc6 (>= 2.17), libdevmapper1.02.1 (>= 2:1.02.97), libltdl7 (>= 2.4.6), libseccomp2 (>= 2.1.0), libsystemd0
	},

	// 1.13.1 - Centos / Rhel7 (two packages)
	{
		DockerVersion: "1.13.1",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.13.1",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.13.1-1.el7.centos.x86_64.rpm",
		Hash:          "b18f7fd8057665e7d2871d29640e214173f70fe1",
		Dependencies:  []string{"libtool-ltdl", "libseccomp", "libcgroup"},
	},
	{
		DockerVersion: "1.13.1",
		Name:          "docker-engine-selinux",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.13.1",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.13.1-1.el7.centos.noarch.rpm",
		Hash:          "948c518a610af631fa98aa32d9bcd43e9ddd5ebc",
		Dependencies:  []string{"policycoreutils-python", "selinux-policy-base", "selinux-policy-targeted"},
	},

	// 17.03.2 - k8s 1.8

	// 17.03.2 - Debian9 (stretch)
	{
		DockerVersion: "17.03.2",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionDebian9},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "17.03.2~ce-0~debian-stretch",
		Source:        "http://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce_17.03.2~ce-0~debian-stretch_amd64.deb",
		Hash:          "36773361cf44817371770cb4e6e6823590d10297",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 17.03.2 - Jessie
	{
		DockerVersion: "17.03.2",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionJessie},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "17.03.2~ce-0~debian-jessie",
		Source:        "http://download.docker.com/linux/debian/dists/jessie/pool/stable/amd64/docker-ce_17.03.2~ce-0~debian-jessie_amd64.deb",
		Hash:          "a7ac54aaa7d33122ca5f7a2df817cbefb5cdbfc7",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 17.03.2 - Jessie on ARM
	{
		DockerVersion: "17.03.2",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionJessie},
		Architectures: []Architecture{ArchitectureArm},
		Version:       "17.03.2~ce-0~debian-jessie",
		Source:        "http://download.docker.com/linux/debian/dists/jessie/pool/stable/armhf/docker-ce_17.03.2~ce-0~debian-jessie_armhf.deb",
		Hash:          "71e425b83ce0ef49d6298d61e61c4efbc76b9c65",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 17.03.2 - Xenial
	{
		DockerVersion: "17.03.2",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionXenial},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "17.03.2~ce-0~ubuntu-xenial",
		Source:        "http://download.docker.com/linux/ubuntu/dists/xenial/pool/stable/amd64/docker-ce_17.03.2~ce-0~ubuntu-xenial_amd64.deb",
		Hash:          "4dcee1a05ec592e8a76e53e5b464ea43085a2849",
		Dependencies:  []string{"bridge-utils", "iptables", "libapparmor1", "libltdl7", "perl"},
	},

	// 17.03.2 - Ubuntu Bionic via binary download (no packages available)
	{
		DockerVersion: "17.03.2",
		PlainBinary:   true,
		Distros:       []distros.Distribution{distros.DistributionBionic},
		Architectures: []Architecture{ArchitectureAmd64},
		Source:        "http://download.docker.com/linux/static/stable/x86_64/docker-17.03.2-ce.tgz",
		Hash:          "141716ae046016a1792ce232a0f4c8eed7fe37d1",
		Dependencies:  []string{"bridge-utils", "iptables", "libapparmor1", "libltdl7", "perl"},
	},

	// 17.03.2 - Centos / Rhel7 (two packages)
	{
		DockerVersion: "17.03.2",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "17.03.2.ce",
		Source:        "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-17.03.2.ce-1.el7.centos.x86_64.rpm",
		Hash:          "494ca888f5b1553f93b9d9a5dad4a67f76cf9eb5",
		Dependencies:  []string{"libtool-ltdl", "libseccomp", "libcgroup"},
	},
	{
		DockerVersion: "17.03.2",
		Name:          "docker-ce-selinux",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "17.03.2.ce",
		Source:        "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-selinux-17.03.2.ce-1.el7.centos.noarch.rpm",
		Hash:          "4659c937b66519c88ef2a82a906bb156db29d191",
		Dependencies:  []string{"policycoreutils-python"},
	},
	// 17.09.0 - k8s 1.8

	// 17.09.0 - Jessie
	{
		DockerVersion: "17.09.0",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionJessie},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "17.09.0~ce-0~debian",
		Source:        "http://download.docker.com/linux/debian/dists/jessie/pool/stable/amd64/docker-ce_17.09.0~ce-0~debian_amd64.deb",
		Hash:          "430ba87f8aa36fedcac1a48e909cbe1830b53845",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 17.09.0 - Jessie on ARM
	{
		DockerVersion: "17.09.0",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionJessie},
		Architectures: []Architecture{ArchitectureArm},
		Version:       "17.09.0~ce-0~debian",
		Source:        "http://download.docker.com/linux/debian/dists/jessie/pool/stable/armhf/docker-ce_17.09.0~ce-0~debian_armhf.deb",
		Hash:          "5001a1defec7c33aa58ddebbd3eae6ebb5f36479",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 17.09.0 - Debian9 (stretch)
	{
		DockerVersion: "17.09.0",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionDebian9},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "17.09.0~ce-0~debian",
		Source:        "http://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce_17.09.0~ce-0~debian_amd64.deb",
		Hash:          "70aa5f96cf00f11374b6593ccf4ed120a65375d2",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 17.09.0 - Xenial
	{
		DockerVersion: "17.09.0",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionXenial},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "17.09.0~ce-0~ubuntu",
		Source:        "http://download.docker.com/linux/ubuntu/dists/xenial/pool/stable/amd64/docker-ce_17.09.0~ce-0~ubuntu_amd64.deb",
		Hash:          "94f6e89be6d45d9988269a237eb27c7d6a844d7f",
		Dependencies:  []string{"bridge-utils", "iptables", "libapparmor1", "libltdl7", "perl"},
		//Depends: iptables, init-system-helpers, lsb-base, libapparmor1, libc6, libdevmapper1.02.1, libltdl7, libeseccomp2, libsystemd0
		//Recommends: aufs-tools, ca-certificates, cgroupfs-mount | cgroup-lite, git, xz-utils, apparmor
	},

	// 17.09.0 - Centos / Rhel7
	{
		DockerVersion: "17.09.0",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "17.09.0.ce",
		Source:        "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-17.09.0.ce-1.el7.centos.x86_64.rpm",
		Hash:          "b4ce72e80ff02926de943082821bbbe73958f87a",
		Dependencies:  []string{"libtool-ltdl", "libseccomp", "libcgroup"},
	},

	// 18.03.1 - Bionic
	{
		DockerVersion: "18.03.1",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionBionic},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "18.03.1~ce~3-0~ubuntu",
		Source:        "https://download.docker.com/linux/ubuntu/dists/bionic/pool/stable/amd64/docker-ce_18.03.1~ce~3-0~ubuntu_amd64.deb",
		Hash:          "b55b32bd0e9176dd32b1e6128ad9fda10a65cc8b",
		Dependencies:  []string{"bridge-utils", "iptables", "libapparmor1", "libltdl7", "perl"},
		//Depends: iptables, init-system-helpers, lsb-base, libapparmor1, libc6, libdevmapper1.02.1, libltdl7, libeseccomp2, libsystemd0
		//Recommends: aufs-tools, ca-certificates, cgroupfs-mount | cgroup-lite, git, xz-utils, apparmor
	},

	// 18.06.1 - Debian Stretch
	{

		DockerVersion: "18.06.1",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionDebian9},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "18.06.1~ce-0~debian",
		Source:        "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce_18.06.1~ce~3-0~debian_amd64.deb",
		Hash:          "18473b80e61b6d4eb8b52d87313abd71261287e5",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},
}

func (d *dockerVersion) matches(arch Architecture, dockerVersion string, distro distros.Distribution) bool {
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

// Build is responsible for configuring the docker daemon
func (b *DockerBuilder) Build(c *fi.ModelBuilderContext) error {

	// @check: neither coreos or containeros need provision docker.service, just the docker daemon options
	switch b.Distribution {
	case distros.DistributionCoreOS:
		glog.Infof("Detected CoreOS; won't install Docker")
		if err := b.buildContainerOSConfigurationDropIn(c); err != nil {
			return err
		}
		return nil

	case distros.DistributionContainerOS:
		glog.Infof("Detected ContainerOS; won't install Docker")
		if err := b.buildContainerOSConfigurationDropIn(c); err != nil {
			return err
		}
		return nil
	}

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
		count := 0
		for i := range dockerVersions {
			dv := &dockerVersions[i]
			if !dv.matches(b.Architecture, dockerVersion, b.Distribution) {
				continue
			}

			count++

			if dv.PlainBinary {
				c.AddTask(&nodetasks.Archive{
					Name:            "docker",
					Source:          dv.Source,
					Hash:            dv.Hash,
					TargetDir:       "/usr/bin/",
					StripComponents: 1,
				})

				c.AddTask(b.buildDockerGroup())
				c.AddTask(b.buildSystemdSocket())
			} else {
				c.AddTask(&nodetasks.Package{
					Name:    dv.Name,
					Version: s(dv.Version),
					Source:  s(dv.Source),
					Hash:    s(dv.Hash),

					// TODO: PreventStart is now unused?
					PreventStart: fi.Bool(true),
				})
			}

			for _, dep := range dv.Dependencies {
				c.AddTask(&nodetasks.Package{Name: dep})
			}

			// Note we do _not_ stop looping... centos/rhel comprises multiple packages
		}

		if count == 0 {
			glog.Warningf("Did not find docker package for %s %s %s", b.Distribution, b.Architecture, dockerVersion)
		}
	}

	// Split into major.minor.(patch+pr+meta)
	parts := strings.SplitN(dockerVersion, ".", 3)
	if len(parts) != 3 {
		return fmt.Errorf("error parsing docker version %q, no Major.Minor.Patch elements found", dockerVersion)
	}

	// Validate major
	dockerVersionMajor, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing major docker version %q: %v", parts[0], err)
	}

	// Validate minor
	dockerVersionMinor, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing minor docker version %q: %v", parts[1], err)
	}

	c.AddTask(b.buildSystemdService(dockerVersionMajor, dockerVersionMinor))

	if err := b.buildSysconfig(c); err != nil {
		return err
	}

	return nil
}

// buildDockerGroup creates the docker group, which owns the docker.socket
func (b *DockerBuilder) buildDockerGroup() *nodetasks.GroupTask {
	return &nodetasks.GroupTask{
		Name:   "docker",
		System: true,
	}
}

// buildSystemdSocket creates docker.socket, for when we're not installing from a package
func (b *DockerBuilder) buildSystemdSocket() *nodetasks.Service {
	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Docker Socket for the API")
	manifest.Set("Unit", "PartOf", "docker.service")

	manifest.Set("Socket", "ListenStream", "/var/run/docker.sock")
	manifest.Set("Socket", "SocketMode", "0660")
	manifest.Set("Socket", "SocketUser", "root")
	manifest.Set("Socket", "SocketGroup", "docker")

	manifest.Set("Install", "WantedBy", "sockets.target")

	manifestString := manifest.Render()
	glog.V(8).Infof("Built docker.socket manifest\n%s", manifestString)

	service := &nodetasks.Service{
		Name:       "docker.socket",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

func (b *DockerBuilder) buildSystemdService(dockerVersionMajor int64, dockerVersionMinor int64) *nodetasks.Service {
	oldDocker := dockerVersionMajor <= 1 && dockerVersionMinor <= 11
	usesDockerSocket := true
	hasDockerBabysitter := false

	var dockerdCommand string
	if oldDocker {
		dockerdCommand = "/usr/bin/docker daemon"
	} else {
		dockerdCommand = "/usr/bin/dockerd"
	}

	if b.Distribution.IsDebianFamily() {
		hasDockerBabysitter = true
	}

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Docker Application Container Engine")
	manifest.Set("Unit", "Documentation", "https://docs.docker.com")

	if b.Distribution.IsRHELFamily() && !oldDocker {
		// See https://github.com/docker/docker/pull/24804
		usesDockerSocket = false
	}

	if usesDockerSocket {
		manifest.Set("Unit", "After", "network.target docker.socket")
		manifest.Set("Unit", "Requires", "docker.socket")
	} else {
		manifest.Set("Unit", "After", "network.target")
	}

	manifest.Set("Service", "Type", "notify")
	manifest.Set("Service", "EnvironmentFile", "/etc/sysconfig/docker")
	manifest.Set("Service", "EnvironmentFile", "/etc/environment")

	if usesDockerSocket {
		manifest.Set("Service", "ExecStart", dockerdCommand+" -H fd:// \"$DOCKER_OPTS\"")
	} else {
		manifest.Set("Service", "ExecStart", dockerdCommand+" \"$DOCKER_OPTS\"")
	}

	if !oldDocker {
		// This was added by docker 1.12
		// TODO: They seem sensible - should we backport them?

		manifest.Set("Service", "ExecReload", "/bin/kill -s HUP $MAINPID")
		// kill only the docker process, not all processes in the cgroup
		manifest.Set("Service", "KillMode", "process")

		manifest.Set("Service", "TimeoutStartSec", "0")
	}

	if oldDocker {
		// Only in older versions of docker (< 1.12)
		manifest.Set("Service", "MountFlags", "slave")
	}

	// Having non-zero Limit*s causes performance problems due to accounting overhead
	// in the kernel. We recommend using cgroups to do container-local accounting.
	// TODO: Should we set this? https://github.com/kubernetes/kubernetes/issues/39682
	//service.Set("Service", "LimitNOFILE", "infinity")
	//service.Set("Service", "LimitNPROC", "infinity")
	//service.Set("Service", "LimitCORE", "infinity")
	manifest.Set("Service", "LimitNOFILE", "1048576")
	manifest.Set("Service", "LimitNPROC", "1048576")
	manifest.Set("Service", "LimitCORE", "infinity")

	//# Uncomment TasksMax if your systemd version supports it.
	//# Only systemd 226 and above support this version.
	//#TasksMax=infinity
	if b.IsKubernetesGTE("1.10") {
		// Equivalent of https://github.com/kubernetes/kubernetes/pull/51986
		manifest.Set("Service", "TasksMax", "infinity")
	}

	manifest.Set("Service", "Restart", "always")
	manifest.Set("Service", "RestartSec", "2s")
	manifest.Set("Service", "StartLimitInterval", "0")

	// set delegate yes so that systemd does not reset the cgroups of docker containers
	manifest.Set("Service", "Delegate", "yes")

	if hasDockerBabysitter {
		manifest.Set("Service", "ExecStartPre", "/opt/kubernetes/helpers/docker-prestart")
	}

	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	glog.V(8).Infof("Built service manifest %q\n%s", "docker", manifestString)

	service := &nodetasks.Service{
		Name:       "docker.service",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

// buildContainerOSConfigurationDropIn is responsible for configuring the docker daemon options
func (b *DockerBuilder) buildContainerOSConfigurationDropIn(c *fi.ModelBuilderContext) error {
	lines := []string{
		"[Service]",
		"EnvironmentFile=/etc/sysconfig/docker",
		"EnvironmentFile=/etc/environment",
	}

	if b.IsKubernetesGTE("1.10") {
		// Equivalent of https://github.com/kubernetes/kubernetes/pull/51986
		lines = append(lines, "TasksMax=infinity")
	}

	contents := strings.Join(lines, "\n")

	c.AddTask(&nodetasks.File{
		AfterFiles: []string{"/etc/sysconfig/docker"},
		Path:       "/etc/systemd/system/docker.service.d/10-kops.conf",
		Contents:   fi.NewStringResource(contents),
		Type:       nodetasks.FileType_File,
		OnChangeExecute: [][]string{
			{"systemctl", "daemon-reload"},
			{"systemctl", "restart", "docker.service"},
			// We need to restart kops-configuration service since nodeup needs to load images
			// into docker with the new overlay storage. Restart is on the background because
			// kops-configuration is of type 'one-shot' so the restart command will wait for
			// nodeup to finish executing
			{"systemctl", "restart", "kops-configuration.service", "&"},
		},
	})

	if err := b.buildSysconfig(c); err != nil {
		return err
	}

	return nil
}

// buildSysconfig is responsible for extracting the docker configuration and writing the sysconfig file
func (b *DockerBuilder) buildSysconfig(c *fi.ModelBuilderContext) error {
	var docker kops.DockerConfig
	if b.Cluster.Spec.Docker != nil {
		docker = *b.Cluster.Spec.Docker
	}

	// ContainerOS now sets the storage flag in /etc/docker/daemon.json, and it is an error to set it twice
	if b.Distribution == distros.DistributionContainerOS {
		// So that we can support older COS images though, we do check for /etc/docker/daemon.json
		if b, err := ioutil.ReadFile("/etc/docker/daemon.json"); err != nil {
			if os.IsNotExist(err) {
				glog.V(2).Infof("/etc/docker/daemon.json not found")
			} else {
				glog.Warningf("error reading /etc/docker/daemon.json: %v", err)
			}
		} else {
			// Maybe we get smarter here?
			data := make(map[string]interface{})
			if err := json.Unmarshal(b, &data); err != nil {
				glog.Warningf("error deserializing /etc/docker/daemon.json: %v", err)
			} else {
				storageDriver := data["storage-driver"]
				glog.Infof("/etc/docker/daemon.json has storage-driver: %q", storageDriver)
			}
			docker.Storage = nil
		}
	}

	flagsString, err := flagbuilder.BuildFlags(&docker)
	if err != nil {
		return fmt.Errorf("error building docker flags: %v", err)
	}

	lines := []string{
		"DOCKER_OPTS=" + flagsString,
		"DOCKER_NOFILE=1000000",
	}
	contents := strings.Join(lines, "\n")

	c.AddTask(&nodetasks.File{
		Path:     "/etc/sysconfig/docker",
		Contents: fi.NewStringResource(contents),
		Type:     nodetasks.FileType_File,
	})

	return nil
}
