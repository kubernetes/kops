/*
Copyright 2023 The Kubernetes Authors.

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

package certresolver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/resolver"
)

type CertResolver struct {
	// RootCAs are used to validate the server certificates.
	RootCAs *x509.CertPool
}

func New(caCertificates string) (resolver.Resolver, error) {
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM([]byte(caCertificates)) {
		klog.Warningf("no CA certs found in boot config")
	}

	return &CertResolver{
		RootCAs: rootCAs,
	}, nil
}

var _ resolver.Resolver = &CertResolver{}

// Resolve resolves the host to IP addresses or alternative hostnames.
func (r *CertResolver) Resolve(ctx context.Context, host string) ([]string, error) {
	tlsConfig := &tls.Config{
		RootCAs: r.RootCAs,
		// This is sort of a hack.  We do want to verify that this is the apiserver.
		ServerName: "kubernetes.default",
	}
	httpTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	httpClient := &http.Client{Transport: httpTransport}

	url := "https://" + host + "/.well-known/cert-discovery"
	response, err := httpClient.Get(url)
	if err != nil {
		klog.Infof("CertResolver unable to resolve %q: %v", err)
		return nil, fmt.Errorf("error doing HTTP get on %q: %w", url, err)
	}

	// We don't really care about the response per-se (we expect a 401), just the certificates

	var records []string
	if response.TLS != nil {
		for _, cert := range response.TLS.PeerCertificates {
			// TODO: Check this cert is signed by our CA?
			for _, ip := range cert.IPAddresses {
				ip := ip.String()
				switch ip {
				case "127.0.0.1":
					// ignore

					// TODO: Ignore others?
					// DNS:kubernetes, DNS:kubernetes.default, DNS:kubernetes.default.svc, DNS:kubernetes.default.svc.cluster.local, DNS:api.internal.foo.k8s.local, IP Address:34.86.xx.xx, IP Address:100.64.0.1, IP Address:127.0.0.1, IP Address:10.0.16.3

				default:
					records = append(records, ip)
				}
			}
		}
	}

	klog.Infof("CertResolver resolved %q to %v", host, records)
	return records, nil
}
