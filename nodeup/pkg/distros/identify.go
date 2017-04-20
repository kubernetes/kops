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

package distros

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// FindDistribution identifies the distribution on which we are running
// We will likely remove this when everything is containerized
func FindDistribution(rootfs string) (Distribution, error) {
	// Ubuntu has /etc/lsb-release (and /etc/debian_version)
	lsbRelease, err := ioutil.ReadFile(path.Join(rootfs, "etc/lsb-release"))
	if err == nil {
		for _, line := range strings.Split(string(lsbRelease), "\n") {
			line = strings.TrimSpace(line)
			if line == "DISTRIB_CODENAME=xenial" {
				return DistributionXenial, nil
			}
		}
		glog.Warningf("unhandled lsb-release info %q", string(lsbRelease))
	} else if !os.IsNotExist(err) {
		glog.Warningf("error reading /etc/lsb-release: %v", err)
	}

	// Debian has /etc/debian_version
	debianVersionBytes, err := ioutil.ReadFile(path.Join(rootfs, "etc/debian_version"))
	if err == nil {
		debianVersion := strings.TrimSpace(string(debianVersionBytes))
		if strings.HasPrefix(debianVersion, "8.") {
			return DistributionJessie, nil
		} else {
			return "", fmt.Errorf("unhandled debian version %q", debianVersion)
		}
	} else if !os.IsNotExist(err) {
		glog.Warningf("error reading /etc/debian_version: %v", err)
	}

	// Redhat has /etc/redhat-release
	// Centos has /etc/centos-release
	redhatRelease, err := ioutil.ReadFile(path.Join(rootfs, "etc/redhat-release"))
	if err == nil {
		for _, line := range strings.Split(string(redhatRelease), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Red Hat Enterprise Linux Server release 7.") {
				return DistributionRhel7, nil
			}
			if strings.HasPrefix(line, "CentOS Linux release 7.") {
				return DistributionCentos7, nil
			}
		}
		glog.Warningf("unhandled redhat-release info %q", string(lsbRelease))
	} else if !os.IsNotExist(err) {
		glog.Warningf("error reading /etc/redhat-release: %v", err)
	}

	// CoreOS uses /usr/lib/os-release
	osRelease, err := ioutil.ReadFile(path.Join(rootfs, "usr/lib/os-release"))
	if err == nil {
		for _, line := range strings.Split(string(osRelease), "\n") {
			line = strings.TrimSpace(line)
			if line == "ID=coreos" {
				return DistributionCoreOS, nil
			}
		}
		glog.Warningf("unhandled os-release info %q", string(osRelease))
	} else if !os.IsNotExist(err) {
		glog.Warningf("error reading /usr/lib/os-release: %v", err)
	}

	// ContainerOS uses /etc/os-release
	{
		osRelease, err := ioutil.ReadFile(path.Join(rootfs, "etc/os-release"))
		if err == nil {
			for _, line := range strings.Split(string(osRelease), "\n") {
				line = strings.TrimSpace(line)
				if line == "ID=cos" {
					return DistributionContainerOS, nil
				}
			}
			glog.Warningf("unhandled /etc/os-release info %q", string(osRelease))
		} else if !os.IsNotExist(err) {
			glog.Warningf("error reading /etc/os-release: %v", err)
		}
	}

	return "", fmt.Errorf("cannot identify distro")
}
