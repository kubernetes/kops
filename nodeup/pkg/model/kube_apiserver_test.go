/*
Copyright 2017 The Kubernetes Authors.

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
	"bytes"
	"crypto/x509"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/vfs"
)

func Test_KubeAPIServer_Builder(t *testing.T) {
	basedir := "tests/apiServer/auditDynamicConfiguration"

	context := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}

	nodeUpModelContext, err := BuildNodeupModelContext(basedir)
	if err != nil {
		t.Fatalf("error loading model %q: %v", basedir, err)
		return
	}
	nodeUpModelContext.KeyStore = &fakeKeyStore{T: t}

	builder := KubeAPIServerBuilder{NodeupModelContext: nodeUpModelContext}

	err = builder.Build(context)
	if err != nil {
		t.Fatalf("error from KubeAPIServerBuilder buildKubeletConfig: %v", err)
		return
	}
	if task, ok := context.Tasks["File//etc/kubernetes/manifests/kube-apiserver.manifest"]; !ok {
		t.Error("did not find the kubernetes API manifest after the build")
	} else {
		nodeTask, _ := task.(*nodetasks.File)
		reader, _ := nodeTask.Contents.Open()
		buf := new(bytes.Buffer)
		buf.ReadFrom(reader)
		s := buf.String()
		if strings.Contains(s, "--audit-dynamic-configuration") {
			t.Error("Older versions of k8s should not have --audit-dynamic-configuration flag")
		}
	}
}

type fakeKeyStore struct {
	T *testing.T
}

func (k fakeKeyStore) FindKeypair(name string) (*pki.Certificate, *pki.PrivateKey, fi.KeysetFormat, error) {
	panic("implement me")
}

func (k fakeKeyStore) CreateKeypair(signer string, name string, template *x509.Certificate, privateKey *pki.PrivateKey) (*pki.Certificate, error) {
	panic("implement me")
}

func (k fakeKeyStore) StoreKeypair(id string, cert *pki.Certificate, privateKey *pki.PrivateKey) error {
	panic("implement me")
}

func (k fakeKeyStore) MirrorTo(basedir vfs.Path) error {
	panic("implement me")
}

func (k fakeKeyStore) FindCertificatePool(name string) (*fi.CertificatePool, error) {
	panic("implement me")
}

func (k fakeKeyStore) FindCertificateKeyset(name string) (*kops.Keyset, error) {
	panic("implement me")
}

func (k fakeKeyStore) FindPrivateKey(name string) (*pki.PrivateKey, error) {
	panic("implement me")
}

func (k fakeKeyStore) FindPrivateKeyset(name string) (*kops.Keyset, error) {
	panic("implement me")
}

func (k fakeKeyStore) FindCert(name string) (*pki.Certificate, error) {
	assert.Equal(k.T, "apiserver-aggregator-ca", name)
	return &pki.Certificate{}, nil
}

func (k fakeKeyStore) ListKeysets() ([]*kops.Keyset, error) {
	panic("implement me")
}

func (k fakeKeyStore) AddCert(name string, cert *pki.Certificate) error {
	panic("implement me")
}

func (k fakeKeyStore) DeleteKeysetItem(item *kops.Keyset, id string) error {
	panic("implement me")
}

func Test_KubeAPIServer_BuildFlags(t *testing.T) {
	grid := []struct {
		config   kops.KubeAPIServerConfig
		expected string
	}{
		{
			kops.KubeAPIServerConfig{},
			"--insecure-port=0 --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				SecurePort: 443,
			},
			"--insecure-port=0 --secure-port=443",
		},
		{
			kops.KubeAPIServerConfig{
				MaxRequestsInflight: 1000,
			},
			"--insecure-port=0 --max-requests-inflight=1000 --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				MaxMutatingRequestsInflight: 900,
			},
			"--insecure-port=0 --max-mutating-requests-inflight=900 --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				InsecurePort: 8080,
				SecurePort:   443,
			},
			"--insecure-port=8080 --secure-port=443",
		},
		{
			kops.KubeAPIServerConfig{
				InsecurePort:        8080,
				SecurePort:          443,
				MaxRequestsInflight: 1000,
			},
			"--insecure-port=8080 --max-requests-inflight=1000 --secure-port=443",
		},
		{
			kops.KubeAPIServerConfig{
				InsecurePort:                8080,
				SecurePort:                  443,
				MaxMutatingRequestsInflight: 900,
			},
			"--insecure-port=8080 --max-mutating-requests-inflight=900 --secure-port=443",
		},
		{
			kops.KubeAPIServerConfig{
				InsecurePort: 8080,
			},
			"--insecure-port=8080 --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				ServiceNodePortRange: "30000-33000",
			},
			"--insecure-port=0 --secure-port=0 --service-node-port-range=30000-33000",
		},
		{
			kops.KubeAPIServerConfig{
				ExperimentalEncryptionProviderConfig: fi.String("/srv/kubernetes/encryptionconfig.yaml"),
			},
			"--experimental-encryption-provider-config=/srv/kubernetes/encryptionconfig.yaml --insecure-port=0 --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				EncryptionProviderConfig: fi.String("/srv/kubernetes/encryptionconfig.yaml"),
			},
			"--encryption-provider-config=/srv/kubernetes/encryptionconfig.yaml --insecure-port=0 --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				TargetRamMb: 320,
			},
			"--insecure-port=0 --secure-port=0 --target-ram-mb=320",
		},
		{
			kops.KubeAPIServerConfig{
				AuditDynamicConfiguration: &[]bool{true}[0],
				ServiceAccountKeyFile:     []string{"/srv/kubernetes/server.key", "/srv/kubernetes/service-account.key"},
			},
			"--audit-dynamic-configuration=true --insecure-port=0 --secure-port=0 --service-account-key-file=/srv/kubernetes/server.key --service-account-key-file=/srv/kubernetes/service-account.key",
		},
		{
			kops.KubeAPIServerConfig{
				AuditDynamicConfiguration: &[]bool{false}[0],
			},
			"--audit-dynamic-configuration=false --insecure-port=0 --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{

				AuditDynamicConfiguration: &[]bool{true}[0],
			},
			"--audit-dynamic-configuration=true --insecure-port=0 --secure-port=0",
		},
		{
			kops.KubeAPIServerConfig{
				EnableProfiling: &[]bool{false}[0],
			},
			"--insecure-port=0 --profiling=false --secure-port=0",
		},
	}

	for _, g := range grid {
		actual, err := flagbuilder.BuildFlags(&g.config)
		if err != nil {
			t.Errorf("error building flags for %v: %v", g.config, err)
			continue
		}
		if actual != g.expected {
			t.Errorf("flags did not match.  actual=%q expected=%q", actual, g.expected)
		}
	}
}

func TestKubeAPIServerBuilder(t *testing.T) {
	RunGoldenTest(t, "tests/golden/minimal", "kube-apiserver", func(nodeupModelContext *NodeupModelContext, target *fi.ModelBuilderContext) error {
		builder := KubeAPIServerBuilder{NodeupModelContext: nodeupModelContext}
		return builder.Build(target)
	})
}
