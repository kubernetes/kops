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

var dockerVersionsDebian = []dockerVersion{
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
		MarkImmutable: []string{"/usr/bin/docker-runc"},
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
		MarkImmutable: []string{"/usr/bin/docker-runc"},
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
		MarkImmutable: []string{"/usr/bin/docker-runc"},
	},

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

	// 18.06.1 - Debian Stretch
	{
		DockerVersion: "18.06.1",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionDebian9},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "18.06.1~ce~3-0~debian",
		Source:        "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce_18.06.1~ce~3-0~debian_amd64.deb",
		Hash:          "18473b80e61b6d4eb8b52d87313abd71261287e5",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 18.06.2 - Jessie
	{
		DockerVersion: "18.06.2",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionJessie},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "18.06.2~ce~3-0~debian",
		Source:        "https://download.docker.com/linux/debian/dists/jessie/pool/stable/amd64/docker-ce_18.06.2~ce~3-0~debian_amd64.deb",
		Hash:          "1a2500311230aff37aa81dd1292a88302fb0a2e1",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 18.06.2 - Debian Stretch
	{

		DockerVersion: "18.06.2",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionDebian9},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "18.06.2~ce~3-0~debian",
		Source:        "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce_18.06.2~ce~3-0~debian_amd64.deb",
		Hash:          "aad1efd2c90725034e996c6a368ccc2bf41ca5b8",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 18.06.3 - Jessie (contains fix for CVE-2019-5736)
	{
		DockerVersion: "18.06.3",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionJessie},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "18.06.3~ce~3-0~debian",
		Source:        "https://download.docker.com/linux/debian/dists/jessie/pool/stable/amd64/docker-ce_18.06.3~ce~3-0~debian_amd64.deb",
		Hash:          "058bcd4b055560866b8cad978c7aa224694602da",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 18.06.3 - Debian Stretch (contains fix for CVE-2019-5736)
	{
		DockerVersion: "18.06.3",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionDebian9},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "18.06.3~ce~3-0~debian",
		Source:        "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce_18.06.3~ce~3-0~debian_amd64.deb",
		Hash:          "93b5a055a39462867d79109b00db1367e3d9e32f",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 18.09.3 - Debian Stretch
	{
		DockerVersion: "18.09.3",
		Name:          "docker-ce",
		Distros:       []distros.Distribution{distros.DistributionDebian9},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "18.09.3~3-0~debian-stretch",
		Source:        "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce_18.09.3~3-0~debian-stretch_amd64.deb",
		Hash:          "009b9a2d8bfaa97c74773fe4ec25b6bb396b10d0",
		ExtraPackages: map[string]packageInfo{
			"docker-ce-cli": {
				Version: "18.09.3~3-0~debian-stretch",
				Source:  "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce-cli_18.09.3~3-0~debian-stretch_amd64.deb",
				Hash:    "557f868ec63e5251639ebd1d8669eb0c61dd555c",
			},
			"containerd.io": {
				Version: "1.2.4-1",
				Source:  "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/containerd.io_1.2.4-1_amd64.deb",
				Hash:    "48c6ab0c908316af9a183de5aad64703bc516bdf",
			},
		},
		Dependencies: []string{"bridge-utils", "libapparmor1", "libltdl7"},
	},

	// TIP: When adding the next version, copy the previous
	// version, string replace the version, run `VERIFY_HASHES=1
	// go test ./nodeup/pkg/model` (you might want to temporarily
	// comment out older versions on a slower connection), and
	// then validate the dependencies etc
}
