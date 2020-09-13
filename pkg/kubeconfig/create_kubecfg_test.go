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

package kubeconfig

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/kops/pkg/testutils"

	"github.com/google/go-cmp/cmp"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

const certData = "-----BEGIN CERTIFICATE-----\nMIIC2DCCAcCgAwIBAgIRALJXAkVj964tq67wMSI8oJQwDQYJKoZIhvcNAQELBQAw\nFTETMBEGA1UEAxMKa3ViZXJuZXRlczAeFw0xNzEyMjcyMzUyNDBaFw0yNzEyMjcy\nMzUyNDBaMBUxEzARBgNVBAMTCmt1YmVybmV0ZXMwggEiMA0GCSqGSIb3DQEBAQUA\nA4IBDwAwggEKAoIBAQDgnCkSmtnmfxEgS3qNPaUCH5QOBGDH/inHbWCODLBCK9gd\nXEcBl7FVv8T2kFr1DYb0HVDtMI7tixRVFDLgkwNlW34xwWdZXB7GeoFgU1xWOQSY\nOACC8JgYTQ/139HBEvgq4sej67p+/s/SNcw34Kk7HIuFhlk1rRk5kMexKIlJBKP1\nYYUYetsJ/QpUOkqJ5HW4GoetE76YtHnORfYvnybviSMrh2wGGaN6r/s4ChOaIbZC\nAn8/YiPKGIDaZGpj6GXnmXARRX/TIdgSQkLwt0aTDBnPZ4XvtpI8aaL8DYJIqAzA\nNPH2b4/uNylat5jDo0b0G54agMi97+2AUrC9UUXpAgMBAAGjIzAhMA4GA1UdDwEB\n/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBVGR2r\nhzXzRMU5wriPQAJScszNORvoBpXfZoZ09FIupudFxBVU3d4hV9StKnQgPSGA5XQO\nHE97+BxJDuA/rB5oBUsMBjc7y1cde/T6hmi3rLoEYBSnSudCOXJE4G9/0f8byAJe\nrN8+No1r2VgZvZh6p74TEkXv/l3HBPWM7IdUV0HO9JDhSgOVF1fyQKJxRuLJR8jt\nO6mPH2UX0vMwVa4jvwtkddqk2OAdYQvH9rbDjjbzaiW0KnmdueRo92KHAN7BsDZy\nVpXHpqo1Kzg7D3fpaXCf5si7lqqrdJVXH4JC72zxsPehqgi8eIuqOBkiDWmRxAxh\n8yGeRx9AbknHh4Ia\n-----END CERTIFICATE-----\n"
const privatekeyData = "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA4JwpEprZ5n8RIEt6jT2lAh+UDgRgx/4px21gjgywQivYHVxH\nAZexVb/E9pBa9Q2G9B1Q7TCO7YsUVRQy4JMDZVt+McFnWVwexnqBYFNcVjkEmDgA\ngvCYGE0P9d/RwRL4KuLHo+u6fv7P0jXMN+CpOxyLhYZZNa0ZOZDHsSiJSQSj9WGF\nGHrbCf0KVDpKieR1uBqHrRO+mLR5zkX2L58m74kjK4dsBhmjeq/7OAoTmiG2QgJ/\nP2IjyhiA2mRqY+hl55lwEUV/0yHYEkJC8LdGkwwZz2eF77aSPGmi/A2CSKgMwDTx\n9m+P7jcpWreYw6NG9BueGoDIve/tgFKwvVFF6QIDAQABAoIBAA0ktjaTfyrAxsTI\nBezb7Zr5NBW55dvuII299cd6MJo+rI/TRYhvUv48kY8IFXp/hyUjzgeDLunxmIf9\n/Zgsoic9Ol44/g45mMduhcGYPzAAeCdcJ5OB9rR9VfDCXyjYLlN8H8iU0734tTqM\n0V13tQ9zdSqkGPZOIcq/kR/pylbOZaQMe97BTlsAnOMSMKDgnftY4122Lq3GYy+t\nvpr+bKVaQZwvkLoSU3rECCaKaghgwCyX7jft9aEkhdJv+KlwbsGY6WErvxOaLWHd\ncuMQjGapY1Fa/4UD00mvrA260NyKfzrp6+P46RrVMwEYRJMIQ8YBAk6N6Hh7dc0G\n8Z6i1m0CgYEA9HeCJR0TSwbIQ1bDXUrzpftHuidG5BnSBtax/ND9qIPhR/FBW5nj\n22nwLc48KkyirlfIULd0ae4qVXJn7wfYcuX/cJMLDmSVtlM5Dzmi/91xRiFgIzx1\nAsbBzaFjISP2HpSgL+e9FtSXaaqeZVrflitVhYKUpI/AKV31qGHf04sCgYEA6zTV\n99Sb49Wdlns5IgsfnXl6ToRttB18lfEKcVfjAM4frnkk06JpFAZeR+9GGKUXZHqs\nz2qcplw4d/moCC6p3rYPBMLXsrGNEUFZqBlgz72QA6BBq3X0Cg1Bc2ZbK5VIzwkg\nST2SSux6ccROfgULmN5ZiLOtdUKNEZpFF3i3qtsCgYADT/s7dYFlatobz3kmMnXK\nsfTu2MllHdRys0YGHu7Q8biDuQkhrJwhxPW0KS83g4JQym+0aEfzh36bWcl+u6R7\nKhKj+9oSf9pndgk345gJz35RbPJYh+EuAHNvzdgCAvK6x1jETWeKf6btj5pF1U1i\nQ4QNIw/QiwIXjWZeubTGsQKBgQCbduLu2rLnlyyAaJZM8DlHZyH2gAXbBZpxqU8T\nt9mtkJDUS/KRiEoYGFV9CqS0aXrayVMsDfXY6B/S/UuZjO5u7LtklDzqOf1aKG3Q\ndGXPKibknqqJYH+bnUNjuYYNerETV57lijMGHuSYCf8vwLn3oxBfERRX61M/DU8Z\nworz/QKBgQDCTJI2+jdXg26XuYUmM4XXfnocfzAXhXBULt1nENcogNf1fcptAVtu\nBAiz4/HipQKqoWVUYmxfgbbLRKKLK0s0lOWKbYdVjhEm/m2ZU8wtXTagNwkIGoyq\nY/C1Lox4f1ROJnCjc/hfcOjcxX5M8A8peecHWlVtUPKTJgxQ7oMKcw==\n-----END RSA PRIVATE KEY-----\n"

// mock a fake status store.
type fakeStatusStore struct {
	FindClusterStatusFn   func(cluster *kops.Cluster) (*kops.ClusterStatus, error)
	GetApiIngressStatusFn func(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error)
}

func (f fakeStatusStore) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	return f.FindClusterStatusFn(cluster)
}

func (f fakeStatusStore) GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error) {
	return f.GetApiIngressStatusFn(cluster)
}

// mock a fake key store
type fakeKeyStore struct {
	FindKeypairFn func(name string) (*pki.Certificate, *pki.PrivateKey, bool, error)

	// StoreKeypair writes the keypair to the store
	StoreKeypairFn func(id string, cert *pki.Certificate, privateKey *pki.PrivateKey) error

	// MirrorTo will copy secrets to a vfs.Path, which is often easier for a machine to read
	MirrorToFn func(basedir vfs.Path) error
}

func (f fakeKeyStore) FindKeypair(name string) (*pki.Certificate, *pki.PrivateKey, bool, error) {
	return f.FindKeypairFn(name)
}

func (f fakeKeyStore) StoreKeypair(id string, cert *pki.Certificate, privateKey *pki.PrivateKey) error {
	return f.StoreKeypairFn(id, cert, privateKey)
}

func (f fakeKeyStore) MirrorTo(basedir vfs.Path) error {
	return f.MirrorToFn(basedir)
}

// build a generic minimal cluster
func buildMinimalCluster(clusterName string, masterPublicName string) *kops.Cluster {
	cluster := testutils.BuildMinimalCluster(clusterName)
	cluster.Spec.MasterPublicName = masterPublicName
	cluster.Spec.MasterInternalName = fmt.Sprintf("internal.%v", masterPublicName)
	return cluster
}

// create a fake certificate
func fakeCertificate() *pki.Certificate {
	cert, _ := pki.ParsePEMCertificate([]byte(certData))
	return cert
}

// create a fake private key
func fakePrivateKey() *pki.PrivateKey {
	key, _ := pki.ParsePEMPrivateKey([]byte(privatekeyData))
	return key
}

func TestBuildKubecfg(t *testing.T) {
	originalPKIDefaultPrivateKeySize := pki.DefaultPrivateKeySize
	pki.DefaultPrivateKeySize = 512
	defer func() {
		pki.DefaultPrivateKeySize = originalPKIDefaultPrivateKeySize
	}()

	type args struct {
		cluster                     *kops.Cluster
		secretStore                 fi.SecretStore
		status                      fakeStatusStore
		admin                       time.Duration
		user                        string
		internal                    bool
		useKopsAuthenticationPlugin bool
	}

	publiccluster := buildMinimalCluster("testcluster", "testcluster.test.com")
	emptyMasterPublicNameCluster := buildMinimalCluster("emptyMasterPublicNameCluster", "")
	gossipCluster := buildMinimalCluster("testgossipcluster.k8s.local", "")

	tests := []struct {
		name           string
		args           args
		want           *KubeconfigBuilder
		wantErr        bool
		wantClientCert bool
	}{
		{
			name: "Test Kube Config Data For Public DNS with admin",
			args: args{
				cluster: publiccluster,
				status:  fakeStatusStore{},
				admin:   DefaultKubecfgAdminLifetime,
				user:    "",
			},
			want: &KubeconfigBuilder{
				Context: "testcluster",
				Server:  "https://testcluster.test.com",
				CACert:  []byte(certData),
				User:    "testcluster",
			},
			wantClientCert: true,
		},
		{
			name: "Test Kube Config Data For Public DNS without admin",
			args: args{
				cluster: publiccluster,
				status:  fakeStatusStore{},
				admin:   0,
				user:    "myuser",
			},
			want: &KubeconfigBuilder{
				Context: "testcluster",
				Server:  "https://testcluster.test.com",
				CACert:  []byte(certData),
				User:    "myuser",
			},
			wantClientCert: false,
		},
		{
			name: "Test Kube Config Data For Public DNS with Empty Master Name",
			args: args{
				cluster: emptyMasterPublicNameCluster,
				status:  fakeStatusStore{},
				admin:   0,
				user:    "",
			},
			want: &KubeconfigBuilder{
				Context: "emptyMasterPublicNameCluster",
				Server:  "https://api.emptyMasterPublicNameCluster",
				CACert:  []byte(certData),
				User:    "emptyMasterPublicNameCluster",
			},
			wantClientCert: false,
		},
		{
			name: "Test Kube Config Data For Gossip cluster",
			args: args{
				cluster: gossipCluster,
				status: fakeStatusStore{
					GetApiIngressStatusFn: func(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error) {
						return []kops.ApiIngressStatus{
							{
								Hostname: "elbHostName",
							},
						}, nil
					},
				},
			},
			want: &KubeconfigBuilder{
				Context: "testgossipcluster.k8s.local",
				Server:  "https://elbHostName",
				CACert:  []byte(certData),
				User:    "testgossipcluster.k8s.local",
			},
			wantClientCert: false,
		},
		{
			name: "Public DNS with kops auth plugin",
			args: args{
				cluster:                     publiccluster,
				status:                      fakeStatusStore{},
				admin:                       0,
				useKopsAuthenticationPlugin: true,
			},
			want: &KubeconfigBuilder{
				Context: "testcluster",
				Server:  "https://testcluster.test.com",
				CACert:  []byte(certData),
				User:    "testcluster",
				AuthenticationExec: []string{
					"kops",
					"helpers",
					"kubectl-auth",
					"--cluster=testcluster",
					"--state=memfs://example-state-store",
				},
			},
			wantClientCert: false,
		},
		{
			name: "Test Kube Config Data For internal DNS name with admin",
			args: args{
				cluster:  publiccluster,
				status:   fakeStatusStore{},
				admin:    DefaultKubecfgAdminLifetime,
				internal: true,
			},
			want: &KubeconfigBuilder{
				Context: "testcluster",
				Server:  "https://internal.testcluster.test.com",
				CACert:  []byte(certData),
				User:    "testcluster",
			},
			wantClientCert: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kopsStateStore := "memfs://example-state-store"

			keyStore := fakeKeyStore{
				FindKeypairFn: func(name string) (*pki.Certificate, *pki.PrivateKey, bool, error) {
					return fakeCertificate(),
						fakePrivateKey(),
						true,
						nil
				},
			}

			got, err := BuildKubecfg(tt.args.cluster, keyStore, tt.args.secretStore, tt.args.status, tt.args.admin, tt.args.user, tt.args.internal, kopsStateStore, tt.args.useKopsAuthenticationPlugin)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildKubecfg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantClientCert {
				if got.ClientCert == nil {
					t.Errorf("Expected ClientCert, got nil")
				}
				if got.ClientKey == nil {
					t.Errorf("Expected ClientKey, got nil")
				}
				tt.want.ClientCert = got.ClientCert
				tt.want.ClientKey = got.ClientKey
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("BuildKubecfg() diff (+got, -want): %s", diff)
			}
		})
	}
}
