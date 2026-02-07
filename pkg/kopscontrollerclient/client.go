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

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/vfs"
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

func New(authenticator bootstrap.Authenticator, cas []byte, baseURL url.URL) *Client {
	client := &Client{
		Authenticator: authenticator,
		CAs:           cas,
		BaseURL:       baseURL,
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(cas)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    certPool,
			MinVersion: tls.VersionTLS12,
		},
	}
	client.httpClient = &http.Client{
		Timeout:   time.Duration(15) * time.Second,
		Transport: transport,
	}

	return client
}

func (b *Client) Query(ctx context.Context, req any, resp any) error {
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

	backoff := wait.Backoff{
		Duration: 1 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    100,
	}

	var response *http.Response
	done, err := vfs.RetryWithBackoff(backoff, func() (bool, error) {
		httpReq, reqErr := http.NewRequestWithContext(ctx, "POST", bootstrapURL.String(), bytes.NewReader(reqBytes))
		if reqErr != nil {
			return false, reqErr
		}
		httpReq.Header.Set("Content-Type", "application/json")

		token, tokenErr := b.Authenticator.CreateToken(reqBytes)
		if tokenErr != nil {
			return false, tokenErr
		}
		httpReq.Header.Set("Authorization", token)

		resp, doErr := b.httpClient.Do(httpReq)
		if doErr != nil {
			return false, fmt.Errorf("request to kops-controller failed: %w", doErr)
		}

		response = resp
		return true, nil
	})
	if !done {
		if err != nil {
			return err
		}
		return fmt.Errorf("timed out waiting for a successful response from kops-controller")
	}

	// if we receive StatusConflict it means that we should exit gracefully
	if response.StatusCode == http.StatusConflict {
		klog.Infof("kops-controller returned status code %d", response.StatusCode)
		if response.Body != nil {
			response.Body.Close()
		}
		os.Exit(0)
	}

	if response.Body != nil {
		defer response.Body.Close()
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

func (b *Client) Close() {
	if b.httpClient != nil {
		b.httpClient.CloseIdleConnections()
	}
}
