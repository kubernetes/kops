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

package hosts

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/kops/pkg/diff"
)

func TestRemovesDuplicateGuardedBlocks(t *testing.T) {
	in := `
10.2.3.4 foo

# Begin host entries managed by etcd-manager[etcd] - do not edit
# End host entries managed by etcd-manager[etcd]
# Begin host entries managed by etcd-manager[etcd] - do not edit
# End host entries managed by etcd-manager[etcd]
# Begin host entries managed by kops - do not edit
# End host entries managed by kops
# Begin host entries managed by kops - do not edit
# End host entries managed by kops
# Begin host entries managed by kops - do not edit
# End host entries managed by kops
# Begin host entries managed by kops - do not edit
# End host entries managed by kops
# Begin host entries managed by kops - do not edit
# End host entries managed by kops
# Begin host entries managed by kops - do not edit
# End host entries managed by kops
# Begin host entries managed by kops - do not edit
# End host entries managed by kops
# Begin host entries managed by kops - do not edit
# End host entries managed by kops
`

	expected := `
10.2.3.4 foo

# Begin host entries managed by etcd-manager[etcd] - do not edit
# End host entries managed by etcd-manager[etcd]
# Begin host entries managed by etcd-manager[etcd] - do not edit
# End host entries managed by etcd-manager[etcd]

# Begin host entries managed by kops - do not edit
10.0.1.1	a
10.0.1.2	a
10.0.2.1	b
# End host entries managed by kops
`

	runTest(t, in, expected)
}

func TestRecoversFromBadNesting(t *testing.T) {
	in := `
10.2.3.4 foo

# End host entries managed by kops
# Begin host entries managed by kops - do not edit
# Begin host entries managed by kops - do not edit
# End host entries managed by kops
# End host entries managed by kops
# End host entries managed by kops
# Begin host entries managed by kops - do not edit
# End host entries managed by kops
# Begin host entries managed by kops - do not edit
# End host entries managed by kops
# Begin host entries managed by kops - do not edit
# End host entries managed by kops
# Begin host entries managed by kops - do not edit
# Begin host entries managed by kops - do not edit
# Begin host entries managed by kops - do not edit
# Begin host entries managed by kops - do not edit
# End host entries managed by kops
# Begin host entries managed by kops - do not edit
# End host entries managed by kops

10.1.2.3 bar
`

	expected := `
10.2.3.4 foo


10.1.2.3 bar

# Begin host entries managed by kops - do not edit
10.0.1.1	a
10.0.1.2	a
10.0.2.1	b
# End host entries managed by kops
`

	runTest(t, in, expected)
}

func runTest(t *testing.T, in string, expected string) {
	dir := t.TempDir()

	p := filepath.Join(dir, "hosts")
	namesToAddresses := map[string][]string{
		"a": {"10.0.1.2", "10.0.1.1"},
		"b": {"10.0.2.1"},
		"c": {},
	}

	if err := os.WriteFile(p, []byte(in), 0o755); err != nil {
		t.Fatalf("error writing hosts file: %v", err)
	}

	// We run it repeatedly to make sure we don't change it accidentally
	for i := 0; i < 100; i++ {
		mutator := func(existing []string) (*HostMap, error) {
			hostMap := &HostMap{}
			badLines := hostMap.Parse(existing)
			if len(badLines) != 0 {
				t.Errorf("unexpected lines in /etc/hosts: %v", badLines)
			}

			for name, addresses := range namesToAddresses {
				hostMap.ReplaceRecords(name, addresses)
			}

			return hostMap, nil
		}
		if err := UpdateHostsFileWithRecords(p, mutator); err != nil {
			t.Fatalf("error updating hosts file: %v", err)
		}

		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("error reading output file: %v", err)
		}

		actual := string(b)
		if actual != expected {
			diffString := diff.FormatDiff(expected, actual)
			t.Logf("diff:\n%s\n", diffString)
			t.Errorf("unexpected output.  expected=%q, actual=%q", expected, actual)
		}
	}
}
