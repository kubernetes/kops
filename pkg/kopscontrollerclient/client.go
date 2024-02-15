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

package kopscontrollerclient

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
)

type Client struct {
	// Authenticator generates authentication credentials for requests.
	Authenticator bootstrap.Authenticator
	// CAs are the CA certificates for kops-controller.
	CAs []byte

	// BaseURL is the base URL for the server
	BaseURL url.URL

	httpClient *http.Client
}

func (b *Client) Query(ctx context.Context, req any, resp any) error {
	if b.httpClient == nil {
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(b.CAs)

		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    certPool,
				MinVersion: tls.VersionTLS12,
			},
		}

		httpClient := &http.Client{
			Timeout:   time.Duration(15) * time.Second,
			Transport: transport,
		}

		b.httpClient = httpClient
	}

	// Sanity-check DNS to provide clearer diagnostic messages.
	if ips, err := net.LookupIP(b.BaseURL.Hostname()); err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
			return fi.NewTryAgainLaterError(fmt.Sprintf("kops-controller DNS not setup yet (not found: %v)", dnsErr))
		}
		return err
	} else if len(ips) == 1 && (ips[0].String() == cloudup.PlaceholderIP || ips[0].String() == cloudup.PlaceholderIPv6) {
		return fi.NewTryAgainLaterError(fmt.Sprintf("kops-controller DNS not setup yet (placeholder IP found: %v)", ips))
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	bootstrapURL := b.BaseURL
	bootstrapURL.Path = path.Join(bootstrapURL.Path, "/bootstrap")
	httpReq, err := http.NewRequestWithContext(ctx, "POST", bootstrapURL.String(), bytes.NewReader(reqBytes))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	token, err := b.Authenticator.CreateToken(reqBytes)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", token)

	response, err := b.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	if response.Body != nil {
		defer response.Body.Close()
	}

	// if we receive StatusConflict it means that we should exit gracefully
	if response.StatusCode == http.StatusConflict {
		klog.Infof("kops-controller returned status code %d\n\n=====%#+v=======\n\n", response.StatusCode, response.Body)
		os.Exit(0)
	}

	if response.StatusCode != http.StatusOK {
		detail := ""
		if response.Body != nil {
			scanner := bufio.NewScanner(response.Body)
			if scanner.Scan() {
				detail = scanner.Text()
			}
		}
		return fmt.Errorf("kops-controller returned status code %d: %s", response.StatusCode, detail)
	}

	return json.NewDecoder(response.Body).Decode(resp)
}
