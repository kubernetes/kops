/*
Copyright 2020 The Kubernetes Authors.

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

package nodetasks

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/resolver"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
)

type BootstrapClientTask struct {
	// Certs are the requested certificates.
	Certs map[string]*BootstrapCert
	// KeypairIDs are the keypair IDs of the CAs to use for issuing certificates.
	KeypairIDs map[string]string

	// Client holds the client wrapper for the kops-bootstrap protocol
	Client *KopsBootstrapClient

	keys map[string]*pki.PrivateKey
}

type BootstrapCert struct {
	Cert *fi.NodeupTaskDependentResource
	Key  *fi.NodeupTaskDependentResource
}

var (
	_ fi.NodeupTask            = &BootstrapClientTask{}
	_ fi.HasName               = &BootstrapClientTask{}
	_ fi.NodeupHasDependencies = &BootstrapClientTask{}
)

func (b *BootstrapClientTask) GetDependencies(tasks map[string]fi.NodeupTask) []fi.NodeupTask {
	// BootstrapClient depends on the protokube service to ensure gossip DNS
	var deps []fi.NodeupTask
	for _, v := range tasks {
		if svc, ok := v.(*Service); ok && svc.Name == protokubeService {
			deps = append(deps, v)
		}
	}
	return deps
}

func (b *BootstrapClientTask) GetName() *string {
	name := "BootstrapClient"
	return &name
}

func (b *BootstrapClientTask) String() string {
	return "BootstrapClientTask"
}

func (b *BootstrapClientTask) Run(c *fi.NodeupContext) error {
	ctx := context.TODO()

	req := nodeup.BootstrapRequest{
		APIVersion: nodeup.BootstrapAPIVersion,
		Certs:      map[string]string{},
		KeypairIDs: b.KeypairIDs,
	}

	if b.keys == nil {
		b.keys = map[string]*pki.PrivateKey{}
	}

	for name, certRequest := range b.Certs {
		key, ok := b.keys[name]
		if !ok {
			var err error
			key, err = pki.GeneratePrivateKey()
			if err != nil {
				return fmt.Errorf("generating private key: %v", err)
			}

			certRequest.Key.Resource = &asBytesResource{key}
			b.keys[name] = key
		}

		pkData, err := x509.MarshalPKIXPublicKey(key.Key.Public())
		if err != nil {
			return fmt.Errorf("marshalling public key: %v", err)
		}
		// TODO perhaps send a CSR instead to prove we own the private key?
		req.Certs[name] = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pkData}))
	}

	resp, err := b.Client.QueryBootstrap(ctx, &req)
	if err != nil {
		return err
	}

	for name, certRequest := range b.Certs {
		cert, ok := resp.Certs[name]
		if !ok {
			return fmt.Errorf("kops-controller did not return a %q certificate", name)
		}
		certificate, err := pki.ParsePEMCertificate([]byte(cert))
		if err != nil {
			return fmt.Errorf("parsing %q certificate: %v", name, err)
		}
		certRequest.Cert.Resource = asBytesResource{certificate}
	}

	return nil
}

type KopsBootstrapClient struct {
	// Authenticator generates authentication credentials for requests.
	Authenticator bootstrap.Authenticator
	// CAs are the CA certificates for kops-controller.
	CAs []byte

	// BaseURL is the base URL for the server
	BaseURL url.URL

	// Resolver is a custom resolver that supports resolution of hostnames without requiring DNS.
	// In particular, this supports gossip mode.
	Resolver resolver.Resolver

	httpClient *http.Client
}

// dial implements a DialContext resolver function, for when a custom resolver is in use
func (b *KopsBootstrapClient) dial(ctx context.Context, network, addr string) (net.Conn, error) {
	var errors []error

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("cannot split host and port from %q: %w", addr, err)
	}

	// TODO: cache?
	addresses, err := b.Resolver.Resolve(ctx, host)
	if err != nil {
		return nil, err
	}

	klog.Infof("resolved %q to %v", host, addresses)

	for _, addr := range addresses {
		timeout := 5 * time.Second
		conn, err := net.DialTimeout(network, addr+":"+port, timeout)
		if err == nil {
			return conn, nil
		}
		if err != nil {
			klog.Warningf("failed to dial %q: %v", addr, err)
			errors = append(errors, err)
		}
	}
	if len(errors) == 0 {
		return nil, fmt.Errorf("no addresses for %q", addr)
	}
	return nil, errors[0]
}

func (b *KopsBootstrapClient) QueryBootstrap(ctx context.Context, req *nodeup.BootstrapRequest) (*nodeup.BootstrapResponse, error) {
	if b.httpClient == nil {
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(b.CAs)

		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    certPool,
				MinVersion: tls.VersionTLS12,
			},
		}

		if b.Resolver != nil {
			transport.DialContext = b.dial
		}

		httpClient := &http.Client{
			Timeout:   time.Duration(15) * time.Second,
			Transport: transport,
		}

		b.httpClient = httpClient
	}

	// Sanity-check DNS to provide clearer diagnostic messages.
	if b.Resolver != nil {
		// Don't check DNS when there's a custom resolver.
	} else if ips, err := net.LookupIP(b.BaseURL.Hostname()); err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
			return nil, fi.NewTryAgainLaterError(fmt.Sprintf("kops-controller DNS not setup yet (not found: %v)", dnsErr))
		}
		return nil, err
	} else if len(ips) == 1 && (ips[0].String() == cloudup.PlaceholderIP || ips[0].String() == cloudup.PlaceholderIPv6) {
		return nil, fi.NewTryAgainLaterError(fmt.Sprintf("kops-controller DNS not setup yet (placeholder IP found: %v)", ips))
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	bootstrapURL := b.BaseURL
	bootstrapURL.Path = path.Join(bootstrapURL.Path, "/bootstrap")
	httpReq, err := http.NewRequestWithContext(ctx, "POST", bootstrapURL.String(), bytes.NewReader(reqBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	token, err := b.Authenticator.CreateToken(reqBytes)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", token)

	resp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		detail := ""
		if resp.Body != nil {
			scanner := bufio.NewScanner(resp.Body)
			if scanner.Scan() {
				detail = scanner.Text()
			}
		}
		return nil, fmt.Errorf("bootstrap returned status code %d: %s", resp.StatusCode, detail)
	}

	var bootstrapResp nodeup.BootstrapResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &bootstrapResp)
	if err != nil {
		return nil, err
	}

	return &bootstrapResp, nil
}
