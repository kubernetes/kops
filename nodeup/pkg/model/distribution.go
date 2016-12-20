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

package model

import (
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi/nodeup/tags"
)

type Distribution string

var (
	DistributionJessie  Distribution = "jessie"
	DistributionXenial  Distribution = "xenial"
	DistributionRhel7   Distribution = "rhel7"
	DistributionCentos7 Distribution = "centos7"
)

func (d Distribution) BuildTags() []string {
	switch d {
	case DistributionJessie:
		return []string{"_jessie", tags.TagOSFamilyDebian, tags.TagSystemd}
	case DistributionXenial:
		return []string{"_xenial", tags.TagOSFamilyDebian, tags.TagSystemd}
	case DistributionCentos7:
		return []string{"_centos7", tags.TagOSFamilyRHEL, tags.TagSystemd}
	case DistributionRhel7:
		return []string{"_rhel7", tags.TagOSFamilyRHEL, tags.TagSystemd}
	default:
		glog.Fatalf("unknown distribution: %s", d)
		return nil
	}
}
