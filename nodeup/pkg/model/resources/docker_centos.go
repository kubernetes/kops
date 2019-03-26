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

package resources

import "k8s.io/kops/nodeup/pkg/distros"

var dockerVersionsCentos = []dockerVersion{
	// 1.11.2 - Centos / Rhel7
	{
		DockerVersion: "1.11.2",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.11.2",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.11.2-1.el7.centos.x86_64.rpm",
		Hash:          "432e6d7948df9e05f4190fce2f423eedbfd673d5",
		ExtraPackages: map[string]packageInfo{
			"docker-engine-selinux": {
				Version: "1.11.2",
				Source:  "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.11.2-1.el7.centos.noarch.rpm",
				Hash:    "f6da608fa8eeb2be8071489086ed9ff035f6daba",
			},
		},
		Dependencies: []string{"libtool-ltdl"},
	},

	// 1.12.1 - Centos / Rhel7
	{
		DockerVersion: "1.12.1",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.1",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.12.1-1.el7.centos.x86_64.rpm",
		Hash:          "636471665665546224444052c3b48001397036be",
		ExtraPackages: map[string]packageInfo{
			"docker-engine-selinux": {
				Version: "1.12.1",
				Source:  "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.12.1-1.el7.centos.noarch.rpm",
				Hash:    "52ec22128e70acc2f76b3a8e87ff96785995116a",
			},
		},
		Dependencies: []string{"libtool-ltdl"},
	},

	// 1.12.3 - Centos / Rhel7
	{
		DockerVersion: "1.12.3",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.3",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.12.3-1.el7.centos.x86_64.rpm",
		Hash:          "67fbb78cfb9526aaf8142c067c10384df199d8f9",
		ExtraPackages: map[string]packageInfo{
			"docker-engine-selinux": {
				Version: "1.12.3",
				Source:  "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.12.3-1.el7.centos.noarch.rpm",
				Hash:    "a6b0243af348140236ed96f2e902b259c590eefa",
			},
		},
		Dependencies: []string{"libtool-ltdl", "libseccomp"},
	},

	// 1.12.6 - Centos / Rhel7
	{
		DockerVersion: "1.12.6",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.6",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.12.6-1.el7.centos.x86_64.rpm",
		Hash:          "776dbefa9dc7733000e46049293555a9a422c50e",
		ExtraPackages: map[string]packageInfo{
			"docker-engine-selinux": {
				Version: "1.12.6",
				Source:  "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.12.6-1.el7.centos.noarch.rpm",
				Hash:    "9a6ee0d631ca911b6927450a3c396e9a5be75047",
			},
		},
		Dependencies: []string{"libtool-ltdl", "libseccomp", "libcgroup", "policycoreutils-python"},
	},

	// 1.13.1 - Centos / Rhel7
	{
		DockerVersion: "1.13.1",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.13.1",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.13.1-1.el7.centos.x86_64.rpm",
		Hash:          "b18f7fd8057665e7d2871d29640e214173f70fe1",
		ExtraPackages: map[string]packageInfo{
			"docker-engine-selinux": {
				Version: "1.13.1",
				Source:  "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.13.1-1.el7.centos.noarch.rpm",
				Hash:    "948c518a610af631fa98aa32d9bcd43e9ddd5ebc",
			},
		},
		Dependencies: []string{"libtool-ltdl", "libseccomp", "libcgroup", "policycoreutils-python", "selinux-policy-base", "selinux-policy-targeted"},
	},

	// 17.03.2 - Centos / Rhel7
	{
		DockerVersion: "17.03.2",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "17.03.2.ce",
		Source:        "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-17.03.2.ce-1.el7.centos.x86_64.rpm",
		Hash:          "494ca888f5b1553f93b9d9a5dad4a67f76cf9eb5",
		ExtraPackages: map[string]packageInfo{
			"docker-ce-selinux": {
				Version: "17.03.2.ce",
				Source:  "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-selinux-17.03.2.ce-1.el7.centos.noarch.rpm",
				Hash:    "4659c937b66519c88ef2a82a906bb156db29d191",
			},
		},
		Dependencies:  []string{"libtool-ltdl", "libseccomp", "libcgroup", "policycoreutils-python"},
		MarkImmutable: []string{"/usr/bin/docker-runc"},
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
		ExtraPackages: map[string]packageInfo{
			"container-selinux": {
				Version: "2.68",
				Source:  "http://mirror.centos.org/centos/7/extras/x86_64/Packages/container-selinux-2.68-1.el7.noarch.rpm",
				Hash:    "d9f87f7f4f2e8e611f556d873a17b8c0c580fec0",
			},
		},
		Dependencies: []string{"libtool-ltdl", "libseccomp", "libcgroup", "policycoreutils-python"},
	},

	// 18.06.1 - CentOS / Rhel7
	{
		DockerVersion: "18.06.1",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "18.06.1.ce",
		Source:        "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-18.06.1.ce-3.el7.x86_64.rpm",
		Hash:          "0a1325e570c5e54111a79623c9fd0c0c714d3a11",
		ExtraPackages: map[string]packageInfo{
			"container-selinux": {
				Version: "2.68",
				Source:  "http://mirror.centos.org/centos/7/extras/x86_64/Packages/container-selinux-2.68-1.el7.noarch.rpm",
				Hash:    "d9f87f7f4f2e8e611f556d873a17b8c0c580fec0",
			},
		},
		Dependencies: []string{"libtool-ltdl", "libseccomp", "libcgroup", "policycoreutils-python"},
	},

	// 18.06.2 - CentOS / Rhel7 (two packages)
	{
		DockerVersion: "18.06.2",
		Name:          "container-selinux",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "2.68",
		Source:        "http://mirror.centos.org/centos/7/extras/x86_64/Packages/container-selinux-2.68-1.el7.noarch.rpm",
		Hash:          "d9f87f7f4f2e8e611f556d873a17b8c0c580fec0",
		Dependencies:  []string{"policycoreutils-python"},
	},
	{
		DockerVersion: "18.06.2",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "18.06.2.ce",
		Source:        "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-18.06.2.ce-3.el7.x86_64.rpm",
		Hash:          "456eb7c5bfb37fac342e9ade21b602c076c5b367",
		Dependencies:  []string{"libtool-ltdl", "libseccomp", "libcgroup"},
	},

	// 18.06.3 - CentOS / Rhel7
	{
		DockerVersion: "18.06.3",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "18.06.3.ce",
		Source:        "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-18.06.3.ce-3.el7.x86_64.rpm",
		Hash:          "5369602f88406d4fb9159dc1d3fd44e76fb4cab8",
		ExtraPackages: map[string]packageInfo{
			"container-selinux": {
				Version: "2.68",
				Source:  "http://mirror.centos.org/centos/7/extras/x86_64/Packages/container-selinux-2.68-1.el7.noarch.rpm",
				Hash:    "d9f87f7f4f2e8e611f556d873a17b8c0c580fec0",
			},
		},
		Dependencies: []string{"libtool-ltdl", "libseccomp", "libcgroup", "policycoreutils-python"},
	},

	// TIP: When adding the next version, copy the previous
	// version, string replace the version, run `VERIFY_HASHES=1
	// go test ./nodeup/pkg/model` (you might want to temporarily
	// comment out older versions on a slower connection), and
	// then validate the dependencies etc
}
