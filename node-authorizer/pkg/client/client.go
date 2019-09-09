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
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"k8s.io/kops/node-authorizer/pkg/utils"

	"go.uber.org/zap"
)

// New returns a client verifier
func New(config *Config) error {
	if err := config.IsValid(); err != nil {
		return err
	}

	utils.Logger.Info("attempting to acquire a node bootstrap configuration",
		zap.Duration("timeout", config.Timeout),
		zap.String("authorizer", config.Authorizer),
		zap.String("kubeapi-url", config.KubeAPI),
		zap.String("kubeconfig", config.KubeConfigPath),
		zap.String("registration-url", config.NodeURL))

	// @step: if we have a kubeconfig already we can skip it
	if utils.FileExists(config.KubeConfigPath) {
		utils.Logger.Info("skipping the client authorization as kubeconfig found",
			zap.String("kubeconfig", config.KubeConfigPath))

		return nil
	}

	// @step: create the verifier
	verifier, err := newNodeVerifier(config.Authorizer)
	if err != nil {
		return err
	}

	hc, err := makeHTTPClient(config)
	if err != nil {
		return err
	}

	// @step: attempt to get the token
	err = utils.Retry(context.TODO(), config.Interval, config.Timeout, func() error {
		token, err := makeTokenRequest(context.TODO(), hc, verifier, config)
		if err != nil {
			utils.Logger.Error("failed to request bootstrap token from node authorizer service", zap.Error(err))

			return err
		}

		// @check if we have been refused
		if !token.IsAllowed() {
			utils.Logger.Error("node has been refused registration",
				zap.String("reason", token.Status.Reason))

			os.Exit(1)
		}

		utils.Logger.Info("successfully requested bootstrap token from service")
		utils.Logger.Info("attempting to write bootstrap configuration")

		kubeconfig, err := makeKubeconfig(context.TODO(), config, token.Status.Token)
		if err != nil {
			utils.Logger.Error("failed to generate the bootstrap token configuration",
				zap.String("path", config.KubeConfigPath),
				zap.Error(err))

			return err
		}

		dirname := filepath.Dir(config.KubeConfigPath)
		if err := os.MkdirAll(dirname, os.FileMode(0770)); err != nil {
			return err
		}

		return ioutil.WriteFile(config.KubeConfigPath, kubeconfig, os.FileMode(0640))
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "[error] %s\n", err)
		os.Exit(1)
	}

	utils.Logger.Info("successfully wrote bootstrap configuration")

	return nil
}

// IsValid validates the client configuration
func (c *Config) IsValid() error {
	if c.Authorizer == "" {
		return errors.New("no authorizer")
	}
	if c.KubeAPI == "" {
		return errors.New("no kubeapi url")
	}
	if c.KubeConfigPath == "" {
		return errors.New("no bootstrap kubeconfig path")
	}
	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than zero")
	}
	if _, err := url.Parse(c.KubeAPI); err != nil {
		return fmt.Errorf("invalid kubeapi url: %s", err)
	}

	return nil
}
