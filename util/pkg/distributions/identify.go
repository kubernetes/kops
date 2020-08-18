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
	"io/ioutil"
	"path"
	"strings"

	"k8s.io/klog/v2"
)

// FindDistribution identifies the distribution on which we are running
func FindDistribution(rootfs string) (Distribution, error) {
	// All supported distros have an /etc/os-release file
	osReleaseBytes, err := ioutil.ReadFile(path.Join(rootfs, "etc/os-release"))
	osRelease := make(map[string]string)
	if err == nil {
		for _, line := range strings.Split(string(osReleaseBytes), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "ID=") {
				osRelease["ID"] = strings.Trim(line[3:], "\"")
			}
			if strings.HasPrefix(line, "VERSION_ID=") {
				osRelease["VERSION_ID"] = strings.Trim(line[11:], "\"")
			}
		}
	} else {
		return "", fmt.Errorf("reading /etc/os-release: %v", err)
	}

	distro := fmt.Sprintf("%s-%s", osRelease["ID"], osRelease["VERSION_ID"])

	// Most distros have a fixed VERSION_ID
	switch distro {
	case "amzn-2":
		return DistributionAmazonLinux2, nil
	case "centos-7":
		return DistributionCentos7, nil
	case "centos-8":
		return DistributionCentos8, nil
	case "debian-9":
		return DistributionDebian9, nil
	case "debian-10":
		return DistributionDebian10, nil
	case "ubuntu-16.04":
		return DistributionUbuntu1604, nil
	case "ubuntu-18.04":
		return DistributionUbuntu1804, nil
	case "ubuntu-20.04":
		return DistributionUbuntu2004, nil
	}

	// Some distros have a more verbose VERSION_ID
	if strings.HasPrefix(distro, "cos-") {
		return DistributionContainerOS, nil
	}
	if strings.HasPrefix(distro, "flatcar-") {
		return DistributionFlatcar, nil
	}
	if strings.HasPrefix(distro, "rhel-7.") {
		return DistributionRhel7, nil
	}
	if strings.HasPrefix(distro, "rhel-8.") {
		return DistributionRhel8, nil
	}

	// Some distros are not supported
	klog.V(2).Infof("Contents of /etc/os-release:\n%s", osReleaseBytes)
	return "", fmt.Errorf("unsupported distro: %s", distro)
}
