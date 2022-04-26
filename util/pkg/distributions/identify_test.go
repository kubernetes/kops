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
			err:      fmt.Errorf("unsupported distro: centos-7"),
			expected: Distribution{},
		},
		{
			rootfs:   "centos8",
			err:      fmt.Errorf("unsupported distro: centos-8"),
			expected: Distribution{},
		},
		{
			rootfs:   "coreos",
			err:      fmt.Errorf("unsupported distro: coreos-2247.7.0"),
			expected: Distribution{},
		},
		{
			rootfs:   "containeros",
			err:      nil,
			expected: DistributionContainerOS,
		},
		{
			rootfs:   "debian8",
			err:      fmt.Errorf("unsupported distro: debian-8"),
			expected: Distribution{},
		},
		{
			rootfs:   "debian9",
			err:      fmt.Errorf("unsupported distro: debian-9"),
			expected: Distribution{},
		},
		{
			rootfs:   "debian10",
			err:      nil,
			expected: DistributionDebian10,
		},
		{
			rootfs:   "debian11",
			err:      nil,
			expected: DistributionDebian11,
		},
		{
			rootfs:   "flatcar",
			err:      nil,
			expected: DistributionFlatcar,
		},
		{
			rootfs:   "rhel7",
			err:      fmt.Errorf("unsupported distro: rhel-7.8"),
			expected: Distribution{},
		},
		{
			rootfs:   "rhel8",
			err:      nil,
			expected: DistributionRhel8,
		},
		{
			rootfs:   "ubuntu1604",
			err:      fmt.Errorf("unsupported distro: ubuntu-16.04"),
			expected: Distribution{},
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
		{
			rootfs:   "ubuntu2010",
			err:      nil,
			expected: DistributionUbuntu2010,
		},
		{
			rootfs:   "ubuntu2104",
			err:      nil,
			expected: DistributionUbuntu2104,
		},
		{
			rootfs:   "ubuntu2110",
			err:      nil,
			expected: DistributionUbuntu2110,
		},
		{
			rootfs:   "ubuntu2204",
			err:      nil,
			expected: DistributionUbuntu2204,
		},
		{
			rootfs:   "notfound",
			err:      fmt.Errorf("reading /etc/os-release: open tests/notfound/etc/os-release: no such file or directory"),
			expected: Distribution{},
		},
	}

	for _, test := range tests {
		actual, err := FindDistribution(path.Join("tests", test.rootfs))
		if !reflect.DeepEqual(err, test.err) {
			t.Errorf("unexpected error, actual=\"%v\", expected=\"%v\"", err, test.err)
			continue
		}
		if actual != test.expected {
			t.Errorf("unexpected distribution, actual=%v, expected=%v", actual, test.expected)
		}
	}
}
