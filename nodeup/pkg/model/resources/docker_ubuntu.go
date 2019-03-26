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

var dockerVersionsUbuntu = []dockerVersion{
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
		MarkImmutable: []string{"/usr/bin/docker-runc"},
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
		MarkImmutable: []string{"/usr/bin/docker-runc"},
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

	// 18.06.2 - Xenial
	{
		DockerVersion: "18.06.2",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionXenial},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "18.06.2~ce~3-0~ubuntu",
		Source:        "https://download.docker.com/linux/ubuntu/dists/xenial/pool/stable/amd64/docker-ce_18.06.2~ce~3-0~ubuntu_amd64.deb",
		Hash:          "03e5eaae9c84b144e1140d9b418e43fce0311892",
		Dependencies:  []string{"bridge-utils", "iptables", "libapparmor1", "libltdl7", "perl"},
		//Depends: iptables, init-system-helpers, lsb-base, libapparmor1, libc6, libdevmapper1.02.1, libltdl7, libeseccomp2, libsystemd0
		//Recommends: aufs-tools, ca-certificates, cgroupfs-mount | cgroup-lite, git, xz-utils, apparmor
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

	// 18.06.2 - Bionic
	{
		DockerVersion: "18.06.2",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionBionic},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "18.06.2~ce~3-0~ubuntu",
		Source:        "https://download.docker.com/linux/ubuntu/dists/bionic/pool/stable/amd64/docker-ce_18.06.2~ce~3-0~ubuntu_amd64.deb",
		Hash:          "9607c67644e3e1ad9661267c99499004f2e84e05",
		Dependencies:  []string{"bridge-utils", "iptables", "libapparmor1", "libltdl7", "perl"},
		//Depends: iptables, init-system-helpers, lsb-base, libapparmor1, libc6, libdevmapper1.02.1, libltdl7, libeseccomp2, libsystemd0
		//Recommends: aufs-tools, ca-certificates, cgroupfs-mount | cgroup-lite, git, xz-utils, apparmor
	},

	// 18.06.3 - Bionic (contains fix for CVE-2019-5736)
	{
		DockerVersion: "18.06.3",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionBionic},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "18.06.3~ce~3-0~ubuntu",
		Source:        "https://download.docker.com/linux/ubuntu/dists/bionic/pool/stable/amd64/docker-ce_18.06.3~ce~3-0~ubuntu_amd64.deb",
		Hash:          "b396678a8b70f0503a7b944fa6e3297ab27b345b",
		Dependencies:  []string{"bridge-utils", "iptables", "libapparmor1", "libltdl7", "perl"},
		//Depends: iptables, init-system-helpers, lsb-base, libapparmor1, libc6, libdevmapper1.02.1, libltdl7, libeseccomp2, libsystemd0
		//Recommends: aufs-tools, ca-certificates, cgroupfs-mount | cgroup-lite, git, xz-utils, apparmor
	},

	// TIP: When adding the next version, copy the previous
	// version, string replace the version, run `VERIFY_HASHES=1
	// go test ./nodeup/pkg/model` (you might want to temporarily
	// comment out older versions on a slower connection), and
	// then validate the dependencies etc
}
