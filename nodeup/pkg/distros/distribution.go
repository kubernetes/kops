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

package distros

import (
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi/nodeup/tags"
)

type Distribution string

var (
	DistributionDebian9      Distribution = "debian9"
	DistributionDebian10     Distribution = "buster"
	DistributionXenial       Distribution = "xenial"
	DistributionBionic       Distribution = "bionic"
	DistributionFocal        Distribution = "focal"
	DistributionAmazonLinux2 Distribution = "amazonlinux2"
	DistributionRhel7        Distribution = "rhel7"
	DistributionCentos7      Distribution = "centos7"
	DistributionRhel8        Distribution = "rhel8"
	DistributionCentos8      Distribution = "centos8"
	DistributionFlatcar      Distribution = "flatcar"
	DistributionContainerOS  Distribution = "containeros"
)

func (d Distribution) BuildTags() []string {
	var t []string

	switch d {
	case DistributionDebian9, DistributionDebian10:
		t = []string{} // trying to move away from tags
	case DistributionXenial:
		t = []string{"_xenial"}
	case DistributionBionic:
		t = []string{"_bionic"}
	case DistributionFocal:
		t = []string{"_focal"}
	case DistributionAmazonLinux2:
		t = []string{"_amazonlinux2"}
	case DistributionCentos7:
		t = []string{"_centos7"}
	case DistributionRhel7:
		t = []string{"_rhel7"}
	case DistributionCentos8:
		t = []string{"_centos8"}
	case DistributionRhel8:
		t = []string{"_rhel8"}
	case DistributionFlatcar:
		t = []string{"_flatcar"}
	case DistributionContainerOS:
		t = []string{"_containeros"}
	default:
		klog.Fatalf("unknown distribution: %s", d)
		return nil
	}

	if d.IsDebianFamily() {
		t = append(t, tags.TagOSFamilyDebian)
	}
	if d.IsRHELFamily() {
		t = append(t, tags.TagOSFamilyRHEL)
	}
	if d.IsSystemd() {
		t = append(t, tags.TagSystemd)
	}

	return t
}

func (d Distribution) IsDebianFamily() bool {
	switch d {
	case DistributionDebian9, DistributionDebian10:
		return true
	case DistributionXenial, DistributionBionic, DistributionFocal:
		return true
	case DistributionCentos7, DistributionRhel7, DistributionCentos8, DistributionRhel8, DistributionAmazonLinux2:
		return false
	case DistributionFlatcar, DistributionContainerOS:
		return false
	default:
		klog.Fatalf("unknown distribution: %s", d)
		return false
	}
}

func (d Distribution) IsUbuntu() bool {
	switch d {
	case DistributionDebian9, DistributionDebian10:
		return false
	case DistributionXenial, DistributionBionic, DistributionFocal:
		return true
	case DistributionCentos7, DistributionRhel7, DistributionCentos8, DistributionRhel8, DistributionAmazonLinux2:
		return false
	case DistributionFlatcar, DistributionContainerOS:
		return false
	default:
		klog.Fatalf("unknown distribution: %s", d)
		return false
	}
}

func (d Distribution) IsRHELFamily() bool {
	switch d {
	case DistributionCentos7, DistributionRhel7, DistributionCentos8, DistributionRhel8, DistributionAmazonLinux2:
		return true
	case DistributionXenial, DistributionBionic, DistributionFocal, DistributionDebian9, DistributionDebian10:
		return false
	case DistributionFlatcar, DistributionContainerOS:
		return false
	default:
		klog.Fatalf("unknown distribution: %s", d)
		return false
	}
}

func (d Distribution) IsSystemd() bool {
	switch d {
	case DistributionXenial, DistributionBionic, DistributionFocal, DistributionDebian9, DistributionDebian10:
		return true
	case DistributionCentos7, DistributionRhel7, DistributionCentos8, DistributionRhel8, DistributionAmazonLinux2:
		return true
	case DistributionFlatcar:
		return true
	case DistributionContainerOS:
		return true
	default:
		klog.Fatalf("unknown distribution: %s", d)
		return false
	}
}
