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
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"

	"k8s.io/klog"
)

// PackagesBuilder adds miscellaneous OS packages that we need
type PackagesBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &DockerBuilder{}

// Build is responsible for installing packages
func (b *PackagesBuilder) Build(c *fi.ModelBuilderContext) error {
	// kubelet needs:
	//   conntrack  - kops #5671
	//   ebtables - kops #1711
	//   ethtool - kops #1830
	if b.Distribution.IsDebianFamily() {
		// From containerd: https://github.com/containerd/cri/blob/master/contrib/ansible/tasks/bootstrap_ubuntu.yaml
		c.AddTask(&nodetasks.Package{Name: "bridge-utils"})
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
	} else if b.Distribution.IsRHELFamily() {
		// From containerd: https://github.com/containerd/cri/blob/master/contrib/ansible/tasks/bootstrap_centos.yaml
		c.AddTask(&nodetasks.Package{Name: "conntrack-tools"})
		c.AddTask(&nodetasks.Package{Name: "ebtables"})
		c.AddTask(&nodetasks.Package{Name: "ethtool"})
		c.AddTask(&nodetasks.Package{Name: "iptables"})
		c.AddTask(&nodetasks.Package{Name: "libseccomp"})
		c.AddTask(&nodetasks.Package{Name: "libtool-ltdl"})
		c.AddTask(&nodetasks.Package{Name: "socat"})
		c.AddTask(&nodetasks.Package{Name: "util-linux"})
		// Handle some packages differently for each distro
		switch b.Distribution {
		case distros.DistributionRhel7:
			// Easier to install container-selinux from CentOS than extras
			c.AddTask(&nodetasks.Package{
				Name:   "container-selinux",
				Source: s("http://vault.centos.org/7.6.1810/extras/x86_64/Packages/container-selinux-2.107-1.el7_6.noarch.rpm"),
				Hash:   s("7de4211fa0dfd240d8827b93763e1eb5f0d56411"),
			})
		case distros.DistributionAmazonLinux2:
			// Amazon Linux 2 doesn't have SELinux enabled by default
		default:
			c.AddTask(&nodetasks.Package{Name: "container-selinux"})
			c.AddTask(&nodetasks.Package{Name: "pigz"})
		}
	} else {
		// Hopefully they are already installed
		klog.Warningf("unknown distribution, skipping required packages install: %v", b.Distribution)
	}

	return nil
}
