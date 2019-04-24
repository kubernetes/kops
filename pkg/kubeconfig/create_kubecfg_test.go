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
	"reflect"
	"testing"
	
	"crypto/x509"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)


type fakeKeyStore struct {
	FindKeypairFn      func(name string) (*pki.Certificate, *pki.PrivateKey, fi.KeysetFormat, error)
	
	CreateKeypairFn    func(signer string, name string, template *x509.Certificate, privateKey *pki.PrivateKey) (*pki.Certificate, error)

	// StoreKeypair writes the keypair to the store
	StoreKeypairFn     func(id string, cert *pki.Certificate, privateKey *pki.PrivateKey) error

	// MirrorTo will copy secrets to a vfs.Path, which is often easier for a machine to read
	MirrorToFn         func(basedir vfs.Path) error
}

func (f fakeKeyStore) FindKeypair(name string) (*pki.Certificate, *pki.PrivateKey, fi.KeysetFormat, error) {
	return f.FindKeypairFn(name)
}

func (f fakeKeyStore) CreateKeypair(signer string, name string, template *x509.Certificate, privateKey *pki.PrivateKey) (*pki.Certificate, error) {
	return f.CreateKeypairFn(signer, name, template, privateKey)
}

func (f fakeKeyStore) StoreKeypair(id string, cert *pki.Certificate, privateKey *pki.PrivateKey) error {
	return f.StoreKeypairFn(id, cert, privateKey)
}

func (f fakeKeyStore) MirrorTo(basedir vfs.Path) error {
	return f.MirrorToFn(basedir)
}

func buildMinimalCluster() *kops.Cluster {
	c := &kops.Cluster{}
	c.ObjectMeta.Name = "testcluster.test.com"
	c.Spec.KubernetesVersion = "1.4.6"
	c.Spec.Subnets = []kops.ClusterSubnetSpec{
		{Name: "subnet-us-mock-1a", Zone: "us-mock-1a", CIDR: "172.20.1.0/24", Type: kops.SubnetTypePrivate},
	}

	c.Spec.MasterPublicName = "testcluster.test.com"
	c.Spec.KubernetesAPIAccess = []string{"0.0.0.0/0"}
	c.Spec.SSHAccess = []string{"0.0.0.0/0"}

	// Default to public topology
	c.Spec.Topology = &kops.TopologySpec{
		Masters: kops.TopologyPublic,
		Nodes:   kops.TopologyPublic,
		DNS:     &kops.DNSSpec{
			Type: kops.DNSTypePublic,
		},
	}

	c.Spec.NetworkCIDR = "172.20.0.0/16"
	c.Spec.NonMasqueradeCIDR = "100.64.0.0/10"
	c.Spec.CloudProvider = "aws"

	c.Spec.ConfigBase = "s3://unittest-bucket/"

	// Required to stop a call to cloud provider
	// TODO: Mock cloudprovider
	c.Spec.DNSZone = "test.com"

	return c
}

func fakeCertificate() (*pki.Certificate) {
	data := "-----BEGIN CERTIFICATE-----\nMIIC2DCCAcCgAwIBAgIRALJXAkVj964tq67wMSI8oJQwDQYJKoZIhvcNAQELBQAw\nFTETMBEGA1UEAxMKa3ViZXJuZXRlczAeFw0xNzEyMjcyMzUyNDBaFw0yNzEyMjcy\nMzUyNDBaMBUxEzARBgNVBAMTCmt1YmVybmV0ZXMwggEiMA0GCSqGSIb3DQEBAQUA\nA4IBDwAwggEKAoIBAQDgnCkSmtnmfxEgS3qNPaUCH5QOBGDH/inHbWCODLBCK9gd\nXEcBl7FVv8T2kFr1DYb0HVDtMI7tixRVFDLgkwNlW34xwWdZXB7GeoFgU1xWOQSY\nOACC8JgYTQ/139HBEvgq4sej67p+/s/SNcw34Kk7HIuFhlk1rRk5kMexKIlJBKP1\nYYUYetsJ/QpUOkqJ5HW4GoetE76YtHnORfYvnybviSMrh2wGGaN6r/s4ChOaIbZC\nAn8/YiPKGIDaZGpj6GXnmXARRX/TIdgSQkLwt0aTDBnPZ4XvtpI8aaL8DYJIqAzA\nNPH2b4/uNylat5jDo0b0G54agMi97+2AUrC9UUXpAgMBAAGjIzAhMA4GA1UdDwEB\n/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBVGR2r\nhzXzRMU5wriPQAJScszNORvoBpXfZoZ09FIupudFxBVU3d4hV9StKnQgPSGA5XQO\nHE97+BxJDuA/rB5oBUsMBjc7y1cde/T6hmi3rLoEYBSnSudCOXJE4G9/0f8byAJe\nrN8+No1r2VgZvZh6p74TEkXv/l3HBPWM7IdUV0HO9JDhSgOVF1fyQKJxRuLJR8jt\nO6mPH2UX0vMwVa4jvwtkddqk2OAdYQvH9rbDjjbzaiW0KnmdueRo92KHAN7BsDZy\nVpXHpqo1Kzg7D3fpaXCf5si7lqqrdJVXH4JC72zxsPehqgi8eIuqOBkiDWmRxAxh\n8yGeRx9AbknHh4Ia\n-----END CERTIFICATE-----\n"

	cert, _ := pki.ParsePEMCertificate([]byte(data))
	return cert
}



func TestBuildKubecfgForPublicDns(t *testing.T) {
	type args struct {
		cluster      *kops.Cluster
		keyStore     fakeKeyStore
		secretStore  fi.SecretStore
		status       kops.StatusStore
		configAccess clientcmd.ConfigAccess
	}

	cluster := buildMinimalCluster()
	
	tests := []struct {
		name    string
		args    args
		want    *KubeconfigBuilder
		wantErr bool
	}{
		{
			"Test Kube Config Data For Public DNS",
			args {
				cluster,
				fakeKeyStore{
					FindKeypairFn: func(name string) (*pki.Certificate, *pki.PrivateKey, fi.KeysetFormat, error) {
							return fakeCertificate(),
							&pki.PrivateKey{

							},
							fi.KeysetFormatLegacy,
							nil
						},
				},
				nil,
				nil,
				nil,
			},
			&KubeconfigBuilder{
				Context: "testcluster.test.com",
				Server: "https://testcluster.test.com",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildKubecfg(tt.args.cluster, tt.args.keyStore, tt.args.secretStore, tt.args.status, tt.args.configAccess)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildKubecfg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildKubecfg() = %v, want %v", got, tt.want)
			}
		})
	}
}
