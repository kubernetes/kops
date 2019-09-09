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
	"time"

	"k8s.io/kops/node-authorizer/pkg/client"

	"github.com/urfave/cli"
)

// addClientCommand creates and returns a client command
func addClientCommand() cli.Command {
	return cli.Command{
		Name:  "client",
		Usage: "starts the service in a client mode and attempts to acquire a bootstrap token",

		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "authorizer",
				Usage:  "provider we should use to authorize the node registration `NAME`",
				EnvVar: "AUTHORIZER",
				Value:  "aws",
			},
			cli.StringFlag{
				Name:   "node-url",
				Usage:  "the url for the node authorizer service `URL`",
				EnvVar: "NODE_AUTHORIZER_URL",
			},
			cli.StringFlag{
				Name:   "kubeapi-url",
				Usage:  "the url for the kubernetes api `URL`",
				EnvVar: "KUBEAPI_URL",
			},
			cli.StringFlag{
				Name:   "kubeconfig",
				Usage:  "location to write bootstrap token config `PATH`",
				EnvVar: "KUBECONFIG_BOOTSTRAP",
				Value:  "/var/lib/kubelet/bootstrap-kubeconfig",
			},
			cli.StringFlag{
				Name:   "tls-client-ca",
				Usage:  "file containing the certificate authority used to verify node endpoint `PATH`",
				EnvVar: "TLS_CLIENT_CA",
			},
			cli.StringFlag{
				Name:   "tls-cert",
				Usage:  "file containing the client certificate `PATH`",
				EnvVar: "TLS_CERT",
			},
			cli.StringFlag{
				Name:   "tls-private-key",
				Usage:  "file containing the client private key `PATH`",
				EnvVar: "TLS_PRIVATE_KEY",
			},
			cli.DurationFlag{
				Name:   "interval",
				Usage:  "an interval to wait between failed request `DURATION`",
				EnvVar: "INTERVAL",
				Value:  3 * time.Second,
			},
			cli.DurationFlag{
				Name:   "timeout",
				Usage:  "the max time to wait before timing out `DURATION`",
				EnvVar: "TIMEOUT",
				Value:  30 * time.Second,
			},
		},

		Action: func(ctx *cli.Context) error {
			return actionClientCommand(ctx)
		},
	}
}

// actionClientCommand is the client action
func actionClientCommand(ctx *cli.Context) error {
	return client.New(&client.Config{
		Authorizer:        ctx.String("authorizer"),
		Interval:          ctx.Duration("interval"),
		KubeAPI:           ctx.String("kubeapi-url"),
		KubeConfigPath:    ctx.String("kubeconfig"),
		NodeURL:           ctx.String("node-url"),
		TLSCertPath:       ctx.String("tls-cert"),
		TLSClientCAPath:   ctx.String("tls-client-ca"),
		TLSPrivateKeyPath: ctx.String("tls-private-key"),
		Timeout:           ctx.Duration("timeout"),
	})
}
