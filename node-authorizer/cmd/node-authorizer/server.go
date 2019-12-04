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

package main

import (
	"errors"
	"fmt"
	"time"

	"k8s.io/kops/node-authorizer/pkg/authorizers/alwaysallow"
	"k8s.io/kops/node-authorizer/pkg/authorizers/aws"
	"k8s.io/kops/node-authorizer/pkg/server"
	"k8s.io/kops/node-authorizer/pkg/utils"

	"github.com/urfave/cli"
)

// addServerCommand creates and returns a server command
func addServerCommand() cli.Command {
	return cli.Command{
		Name:  "server",
		Usage: "starts the node-authorizer in server mode",

		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "authorizer",
				Usage:  "provider we should use to authorize the node registration `NAME`",
				EnvVar: "AUTHORIZER",
			},
			cli.StringFlag{
				Name:   "listen",
				Usage:  "interface to bind the service `INTERFACE`",
				EnvVar: "LISTEN",
				Value:  ":10443",
			},
			cli.StringFlag{
				Name:   "tls-client-ca",
				Usage:  "file containing the client certificate authority, required for mutual tls `PATH`",
				EnvVar: "TLS_CLIENT_CA",
			},
			cli.StringFlag{
				Name:   "tls-cert",
				Usage:  "file containing the certificate `PATH`",
				EnvVar: "TLS_CERT",
				Value:  "/certs/node-authorizer.pem",
			},
			cli.StringFlag{
				Name:   "tls-private-key",
				Usage:  "file containing the private key `PATH`",
				EnvVar: "TLS_PRIVATE_KEY",
				Value:  "/certs/node-authorizer-key.pem",
			},
			cli.StringFlag{
				Name:   "cluster-name",
				Usage:  "name of the kubernetes cluster we are provisioning `NAME`",
				EnvVar: "CLUSTER_NAME",
			},
			cli.StringFlag{
				Name:   "cluster-tag",
				Usage:  "name of the cloud tag used to identify the cluster name `NAME`",
				EnvVar: "CLUSTER_TAG",
				Value:  "KubernetesCluster",
			},
			cli.StringSliceFlag{
				Name:  "feature",
				Usage: "enables or disables a feature in the chosen authorizer `NAME`",
			},
			cli.DurationFlag{
				Name:   "token-ttl",
				Usage:  "expiration on created bootstrap token `DURATION`",
				EnvVar: "TOKEN_TTL",
				Value:  5 * time.Minute,
			},
			cli.StringFlag{
				Name:   "client-common-name",
				Usage:  "the common name of the client certificate when use mutual tls `NAME`",
				EnvVar: "CLIENT_COMMON_NAME",
				Value:  "node-authorizer-client",
			},
			cli.DurationFlag{
				Name:   "certificate-ttl",
				Usage:  "check the certificates exist and if not wait for x period `DURATION`",
				EnvVar: "CERTIFICATE_TTL",
				Value:  1 * time.Hour,
			},
			cli.DurationFlag{
				Name:   "authorization-timeout",
				Usage:  "max time permitted for a authorization `DURATION`",
				EnvVar: "AUTHORIZATION_TIMEOUT",
				Value:  15 * time.Second,
			},
		},

		Action: func(ctx *cli.Context) error {
			return actionServerCommand(ctx)
		},
	}
}

// actionServerCommand is responsible for performing the server action
func actionServerCommand(ctx *cli.Context) error {
	config := &server.Config{
		AuthorizationTimeout: ctx.Duration("authorization-timeout"),
		ClientCommonName:     ctx.String("client-common-name"),
		ClusterName:          ctx.String("cluster-name"),
		ClusterTag:           ctx.String("cluster-tag"),
		Features:             ctx.StringSlice("feature"),
		Listen:               ctx.String("listen"),
		TLSCertPath:          ctx.String("tls-cert"),
		TLSClientCAPath:      ctx.String("tls-client-ca"),
		TLSPrivateKeyPath:    ctx.String("tls-private-key"),
		TokenDuration:        ctx.Duration("token-ttl"),
	}

	if ctx.String("authorizer") == "" {
		return errors.New("no authorizer specified")
	}

	// @step: should we wait for the certificates to appear
	if ctx.Duration("certificate-ttl") > 0 {
		var files = []string{ctx.String("tls-cert"), ctx.String("tls-client-ca"), ctx.String("tls-private-key")}
		var timeout = ctx.Duration("certificate-ttl")
		if err := waitForCertificates(files, timeout); err != nil {
			return err
		}
	}

	// @step: create the authorizers
	auth, err := createAuthorizer(ctx.String("authorizer"), config)
	if err != nil {
		return fmt.Errorf("failed to create authorizer: %v", err)
	}

	svc, err := server.New(config, auth)
	if err != nil {
		return err
	}

	return svc.Run()
}

// waitForCertificates is responsible for waiting for the certificates to appear
func waitForCertificates(files []string, timeout time.Duration) error {
	doneCh := make(chan struct{})

	go func() {
		expires := time.Now().Add(timeout)

		// @step: iterate the file we are looking for
		for _, x := range files {
			if x == "" {
				continue
			}
			// @step: iterate until we find the file
			for {
				if utils.FileExists(x) {
					break
				}
				fmt.Printf("waiting for file: %s to appear, timeouts in %s\n", x, time.Until(expires))
				time.Sleep(5 * time.Second)
			}
		}
		doneCh <- struct{}{}
	}()

	select {
	case <-doneCh:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("unable to find the certificates after %s timeout", timeout)
	}
}

// createAuthorizer creates and returns a authorizer
func createAuthorizer(name string, config *server.Config) (server.Authorizer, error) {
	switch name {
	case "alwaysallow":
		return alwaysallow.NewAuthorizer()
	case "aws":
		return aws.NewAuthorizer(config)
	}

	return nil, errors.New("unknown authorizer")
}
