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

package main

import (
	"reflect"
	"testing"
)

func TestParseHashFile(t *testing.T) {
	grid := []struct {
		name   string
		prefix string
		input  string
		want   []file
	}{
		{
			// Excerpt of https://github.com/opencontainers/runc/releases/download/v1.2.0/runc.sha256sum
			name:   "clearsigned with SHA256",
			prefix: "v1.2.0/",
			input: `-----BEGIN PGP SIGNED MESSAGE-----
Hash: SHA256

3bbb68e49bc89dd2607f11d2ff0fa699963ebada39c32ad8a6aab0d40435c1ed  runc.amd64
3d4f66dc1d91f1b2a46713d185a506a604f1fe9f2f2b89c281eb1c5c13677ff0  runc.arm64
-----BEGIN PGP SIGNATURE-----

iQJEBAEBCAAuFiEEXzbGxhtUYBJKdfWmnhiqJn3bjbQFAmcW1hAQHGFzYXJhaUBz
-----END PGP SIGNATURE-----
`,
			want: []file{
				{Name: "v1.2.0/runc.amd64", SHA256: "3bbb68e49bc89dd2607f11d2ff0fa699963ebada39c32ad8a6aab0d40435c1ed"},
				{Name: "v1.2.0/runc.arm64", SHA256: "3d4f66dc1d91f1b2a46713d185a506a604f1fe9f2f2b89c281eb1c5c13677ff0"},
			},
		},
		{
			// Excerpt of https://github.com/opencontainers/runc/releases/download/v1.4.0/runc.sha256sum
			name:   "clearsigned with SHA512",
			prefix: "v1.4.0/",
			input: `-----BEGIN PGP SIGNED MESSAGE-----
Hash: SHA512

c5d4995c5aec204d7e1827d9d9a6b45042602736f7f415f484252e576dcdac28  runc.amd64
2adbeed4c751d6f2201c642ed06269ff4370fcc4165abd3f323e19c653716c31  runc.arm64
-----BEGIN PGP SIGNATURE-----

iJEEARYKADkWIQS2TklVsp+j1GPyqQYol/rSt+lEbwUCaSjfMhsUgAAAAAAEAA5t
-----END PGP SIGNATURE-----
`,
			want: []file{
				{Name: "v1.4.0/runc.amd64", SHA256: "c5d4995c5aec204d7e1827d9d9a6b45042602736f7f415f484252e576dcdac28"},
				{Name: "v1.4.0/runc.arm64", SHA256: "2adbeed4c751d6f2201c642ed06269ff4370fcc4165abd3f323e19c653716c31"},
			},
		},
	}

	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			m, err := parseHashFile(g.input, g.prefix, nil)
			if err != nil {
				t.Fatalf("parsing hash file: %v", err)
			}
			if !reflect.DeepEqual(m.Files, g.want) {
				t.Errorf("unexpected files; got %+v, want %+v", m.Files, g.want)
			}
		})
	}
}
