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

package deployer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/kops/tests/e2e/kubetest2-kops/builder"
)

func TestMaybeGSURL(t *testing.T) {
	cases := []struct {
		name          string
		cloudProvider string
		baseURL       string
		expected      string
	}{
		{
			name:          "gce staged artifacts from a CI build",
			cloudProvider: "gce",
			baseURL:       "https://storage.googleapis.com/k8s-staging-kops/kops/releases/1.37.0-alpha.2+v1.37.0-alpha.1-42-gdeadbeef01",
			expected:      "gs://k8s-staging-kops/kops/releases/1.37.0-alpha.2+v1.37.0-alpha.1-42-gdeadbeef01",
		},
		{
			name:          "gce staged artifacts from a release branch without gs support are left alone",
			cloudProvider: "gce",
			baseURL:       "https://storage.googleapis.com/k8s-staging-kops/kops/releases/1.35.3+v1.35.1-39-gc89e13599b",
			expected:      "https://storage.googleapis.com/k8s-staging-kops/kops/releases/1.35.3+v1.35.1-39-gc89e13599b",
		},
		{
			name:          "gce with no parseable version is left alone",
			cloudProvider: "gce",
			baseURL:       "https://storage.googleapis.com/k8s-staging-kops/kops/latest/",
			expected:      "https://storage.googleapis.com/k8s-staging-kops/kops/latest/",
		},
		{
			name:          "aws staged artifacts are left alone",
			cloudProvider: "aws",
			baseURL:       "https://storage.googleapis.com/k8s-staging-kops/kops/releases/1.37.0-alpha.2+v1.37.0-alpha.1-42-gdeadbeef01",
			expected:      "https://storage.googleapis.com/k8s-staging-kops/kops/releases/1.37.0-alpha.2+v1.37.0-alpha.1-42-gdeadbeef01",
		},
		{
			name:          "gce with a non-GCS url",
			cloudProvider: "gce",
			baseURL:       "https://artifacts.k8s.io/binaries/kops/1.37.0/",
			expected:      "https://artifacts.k8s.io/binaries/kops/1.37.0/",
		},
		{
			name:          "gce with an already converted url",
			cloudProvider: "gce",
			baseURL:       "gs://k8s-staging-kops/kops/releases/1.37.0-alpha.2+v1.37.0-alpha.1-42-gdeadbeef01",
			expected:      "gs://k8s-staging-kops/kops/releases/1.37.0-alpha.2+v1.37.0-alpha.1-42-gdeadbeef01",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := &deployer{CloudProvider: tc.cloudProvider}
			if actual := d.maybeGSURL(tc.baseURL); actual != tc.expected {
				t.Errorf("maybeGSURL(%q) = %q, expected %q", tc.baseURL, actual, tc.expected)
			}
		})
	}
}

func TestResolveSSHKeys(t *testing.T) {
	// Every SSH key env var resolveSSHKeys consults, cleared before each case so
	// that the ambient environment cannot influence the result.
	allEnvVars := []string{
		"KUBE_SSH_KEY_PATH", "KUBE_SSH_PUBLIC_KEY_PATH",
		"AWS_SSH_PRIVATE_KEY_FILE", "AWS_SSH_PUBLIC_KEY_FILE",
		"DO_SSH_PRIVATE_KEY_FILE", "DO_SSH_PUBLIC_KEY_FILE",
		"GCE_SSH_PRIVATE_KEY_FILE", "GCE_SSH_PUBLIC_KEY_FILE",
	}

	dir := t.TempDir()
	// A keypair that exists on disk, so that the .pub sibling lookup can find it.
	onDiskPrivate := filepath.Join(dir, "id_ed25519")
	for _, p := range []string{onDiskPrivate, onDiskPrivate + ".pub"} {
		if err := os.WriteFile(p, []byte("test"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	cases := []struct {
		name            string
		cloudProvider   string
		flagPrivate     string
		flagPublic      string
		env             map[string]string
		expectedPrivate string
		expectedPublic  string
		expectGenerated bool
	}{
		{
			name:            "flags win over everything",
			cloudProvider:   "aws",
			flagPrivate:     "/flag/private",
			flagPublic:      "/flag/public",
			env:             map[string]string{"KUBE_SSH_KEY_PATH": "/env/private", "AWS_SSH_PRIVATE_KEY_FILE": "/legacy/private"},
			expectedPrivate: "/flag/private",
			expectedPublic:  "/flag/public",
		},
		{
			name:            "cloud-agnostic env vars are honored",
			cloudProvider:   "aws",
			env:             map[string]string{"KUBE_SSH_KEY_PATH": "/env/private", "KUBE_SSH_PUBLIC_KEY_PATH": "/env/public"},
			expectedPrivate: "/env/private",
			expectedPublic:  "/env/public",
		},
		{
			name:            "cloud-agnostic env vars win over the legacy per-cloud ones",
			cloudProvider:   "aws",
			env:             map[string]string{"KUBE_SSH_KEY_PATH": "/env/private", "KUBE_SSH_PUBLIC_KEY_PATH": "/env/public", "AWS_SSH_PRIVATE_KEY_FILE": "/legacy/private", "AWS_SSH_PUBLIC_KEY_FILE": "/legacy/public"},
			expectedPrivate: "/env/private",
			expectedPublic:  "/env/public",
		},
		{
			// The preset-aws-ssh preset, which must keep working unchanged.
			name:            "legacy aws env vars still resolve",
			cloudProvider:   "aws",
			env:             map[string]string{"AWS_SSH_PRIVATE_KEY_FILE": "/etc/aws-ssh/aws-ssh-private", "AWS_SSH_PUBLIC_KEY_FILE": "/etc/aws-ssh/aws-ssh-public"},
			expectedPrivate: "/etc/aws-ssh/aws-ssh-private",
			expectedPublic:  "/etc/aws-ssh/aws-ssh-public",
		},
		{
			// The preset-k8s-ssh preset.
			name:            "legacy gce env vars still resolve",
			cloudProvider:   "gce",
			env:             map[string]string{"GCE_SSH_PRIVATE_KEY_FILE": "/etc/ssh-key-secret/ssh-private", "GCE_SSH_PUBLIC_KEY_FILE": "/etc/ssh-key-secret/ssh-public"},
			expectedPrivate: "/etc/ssh-key-secret/ssh-private",
			expectedPublic:  "/etc/ssh-key-secret/ssh-public",
		},
		{
			// The preset-do-ssh preset.
			name:            "legacy digitalocean env vars still resolve",
			cloudProvider:   "digitalocean",
			env:             map[string]string{"DO_SSH_PRIVATE_KEY_FILE": "/etc/do-ssh/private-ssh-key", "DO_SSH_PUBLIC_KEY_FILE": "/etc/do-ssh/public-ssh-key"},
			expectedPrivate: "/etc/do-ssh/private-ssh-key",
			expectedPublic:  "/etc/do-ssh/public-ssh-key",
		},
		{
			name:            "another cloud's legacy env vars are ignored",
			cloudProvider:   "gce",
			env:             map[string]string{"AWS_SSH_PRIVATE_KEY_FILE": "/legacy/private", "AWS_SSH_PUBLIC_KEY_FILE": "/legacy/public"},
			expectGenerated: true,
		},
		{
			name:            "a private key with a .pub sibling resolves the pair",
			cloudProvider:   "aws",
			env:             map[string]string{"KUBE_SSH_KEY_PATH": onDiskPrivate},
			expectedPrivate: onDiskPrivate,
			expectedPublic:  onDiskPrivate + ".pub",
		},
		{
			name:            "a private key with no public half is discarded and a pair is generated",
			cloudProvider:   "aws",
			env:             map[string]string{"KUBE_SSH_KEY_PATH": "/nonexistent/private"},
			expectGenerated: true,
		},
		{
			name:            "nothing set generates a pair",
			cloudProvider:   "aws",
			expectGenerated: true,
		},
		{
			// Azure has no legacy env vars, so it always generated before and still does.
			name:            "azure generates",
			cloudProvider:   "azure",
			expectGenerated: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, k := range allEnvVars {
				t.Setenv(k, "")
				os.Unsetenv(k)
			}
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			d := &deployer{
				CloudProvider:     tc.cloudProvider,
				ClusterName:       "test-resolve-ssh-keys.k8s.local",
				SSHPrivateKeyPath: tc.flagPrivate,
				SSHPublicKeyPath:  tc.flagPublic,
			}
			if err := d.resolveSSHKeys(); err != nil {
				t.Fatalf("resolveSSHKeys() returned error: %v", err)
			}

			if tc.expectGenerated {
				// The generated pair lives under the temp dir and is named for the cluster.
				if !strings.Contains(d.SSHPrivateKeyPath, "test-resolve-ssh-keys.k8s.local") {
					t.Errorf("expected a generated private key, got %q", d.SSHPrivateKeyPath)
				}
				if d.SSHPublicKeyPath != d.SSHPrivateKeyPath+".pub" {
					t.Errorf("expected generated public key %q, got %q", d.SSHPrivateKeyPath+".pub", d.SSHPublicKeyPath)
				}
				return
			}

			if d.SSHPrivateKeyPath != tc.expectedPrivate {
				t.Errorf("private key = %q, expected %q", d.SSHPrivateKeyPath, tc.expectedPrivate)
			}
			if d.SSHPublicKeyPath != tc.expectedPublic {
				t.Errorf("public key = %q, expected %q", d.SSHPublicKeyPath, tc.expectedPublic)
			}
		})
	}
}

// The kubernetes e2e framework reads the SSH key and user from the environment
// rather than from flags, so the deployer has to export both for the tests it
// shells out to. Getting this wrong is quiet: when KUBE_SSH_KEY_PATH is absent
// SkipUnlessSSHKeyPresent() skips the SSH tests, and when KUBE_SSH_USER is absent
// the framework falls back to $USER and fails to authenticate.
func TestEnvExportsSSHKeyAndUser(t *testing.T) {
	cases := []struct {
		name          string
		cloudProvider string
		sshUser       string
		privateKey    string
		expected      map[string]string
	}{
		{
			name:          "azure exports the internally assigned user",
			cloudProvider: "azure",
			sshUser:       "kops",
			privateKey:    "/tmp/kops/cluster/id_ed25519",
			expected: map[string]string{
				"KUBE_SSH_KEY_PATH": "/tmp/kops/cluster/id_ed25519",
				"KUBE_SSH_USER":     "kops",
			},
		},
		{
			name:          "digitalocean exports the internally assigned user",
			cloudProvider: "digitalocean",
			sshUser:       "root",
			privateKey:    "/etc/do-ssh/private-ssh-key",
			expected: map[string]string{
				"KUBE_SSH_KEY_PATH": "/etc/do-ssh/private-ssh-key",
				"KUBE_SSH_USER":     "root",
			},
		},
		{
			name:          "aws exports the user resolved from the job config",
			cloudProvider: "aws",
			sshUser:       "ubuntu",
			privateKey:    "/etc/aws-ssh/aws-ssh-private",
			expected: map[string]string{
				"KUBE_SSH_KEY_PATH": "/etc/aws-ssh/aws-ssh-private",
				"KUBE_SSH_USER":     "ubuntu",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := &deployer{
				CloudProvider:     tc.cloudProvider,
				ClusterName:       "test-env.k8s.local",
				SSHUser:           tc.sshUser,
				SSHPrivateKeyPath: tc.privateKey,
				BuildOptions:      &builder.BuildOptions{},
			}

			actual := map[string]string{}
			for _, kv := range d.env() {
				if k, v, ok := strings.Cut(kv, "="); ok {
					actual[k] = v
				}
			}
			for k, want := range tc.expected {
				if actual[k] != want {
					t.Errorf("env() %s = %q, expected %q", k, actual[k], want)
				}
			}
		})
	}
}
