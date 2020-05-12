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

package hashing

import "testing"

func TestNewHasher(t *testing.T) {
	cases := []struct {
		name        string
		HA          HashAlgorithm
		expectedNil bool
	}{
		{
			name:        "sha256",
			HA:          "sha256",
			expectedNil: false,
		},
		{
			name:        "sha1",
			HA:          "sha1",
			expectedNil: false,
		},
		{
			name:        "md5",
			HA:          "md5",
			expectedNil: false,
		},
		/*{
			name:        "unknown",
			HA:          "unknown",
			expectedNil: true,
		},*/
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			hasher := c.HA.NewHasher()
			if (hasher == nil) != c.expectedNil {
				t.Errorf("expectedNil: %v, got: %v", c.expectedNil, hasher)
			}
		})
	}
}

func TestMemberFromString(t *testing.T) {
	cases := []struct {
		name     string
		HA       HashAlgorithm
		parm     string
		expected string
	}{
		// sha256
		{
			name:     "sha256 1",
			HA:       "sha256",
			parm:     "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc",
			expected: "invalid \"sha256\" hash - unexpected length 63",
		},
		{
			name:     "sha256 2",
			HA:       "sha256",
			parm:     "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5",
			expected: "sha256:5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5",
		},
		{
			name:     "sha256 3",
			HA:       "sha256",
			parm:     "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc51",
			expected: "invalid \"sha256\" hash - unexpected length 65",
		},
		// sha1
		{
			name:     "sha1 1",
			HA:       "sha1",
			parm:     "5e37efab5fdf59c1c508780ce5e83a3219c981d",
			expected: "invalid \"sha1\" hash - unexpected length 39",
		},
		{
			name:     "sha1 2",
			HA:       "sha1",
			parm:     "5e37efab5fdf59c1c508780ce5e83a3219c981d9",
			expected: "sha1:5e37efab5fdf59c1c508780ce5e83a3219c981d9",
		},
		{
			name:     "sha1 3",
			HA:       "sha1",
			parm:     "5e37efab5fdf59c1c508780ce5e83a3219c981d91",
			expected: "invalid \"sha1\" hash - unexpected length 41",
		},
		// md5
		{
			name:     "md5 1",
			HA:       "md5",
			parm:     "69a5f5f7d106c6c22710a15743d5810",
			expected: "invalid \"md5\" hash - unexpected length 31",
		},
		{
			name:     "md5 2",
			HA:       "md5",
			parm:     "69a5f5f7d106c6c22710a15743d58102",
			expected: "md5:69a5f5f7d106c6c22710a15743d58102",
		},
		{
			name:     "md5 3",
			HA:       "md5",
			parm:     "69a5f5f7d106c6c22710a15743d581021",
			expected: "invalid \"md5\" hash - unexpected length 33",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			hash, err := c.HA.FromString(c.parm)
			if err != nil {
				if err.Error() != c.expected {
					t.Errorf("got unexpected error: %v", err)
				}
				return
			}

			if hash.String() != c.expected {
				t.Errorf("expected: %v, got: %v", c.expected, hash.String())
			}
		})
	}
}

func TestFromString(t *testing.T) {
	cases := []struct {
		name     string
		HA       HashAlgorithm
		parm     string
		expected string
	}{
		{
			name:     "sha256",
			parm:     "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5",
			expected: "sha256:5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5",
		},
		{
			name:     "sha1",
			parm:     "5e37efab5fdf59c1c508780ce5e83a3219c981d9",
			expected: "sha1:5e37efab5fdf59c1c508780ce5e83a3219c981d9",
		},
		{
			name:     "md5",
			HA:       "md5",
			parm:     "69a5f5f7d106c6c22710a15743d58102",
			expected: "md5:69a5f5f7d106c6c22710a15743d58102",
		},
		{
			name:     "unknown",
			parm:     "69a5f5f7d106c6c22710a15743d581021",
			expected: "cannot determine algorithm for hash length: 33",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			hash, err := FromString(c.parm)
			if err != nil {
				if err.Error() != c.expected {
					t.Errorf("got unexpected error: %v", err)
				}
				return
			}

			if hash.String() != c.expected {
				t.Errorf("expected: %v, got: %v", c.expected, hash.String())
			}
		})
	}
}
