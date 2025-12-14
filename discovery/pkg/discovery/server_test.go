/*
Copyright 2025 The Kubernetes Authors.

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

package discovery

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	api "k8s.io/kops/discovery/apis/discovery.kops.k8s.io/v1alpha1"
)

// ServerHarness wraps the test server setup
type ServerHarness struct {
	Store     Store
	Server    *httptest.Server
	ServerURL string
}

func NewServerHarness() *ServerHarness {
	store := NewMemoryStore()
	handler := NewServer(store)
	server := httptest.NewUnstartedServer(handler)
	server.TLS = &tls.Config{
		ClientAuth: tls.RequestClientCert,
	}
	server.StartTLS()

	return &ServerHarness{
		Store:     store,
		Server:    server,
		ServerURL: server.URL,
	}
}

func (h *ServerHarness) Close() {
	h.Server.Close()
}

func (h *ServerHarness) Certificate() *x509.Certificate {
	return h.Server.TLS.Certificates[0].Leaf
}

func (h *ServerHarness) NewUniverse(t *testing.T, name string) *UniverseHarness {
	ca, caKey, err := generateCA(name)
	if err != nil {
		t.Fatalf("Failed to generate CA %s: %v", name, err)
	}
	hash := sha256.Sum256(ca.RawSubjectPublicKeyInfo)
	id := hex.EncodeToString(hash[:])

	return &UniverseHarness{
		T:      t,
		Server: h,
		CA:     ca,
		CAKey:  caKey,
		ID:     id,
	}
}

func generateCA(cn string) (*x509.Certificate, *rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	cert, err := x509.ParseCertificate(der)
	return cert, key, err
}

// UniverseHarness wraps a Universe (CA) context
type UniverseHarness struct {
	T *testing.T

	Server *ServerHarness
	CA     *x509.Certificate
	CAKey  *rsa.PrivateKey
	ID     string
}

// ClientHarness wraps the client-side logic
type ClientHarness struct {
	KubeClient dynamic.Interface
	HTTPClient *http.Client
	UniverseID string
	ServerURL  string
	Name       string // Client Identity (CN)
}

func pemEncodeCerts(certs ...*x509.Certificate) []byte {
	var b bytes.Buffer
	for _, cert := range certs {
		pem.Encode(&b, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	}
	return b.Bytes()
}

func pemEncodeKey(key *rsa.PrivateKey) []byte {
	b := x509.MarshalPKCS1PrivateKey(key)
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: b})
}

func (h *UniverseHarness) NewClient(clientName string) *ClientHarness {
	clientCert, clientKey, err := generateClientCert(clientName, h.CA, h.CAKey)
	if err != nil {
		h.T.Fatalf("failed to generate client cert: %v", err)
	}

	hash := sha256.Sum256(h.CA.RawSubjectPublicKeyInfo)
	universeID := hex.EncodeToString(hash[:])

	// 1. Configure Dynamic Client (k8s protocol)
	config := &rest.Config{
		Host: h.Server.ServerURL + "/" + universeID,
		TLSClientConfig: rest.TLSClientConfig{
			CAData:   pemEncodeCerts(h.Server.Certificate()),
			CertData: pemEncodeCerts(clientCert, h.CA),
			KeyData:  pemEncodeKey(clientKey),
		},
	}
	kubeClient, err := dynamic.NewForConfig(config)
	if err != nil {
		h.T.Fatalf("failed to create dynamic client: %v", err)
	}

	// 2. Configure HTTP Client (for OIDC/raw requests)
	tlsCert := tls.Certificate{
		Certificate: [][]byte{clientCert.Raw, h.CA.Raw},
		PrivateKey:  clientKey,
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: func() *x509.CertPool {
					pool := x509.NewCertPool()
					pool.AddCert(h.Server.Certificate())
					return pool
				}(),
				Certificates: []tls.Certificate{tlsCert},
			},
		},
	}

	return &ClientHarness{
		KubeClient: kubeClient,
		HTTPClient: httpClient,
		UniverseID: universeID,
		ServerURL:  h.Server.ServerURL,
		Name:       clientName,
	}
}

func generateClientCert(cn string, ca *x509.Certificate, caKey *rsa.PrivateKey) (*x509.Certificate, *rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(2), // Randomize if needed
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, ca, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}
	cert, err := x509.ParseCertificate(der)
	return cert, key, err
}

// NewAnonymousClient creates a client without client certs (for public endpoints)
func (h *UniverseHarness) NewAnonymousClient(serverURL string) *ClientHarness {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: func() *x509.CertPool {
					pool := x509.NewCertPool()
					pool.AddCert(h.Server.Certificate())
					return pool
				}(),
			},
		},
	}
	return &ClientHarness{
		HTTPClient: client,
		UniverseID: h.ID,
		ServerURL:  serverURL,
		Name:       "public",
	}
}

func (c *ClientHarness) Get(path string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s/%s", c.ServerURL, c.UniverseID, path)
	return c.HTTPClient.Get(url)
}

// RegisterDiscoveryEndpoint uses the dynamic client to create/apply the endpoint
func (c *ClientHarness) RegisterDiscoveryEndpoint(ns string, spec api.DiscoveryEndpointSpec) (*api.DiscoveryEndpoint, error) {
	ep := &api.DiscoveryEndpoint{
		Spec: spec,
	}

	ep.Kind = "DiscoveryEndpoint"
	ep.APIVersion = "discovery.kops.k8s.io/v1alpha1"

	ep.Name = c.Name
	ep.Namespace = ns

	// Convert to Unstructured
	uContent, err := runtime.DefaultUnstructuredConverter.ToUnstructured(ep)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to unstructured: %v", err)
	}
	u := &unstructured.Unstructured{Object: uContent}

	gvr := api.DiscoveryEndpointGVR

	// Use Server-Side Apply to Create/Update
	// Note: We use "Apply" to mimic kubectl apply --server-side
	// But standard Create is also fine. Let's use Create for Registration.
	// Update: The previous test used POST (Create).
	created, err := c.KubeClient.Resource(gvr).Namespace(ns).Create(context.Background(), u, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	var result api.DiscoveryEndpoint
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(created.Object, &result); err != nil {
		return nil, fmt.Errorf("failed to convert from unstructured: %v", err)
	}
	return &result, nil
}

func (c *ClientHarness) ListDiscoveryEndpoints(ns string) (*api.DiscoveryEndpointList, error) {
	gvr := api.DiscoveryEndpointGVR

	list, err := c.KubeClient.Resource(gvr).Namespace(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := &api.DiscoveryEndpointList{
		Items: make([]api.DiscoveryEndpoint, len(list.Items)),
	}
	for i, item := range list.Items {
		var ep api.DiscoveryEndpoint
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &ep); err != nil {
			return nil, fmt.Errorf("failed to convert item %d: %v", i, err)
		}
		result.Items[i] = ep
	}
	return result, nil
}

// Tests

func TestDiscoveryIsolation(t *testing.T) {
	h := NewServerHarness()
	defer h.Close()

	// 1. Setup Universe 1
	u1 := h.NewUniverse(t, "Universe 1 CA")
	client1 := u1.NewClient("client1")

	// 2. Setup Universe 2
	u2 := h.NewUniverse(t, "Universe 2 CA")
	client2 := u2.NewClient("client2")

	// Client 1 Registers
	_, err := client1.RegisterDiscoveryEndpoint("default", api.DiscoveryEndpointSpec{Addresses: []string{"1.2.3.4"}})
	if err != nil {
		t.Fatalf("Client 1 Register failed: %v", err)
	}

	// Client 2 Registers
	_, err = client2.RegisterDiscoveryEndpoint("default", api.DiscoveryEndpointSpec{Addresses: []string{"5.6.7.8"}})
	if err != nil {
		t.Fatalf("Client 2 Register failed: %v", err)
	}

	// Client 1 Lists - Should see only Client 1
	list1, err := client1.ListDiscoveryEndpoints("default")
	if err != nil {
		t.Fatalf("Client 1 List failed: %v", err)
	}
	if len(list1.Items) != 1 {
		t.Errorf("Client 1 should see 1 node, saw %d", len(list1.Items))
	} else if list1.Items[0].ObjectMeta.Name != "client1" {
		t.Errorf("Client 1 saw wrong node: %s", list1.Items[0].ObjectMeta.Name)
	}

	// Client 2 Lists - Should see only Client 2
	list2, err := client2.ListDiscoveryEndpoints("default")
	if err != nil {
		t.Fatalf("Client 2 List failed: %v", err)
	}
	if len(list2.Items) != 1 {
		t.Errorf("Client 2 should see 1 node, saw %d", len(list2.Items))
	} else if list2.Items[0].ObjectMeta.Name != "client2" {
		t.Errorf("Client 2 saw wrong node: %s", list2.Items[0].ObjectMeta.Name)
	}
}

func TestOIDCDiscovery(t *testing.T) {
	h := NewServerHarness()
	defer h.Close()

	u := h.NewUniverse(t, "Universe CA")
	client := u.NewClient("client1")

	// Register Endpoint with OIDC
	jwksKey := api.JSONWebKey{KeyType: "RSA", KeyID: "1"}

	_, err := client.RegisterDiscoveryEndpoint("default", api.DiscoveryEndpointSpec{
		Addresses: []string{"1.2.3.4"},
		OIDC: &api.OIDCSpec{
			Keys: []api.JSONWebKey{jwksKey},
		},
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Query Public OIDC Endpoint
	anonymousClient := u.NewAnonymousClient(h.ServerURL)

	// 1. OIDC Discovery
	resp, err := anonymousClient.Get(".well-known/openid-configuration")
	if err != nil {
		t.Fatalf("Get OIDC failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Get OIDC status: %v", resp.Status)
	}
	defer resp.Body.Close()

	var oidcResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&oidcResp); err != nil {
		t.Fatalf("Decode OIDC failed: %v", err)
	}
	oidcIssuer := anonymousClient.ServerURL + "/" + anonymousClient.UniverseID + "/"
	if oidcResp["issuer"] != oidcIssuer {
		t.Errorf("Expected issuer %s, got %s", oidcIssuer, oidcResp["issuer"])
	}
	expectedJWKSURI := oidcIssuer + "openid/v1/jwks"
	if oidcResp["jwks_uri"] != expectedJWKSURI {
		t.Errorf("Expected jwks_uri %s, got %s", expectedJWKSURI, oidcResp["jwks_uri"])
	}

	// 2. JWKS
	resp, err = anonymousClient.Get("openid/v1/jwks")
	if err != nil {
		t.Fatalf("Get JWKS failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Get JWKS status: %v", resp.Status)
	}
	defer resp.Body.Close()

	var jwksResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jwksResp); err != nil {
		t.Fatalf("Decode JWKS failed: %v", err)
	}
	keys := jwksResp["keys"].([]interface{})
	if len(keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(keys))
	} else {
		key1 := keys[0].(map[string]interface{})
		if key1["kid"] != "1" {
			t.Errorf("Expected kid 1, got %v", key1["kid"])
		}
	}
}

func TestOIDCMerging(t *testing.T) {
	h := NewServerHarness()
	defer h.Close()

	u := h.NewUniverse(t, "Universe CA")

	// Helper to register endpoint
	register := func(name string, keys []api.JSONWebKey) {
		client := u.NewClient(name)
		_, err := client.RegisterDiscoveryEndpoint("default", api.DiscoveryEndpointSpec{
			OIDC: &api.OIDCSpec{
				Keys: keys,
			},
		})
		if err != nil {
			t.Fatalf("Register %s failed: %v", name, err)
		}
	}

	// Scenario:
	// Client A: Has Key 1 (old) and Key 2 (new)
	// Client B: Has Key 1 (newer)

	key1Old := api.JSONWebKey{KeyID: "1", N: "old"}
	key1New := api.JSONWebKey{KeyID: "1", N: "new"}
	key2 := api.JSONWebKey{KeyID: "2", N: "static"}

	register("clientA", []api.JSONWebKey{key1Old, key2})
	// Ensure time passes so LastSeen is updated
	time.Sleep(2 * time.Second)
	register("clientB", []api.JSONWebKey{key1New})

	// Query JWKS using a public client harness
	anonymousClient := u.NewAnonymousClient(h.ServerURL)

	resp, err := anonymousClient.Get("openid/v1/jwks")
	if err != nil {
		t.Fatalf("Get JWKS failed: %v", err)
	}
	defer resp.Body.Close()

	var jwksResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&jwksResp)

	keys := jwksResp["keys"].([]interface{})
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}

	foundNew := false
	for _, k := range keys {
		km := k.(map[string]interface{})
		if km["kid"] == "1" {
			if km["n"] == "new" {
				foundNew = true
			} else {
				t.Errorf("Expected key 1 to be 'new', got '%v'", km["n"])
			}
		}
	}
	if !foundNew {
		t.Error("Key 1 not found")
	}
}

func TestServerSideApply(t *testing.T) {
	h := NewServerHarness()
	defer h.Close()

	u := h.NewUniverse(t, "Universe CA")
	client := u.NewClient("client1")

	// 1. Initial Create (using Register helper which uses dynamic Create)
	ep, err := client.RegisterDiscoveryEndpoint("default", api.DiscoveryEndpointSpec{Addresses: []string{"1.2.3.4"}})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if len(ep.Spec.Addresses) != 1 || ep.Spec.Addresses[0] != "1.2.3.4" {
		t.Errorf("Unexpected addresses: %v", ep.Spec.Addresses)
	}

	// 2. Patch (Apply) - Update addresses using Patch with ApplyPatchType
	ep.Spec.Addresses = []string{"5.6.7.8"}
	// Ensure TypeMeta is set for conversion/application
	ep.TypeMeta.Kind = "DiscoveryEndpoint"
	ep.TypeMeta.APIVersion = "discovery.kops.k8s.io/v1alpha1"

	uContent, err := runtime.DefaultUnstructuredConverter.ToUnstructured(ep)
	if err != nil {
		t.Fatalf("Failed to convert to unstructured: %v", err)
	}

	patchData, err := json.Marshal(uContent)
	if err != nil {
		t.Fatalf("Failed to marshal patch data: %v", err)
	}

	gvr := api.DiscoveryEndpointGVR

	// Use server-side apply patch
	patchOpts := metav1.PatchOptions{
		FieldManager: "client1",
	}

	_, err = client.KubeClient.Resource(gvr).Namespace("default").Patch(context.Background(), client.Name, types.ApplyPatchType, patchData, patchOpts)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}

	// 3. Verify Update
	list, err := client.ListDiscoveryEndpoints("default")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list.Items) == 0 {
		t.Fatalf("List returned no items")
	}
	if list.Items[0].Spec.Addresses[0] != "5.6.7.8" {
		t.Errorf("Patch did not update addresses. Got: %v", list.Items[0].Spec.Addresses)
	}
}
