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

package tomlwriter

import (
	"testing"

	"k8s.io/kops/pkg/diff"
)

// Expected strings match go-toml v1 byte for byte, keeping containerd configuration stable.

func TestEmptyTree(t *testing.T) {
	if got := NewTree().String(); got != "" {
		t.Errorf("empty tree serialized to %q, want empty string", got)
	}
}

func TestScalarsAndTables(t *testing.T) {
	tree := NewTree()
	tree.SetPath([]string{"version"}, int64(3))
	tree.SetPath([]string{"plugins", "io.containerd.nri.v1.nri", "disable"}, false)
	tree.SetPath([]string{"plugins", "io.containerd.cri.v1.runtime", "containerd", "default_runtime_name"}, "runc")
	tree.SetPath([]string{"plugins", "io.containerd.cri.v1.runtime", "containerd", "runtimes", "runc", "runtime_type"}, "io.containerd.runc.v2")
	tree.SetPath([]string{"plugins", "io.containerd.cri.v1.runtime", "containerd", "runtimes", "runc", "options", "SystemdCgroup"}, true)
	tree.SetPath([]string{"plugins", "io.containerd.cri.v1.images", "pinned_images", "sandbox"}, "registry.k8s.io/pause:3.10.1")

	expected := `version = 3

[plugins]

  [plugins."io.containerd.cri.v1.images"]

    [plugins."io.containerd.cri.v1.images".pinned_images]
      sandbox = "registry.k8s.io/pause:3.10.1"

  [plugins."io.containerd.cri.v1.runtime"]

    [plugins."io.containerd.cri.v1.runtime".containerd]
      default_runtime_name = "runc"

      [plugins."io.containerd.cri.v1.runtime".containerd.runtimes]

        [plugins."io.containerd.cri.v1.runtime".containerd.runtimes.runc]
          runtime_type = "io.containerd.runc.v2"

          [plugins."io.containerd.cri.v1.runtime".containerd.runtimes.runc.options]
            SystemdCgroup = true

  [plugins."io.containerd.nri.v1.nri"]
    disable = false
`

	if got := tree.String(); got != expected {
		t.Errorf("unexpected serialization:\n%s", diff.FormatDiff(expected, got))
	}
}

func TestTableAtDocumentStart(t *testing.T) {
	// Match go-toml v1 by starting with a blank line when the document root has no scalars.
	tree := NewTree()
	tree.SetPath([]string{"runsc_config", "platform"}, "systrap")

	expected := "\n[runsc_config]\n  platform = \"systrap\"\n"
	if got := tree.String(); got != expected {
		t.Errorf("unexpected serialization:\n%s", diff.FormatDiff(expected, got))
	}
}

func TestKeyQuotingAndStringEscaping(t *testing.T) {
	tree := NewTree()
	tree.SetPath([]string{"bare-KEY_09"}, "plain")
	tree.SetPath([]string{"needs quoting"}, "a\tb\nc\"d\\e\x01f")
	tree.SetPath([]string{"table", `dotted.and"quoted`}, int64(-42))

	expected := `bare-KEY_09 = "plain"
"needs quoting" = "a\tb\nc\"d\\e\u0001f"

[table]
  "dotted.and\"quoted" = -42
`
	if got := tree.String(); got != expected {
		t.Errorf("unexpected serialization:\n%s", diff.FormatDiff(expected, got))
	}
}

func TestPreQuotedKeyPassthrough(t *testing.T) {
	// Match go-toml v1 by emitting a key enclosed in double quotes verbatim, without escaping.
	tree := NewTree()
	tree.SetPath([]string{`"pre.quoted"`}, int64(1))

	expected := "\"pre.quoted\" = 1\n"
	if got := tree.String(); got != expected {
		t.Errorf("unexpected serialization:\n%s", diff.FormatDiff(expected, got))
	}
}

func TestEmptyKeyIsQuoted(t *testing.T) {
	// Unlike go-toml v1, quote an empty key instead of emitting invalid TOML.
	tree := NewTree()
	tree.SetPath([]string{"a", ""}, true)

	expected := "\n[a]\n  \"\" = true\n"
	if got := tree.String(); got != expected {
		t.Errorf("unexpected serialization:\n%s", diff.FormatDiff(expected, got))
	}
}

func TestEscapeBoundaries(t *testing.T) {
	// Match go-toml v1 by escaping characters below U+001F while emitting U+001F and U+007F raw;
	// TOML requires the latter two to be escaped, so this output is invalid.
	tree := NewTree()
	tree.SetPath([]string{"k"}, "\b\f\r\x1e\x1f\x7f")

	expected := "k = \"\\b\\f\\r\\u001E\x1f\x7f\"\n"
	if got := tree.String(); got != expected {
		t.Errorf("unexpected serialization:\n%s", diff.FormatDiff(expected, got))
	}
}

func TestSetPathThroughScalar(t *testing.T) {
	// Match go-toml v1 by leaving traversal in the current table when an intermediate value is scalar.
	tree := NewTree()
	tree.SetPath([]string{"a"}, int64(1))
	tree.SetPath([]string{"a", "b"}, int64(2))

	expected := "a = 1\nb = 2\n"
	if got := tree.String(); got != expected {
		t.Errorf("unexpected serialization:\n%s", diff.FormatDiff(expected, got))
	}
}

func TestSetPathOverwritesTable(t *testing.T) {
	tree := NewTree()
	tree.SetPath([]string{"a", "b"}, int64(1))
	tree.SetPath([]string{"a"}, "scalar")

	expected := "a = \"scalar\"\n"
	if got := tree.String(); got != expected {
		t.Errorf("unexpected serialization:\n%s", diff.FormatDiff(expected, got))
	}
}

func TestSetPathRejectsUnsupportedType(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic for unsupported value type")
		}
	}()
	NewTree().SetPath([]string{"a"}, 3.14)
}
