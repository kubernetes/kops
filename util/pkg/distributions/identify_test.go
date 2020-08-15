/*
Copyright 2020 The Kubernetes Authors.

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
	"path"
	"reflect"
	"testing"
)

func TestFindDistribution(t *testing.T) {
	tests := []struct {
		rootfs   string
		err      error
		expected Distribution
	}{
		{
			rootfs:   "amazonlinux2",
			err:      nil,
			expected: DistributionAmazonLinux2,
		},
		{
			rootfs:   "centos7",
			err:      nil,
			expected: DistributionCentos7,
		},
		{
			rootfs:   "centos8",
			err:      nil,
			expected: DistributionCentos8,
		},
		{
			rootfs:   "coreos",
			err:      fmt.Errorf("distribution CoreOS is no longer supported"),
			expected: "",
		},
		{
			rootfs:   "containeros",
			err:      nil,
			expected: DistributionContainerOS,
		},
		{
			rootfs:   "debian8",
			err:      fmt.Errorf("distribution Degian 8 (Jessie) is no longer supported"),
			expected: "",
		},
		{
			rootfs:   "debian9",
			err:      nil,
			expected: DistributionDebian9,
		},
		{
			rootfs:   "debian10",
			err:      nil,
			expected: DistributionDebian10,
		},
		{
			rootfs:   "flatcar",
			err:      nil,
			expected: DistributionFlatcar,
		},
		{
			rootfs:   "rhel7",
			err:      nil,
			expected: DistributionRhel7,
		},
		{
			rootfs:   "rhel8",
			err:      nil,
			expected: DistributionRhel8,
		},
		{
			rootfs:   "ubuntu1604",
			err:      nil,
			expected: DistributionUbuntu1604,
		},
		{
			rootfs:   "ubuntu1804",
			err:      nil,
			expected: DistributionUbuntu1804,
		},
		{
			rootfs:   "ubuntu2004",
			err:      nil,
			expected: DistributionUbuntu2004,
		},
	}

	for _, test := range tests {
		actual, err := FindDistribution(path.Join("tests", test.rootfs))
		if !reflect.DeepEqual(err, test.err) {
			t.Errorf("unexpected error, actual=\"%v\", expected=\"%v\"", err, test.err)
			continue
		}
		if actual != test.expected {
			t.Errorf("unexpected distribution, actual=%q, expected=%q", actual, test.expected)
		}
	}
}
