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
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/distributions"

	"k8s.io/klog/v2"
)

// PackagesBuilder adds miscellaneous OS packages that we need
type PackagesBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &PackagesBuilder{}

// Build is responsible for installing packages
func (b *PackagesBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	// kubelet needs:
	//   conntrack  - kops #5671
	//   ebtables - kops #1711
	//   ethtool - kops #1830
	if b.Distribution.IsDebianFamily() {
		// From containerd: https://github.com/containerd/cri/blob/master/contrib/ansible/tasks/bootstrap_ubuntu.yaml
		c.AddTask(&nodetasks.Package{Name: "bridge-utils"})
		if (b.Distribution.IsDebian() && b.Distribution.Version() < 13) || (b.Distribution.IsUbuntu() && b.Distribution.Version() < 25.10) {
			c.AddTask(&nodetasks.Package{Name: "cgroupfs-mount"})
		}
		c.AddTask(&nodetasks.Package{Name: "conntrack"})
		c.AddTask(&nodetasks.Package{Name: "ebtables"})
		c.AddTask(&nodetasks.Package{Name: "ethtool"})
		c.AddTask(&nodetasks.Package{Name: "iptables"})
		c.AddTask(&nodetasks.Package{Name: "libapparmor1"})
		c.AddTask(&nodetasks.Package{Name: "libseccomp2"})
		c.AddTask(&nodetasks.Package{Name: "libltdl7"})
		c.AddTask(&nodetasks.Package{Name: "pigz"})
		c.AddTask(&nodetasks.Package{Name: "socat"})
		c.AddTask(&nodetasks.Package{Name: "util-linux"})
		// Additional packages
		for _, additionalPackage := range b.NodeupConfig.Packages {
			c.EnsureTask(&nodetasks.Package{Name: additionalPackage})
		}
	} else if b.Distribution.IsRHELFamily() {
		// From containerd: https://github.com/containerd/cri/blob/master/contrib/ansible/tasks/bootstrap_centos.yaml
		c.AddTask(&nodetasks.Package{Name: "conntrack-tools"})
		c.AddTask(&nodetasks.Package{Name: "ebtables"})
		c.AddTask(&nodetasks.Package{Name: "ethtool"})
		if b.Distribution == distributions.DistributionAmazonLinux2023 {
			// install iptables-nft in al2023 (NOT the iptables-legacy!)
			c.AddTask(&nodetasks.Package{Name: "iptables-nft"})
		} else {
			c.AddTask(&nodetasks.Package{Name: "iptables"})
		}
		c.AddTask(&nodetasks.Package{Name: "libseccomp"})
		c.AddTask(&nodetasks.Package{Name: "libtool-ltdl"})
		c.AddTask(&nodetasks.Package{Name: "socat"})
		c.AddTask(&nodetasks.Package{Name: "util-linux"})
		// Handle some packages differently for each distro
		// Amazon Linux 2 doesn't have SELinux enabled by default
		if b.Distribution != distributions.DistributionAmazonLinux2 {
			c.AddTask(&nodetasks.Package{Name: "container-selinux"})
			c.AddTask(&nodetasks.Package{Name: "pigz"})
		}
		// RHEL9 and Rocky 9 do not have libcgroup
		if b.Distribution != distributions.DistributionRhel9 && b.Distribution != distributions.DistributionRocky9 {
			c.AddTask(&nodetasks.Package{Name: "libcgroup"})
		}
		// Additional packages
		for _, additionalPackage := range b.NodeupConfig.Packages {
			c.EnsureTask(&nodetasks.Package{Name: additionalPackage})
		}
	} else {
		// Hopefully they are already installed
		klog.Warningf("unknown distribution, skipping required packages install: %v", b.Distribution)
	}

	return nil
}
