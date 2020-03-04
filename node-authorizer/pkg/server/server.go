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

package server

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/kops/node-authorizer/pkg/utils"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

const (
	// Version is the server version
	Version = "v0.0.4"
	// the namespace to place the secrets
	tokenNamespace = "kube-system"
)

// NodeAuthorizer retains the authorizer state
type NodeAuthorizer struct {
	// authorizer is a collection of authorizers
	authorizer Authorizer
	// client is the kubernetes api client
	client kubernetes.Interface
	// config is the configuration
	config *Config
}

// New creates and returns a node authorizer
func New(config *Config, authorizer Authorizer) (*NodeAuthorizer, error) {
	utils.Logger.Info("starting the node authorization service",
		zap.String("listen", config.Listen),
		zap.String("version", Version))

	if authorizer == nil {
		return nil, errors.New("no authorizer")
	}

	if err := config.IsValid(); err != nil {
		return nil, fmt.Errorf("configuration error: %s", err)
	}

	return &NodeAuthorizer{
		authorizer: authorizer,
		config:     config,
	}, nil
}

// Run is responsible for starting the node authorizer service
func (n *NodeAuthorizer) Run() error {
	// @step: create the kubernetes client
	client, err := utils.GetKubernetesClient()
	if err != nil {
		return err
	}
	n.client = client

	// @step: create the http service
	server := &http.Server{
		Addr: n.config.Listen,
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
		},
	}

	// @step: are we using mutual tls?
	if n.config.TLSClientCAPath != "" {
		// a client certificate is not required, but if given we need to verify it
		server.TLSConfig.ClientAuth = tls.VerifyClientCertIfGiven
		caCert, err := ioutil.ReadFile(n.config.TLSClientCAPath)
		if err != nil {
			return err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		server.TLSConfig.ClientCAs = caCertPool
	}

	// @step: add the routing
	r := mux.NewRouter()
	r.Handle("/authorize/{name}", authorized(n.authorizeHandler, n.config.ClientCommonName, n.useMutualTLS())).Methods(http.MethodPost)
	r.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
	r.HandleFunc("/health", n.healthHandler).Methods(http.MethodGet)
	server.Handler = recovery(r)

	// @step: wait for either an error or a termination signal
	errs := make(chan error, 2)
	go func() {
		errs <- server.ListenAndServeTLS(n.config.TLSCertPath, n.config.TLSPrivateKeyPath)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("received termination signal: %s", <-c)
	}()

	return <-errs
}

// useMutualTLS checks if we are using mutual tls
func (n *NodeAuthorizer) useMutualTLS() bool {
	return n.config.TLSClientCAPath != ""
}
