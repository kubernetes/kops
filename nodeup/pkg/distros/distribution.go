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
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi/nodeup/tags"
)

type Distribution string

var (
	DistributionJessie      Distribution = "jessie"
	DistributionXenial      Distribution = "xenial"
	DistributionRhel7       Distribution = "rhel7"
	DistributionCentos7     Distribution = "centos7"
	DistributionCoreOS      Distribution = "coreos"
	DistributionContainerOS Distribution = "containeros"
)

func (d Distribution) BuildTags() []string {
	var t []string

	switch d {
	case DistributionJessie:
		t = []string{"_jessie"}
	case DistributionXenial:
		t = []string{"_xenial"}
	case DistributionCentos7:
		t = []string{"_centos7"}
	case DistributionRhel7:
		t = []string{"_rhel7"}
	case DistributionCoreOS:
		t = []string{"_coreos"}
	case DistributionContainerOS:
		t = []string{"_containeros"}
	default:
		glog.Fatalf("unknown distribution: %s", d)
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
	case DistributionJessie, DistributionXenial:
		return true
	case DistributionCentos7, DistributionRhel7, DistributionCoreOS, DistributionContainerOS:
		return false
	default:
		glog.Fatalf("unknown distribution: %s", d)
		return false
	}
}

func (d Distribution) IsRHELFamily() bool {
	switch d {
	case DistributionCentos7, DistributionRhel7:
		return true
	case DistributionJessie, DistributionXenial, DistributionCoreOS, DistributionContainerOS:
		return false
	default:
		glog.Fatalf("unknown distribution: %s", d)
		return false
	}
}

func (d Distribution) IsSystemd() bool {
	switch d {
	case DistributionJessie, DistributionXenial:
		return true
	case DistributionCentos7, DistributionRhel7:
		return true
	case DistributionCoreOS:
		return true
	case DistributionContainerOS:
		return true
	default:
		glog.Fatalf("unknown distribution: %s", d)
		return false
	}
}
