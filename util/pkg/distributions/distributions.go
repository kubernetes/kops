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

package distributions

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
)

// Distribution represents a particular version of an operating system.
// This enables OS-dependent logic.
type Distribution struct {
	// packageFormat is the packaging format used by this distro; either deb or rpm, or "" for immutable OSes
	packageFormat string

	// project is the entity that produces the distribution e.g. "debian" or "ubuntu" or "rhel" or "centos"
	project string

	// id is the name of the actual distribution version e.g. "buster" or "xenial"
	id string

	// version is a numeric identifier for comparison purposes within a particular project
	version float32
}

var (
	DistributionDebian10        = Distribution{packageFormat: "deb", project: "debian", id: "buster", version: 10}
	DistributionDebian11        = Distribution{packageFormat: "deb", project: "debian", id: "bullseye", version: 11}
	DistributionDebian12        = Distribution{packageFormat: "deb", project: "debian", id: "bookworm", version: 12}
	DistributionUbuntu2004      = Distribution{packageFormat: "deb", project: "ubuntu", id: "focal", version: 20.04}
	DistributionUbuntu2010      = Distribution{packageFormat: "deb", project: "ubuntu", id: "groovy", version: 20.10}
	DistributionUbuntu2104      = Distribution{packageFormat: "deb", project: "ubuntu", id: "hirsute", version: 21.04}
	DistributionUbuntu2110      = Distribution{packageFormat: "deb", project: "ubuntu", id: "impish", version: 21.10}
	DistributionUbuntu2204      = Distribution{packageFormat: "deb", project: "ubuntu", id: "jammy", version: 22.04}
	DistributionAmazonLinux2    = Distribution{packageFormat: "rpm", project: "amazonlinux2", id: "amazonlinux2", version: 0}
	DistributionAmazonLinux2023 = Distribution{packageFormat: "rpm", project: "amazonlinux2023", id: "amzn", version: 2023}
	DistributionRhel8           = Distribution{packageFormat: "rpm", project: "rhel", id: "rhel8", version: 8}
	DistributionRhel9           = Distribution{packageFormat: "rpm", project: "rhel", id: "rhel9", version: 9}
	DistributionRocky8          = Distribution{packageFormat: "rpm", project: "rocky", id: "rocky8", version: 8}
	DistributionFlatcar         = Distribution{packageFormat: "", project: "flatcar", id: "flatcar", version: 0}
	DistributionContainerOS     = Distribution{packageFormat: "", project: "containeros", id: "containeros", version: 0}
)

// IsDebianFamily returns true if this distribution uses deb packages and generally follows debian package names
func (d *Distribution) IsDebianFamily() bool {
	return d.packageFormat == "deb"
}

// IsUbuntu returns true if this distribution is Ubuntu (but not debian)
func (d *Distribution) IsUbuntu() bool {
	return d.project == "ubuntu"
}

// IsRHELFamily returns true if this distribution uses rpm packages and generally follows rhel package names
func (d *Distribution) IsRHELFamily() bool {
	return d.packageFormat == "rpm"
}

// IsSystemd returns true if this distribution uses systemd
func (d *Distribution) IsSystemd() bool {
	return true
}

// DefaultUsers returns the name of the system users for this distribution
func (d *Distribution) DefaultUsers() ([]string, error) {
	switch d.project {
	case "debian":
		return []string{"admin", "root"}, nil
	case "ubuntu":
		return []string{"ubuntu", "root"}, nil
	case "centos":
		return []string{"centos"}, nil
	case "rhel", "amazonlinux2", "amazonlinux2023":
		return []string{"ec2-user"}, nil
	case "rocky":
		return []string{"rocky"}, nil
	case "flatcar":
		return []string{"core"}, nil
	default:
		return nil, fmt.Errorf("unknown distro %v", d)
	}
}

// HasLoopbackEtcResolvConf is true if systemd-resolved has put the loopback address 127.0.0.53 as a nameserver in /etc/resolv.conf
// See https://github.com/coredns/coredns/blob/master/plugin/loop/README.md#troubleshooting-loops-in-kubernetes-clusters
// Returns the recommended target if we should not use /etc/resolv.conf
//
// There are actually 3 main configurations:
//
// * "Classic" DNS:
//     upstream nameservers are in /etc/resolv.conf.
//     We can use /etc/resolv.conf
//
// * systemd-resolved without libnss_resolve:
//     /etc/resolv.conf is a symlink to /run/systemd/resolve/stub-resolv.conf
//     which contains 127.0.0.53 (that won't work in a container), and upstream
//     nameservers are in /run/systemd/resolve/resolv.conf.
//     We must use /run/systemd/resolve/resolv.conf
//
// * systemd-resolved with libnss_resolve:
//     /etc/resolv.conf has upstream nameservers. /etc/nsswitch.conf includes
//     "resolve", which is a direct-path to systemd-resolved.
//     We can use /etc/resolv.conf

func (d *Distribution) HasLoopbackEtcResolvConf() (string, bool) {
	resolvConfPath := "/etc/resolv.conf"

	// Check if it's a symlink
	fileInfo, err := os.Lstat(resolvConfPath)
	if err != nil {
		klog.Warningf("error from stat(%q): %v", resolvConfPath, err)
		return "", false
	}
	if fileInfo.Mode()&os.ModeSymlink == 0 {
		klog.Infof("resolver config %q is not a symlink, will use it for kubelet", resolvConfPath)
		return "", false
	}

	// Check if it's one of the known symlink targets
	dest, err := filepath.EvalSymlinks(resolvConfPath)
	if err != nil {
		klog.Warningf("error from EvalSymlinks(%q): %v", resolvConfPath, err)
		return "", false
	}

	if dest == "/run/systemd/resolve/stub-resolv.conf" {
		klog.Infof("detected systemd-resolved, will use %q for resolv.conf", "/run/systemd/resolve/resolv.conf")
		return "/run/systemd/resolve/resolv.conf", true
	}

	// Although this is a symlink, it's probably resolvconf, rather than systemd-resolved.
	// Thus the /etc/resolv.conf configuration will (probably) work from inside a container.
	klog.Warningf("detected symlink for %q => %q; not a symlink to systemd-resolved, so will not change resolver", resolvConfPath, dest)
	return "", false
}

// Version returns the (project scoped) numeric version
func (d *Distribution) Version() float32 {
	return d.version
}
