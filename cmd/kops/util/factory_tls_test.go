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

package util_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
)

type tlsTestKeyStore struct {
	fi.KeystoreReader
	keyset *fi.Keyset
}

func (s tlsTestKeyStore) FindKeyset(context.Context, string) (*fi.Keyset, error) {
	return s.keyset, nil
}

type tlsTestCloud struct {
	fi.Cloud
	ingress fi.ApiIngressStatus
}

func (c tlsTestCloud) GetApiIngressStatus(*kops.Cluster) ([]fi.ApiIngressStatus, error) {
	return []fi.ApiIngressStatus{c.ingress}, nil
}

func issueTLSCertificate(t *testing.T, keyStore fi.KeystoreReader, names ...string) (*pki.Certificate, *pki.PrivateKey) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	request := &pki.IssueCertRequest{
		Type: "server", Signer: fi.CertificateIDCA, AlternateNames: names,
		PrivateKey: &pki.PrivateKey{Key: key}, Validity: time.Hour,
	}
	if keyStore == nil {
		request.Type = "ca"
	}
	cert, privateKey, _, err := pki.IssueCert(context.Background(), request, fi.NewPKIKeystoreAdapter(keyStore))
	require.NoError(t, err)
	return cert, privateKey
}

func newTLSKeyStore(t *testing.T) tlsTestKeyStore {
	t.Helper()
	cert, key := issueTLSCertificate(t, nil)
	keyset, err := fi.NewKeyset(cert, key)
	require.NoError(t, err)
	return tlsTestKeyStore{keyset: keyset}
}

// Inspect ClientHello, including failed handshakes, rather than just rest.Config.
func checkTLSConnection(t *testing.T, config *rest.Config, signer fi.KeystoreReader, names []string, wantSNI, requiredSNI, wantError, wantAuthorization string) {
	t.Helper()
	cert, key := issueTLSCertificate(t, signer, names...)
	certPEM, err := cert.AsBytes()
	require.NoError(t, err)
	keyPEM, err := key.AsBytes()
	require.NoError(t, err)
	keypair, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)
	hellos := make(chan string, 1)
	requests := make(chan *http.Request, 1)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r
		w.WriteHeader(http.StatusOK)
	}))
	server.Config.ErrorLog = log.New(io.Discard, "", 0)
	server.TLS = &tls.Config{
		Certificates: []tls.Certificate{keypair},
		GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			hellos <- hello.ServerName
			if requiredSNI != "" && hello.ServerName != requiredSNI {
				return nil, fmt.Errorf("proxy has no route for SNI %q", hello.ServerName)
			}
			return nil, nil
		},
	}
	server.StartTLS()
	t.Cleanup(server.Close)

	config = rest.CopyConfig(config)
	config.Timeout = 5 * time.Second
	config.Proxy = func(*http.Request) (*url.URL, error) { return nil, nil }
	// Redirect only TCP dialing, keeping the URL and TLS identity under test.
	config.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp", server.Listener.Addr().String())
	}
	client, err := rest.HTTPClientFor(config)
	require.NoError(t, err)
	t.Cleanup(client.CloseIdleConnections)
	response, err := client.Get(config.Host)
	switch wantError {
	case "":
		require.NoError(t, err)
		defer response.Body.Close()
		require.Equal(t, http.StatusOK, response.StatusCode)
		require.NotEmpty(t, response.TLS.VerifiedChains)
		request := <-requests
		require.Equal(t, wantAuthorization, request.Header.Get("Authorization"))
		endpoint, err := url.Parse(config.Host)
		require.NoError(t, err)
		require.Equal(t, endpoint.Host, request.Host, "TLS identity must not replace the HTTP Host")
	case "hostname":
		var hostnameError x509.HostnameError
		require.ErrorAs(t, err, &hostnameError)
	case "authority":
		var authorityError x509.UnknownAuthorityError
		require.ErrorAs(t, err, &authorityError)
	case "routing":
		require.ErrorContains(t, err, "remote error: tls:")
		var verificationError *tls.CertificateVerificationError
		require.False(t, errors.As(err, &verificationError), "proxy should reject SNI before certificate verification")
	default:
		t.Fatalf("unknown expected error %q", wantError)
	}
	select {
	case sni := <-hellos:
		require.Equal(t, wantSNI, sni)
	default:
		t.Fatal("server did not receive ClientHello")
	}
}

func TestGeneratedKubeconfigTLSIdentity(t *testing.T) {
	keyStore := newTLSKeyStore(t)
	const internalName = "api.internal.tls.example.com"
	for _, tt := range []struct {
		name        string
		override    string
		internal    bool
		ingress     *fi.ApiIngressStatus
		certificate string
		requiredSNI string
		wantError   string
	}{
		{name: "public DNS"},
		{name: "internal DNS", internal: true},
		{name: "load balancer DNS", ingress: &fi.ApiIngressStatus{Hostname: "load-balancer.example.com"}},
		{name: "load balancer IP", ingress: &fi.ApiIngressStatus{IP: "192.0.2.10"}},
		{name: "explicit DNS alias", override: "https://alias.example.com"},
		{name: "IPv4 tunnel", override: "https://127.0.0.1:6443"},
		{name: "IPv6 tunnel", override: "https://[::1]:6443"},
		{name: "wrong certificate name", override: "https://proxy.example.com", certificate: "proxy.example.com", wantError: "hostname"},
		{name: "proxy requires endpoint SNI", override: "https://proxy.example.com", requiredSNI: "proxy.example.com", wantError: "routing"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cluster := testutils.BuildMinimalClusterAWS("tls.example.com")
			cluster.Spec.API.LoadBalancer = nil
			var cloud fi.Cloud
			if tt.ingress != nil {
				cluster.Spec.API.LoadBalancer = &kops.LoadBalancerAccessSpec{}
				cloud = tlsTestCloud{ingress: *tt.ingress}
			}
			builder, err := kubeconfig.BuildKubecfg(context.Background(), cluster, keyStore, nil, cloud,
				kubeconfig.CreateKubecfgOptions{OverrideAPIServer: tt.override, Internal: tt.internal}, "")
			require.NoError(t, err)
			certificateName := tt.certificate
			if certificateName == "" {
				certificateName = internalName
			}
			certificateNames := []string{certificateName}
			if tt.requiredSNI != "" {
				// Both names are valid: this proxy fails on SNI routing alone.
				certificateNames = append(certificateNames, tt.requiredSNI)
			}
			for _, exported := range []bool{false, true} {
				t.Run(fmt.Sprintf("exported=%t", exported), func(t *testing.T) {
					var config *rest.Config
					if exported {
						path := filepath.Join(t.TempDir(), "config")
						access := clientcmd.NewDefaultPathOptions()
						access.LoadingRules.ExplicitPath = path
						require.NoError(t, builder.WriteKubecfg(access))
						config, err = clientcmd.BuildConfigFromFlags("", path)
					} else {
						config, err = builder.ToRESTConfig()
					}
					require.NoError(t, err)
					checkTLSConnection(t, config, keyStore, certificateNames, internalName, tt.requiredSNI, tt.wantError, "")
				})
			}
		})
	}
}

func TestUseKubeconfigTLSIdentity(t *testing.T) {
	h := testutils.NewIntegrationTestHarness(t)
	t.Cleanup(h.Close)
	h.SetupMockAWS()
	const stateStore = "memfs://tls-identity"
	f := util.NewFactory(&util.FactoryOptions{RegistryPath: stateStore})
	clientset, err := f.KopsClient()
	require.NoError(t, err)
	cluster, err := clientset.CreateCluster(context.Background(), testutils.BuildMinimalClusterAWS("tls.example.com"))
	require.NoError(t, err)
	clusterKeyStore, err := clientset.KeyStore(cluster)
	require.NoError(t, err)
	clusterCA := newTLSKeyStore(t)
	require.NoError(t, clusterKeyStore.StoreKeyset(context.Background(), fi.CertificateIDCA, clusterCA.keyset))
	// The proxy uses its own CA and bearer token, not the cluster's CA and admin certificate.
	proxyKeyStore := newTLSKeyStore(t)
	proxyCA, err := proxyKeyStore.keyset.ToCertificateBytes()
	require.NoError(t, err)

	for _, tt := range []struct {
		name        string
		tlsName     string
		override    string
		certificate string
		wantSNI     string
		wrongCA     bool
		wantError   string
	}{
		{name: "implicit DNS", certificate: "proxy.example.com", wantSNI: "proxy.example.com"},
		{name: "explicit DNS", tlsName: "identity.example.com", certificate: "identity.example.com", wantSNI: "identity.example.com"},
		{name: "override implicit DNS", override: "https://other.example.com:6443", certificate: "other.example.com", wantSNI: "other.example.com"},
		{name: "override explicit DNS", tlsName: "identity.example.com", override: "https://other.example.com:6443", certificate: "identity.example.com", wantSNI: "identity.example.com"},
		{name: "override IPv4", override: "https://127.0.0.1:6443", certificate: "127.0.0.1"},
		{name: "override IPv6", override: "https://[::1]:6443", certificate: "::1"},
		{name: "override IP with explicit DNS", tlsName: "identity.example.com", override: "https://127.0.0.1:6443", certificate: "identity.example.com", wantSNI: "identity.example.com"},
		{name: "explicit IP", tlsName: "192.0.2.10", override: "https://other.example.com:6443", certificate: "192.0.2.10"},
		{name: "wrong implicit name", override: "https://other.example.com", certificate: "proxy.example.com", wantSNI: "other.example.com", wantError: "hostname"},
		{name: "wrong explicit name", tlsName: "identity.example.com", override: "https://other.example.com", certificate: "other.example.com", wantSNI: "identity.example.com", wantError: "hostname"},
		{name: "wrong IP", override: "https://127.0.0.1:6443", certificate: "192.0.2.10", wantError: "hostname"},
		{name: "wrong CA", override: "https://other.example.com", certificate: "other.example.com", wantSNI: "other.example.com", wrongCA: true, wantError: "authority"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config")
			t.Setenv("KUBECONFIG", path)
			config := clientcmdapi.NewConfig()
			config.CurrentContext = "unrelated-context"
			config.Contexts[cluster.Name] = &clientcmdapi.Context{Cluster: "proxy", AuthInfo: "proxy-user"}
			config.Clusters["proxy"] = &clientcmdapi.Cluster{
				Server: "https://proxy.example.com", TLSServerName: tt.tlsName, CertificateAuthorityData: proxyCA,
			}
			config.AuthInfos["proxy-user"] = &clientcmdapi.AuthInfo{Token: "synthetic-proxy-token"}
			if tt.wrongCA {
				config.Clusters["proxy"].CertificateAuthorityData, err = clusterCA.keyset.ToCertificateBytes()
				require.NoError(t, err)
			}
			require.NoError(t, clientcmd.WriteToFile(*config, path))
			// Each case loads a fresh kubeconfig; factory caching is tested separately.
			factory := util.NewFactory(&util.FactoryOptions{RegistryPath: stateStore})
			restConfig, err := factory.RESTConfig(context.Background(), cluster,
				kubeconfig.CreateKubecfgOptions{UseKubeconfig: true, OverrideAPIServer: tt.override})
			require.NoError(t, err)
			checkTLSConnection(t, restConfig, proxyKeyStore, []string{tt.certificate}, tt.wantSNI, tt.wantSNI, tt.wantError, "Bearer synthetic-proxy-token")
		})
	}
}
