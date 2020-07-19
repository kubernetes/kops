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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
)

type BootstrapClient struct {
	// Authenticator generates authentication credentials for requests.
	Authenticator fi.Authenticator
	// CA is the CA certificate for kops-controller.
	CA []byte

	client *http.Client
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
	}

	err := b.queryBootstrap(c, req)
	if err != nil {
		return err
	}

	return nil
}

func (b *BootstrapClient) queryBootstrap(c *fi.Context, req nodeup.BootstrapRequest) error {
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
		return err
	}

	bootstrapUrl := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(c.Cluster.Spec.MasterInternalName, strconv.Itoa(wellknownports.KopsControllerPort)),
		Path:   "/bootstrap",
	}
	httpReq, err := http.NewRequest("POST", bootstrapUrl.String(), bytes.NewReader(reqBytes))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	token, err := b.Authenticator.CreateToken(reqBytes)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", token)

	resp, err := b.client.Do(httpReq)
	if err != nil {
		return err
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
		return fmt.Errorf("bootstrap returned status code %d: %s", resp.StatusCode, detail)
	}

	var bootstrapResp nodeup.BootstrapResponse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &bootstrapResp)
	if err != nil {
		return err
	}

	return nil
}
