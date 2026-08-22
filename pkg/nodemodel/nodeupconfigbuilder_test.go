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

package nodemodel

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"reflect"
	"testing"
	"time"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
)

func TestSelectControlPlaneIPs(t *testing.T) {
	// The addresses of an API load balancer, as FindAddresses returns them: the DNS name,
	// then the ENI addresses sorted as strings, which puts IPv4 ahead of IPv6.
	addresses := []string{
		"api-abc123.elb.us-test-1.amazonaws.com",
		"172.20.6.26",
		"2001:db8:0:113::a",
	}

	grid := []struct {
		name              string
		nonMasqueradeCIDR string
		addresses         []string
		want              []string
	}{
		{
			name:              "ipv4 cluster keeps the private IPv4 address",
			nonMasqueradeCIDR: "100.64.0.0/10",
			addresses:         addresses,
			want:              []string{"172.20.6.26", "2001:db8:0:113::a"},
		},
		{
			// Nodes in an IPv6-only cluster have no IPv4 address at all, so 172.20.6.26 is
			// unroutable from them even though it is inside the network CIDR.
			name:              "ipv6-only cluster drops the IPv4 address",
			nonMasqueradeCIDR: "::/0",
			addresses:         addresses,
			want:              []string{"2001:db8:0:113::a"},
		},
		{
			name:              "public IPv4 addresses are excluded",
			nonMasqueradeCIDR: "100.64.0.0/10",
			addresses:         []string{"203.0.113.7"},
			want:              nil,
		},
	}

	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			cluster := &kops.Cluster{
				Spec: kops.ClusterSpec{
					CloudProvider: kops.CloudProviderSpec{AWS: &kops.AWSSpec{}},
					Networking: kops.NetworkingSpec{
						NetworkCIDR:       "172.20.0.0/16",
						NonMasqueradeCIDR: g.nonMasqueradeCIDR,
					},
				},
			}

			got, err := selectControlPlaneIPs(cluster, g.addresses)
			if err != nil {
				t.Fatalf("selectControlPlaneIPs failed: %v", err)
			}
			if !reflect.DeepEqual(got, g.want) {
				t.Errorf("selectControlPlaneIPs = %v, want %v", got, g.want)
			}
		})
	}
}

func TestBuildConfigServerOptionsUsesTLSServerNameForIPServers(t *testing.T) {
	options := buildConfigServerOptions("cluster.k8s.local", "ca-data", []string{"10.0.1.2"})

	if got, want := options.TLSServerName, "kops-controller.internal.cluster.k8s.local"; got != want {
		t.Fatalf("TLSServerName = %q, want %q", got, want)
	}
	if got, want := options.Servers, []string{"https://10.0.1.2:3988/"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Servers = %v, want %v", got, want)
	}
}

func TestBuildConfigServerOptionsUsesDNSNameByDefault(t *testing.T) {
	options := buildConfigServerOptions("cluster.k8s.local", "ca-data", nil)

	if options.TLSServerName != "" {
		t.Fatalf("TLSServerName = %q, want empty", options.TLSServerName)
	}
	if got, want := options.Servers, []string{"https://kops-controller.internal.cluster.k8s.local:3988/"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Servers = %v, want %v", got, want)
	}
}

func selfSignedTestCert(t *testing.T, commonName string) *pki.Certificate {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating test key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("creating test certificate: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parsing test certificate: %v", err)
	}
	return &pki.Certificate{Certificate: cert}
}

// TestLoadCertificatesNeverEmitsBlankCAForPlaceholderKeyset is a regression test for
// https://github.com/kubernetes/kops/issues/18680: an unresolved (not-yet-created)
// keypair must never silently turn into an empty-string CA bundle, which is
// indistinguishable from a real, dangerous blanking of an existing trust bundle.
func TestLoadCertificatesNeverEmitsBlankCAForPlaceholderKeyset(t *testing.T) {
	config := &nodeup.Config{
		CAs:        map[string]string{},
		KeypairIDs: map[string]string{},
	}
	keysets := map[string]*fi.Keyset{
		fi.CertificateIDCA: {
			Primary: &fi.KeysetItem{Id: fi.PlaceholderKeypairID},
		},
	}

	if err := loadCertificates(keysets, fi.CertificateIDCA, config, true); err != nil {
		t.Fatalf("unexpected error for a placeholder (not-yet-created) keypair: %v", err)
	}

	if got := config.CAs[fi.CertificateIDCA]; got == "" {
		t.Errorf("CAs[%q] = %q; must not be blank for a placeholder keyset", fi.CertificateIDCA, got)
	}
	if got, want := config.CAs[fi.CertificateIDCA], fi.PlaceholderKeypairID; got != want {
		t.Errorf("CAs[%q] = %q, want unambiguous placeholder %q", fi.CertificateIDCA, got, want)
	}
	if got, want := config.KeypairIDs[fi.CertificateIDCA], fi.PlaceholderKeypairID; got != want {
		t.Errorf("KeypairIDs[%q] = %q, want %q", fi.CertificateIDCA, got, want)
	}
}

// TestLoadCertificatesPreservesRealCertificate is the control case: a real,
// resolved keyset must still round-trip its actual certificate bytes unchanged.
func TestLoadCertificatesPreservesRealCertificate(t *testing.T) {
	cert := selfSignedTestCert(t, "kubernetes-ca")

	config := &nodeup.Config{
		CAs:        map[string]string{},
		KeypairIDs: map[string]string{},
	}
	keysets := map[string]*fi.Keyset{
		fi.CertificateIDCA: {
			Primary: &fi.KeysetItem{Id: "123", Certificate: cert},
			Items: map[string]*fi.KeysetItem{
				"123": {Id: "123", Certificate: cert},
			},
		},
	}

	if err := loadCertificates(keysets, fi.CertificateIDCA, config, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := config.CAs[fi.CertificateIDCA]; got == "" || got == fi.PlaceholderKeypairID {
		t.Errorf("CAs[%q] = %q, want the real certificate PEM", fi.CertificateIDCA, got)
	}
	if got, want := config.KeypairIDs[fi.CertificateIDCA], "123"; got != want {
		t.Errorf("KeypairIDs[%q] = %q, want %q", fi.CertificateIDCA, got, want)
	}
}
