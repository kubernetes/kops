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
	"context"
	"fmt"
	"time"

	"k8s.io/kops/node-authorizer/pkg/utils"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// CheckRegistration indicates we should validate the node is not regestered
	CheckRegistration = "verify-registration"
)

// authorizeNodeRequest is responsible for handling the incoming authorization request
func (n *NodeAuthorizer) authorizeNodeRequest(ctx context.Context, request *NodeRegistration) error {
	doneCh := make(chan error)

	// @step: create a context to run under
	ctx, cancel := context.WithTimeout(ctx, n.config.AuthorizationTimeout)
	defer cancel()

	// @step: background the request and wait for either a timeout or a token
	go func() {
		doneCh <- func() error {
			// @step: check if the node request is authorized
			if err := n.safelyAuthorizeNode(ctx, request); err != nil {
				return err
			}
			if request.IsAllowed() {
				return n.safelyProvisionBootstrapToken(ctx, request)
			}

			return nil
		}()
	}()

	// @step: we either wait for the context to timeout or cancel, or we receive a done signal
	select {
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			utils.Logger.Error("operation has either timed out or been cancelled",
				zap.String("client", request.Spec.RemoteAddr),
				zap.String("node", request.Spec.NodeName),
				zap.Error(err))
		}

		return nil

	case err := <-doneCh:
		if err != nil {
			utils.Logger.Error("failed to provision a bootstrap token",
				zap.String("client", request.Spec.RemoteAddr),
				zap.String("node", request.Spec.NodeName),
				zap.Error(err))
		}
	}

	if !request.IsAllowed() {
		utils.Logger.Error("the node has been denied authorization",
			zap.String("client", request.Spec.RemoteAddr),
			zap.String("node", request.Spec.NodeName),
			zap.String("reason", request.Status.Reason))

		nodeAuthorizationMetric.WithLabelValues("denied").Inc()

		return nil
	}

	utils.Logger.Info("node has been authorized access",
		zap.String("client", request.Spec.RemoteAddr),
		zap.String("node", request.Spec.NodeName))

	nodeAuthorizationMetric.WithLabelValues("allowed").Inc()

	return nil
}

// safelyAuthorizeNode checks if the request is permitted
func (n *NodeAuthorizer) safelyAuthorizeNode(ctx context.Context, request *NodeRegistration) error {
	// @step: attempt to authorize the request
	now := time.Now()
	if err := n.authorizer.Authorize(ctx, request); err != nil {
		authorizerErrorMetric.Inc()

		return err
	}
	authorizerLatencyMetric.Observe(time.Since(now).Seconds())

	// @check if the node is registered already
	if n.config.UseFeature(CheckRegistration) {
		if found, err := isNodeRegistered(ctx, n.client, request.Spec.NodeName); err != nil {
			return fmt.Errorf("unable to check node registration status: %s", err)
		} else if found {
			request.Deny(fmt.Sprintf("node %s already registered", request.Spec.NodeName))
		}
	}

	return nil
}

// safelyProvisionBootstrapToken is responsible for generating a bootstrap token for us
func (n *NodeAuthorizer) safelyProvisionBootstrapToken(ctx context.Context, request *NodeRegistration) error {
	maxInterval := 500 * time.Millisecond
	maxTime := 10 * time.Second
	usages := []string{"authentication", "signing"}
	now := time.Now()

	if err := utils.Retry(ctx, maxInterval, maxTime, func() error {
		token, err := n.createToken(n.config.TokenDuration, usages)
		if err != nil {
			return err
		}
		request.Status.Token = token.String()

		return err
	}); err != nil {
		return err
	}

	tokenLatencyMetric.Observe(time.Since(now).Seconds())

	return nil
}

// createToken generates a token for the instance
func (n *NodeAuthorizer) createToken(expiration time.Duration, usages []string) (*Token, error) {
	var err error
	var token *Token

	ctx := context.TODO()

	err = utils.Retry(ctx, 2000*time.Millisecond, 10*time.Second, func() error {
		// @step: generate a random token for them
		if token, err = NewToken(); err != nil {
			return err
		}

		// @step: check if the token already exist, remote but a possibility
		if found, err := n.hasToken(ctx, token); err != nil {
			return err
		} else if found {
			return fmt.Errorf("duplicate token found: %s, skipping", token.ID)
		}

		// @step: add the secret to the namespace
		v1secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: token.Name(),
				Labels: map[string]string{
					"name": token.Name(),
				},
			},
			Type: v1.SecretType(secretTypeBootstrapToken),
			Data: encodeTokenSecretData(token, usages, expiration),
		}

		if _, err := n.client.CoreV1().Secrets(tokenNamespace).Create(ctx, v1secret, metav1.CreateOptions{}); err != nil {
			return err
		}

		return nil
	})

	return token, err
}

// hasToken checks if the tokens already exists
func (n *NodeAuthorizer) hasToken(ctx context.Context, token *Token) (bool, error) {
	resp, err := n.client.CoreV1().Secrets(tokenNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: "name=" + token.Name(),
		Limit:         1,
	})
	if err != nil {
		return false, err
	}

	return len(resp.Items) > 0, nil
}

// encodeTokenSecretData takes the token discovery object and an optional duration and returns the .Data for the Secret
func encodeTokenSecretData(token *Token, usages []string, ttl time.Duration) map[string][]byte {
	data := map[string][]byte{
		bootstrapTokenIDKey:     []byte(token.ID),
		bootstrapTokenSecretKey: []byte(token.Secret),
	}

	if ttl > 0 {
		expire := time.Now().Add(ttl).Format(time.RFC3339)
		data[bootstrapTokenExpirationKey] = []byte(expire)
	}

	for _, usage := range usages {
		data[bootstrapTokenUsagePrefix+usage] = []byte("true")
	}

	return data
}
