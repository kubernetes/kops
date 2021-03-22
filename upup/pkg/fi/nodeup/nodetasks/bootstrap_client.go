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
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/deprecations"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
)

type BootstrapClientTask struct {
	// Certs are the requested certificates.
	Certs map[string]*BootstrapCert

	// Client holds the client wrapper for the kops-bootstrap protocol
	Client *KopsBootstrapClient

	keys map[string]*pki.PrivateKey
}

type BootstrapCert struct {
	Cert *fi.TaskDependentResource
	Key  *fi.TaskDependentResource
}

var _ fi.Task = &BootstrapClientTask{}
var _ fi.HasName = &BootstrapClientTask{}
var _ fi.HasDependencies = &BootstrapClientTask{}

func (b *BootstrapClientTask) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	// BootstrapClient depends on the protokube service to ensure gossip DNS
	var deps []fi.Task
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

func (b *BootstrapClientTask) Run(c *fi.Context) error {
	ctx := context.TODO()

	req := nodeup.BootstrapRequest{
		APIVersion:                 nodeup.BootstrapAPIVersion,
		Certs:                      map[string]string{},
		CertificateSigningRequests: make(map[string]*nodeup.CertificateSigningRequest),
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

		rsaKey := key.Key.(*rsa.PrivateKey)

		if deprecations.ShouldIssueWithCSRs.IsEnabled() {
			pkData, err := x509.MarshalPKIXPublicKey(rsaKey.Public())
			if err != nil {
				return fmt.Errorf("error marshalling public key: %w", err)
			}

			req.Certs[name] = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pkData}))
		}

		csrTemplate := &x509.CertificateRequest{
			SignatureAlgorithm: x509.SHA512WithRSA,
		}

		csrCertificate, err := x509.CreateCertificateRequest(cryptorand.Reader, csrTemplate, rsaKey)
		if err != nil {
			return fmt.Errorf("error building certificate request: %w", err)
		}

		req.CertificateSigningRequests[name] = &nodeup.CertificateSigningRequest{
			PEMData: string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrCertificate})),
		}
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
	Authenticator fi.Authenticator
	// CA is the CA certificate for kops-controller.
	CA []byte

	// BaseURL is the base URL for the server
	BaseURL url.URL

	httpClient *http.Client
}

func (b *KopsBootstrapClient) QueryBootstrap(ctx context.Context, req *nodeup.BootstrapRequest) (*nodeup.BootstrapResponse, error) {
	if b.httpClient == nil {
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(b.CA)

		b.httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:    certPool,
					MinVersion: tls.VersionTLS12,
				},
			},
		}
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &bootstrapResp)
	if err != nil {
		return nil, err
	}

	return &bootstrapResp, nil
}
