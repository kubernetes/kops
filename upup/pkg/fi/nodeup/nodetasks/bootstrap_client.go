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
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
)

type BootstrapClient struct {
	// Authenticator generates authentication credentials for requests.
	Authenticator fi.Authenticator
	// CA is the CA certificate for kops-controller.
	CA []byte
	// Certs are the requested certificates.
	Certs map[string]*BootstrapCert

	client *http.Client
	keys   map[string]*pki.PrivateKey
}

type BootstrapCert struct {
	Cert *fi.TaskDependentResource
	Key  *fi.TaskDependentResource
}

var _ fi.Task = &BootstrapClient{}
var _ fi.HasName = &BootstrapClient{}
var _ fi.HasDependencies = &BootstrapClient{}

func (b *BootstrapClient) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

func (b *BootstrapClient) GetName() *string {
	name := "BootstrapClient"
	return &name
}

func (b *BootstrapClient) String() string {
	return "BootstrapClient"
}

func (b *BootstrapClient) Run(c *fi.Context) error {
	req := nodeup.BootstrapRequest{
		APIVersion: nodeup.BootstrapAPIVersion,
		Certs:      map[string]string{},
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

		pkData, err := x509.MarshalPKIXPublicKey(key.Key.(*rsa.PrivateKey).Public())
		if err != nil {
			return fmt.Errorf("marshalling public key: %v", err)
		}
		// TODO perhaps send a CSR instead to prove we own the private key?
		req.Certs[name] = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pkData}))
	}

	resp, err := b.queryBootstrap(c, &req)
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

func (b *BootstrapClient) queryBootstrap(c *fi.Context, req *nodeup.BootstrapRequest) (*nodeup.BootstrapResponse, error) {
	if b.client == nil {
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(b.CA)

		b.client = &http.Client{
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

	bootstrapUrl := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(c.Cluster.Spec.MasterInternalName, strconv.Itoa(wellknownports.KopsControllerPort)),
		Path:   "/bootstrap",
	}
	httpReq, err := http.NewRequest("POST", bootstrapUrl.String(), bytes.NewReader(reqBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	token, err := b.Authenticator.CreateToken(reqBytes)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", token)

	resp, err := b.client.Do(httpReq)
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
