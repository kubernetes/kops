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
	// Debian-family distros
	DistributionDebian10   = Distribution{packageFormat: "deb", project: "debian", id: "buster", version: 10}
	DistributionDebian11   = Distribution{packageFormat: "deb", project: "debian", id: "bullseye", version: 11}
	DistributionDebian12   = Distribution{packageFormat: "deb", project: "debian", id: "bookworm", version: 12}
	DistributionDebian13   = Distribution{packageFormat: "deb", project: "debian", id: "trixie", version: 13}
	DistributionUbuntu2004 = Distribution{packageFormat: "deb", project: "ubuntu", id: "focal", version: 20.04}
	DistributionUbuntu2204 = Distribution{packageFormat: "deb", project: "ubuntu", id: "jammy", version: 22.04}
	DistributionUbuntu2404 = Distribution{packageFormat: "deb", project: "ubuntu", id: "noble", version: 24.04}

	// Redhat-family distros
	DistributionRhel8           = Distribution{packageFormat: "rpm", project: "rhel", id: "rhel8", version: 8}
	DistributionRhel9           = Distribution{packageFormat: "rpm", project: "rhel", id: "rhel9", version: 9}
	DistributionRocky8          = Distribution{packageFormat: "rpm", project: "rocky", id: "rocky8", version: 8}
	DistributionRocky9          = Distribution{packageFormat: "rpm", project: "rocky", id: "rocky9", version: 9}
	DistributionFedora41        = Distribution{packageFormat: "rpm", project: "fedora", id: "fedora41", version: 41}
	DistributionAmazonLinux2    = Distribution{packageFormat: "rpm", project: "amazonlinux2", id: "amazonlinux2", version: 0}
	DistributionAmazonLinux2023 = Distribution{packageFormat: "rpm", project: "amazonlinux2023", id: "amzn", version: 2023}

	// Immutable distros
	DistributionFlatcar     = Distribution{packageFormat: "", project: "flatcar", id: "flatcar", version: 0}
	DistributionContainerOS = Distribution{packageFormat: "", project: "containeros", id: "containeros", version: 0}
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

// HasDNF returns true if this distribution uses dnf
func (d *Distribution) HasDNF() bool {
	if !d.IsRHELFamily() {
		return false
	}
	// All our RHEL distros support DNF at this point, it seems
	switch d.project {
	case "rhel":
		return d.version >= 8
	case "rocky":
		return d.version >= 8
	case "fedora":
		return d.version >= 22
	case "amazonlinux2":
		return false
	default:
		klog.Warningf("unknown project for HasDNF (%q), assuming does support dnf", d.project)
		return true
	}
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
func (d *Distribution) HasLoopbackEtcResolvConf() bool {
	switch d.project {
	case "ubuntu", "flatcar":
		return true
	default:
		if _, err := os.Stat("/run/systemd/resolve/resolv.conf"); err == nil {
			return true
		}
		return false
	}
}

// Version returns the (project scoped) numeric version
func (d *Distribution) Version() float32 {
	return d.version
}
