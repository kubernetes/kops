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

package model

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/nodeup/pkg/model/resources"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/architectures"
)

// DockerBuilder install docker (just the packages at the moment)
type DockerBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &DockerBuilder{}

var dockerVersions = []packageVersion{
	// 17.03.2 - k8s 1.8

	// 17.03.2 - Debian9 (stretch)
	{
		PackageVersion: "17.03.2",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionDebian9},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "17.03.2~ce-0~debian-stretch",
		Source:         "http://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce_17.03.2~ce-0~debian-stretch_amd64.deb",
		Hash:           "6f19489aba744dc02ce5fd9a65c0a2e3049b9f7a61cf70747ce33752094b0961",
		MarkImmutable:  []string{"/usr/bin/docker-runc"},
	},

	// 17.03.2 - Xenial
	{
		PackageVersion: "17.03.2",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionXenial},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "17.03.2~ce-0~ubuntu-xenial",
		Source:         "http://download.docker.com/linux/ubuntu/dists/xenial/pool/stable/amd64/docker-ce_17.03.2~ce-0~ubuntu-xenial_amd64.deb",
		Hash:           "68851f4a395c63b79b34e17ba5582379621389bbc9ea53cf34f70ea9839888fb",
		MarkImmutable:  []string{"/usr/bin/docker-runc"},
	},

	// 17.03.2 - Ubuntu Bionic via binary download (no packages available)
	{
		PackageVersion: "17.03.2",
		PlainBinary:    true,
		Distros:        []distros.Distribution{distros.DistributionBionic},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Source:         "http://download.docker.com/linux/static/stable/x86_64/docker-17.03.2-ce.tgz",
		Hash:           "183b31b001e7480f3c691080486401aa519101a5cfe6e05ad01b9f5521c4112d",
		MarkImmutable:  []string{"/usr/bin/docker-runc"},
	},

	// 17.03.2 - Centos / Rhel7 (two packages)
	{
		PackageVersion: "17.03.2",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7, distros.DistributionAmazonLinux2},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "17.03.2.ce",
		Source:         "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-17.03.2.ce-1.el7.centos.x86_64.rpm",
		Hash:           "0ead9d0db5c15e3123d3194f71f716a1d6e2a70c984b12a5dde4a72e6e483aca",
		ExtraPackages: map[string]packageInfo{
			"docker-ce-selinux": {
				Version: "17.03.2.ce",
				Source:  "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-selinux-17.03.2.ce-1.el7.centos.noarch.rpm",
				Hash:    "07e6cbaf0133468769f5bc7b8b14b2ef72b812ce62948be0989a2ea28463e4df",
			},
		},
		MarkImmutable: []string{"/usr/bin/docker-runc"},
	},
	// 17.09.0 - k8s 1.8

	// 17.09.0 - Debian9 (stretch)
	{
		PackageVersion: "17.09.0",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionDebian9},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "17.09.0~ce-0~debian",
		Source:         "http://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce_17.09.0~ce-0~debian_amd64.deb",
		Hash:           "80aa1429dc4d57eb6d73c291ab5feff5005f21d8402b1979e1e49db06eef52b0",
	},

	// 17.09.0 - Xenial
	{
		PackageVersion: "17.09.0",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionXenial},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "17.09.0~ce-0~ubuntu",
		Source:         "http://download.docker.com/linux/ubuntu/dists/xenial/pool/stable/amd64/docker-ce_17.09.0~ce-0~ubuntu_amd64.deb",
		Hash:           "d33f6eb134f0ab0876148bd96de95ea47d583d7f2cddfdc6757979453f9bd9bf",
	},

	// 18.06.2 - Xenial
	{
		PackageVersion: "18.06.2",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionXenial},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "18.06.2~ce~3-0~ubuntu",
		Source:         "https://download.docker.com/linux/ubuntu/dists/xenial/pool/stable/amd64/docker-ce_18.06.2~ce~3-0~ubuntu_amd64.deb",
		Hash:           "1c52a80430d4dda213a01e6859e7c403b4bebe642accaa6358f5c75f5f2ba682",
	},

	// 18.06.3 - Xenial
	{
		PackageVersion: "18.06.3",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionXenial},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "18.06.3~ce~3-0~ubuntu",
		Source:         "https://download.docker.com/linux/ubuntu/dists/xenial/pool/stable/amd64/docker-ce_18.06.3~ce~3-0~ubuntu_amd64.deb",
		Hash:           "6e9da7303cfa7ef7d4d8035bdc205229dd84e572f29957a9fb36e1351fe88a24",
	},

	// 17.09.0 - Centos / Rhel7 (two packages)
	{
		PackageVersion: "17.09.0",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7, distros.DistributionAmazonLinux2},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "17.09.0.ce",
		Source:         "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-17.09.0.ce-1.el7.centos.x86_64.rpm",
		Hash:           "be342f205c3fc99258e3903bfd3c79dc7f7c337c9321b217f4789dfdfbcac8f9",
	},

	// 18.03.1 - Bionic
	{
		PackageVersion: "18.03.1",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionBionic},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "18.03.1~ce~3-0~ubuntu",
		Source:         "https://download.docker.com/linux/ubuntu/dists/bionic/pool/stable/amd64/docker-ce_18.03.1~ce~3-0~ubuntu_amd64.deb",
		Hash:           "a8d69913a38df46d768f5d4e87e1230d6a1b7ccb4f9098a4fd9357a518f34be0",
	},

	// 18.06.2 - Bionic
	{
		PackageVersion: "18.06.2",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionBionic},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "18.06.2~ce~3-0~ubuntu",
		Source:         "https://download.docker.com/linux/ubuntu/dists/bionic/pool/stable/amd64/docker-ce_18.06.2~ce~3-0~ubuntu_amd64.deb",
		Hash:           "056afb4440b8f2ae52841ee228d7794176fcb81aae0ba5614ecb7b4de6e4db9d",
	},

	// 18.06.1 - Debian Stretch
	{
		PackageVersion: "18.06.1",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionDebian9},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "18.06.1~ce~3-0~debian",
		Source:         "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce_18.06.1~ce~3-0~debian_amd64.deb",
		Hash:           "00a09a8993efd8095bd1817442db86c27de9720d7d5ade36aa52cd91198fa83d",
	},

	// 18.06.2 - Debian Stretch
	{

		PackageVersion: "18.06.2",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionDebian9},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "18.06.2~ce~3-0~debian",
		Source:         "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce_18.06.2~ce~3-0~debian_amd64.deb",
		Hash:           "cbbd2afc85b2a46d55abfd5d362595e39a54022b6c6baab0a5ddc4a85a74e318",
	},

	// 18.06.3 - Debian Buster
	{

		PackageVersion: "18.06.3",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionDebian10},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "18.06.3~ce~3-0~debian",
		Source:         "https://download.docker.com/linux/debian/dists/buster/pool/stable/amd64/docker-ce_18.06.3~ce~3-0~debian_amd64.deb",
		Hash:           "0c8ca09635553f0c6cb70a08bdef6f3b8d89b1247e4dab54896c93aad3bf3f25",
	},

	// 18.06.1 - CentOS / Rhel7 (two packages)
	{
		PackageVersion: "18.06.1",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7, distros.DistributionAmazonLinux2},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "18.06.1.ce",
		Source:         "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-18.06.1.ce-3.el7.x86_64.rpm",
		Hash:           "352909b3df327d10a6ee27e2c6ee8638d90481ee93580ae79c9d1ff7530a196e",
	},

	// 18.09.3 - Debian Stretch
	{
		PackageVersion: "18.09.3",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionDebian9},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "5:18.09.3~3-0~debian-stretch",
		Source:         "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce_18.09.3~3-0~debian-stretch_amd64.deb",
		Hash:           "a941c03d0e7027481e4ff6cd5c77b871c4bf97df76e6444396e004adb759795d",
		ExtraPackages: map[string]packageInfo{
			"docker-ce-cli": {
				Version: "5:18.09.3~3-0~debian-stretch",
				Source:  "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce-cli_18.09.3~3-0~debian-stretch_amd64.deb",
				Hash:    "6102a5de3d1039226fd3d7ec44316371455efb211cfaacda8346d8d5155ffb0c",
			},
		},
	},

	// 18.06.2 - CentOS / Rhel7 (two packages)
	{
		PackageVersion: "18.06.2",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7, distros.DistributionAmazonLinux2},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "18.06.2.ce",
		Source:         "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-18.06.2.ce-3.el7.x86_64.rpm",
		Hash:           "0e5d98c359d93e8a892a07ab1f8eb8153964b535cadda61a8791ca2db3c6b76c",
	},

	// 18.06.3 (contains fix for CVE-2019-5736)

	// 18.06.3 - Bionic
	{
		PackageVersion: "18.06.3",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionBionic},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "18.06.3~ce~3-0~ubuntu",
		Source:         "https://download.docker.com/linux/ubuntu/dists/bionic/pool/stable/amd64/docker-ce_18.06.3~ce~3-0~ubuntu_amd64.deb",
		Hash:           "f8cc02112a125007f5c70f009ce9a91dd536018f139131074ee55cea555ba85d",
	},

	// 18.06.3 - Debian Stretch
	{

		PackageVersion: "18.06.3",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionDebian9},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "18.06.3~ce~3-0~debian",
		Source:         "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce_18.06.3~ce~3-0~debian_amd64.deb",
		Hash:           "0de184cc79d9f9c99b2a6fa4fdd8b29645e9a858106a9814bb11047073a4e8cb",
	},

	// 18.06.3 - CentOS / Rhel7 (two packages)
	{
		PackageVersion: "18.06.3",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7, distros.DistributionAmazonLinux2},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "18.06.3.ce",
		Source:         "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-18.06.3.ce-3.el7.x86_64.rpm",
		Hash:           "f3703698cab918ab41b1244f699c8718a5e3bf4070fdf4894b5b6e8d92545a62",
	},
	// 18.06.3 - CentOS / Rhel8 (two packages)
	{
		PackageVersion: "18.06.3",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionRhel8, distros.DistributionCentos8},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "18.06.3.ce",
		Source:         "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-18.06.3.ce-3.el7.x86_64.rpm",
		Hash:           "f3703698cab918ab41b1244f699c8718a5e3bf4070fdf4894b5b6e8d92545a62",
	},

	// 18.09.9 - k8s 1.14 - https://github.com/kubernetes/kubernetes/pull/72823

	// 18.09.9 - Debian Stretch
	{
		PackageVersion: "18.09.9",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionDebian9},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "5:18.09.9~3-0~debian-stretch",
		Source:         "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce_18.09.9~3-0~debian-stretch_amd64.deb",
		Hash:           "53d9d25bb7d55c05a6c5829606122257ada8863ccb222ff0293fcf1d75990058",
		ExtraPackages: map[string]packageInfo{
			"docker-ce-cli": {
				Version: "5:18.09.9~3-0~debian-stretch",
				Source:  "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce-cli_18.09.9~3-0~debian-stretch_amd64.deb",
				Hash:    "1cc46c8634704e192f402844747a82b986b2461beb3da748f4ca6a36918e6442",
			},
		},
	},

	// 18.09.9 - Debian Buster
	{
		PackageVersion: "18.09.9",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionDebian10},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "5:18.09.9~3-0~debian-buster",
		Source:         "https://download.docker.com/linux/debian/dists/buster/pool/stable/amd64/docker-ce_18.09.9~3-0~debian-buster_amd64.deb",
		Hash:           "b0f4ce24089593ef6335e53e4c78d619a58539492121340da963c1a88687a059",
		ExtraPackages: map[string]packageInfo{
			"docker-ce-cli": {
				Version: "5:18.09.9~3-0~debian-buster",
				Source:  "https://download.docker.com/linux/debian/dists/buster/pool/stable/amd64/docker-ce-cli_18.09.9~3-0~debian-buster_amd64.deb",
				Hash:    "e2b0543de09206072691c0c09fc2ad64acea988eb56e31e3bd02889f1435befd",
			},
		},
	},

	// 18.09.9 - Xenial
	{
		PackageVersion: "18.09.9",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionXenial},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "5:18.09.9~3-0~ubuntu-xenial",
		Source:         "https://download.docker.com/linux/ubuntu/dists/xenial/pool/stable/amd64/docker-ce_18.09.9~3-0~ubuntu-xenial_amd64.deb",
		Hash:           "30885e58747eff619dc22b074307e21bc176c71396c5d54a32764ffcc359beaf",
		ExtraPackages: map[string]packageInfo{
			"docker-ce-cli": {
				Version: "5:18.09.9~3-0~ubuntu-xenial",
				Source:  "https://download.docker.com/linux/ubuntu/dists/xenial/pool/stable/amd64/docker-ce-cli_18.09.9~3-0~ubuntu-xenial_amd64.deb",
				Hash:    "927c6df4fd2bc380be4f315169114cfd34d53856df004eeac3de35360f3eca9f",
			},
		},
	},

	// 18.09.9 - Bionic
	{
		PackageVersion: "18.09.9",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionBionic},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "5:18.09.9~3-0~ubuntu-bionic",
		Source:         "https://download.docker.com/linux/ubuntu/dists/bionic/pool/stable/amd64/docker-ce_18.09.9~3-0~ubuntu-bionic_amd64.deb",
		Hash:           "95160362599c506375c36f324f00404ad066ab4d94c840336781b5930d893467",
		ExtraPackages: map[string]packageInfo{
			"docker-ce-cli": {
				Version: "5:18.09.9~3-0~ubuntu-bionic",
				Source:  "https://download.docker.com/linux/ubuntu/dists/bionic/pool/stable/amd64/docker-ce-cli_18.09.9~3-0~ubuntu-bionic_amd64.deb",
				Hash:    "10abf1e3c25882d5a099ffda2a5a54168f600eb3e056b67c4fa4e20ecf5a03df",
			},
		},
	},

	// 18.09.9 - CentOS7 / Rhel7
	{
		PackageVersion: "18.09.9",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "18.09.9",
		Source:         "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-18.09.9-3.el7.x86_64.rpm",
		Hash:           "f4be41bf8093c076462a9a2d7669d1b3158e4c3799759dbf9689b77de49385a8",
		ExtraPackages: map[string]packageInfo{
			"docker-ce-cli": {
				Version: "18.09.9",
				Source:  "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-cli-18.09.9-3.el7.x86_64.rpm",
				Hash:    "b1658ece6b8524a9c23a8623a7485b361c61a49ba887b51d9cc4ef58cfeb878a",
			},
		},
	},

	// 18.09.9 - CentOS / Rhel8
	{
		PackageVersion: "18.09.9",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionRhel8, distros.DistributionCentos8},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "18.09.9",
		Source:         "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-18.09.9-3.el7.x86_64.rpm",
		Hash:           "f4be41bf8093c076462a9a2d7669d1b3158e4c3799759dbf9689b77de49385a8",
		ExtraPackages: map[string]packageInfo{
			"docker-ce-cli": {
				Version: "18.09.9",
				Source:  "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-cli-18.09.9-3.el7.x86_64.rpm",
				Hash:    "b1658ece6b8524a9c23a8623a7485b361c61a49ba887b51d9cc4ef58cfeb878a",
			},
		},
	},

	// 18.09.9 - Linux Generic
	//
	// * AmazonLinux2: the Centos7 package depends on container-selinux, but selinux isn't used on amazonlinux2
	// * UbuntuFocal: no focal version available at download.docker.com
	{
		PackageVersion: "18.09.9",
		PlainBinary:    true,
		Distros:        []distros.Distribution{distros.DistributionAmazonLinux2, distros.DistributionFocal},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Source:         "https://download.docker.com/linux/static/stable/x86_64/docker-18.09.9.tgz",
		Hash:           "82a362af7689038c51573e0fd0554da8703f0d06f4dfe95dd5bda5acf0ae45fb",
	},

	// 19.03.4 - k8s 1.17 - https://github.com/kubernetes/kubernetes/pull/84476

	// 19.03.4 - Linux Generic
	//
	// * AmazonLinux2: the Centos7 package depends on container-selinux, but selinux isn't used on amazonlinux2
	// * UbuntuFocal: no focal version available at download.docker.com
	{
		PackageVersion: "19.03.4",
		PlainBinary:    true,
		Distros:        []distros.Distribution{distros.DistributionAmazonLinux2, distros.DistributionFocal},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Source:         "https://download.docker.com/linux/static/stable/x86_64/docker-19.03.4.tgz",
		Hash:           "efef2ad32d262674501e712351be0df9dd31d6034b175d0020c8f5d5c9c3fd10",
	},

	// 19.03.4 - Debian Stretch
	{
		PackageVersion: "19.03.4",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionDebian9},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "5:19.03.4~3-0~debian-stretch",
		Source:         "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce_19.03.4~3-0~debian-stretch_amd64.deb",
		Hash:           "a5fedef212914c443ed71c9ba2fbe0cdf39e0a6e2da8dfcc29881c6c536877ce",
		ExtraPackages: map[string]packageInfo{
			"docker-ce-cli": {
				Version: "5:19.03.4~3-0~debian-stretch",
				Source:  "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/docker-ce-cli_19.03.4~3-0~debian-stretch_amd64.deb",
				Hash:    "f0f3c9c91a9482b0fe120cd9e404c3ade342ce01d0d98a7f6bce3e16b7c57a11",
			},
		},
	},

	// 19.03.4 - Debian Buster
	{
		PackageVersion: "19.03.4",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionDebian10},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "5:19.03.4~3-0~debian-buster",
		Source:         "https://download.docker.com/linux/debian/dists/buster/pool/stable/amd64/docker-ce_19.03.4~3-0~debian-buster_amd64.deb",
		Hash:           "cdd9d2147a6f6c9c38a6addfdd56d7d65d688a83f44ff3a289de7e15c796b87c",
		ExtraPackages: map[string]packageInfo{
			"docker-ce-cli": {
				Version: "5:19.03.4~3-0~debian-buster",
				Source:  "https://download.docker.com/linux/debian/dists/buster/pool/stable/amd64/docker-ce-cli_19.03.4~3-0~debian-buster_amd64.deb",
				Hash:    "92c681c324f3d24517dc25daf9f4cd52034a24a72bb98827a4bcf4f6b56e6088",
			},
		},
	},

	// 19.03.4 - Xenial
	{
		PackageVersion: "19.03.4",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionXenial},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "5:19.03.4~3-0~ubuntu-xenial",
		Source:         "https://download.docker.com/linux/ubuntu/dists/xenial/pool/stable/amd64/docker-ce_19.03.4~3-0~ubuntu-xenial_amd64.deb",
		Hash:           "7bf9d7c3127dc910b8364c5799c667ff8a45e4c8bd859f908ea4a66944312ff3",
		ExtraPackages: map[string]packageInfo{
			"docker-ce-cli": {
				Version: "5:19.03.4~3-0~ubuntu-xenial",
				Source:  "https://download.docker.com/linux/ubuntu/dists/xenial/pool/stable/amd64/docker-ce-cli_19.03.4~3-0~ubuntu-xenial_amd64.deb",
				Hash:    "00622505c8f47e0b711ba7f7582473d55b38dd8d7bae20d286aa473595c5f6cf",
			},
		},
	},

	// 19.03.4 - Bionic
	{
		PackageVersion: "19.03.4",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionBionic},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "5:19.03.4~3-0~ubuntu-bionic",
		Source:         "https://download.docker.com/linux/ubuntu/dists/bionic/pool/stable/amd64/docker-ce_19.03.4~3-0~ubuntu-bionic_amd64.deb",
		Hash:           "31ee4b40cc6b76966318e007a1c7cedd64c6a3dd957de1de40734eb06320b8d3",
		ExtraPackages: map[string]packageInfo{
			"docker-ce-cli": {
				Version: "5:19.03.4~3-0~ubuntu-bionic",
				Source:  "https://download.docker.com/linux/ubuntu/dists/bionic/pool/stable/amd64/docker-ce-cli_19.03.4~3-0~ubuntu-bionic_amd64.deb",
				Hash:    "d364ba24b3756c5e1f7b860cef5361ce717a99bb982aa76dbd6d8a928a2de056",
			},
		},
	},

	// 19.03.4 - CentOS / Rhel7
	{
		PackageVersion: "19.03.4",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "19.03.4",
		Source:         "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-19.03.4-3.el7.x86_64.rpm",
		Hash:           "46ebc08b3740bfb532f686a143e144a4c73ddcd600e83104ae4617b301b83f42",
		ExtraPackages: map[string]packageInfo{
			"docker-ce-cli": {
				Version: "19.03.4",
				Source:  "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-cli-19.03.4-3.el7.x86_64.rpm",
				Hash:    "1b34e1dd1ec5af7e0e37e80bb1ddf0e36006639e8964cf8fc308683f90d38b7a",
			},
		},
	},

	// 19.03.4 - CentOS / Rhel8
	{
		PackageVersion: "19.03.4",
		Name:           "docker-ce",
		Distros:        []distros.Distribution{distros.DistributionRhel8, distros.DistributionCentos8},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "19.03.4",
		Source:         "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-19.03.4-3.el7.x86_64.rpm",
		Hash:           "46ebc08b3740bfb532f686a143e144a4c73ddcd600e83104ae4617b301b83f42",
		ExtraPackages: map[string]packageInfo{
			"docker-ce-cli": {
				Version: "19.03.4",
				Source:  "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-cli-19.03.4-3.el7.x86_64.rpm",
				Hash:    "1b34e1dd1ec5af7e0e37e80bb1ddf0e36006639e8964cf8fc308683f90d38b7a",
			},
		},
	},

	// 19.03.8 - Linux Generic
	{
		PackageVersion: "19.03.8",
		PlainBinary:    true,
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Source:         "https://download.docker.com/linux/static/stable/x86_64/docker-19.03.8.tgz",
		Hash:           "7f4115dc6a3c19c917f8b9664d7b51c904def1c984e082c4600097433323cf6f",
	},

	// 19.03.11 - Linux Generic
	{
		PackageVersion: "19.03.11",
		PlainBinary:    true,
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Source:         "https://download.docker.com/linux/static/stable/x86_64/docker-19.03.11.tgz",
		Hash:           "0f4336378f61ed73ed55a356ac19e46699a995f2aff34323ba5874d131548b9e",
	},

	// TIP: When adding the next version, copy the previous version, string replace the version and run:
	//   VERIFY_HASHES=1 go test -v ./nodeup/pkg/model -run TestDockerPackageHashes
	// (you might want to temporarily comment out older versions on a slower connection and then validate)
}

func (b *DockerBuilder) dockerVersion() (string, error) {
	dockerVersion := ""
	if b.Cluster.Spec.Docker != nil {
		dockerVersion = fi.StringValue(b.Cluster.Spec.Docker.Version)
	}
	if dockerVersion == "" {
		return "", fmt.Errorf("error finding Docker version")
	}
	return dockerVersion, nil
}

// Build is responsible for configuring the docker daemon
func (b *DockerBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.skipInstall() {
		klog.Infof("SkipInstall is set to true; won't install Docker")
		return nil
	}

	// @check: neither flatcar nor containeros need provision docker.service, just the docker daemon options
	switch b.Distribution {
	case distros.DistributionFlatcar:
		klog.Infof("Detected Flatcar; won't install Docker")
		if err := b.buildContainerOSConfigurationDropIn(c); err != nil {
			return err
		}
		return nil

	case distros.DistributionContainerOS:
		klog.Infof("Detected ContainerOS; won't install Docker")
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

	dockerVersion, err := b.dockerVersion()
	if err != nil {
		return err
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

			var packageTask fi.Task
			if dv.PlainBinary {
				packageTask = &nodetasks.Archive{
					Name:      "docker-ce",
					Source:    dv.Source,
					Hash:      dv.Hash,
					TargetDir: "/",
					MapFiles: map[string]string{
						"docker/docker*": "/usr/bin",
					},
				}
				c.AddTask(packageTask)

				c.AddTask(b.buildDockerGroup())
				if b.Distribution.IsDebianFamily() {
					c.AddTask(b.buildSystemdSocket())
				}
			} else {
				var extraPkgs []*nodetasks.Package
				for name, pkg := range dv.ExtraPackages {
					dep := &nodetasks.Package{
						Name:         name,
						Version:      s(pkg.Version),
						Source:       s(pkg.Source),
						Hash:         s(pkg.Hash),
						PreventStart: fi.Bool(true),
					}
					extraPkgs = append(extraPkgs, dep)
				}
				packageTask = &nodetasks.Package{
					Name:    dv.Name,
					Version: s(dv.Version),
					Source:  s(dv.Source),
					Hash:    s(dv.Hash),
					Deps:    extraPkgs,

					// TODO: PreventStart is now unused?
					PreventStart: fi.Bool(true),
				}
				c.AddTask(packageTask)
			}

			// As a mitigation for CVE-2019-5736 (possibly a fix, definitely defense-in-depth) we chattr docker-runc to be immutable
			for _, f := range dv.MarkImmutable {
				c.AddTask(&nodetasks.Chattr{
					File: f,
					Mode: "+i",
					Deps: []fi.Task{packageTask},
				})
			}

			for _, dep := range dv.Dependencies {
				c.AddTask(&nodetasks.Package{Name: dep})
			}

			// Note we do _not_ stop looping... centos/rhel comprises multiple packages
		}

		if count == 0 {
			klog.Warningf("Did not find docker package for %s %s %s", b.Distribution, b.Architecture, dockerVersion)
		}
	}

	// Split into major.minor.(patch+pr+meta)
	parts := strings.SplitN(dockerVersion, ".", 3)
	if len(parts) != 3 {
		return fmt.Errorf("error parsing docker version %q, no Major.Minor.Patch elements found", dockerVersion)
	}

	// Validate major
	dockerVersionMajor, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("error parsing major docker version %q: %v", parts[0], err)
	}

	// Validate minor
	dockerVersionMinor, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("error parsing minor docker version %q: %v", parts[1], err)
	}

	c.AddTask(b.buildSystemdService(dockerVersionMajor, dockerVersionMinor))

	if err := b.buildSysconfig(c); err != nil {
		return err
	}

	// Enable health-check
	if b.healthCheck() || (b.IsKubernetesLT("1.18") && b.Distribution.IsDebianFamily()) {
		c.AddTask(b.buildSystemdHealthCheckScript())
		c.AddTask(b.buildSystemdHealthCheckService())
		c.AddTask(b.buildSystemdHealthCheckTimer())
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
	klog.V(8).Infof("Built docker.socket manifest\n%s", manifestString)

	service := &nodetasks.Service{
		Name:       "docker.socket",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

func (b *DockerBuilder) buildSystemdService(dockerVersionMajor int, dockerVersionMinor int) *nodetasks.Service {
	usesDockerSocket := true

	var dockerdCommand = "/usr/bin/dockerd"

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Docker Application Container Engine")
	manifest.Set("Unit", "Documentation", "https://docs.docker.com")

	if b.Distribution.IsRHELFamily() {
		// See https://github.com/docker/docker/pull/24804
		usesDockerSocket = false
	}

	if usesDockerSocket {
		manifest.Set("Unit", "BindsTo", "containerd.service")
		manifest.Set("Unit", "After", "network-online.target firewalld.service containerd.service")
		manifest.Set("Unit", "Wants", "network-online.target")
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

	manifest.Set("Service", "ExecReload", "/bin/kill -s HUP $MAINPID")
	// kill only the docker process, not all processes in the cgroup
	manifest.Set("Service", "KillMode", "process")

	manifest.Set("Service", "TimeoutStartSec", "0")

	manifest.Set("Service", "LimitNOFILE", "infinity")
	manifest.Set("Service", "LimitNPROC", "infinity")
	manifest.Set("Service", "LimitCORE", "infinity")

	manifest.Set("Service", "TasksMax", "infinity")

	manifest.Set("Service", "Restart", "always")
	manifest.Set("Service", "RestartSec", "2s")
	manifest.Set("Service", "StartLimitInterval", "0")

	// set delegate yes so that systemd does not reset the cgroups of docker containers
	manifest.Set("Service", "Delegate", "yes")

	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built service manifest %q\n%s", "docker", manifestString)

	service := &nodetasks.Service{
		Name:       "docker.service",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

func (b *DockerBuilder) buildSystemdHealthCheckScript() *nodetasks.File {
	script := &nodetasks.File{
		Path:     "/opt/kops/bin/docker-healthcheck",
		Contents: fi.NewStringResource(resources.DockerHealthCheck),
		Type:     nodetasks.FileType_File,
		Mode:     s("0755"),
	}

	return script
}

func (b *DockerBuilder) buildSystemdHealthCheckService() *nodetasks.Service {
	manifest := &systemd.Manifest{}

	manifest.Set("Unit", "Description", "Run docker-healthcheck once")
	manifest.Set("Unit", "Documentation", "https://kops.sigs.k8s.io")
	manifest.Set("Service", "Type", "oneshot")
	manifest.Set("Service", "ExecStart", "/opt/kops/bin/docker-healthcheck")
	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built service manifest %q\n%s", "docker-healthcheck.service", manifestString)

	service := &nodetasks.Service{
		Name:       "docker-healthcheck.service",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

func (b *DockerBuilder) buildSystemdHealthCheckTimer() *nodetasks.Service {
	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Trigger docker-healthcheck periodically")
	manifest.Set("Unit", "Documentation", "https://kops.sigs.k8s.io")
	manifest.Set("Timer", "OnUnitInactiveSec", "10s")
	manifest.Set("Timer", "Unit", "docker-healthcheck.service")
	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built timer manifest %q\n%s", "docker-healthcheck.timer", manifestString)

	service := &nodetasks.Service{
		Name:       "docker-healthcheck.timer",
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

	// Equivalent of https://github.com/kubernetes/kubernetes/pull/51986
	lines = append(lines, "TasksMax=infinity")

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
				klog.V(2).Infof("/etc/docker/daemon.json not found")
			} else {
				klog.Warningf("error reading /etc/docker/daemon.json: %v", err)
			}
		} else {
			// Maybe we get smarter here?
			data := make(map[string]interface{})
			if err := json.Unmarshal(b, &data); err != nil {
				klog.Warningf("error deserializing /etc/docker/daemon.json: %v", err)
			} else {
				storageDriver := data["storage-driver"]
				klog.Infof("/etc/docker/daemon.json has storage-driver: %q", storageDriver)
			}
			docker.Storage = nil
		}
	}

	// RHEL-family / docker has a bug with 17.x where it fails to use overlay2 because it does a broken kernel check
	if b.Distribution.IsRHELFamily() {
		dockerVersion, err := b.dockerVersion()
		if err != nil {
			return err
		}
		if strings.HasPrefix(dockerVersion, "17.") {
			storageOpts := strings.Join(docker.StorageOpts, ",")
			if strings.Contains(storageOpts, "overlay2.override_kernel_check=1") {
				// Already there
			} else if !strings.Contains(storageOpts, "overlay2.override_kernel_check") {
				docker.StorageOpts = append(docker.StorageOpts, "overlay2.override_kernel_check=1")
			} else {
				klog.Infof("detected image was RHEL and overlay2.override_kernel_check=1 was probably needed, but overlay2.override_kernel_check was already set (%q) so won't set", storageOpts)
			}
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

// skipInstall determines if kops should skip the installation and configuration of Docker
func (b *DockerBuilder) skipInstall() bool {
	d := b.Cluster.Spec.Docker

	// don't skip install if the user hasn't specified anything
	if d == nil {
		return false
	}

	return d.SkipInstall
}

// healthCheck determines if kops should enable the health-check for Docker
func (b *DockerBuilder) healthCheck() bool {
	d := b.Cluster.Spec.Docker

	// don't enable the health-check if the user hasn't specified anything
	if d == nil {
		return false
	}

	return d.HealthCheck
}
