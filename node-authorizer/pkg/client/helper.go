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

package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"k8s.io/kops/node-authorizer/pkg/authorizers/alwaysallow"
	"k8s.io/kops/node-authorizer/pkg/authorizers/aws"
	"k8s.io/kops/node-authorizer/pkg/server"

	v1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

// makeHTTPClient is responsible for making a http client
func makeHTTPClient(config *Config) (*http.Client, error) {
	tlsConfig := &tls.Config{}

	if config.TLSClientCAPath != "" {
		ca, err := ioutil.ReadFile(config.TLSClientCAPath)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(ca)
		tlsConfig.RootCAs = caCertPool
	}

	if config.TLSCertPath != "" && config.TLSPrivateKeyPath != "" {
		certs, err := tls.LoadX509KeyPair(config.TLSCertPath, config.TLSPrivateKeyPath)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{certs}
	}
	tlsConfig.BuildNameToCertificate()

	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Dial:                (&net.Dialer{Timeout: 5 * time.Second}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
			TLSClientConfig:     tlsConfig,
		},
	}, nil
}

// makeKubeconfig is responsible for generating a bootstrap config
func makeKubeconfig(ctx context.Context, config *Config, token string) ([]byte, error) {
	// @step: load the certificate authority
	content, err := ioutil.ReadFile(config.TLSClientCAPath)
	if err != nil {
		return []byte{}, err
	}

	// @step: we need to write out the kubeconfig
	name := "bootstrap-context"
	clusterName := "cluster"

	cfg := &v1.Config{
		APIVersion: "v1",
		Kind:       "Config",
		AuthInfos: []v1.NamedAuthInfo{
			{
				Name: name,
				AuthInfo: v1.AuthInfo{
					Token: token,
				},
			},
		},
		Clusters: []v1.NamedCluster{
			{
				Name: clusterName,
				Cluster: v1.Cluster{
					Server:                   config.KubeAPI,
					CertificateAuthorityData: content,
				},
			},
		},
		Contexts: []v1.NamedContext{
			{
				Name: name,
				Context: v1.Context{
					Cluster:  clusterName,
					AuthInfo: name,
				},
			},
		},
		CurrentContext: name,
	}

	return json.MarshalIndent(cfg, "", "  ")
}

// makeTokenRequest makes a request for a bootstrap token
func makeTokenRequest(ctx context.Context, client *http.Client, verifier server.Verifier, config *Config) (*server.NodeRegistration, error) {
	registration := &server.NodeRegistration{}

	// @step: create the request payload
	req, err := verifier.VerifyIdentity(ctx)
	if err != nil {
		return nil, err
	}

	hostname, err := getHostname()
	if err != nil {
		return nil, err
	}

	// @step: make the request to the node-authozier
	url := fmt.Sprintf("%s/authorize/%s", strings.TrimSuffix(config.NodeURL, "/authorize"), hostname)
	resp, err := client.Post(url, "application/json", bytes.NewReader(req))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response code: %d", resp.StatusCode)
	}

	// @step: read in the response and decode
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(content, registration); err != nil {
		return nil, err
	}

	return registration, nil
}

// getHostname attempts to return the instance hostname
func getHostname() (string, error) {
	return os.Hostname()
}

// newNodeVerifier returns a new verifier
func newNodeVerifier(name string) (server.Verifier, error) {
	switch name {
	case "aws":
		return aws.NewVerifier()
	case "alwaysallow":
		return alwaysallow.NewVerifier()
	}

	return nil, errors.New("unsupported authorizer")
}
