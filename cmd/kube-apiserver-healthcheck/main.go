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

package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"k8s.io/klog"
)

// healthCheckServer is the http server
type healthCheckServer struct {
	transport *http.Transport
}

// handler processes a single http request
func (s *healthCheckServer) handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" && r.URL.Path == "/.kube-apiserver-healthcheck/healthz" {
		// This is a check for our own health
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
		return
	}

	if proxyRequest := mapToProxyRequest(r); proxyRequest != nil {
		s.proxyRequest(w, proxyRequest)
		return
	}

	klog.Infof("unknown request: %s %s", r.Method, r.URL.Path)
	http.Error(w, "not found", http.StatusNotFound)
}

// httpClient builds an isolated http.Client
func (s *healthCheckServer) httpClient() *http.Client {
	return &http.Client{Transport: s.transport}
}

// mapToProxyRequest returns the request we should make to the apiserver,
// or nil if the query is not on the safelist
func mapToProxyRequest(r *http.Request) *http.Request {
	if r.Method == "GET" {
		switch r.URL.Path {
		case "/livez", "/healthz", "/readyz":
			// This is a health-check we will proxy
			return sanitizeRequest(r, []string{"exclude"})
		}
	}
	return nil
}

// sanitizeRequest builds the request we should pass to the target apiserver,
// passing through only allowedQueryParameters
func sanitizeRequest(r *http.Request, allowedQueryParameters []string) *http.Request {
	u := &url.URL{
		Scheme: "https",
		Host:   "127.0.0.1",
		Path:   r.URL.Path,
	}

	// Pass-through (only) the parameters in allowedQueryParameters
	{
		in := r.URL.Query()
		out := make(url.Values)

		for _, k := range allowedQueryParameters {
			for _, v := range in[k] {
				out.Add(k, v)
			}
		}
		u.RawQuery = out.Encode()
	}

	req := &http.Request{
		Method: r.Method,
		URL:    u,
	}

	return req
}

// proxyRequest forwards a request, that has been sanitized by mapToProxyRequest/buildProxyRequest
func (s *healthCheckServer) proxyRequest(w http.ResponseWriter, forwardRequest *http.Request) {
	httpClient := s.httpClient()

	resp, err := httpClient.Do(forwardRequest)
	if err != nil {
		klog.Infof("error from %s: %v", forwardRequest.URL, err)
		http.Error(w, "internal error", http.StatusBadGateway)
		return
	}

	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		klog.Warningf("error writing response body: %v", err)
		return
	}

	switch resp.StatusCode {
	case 200:
		klog.V(2).Infof("proxied to %s %s: %s", forwardRequest.Method, forwardRequest.URL, resp.Status)
	default:
		klog.Infof("proxied to %s %s: %s", forwardRequest.Method, forwardRequest.URL, resp.Status)
	}
}

func run() error {
	listen := ":8080"

	clientCert := ""
	clientKey := ""
	caCert := ""

	flag.StringVar(&clientCert, "client-cert", clientCert, "path to client certificate")
	flag.StringVar(&clientKey, "client-key", clientKey, "path to client key")
	flag.StringVar(&caCert, "ca-cert", caCert, "path to ca certificate")

	klog.InitFlags(nil)

	flag.Parse()

	tlsConfig := &tls.Config{}

	if caCert != "" {
		b, err := ioutil.ReadFile(caCert)
		if err != nil {
			return fmt.Errorf("error reading certificate %q: %v", caCert, err)
		}
		rootCAs := x509.NewCertPool()
		rootCAs.AppendCertsFromPEM(b)
		tlsConfig.RootCAs = rootCAs
	}

	if clientKey != "" {
		keypair, err := tls.LoadX509KeyPair(clientCert, clientKey)
		if err != nil {
			return fmt.Errorf("error reading client keypair: %v", err)
		}

		tlsConfig.Certificates = []tls.Certificate{keypair}
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	s := &healthCheckServer{
		transport: transport,
	}

	http.HandleFunc("/", s.handler)

	klog.Infof("listening on %s", listen)

	if err := http.ListenAndServe(listen, nil); err != nil {
		return fmt.Errorf("error listening on %q: %v", listen, err)
	}

	return fmt.Errorf("unexpected return from ListenAndServe")
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
