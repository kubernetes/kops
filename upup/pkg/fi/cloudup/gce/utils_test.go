/*
Copyright 2026 The Kubernetes Authors.

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

package gce

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSSHUsernameForImage(t *testing.T) {
	testcases := []struct {
		image    string
		expected string
	}{
		{
			image:    "ubuntu-os-cloud/ubuntu-2404-noble-amd64-v20260615",
			expected: "ubuntu",
		},
		{
			image:    "ubuntu-2404-noble-amd64-v20260615",
			expected: "ubuntu",
		},
		{
			image:    "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-2404-noble-amd64-v20260615",
			expected: "ubuntu",
		},
		{
			image:    "cos-cloud/cos-stable-77-12371-114-0",
			expected: "admin",
		},
		{
			image:    "debian-cloud/debian-12-bookworm-v20260615",
			expected: "admin",
		},
		{
			image:    "my-project/my-custom-image",
			expected: "admin",
		},
	}

	for _, g := range testcases {
		t.Run(g.image, func(t *testing.T) {
			assert.Equal(t, g.expected, SSHUsernameForImage(g.image))
		})
	}
}
