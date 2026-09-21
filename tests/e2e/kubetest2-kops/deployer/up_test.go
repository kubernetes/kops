/*
Copyright 2021 The Kubernetes Authors.

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

package deployer

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"k8s.io/kops/tests/e2e/kubetest2-kops/builder"
)

func TestAppendIfUnset(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		arg      string
		val      string
		expected []string
	}{
		{
			"empty",
			[]string{},
			"--foo",
			"bar",
			[]string{"--foo", "bar"},
		},
		{
			"unset",
			[]string{"--baz"},
			"--foo",
			"bar",
			[]string{"--baz", "--foo", "bar"},
		},
		{
			"set without value",
			[]string{"--foo"},
			"--foo",
			"bar",
			[]string{"--foo"},
		},
		{
			"set with different value",
			[]string{"--foo", "123"},
			"--foo",
			"bar",
			[]string{"--foo", "123"},
		},
		{
			"set with same value",
			[]string{"--foo", "bar"},
			"--foo",
			"bar",
			[]string{"--foo", "bar"},
		},
		{
			"set with same value and equals sign",
			[]string{"--foo=bar", "--baz=bar"},
			"--foo",
			"bar",
			[]string{"--foo=bar", "--baz=bar"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := appendIfUnset(tc.args, tc.arg, tc.val)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("arguments didn't match: %v vs %v", actual, tc.expected)
			}
		})
	}
}

func TestBuiltFromKopsRoot(t *testing.T) {
	const root = "/home/prow/go/src/k8s.io/kops"
	cases := []struct {
		name       string
		binaryPath string
		kopsRoot   string
		want       bool
	}{
		{"binary built under the checkout", root + "/.build/dist/linux/amd64/kops", root, true},
		{"binary downloaded to a temp dir", "/tmp/kops.abc123", root, false},
		{"sibling dir is not under the checkout", "/home/prow/go/src/k8s.io/kops-other/kops", root, false},
		{"empty kops root", root + "/.build/dist/linux/amd64/kops", "", false},
		{"empty binary path", "", root, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := builtFromKopsRoot(tc.binaryPath, tc.kopsRoot); got != tc.want {
				t.Errorf("builtFromKopsRoot(%q, %q) = %v, want %v", tc.binaryPath, tc.kopsRoot, got, tc.want)
			}
		})
	}
}

func TestLocalChannelArgs(t *testing.T) {
	const kopsRoot = "/home/prow/go/src/k8s.io/kops"
	base := "file://" + kopsRoot + "/channels/"
	cases := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "rewrites shorthand in equals form",
			args:     []string{"create", "cluster", "--channel=alpha", "--networking=cilium"},
			expected: []string{"create", "cluster", "--channel=" + base + "alpha", "--networking=cilium"},
		},
		{
			name:     "rewrites shorthand in space form",
			args:     []string{"--channel", "stable"},
			expected: []string{"--channel", base + "stable"},
		},
		{
			name:     "leaves absolute channel URL unchanged",
			args:     []string{"--channel=https://example.com/channels/alpha"},
			expected: []string{"--channel=https://example.com/channels/alpha"},
		},
		{
			name:     "leaves none unchanged",
			args:     []string{"--channel=none"},
			expected: []string{"--channel=none"},
		},
		{
			name:     "absent channel defaults to alpha pointed at the checkout",
			args:     []string{"create", "cluster", "--networking=cilium"},
			expected: []string{"create", "cluster", "--networking=cilium", "--channel=" + base + "alpha"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := localChannelArgs(slices.Clone(tc.args), kopsRoot)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("localChannelArgs() = %v, want %v", actual, tc.expected)
			}
		})
	}
}

func TestProwJobLabel(t *testing.T) {
	cases := []struct {
		name          string
		cloudProvider string
		jobName       string
		expected      string
		expectedOK    bool
	}{
		{
			name:          "empty job name omits label",
			cloudProvider: "aws",
			jobName:       "",
			expectedOK:    false,
		},
		{
			name:          "aws preserves slash and dots",
			cloudProvider: "aws",
			jobName:       "pull-kops-e2e-k8s-aws",
			expected:      "prow.k8s.io/job=pull-kops-e2e-k8s-aws",
			expectedOK:    true,
		},
		{
			name:          "azure replaces slash with underscore",
			cloudProvider: "azure",
			jobName:       "pull-kops-e2e-k8s-azure",
			expected:      "prow.k8s.io_job=pull-kops-e2e-k8s-azure",
			expectedOK:    true,
		},
		{
			name:          "gce normalizes key and value",
			cloudProvider: "gce",
			jobName:       "Pull-Kops-E2E-K8s-GCE",
			expected:      "prow_k8s_io_job=pull-kops-e2e-k8s-gce",
			expectedOK:    true,
		},
		{
			name:          "gce truncates value to 63 chars",
			cloudProvider: "gce",
			jobName:       "pull-kops-e2e-kubernetes-aws-canary-very-long-suffix-that-exceeds-the-gce-label-limit",
			expected:      "prow_k8s_io_job=" + "pull-kops-e2e-kubernetes-aws-canary-very-long-suffix-that-excee",
			expectedOK:    true,
		},
		{
			name:          "gce replaces invalid chars with dash",
			cloudProvider: "gce",
			jobName:       "job.name/with:invalid",
			expected:      "prow_k8s_io_job=job-name-with-invalid",
			expectedOK:    true,
		},
		{
			name:          "digitalocean omits label",
			cloudProvider: "digitalocean",
			jobName:       "pull-kops-e2e-k8s-do",
			expectedOK:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, ok := prowJobLabel(tc.cloudProvider, tc.jobName)
			if ok != tc.expectedOK {
				t.Errorf("ok mismatch: got %v, want %v", ok, tc.expectedOK)
			}
			if actual != tc.expected {
				t.Errorf("label mismatch: got %q, want %q", actual, tc.expected)
			}
		})
	}
}

func TestSetInstanceGroupOverridesControlPlane(t *testing.T) {
	dir := t.TempDir()

	kopsScript := filepath.Join(dir, "kops")
	argsFile := filepath.Join(dir, "args")

	script := `#!/bin/sh
if [ "$1" = "get" ] && [ "$2" = "instancegroups" ]; then
  cat <<'EOF'
[
  {
    "metadata": {
      "name": "control-plane-us-east-2a.masters.test.k8s.local"
    },
    "spec": {
      "role": "ControlPlane"
    }
  }
]
EOF
  exit 0
fi

if [ "$1" = "edit" ] && [ "$2" = "instancegroup" ]; then
  printf '%s\n' "$@" > "$KOPS_TEST_ARGS_FILE"
  exit 0
fi

echo "unexpected kops invocation: $@" >&2
exit 1
`

	if err := os.WriteFile(kopsScript, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	d := &deployer{
		BuildOptions:            &builder.BuildOptions{BuildKubernetes: false},
		KopsBinaryPath:          kopsScript,
		ClusterName:             "test.k8s.local",
		ControlPlaneIGOverrides: []string{"spec.rootVolume.type=io2"},
		Env:                     []string{"KOPS_TEST_ARGS_FILE=" + argsFile},
		stateStoreName:          "s3://test-state-store",
	}

	if err := d.setInstanceGroupOverrides(); err != nil {
		t.Fatalf("setInstanceGroupOverrides() failed: %v", err)
	}

	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read recorded kops arguments: %v", err)
	}

	want := "edit\ninstancegroup\n--name\ntest.k8s.local\ncontrol-plane-us-east-2a.masters.test.k8s.local\n--set\nspec.rootVolume.type=io2\n"
	if got := string(args); got != want {
		t.Errorf("unexpected kops edit arguments:\n%s", got)
	}
}
