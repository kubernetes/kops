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

package nodebootstrap

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/peer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	rest "k8s.io/client-go/rest"
	"k8s.io/klog"
	"k8s.io/kops/node-authorizer/pkg/server"
	pb "k8s.io/kops/pkg/proto/nodebootstrap"
)

const (
	// the namespace to place the secrets
	tokenNamespace = "kube-system"
)

type nodeBootstrapService struct {
	authorizer server.Authorizer

	corev1Client corev1client.CoreV1Interface

	options Options
}

// Options is the configuration for the NodeBootstrap server
type Options struct {
	TokenTTL time.Duration `json:"tokenTTL,omitempty"`
}

// PopulateDefaults sets the default configuration values
func (o *Options) PopulateDefaults() {
	o.TokenTTL = 5 * time.Minute
}

func NewNodeBootstrapService(restConfig *rest.Config, authorizer server.Authorizer, options *Options) (*nodeBootstrapService, error) {
	s := &nodeBootstrapService{}

	s.authorizer = authorizer

	c, err := corev1client.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error connecting to kubernetes: %v", err)
	}
	s.corev1Client = c

	s.options = *options

	return s, nil
}

var _ pb.NodeBootstrapServiceServer = &nodeBootstrapService{}

func (s *nodeBootstrapService) CreateKubeletBootstrapToken(ctx context.Context, request *pb.CreateKubeletBootstrapTokenRequest) (*pb.CreateKubeletBootstrapTokenResponse, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to get peer context")
	}

	nodeRegistration := &server.NodeRegistration{}
	nodeRegistration.Spec.RemoteAddr = peer.Addr.String()

	if err := s.authorizer.Authorize(ctx, nodeRegistration); err != nil {
		// In general we prefer to log error details here, and only return minimal information to the client, because it may be an attacker
		klog.Warningf("internal error during authorization: %v", err)
		return nil, fmt.Errorf("internal error during authorization")
	}

	if !nodeRegistration.Status.Allowed {
		// TODO: Use grpc Status errors?
		return nil, fmt.Errorf("node registration not allowed")
	}

	token, err := s.createBootstrapToken(ctx, nodeRegistration)
	if err != nil {
		klog.Warningf("error creating bootstrap token: %v", err)
		return nil, fmt.Errorf("error creating bootstrap token")
	}

	response := &pb.CreateKubeletBootstrapTokenResponse{}
	if token != nil {
		response.Token = &pb.Token{
			BearerToken: token.ID + "." + token.Secret,
		}
	}

	return response, nil
}

// createBootstrapToken generates a bootstrap token for the node, inserting it into k8s
func (s *nodeBootstrapService) createBootstrapToken(ctx context.Context, request *server.NodeRegistration) (*server.Token, error) {
	usages := []string{"authentication", "signing"}

	// @step: generate a random token for them
	token, err := server.NewToken()
	if err != nil {
		return nil, err
	}

	// @step: add the secret to the namespace
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: token.Name(),
		},
		Type: corev1.SecretType(corev1.SecretTypeBootstrapToken),
		Data: token.EncodeTokenSecretData(usages, s.options.TokenTTL),
	}

	if _, err := s.corev1Client.Secrets(tokenNamespace).Create(secret); err != nil {
		// A token collision is very unlikely, so we'll return an error and let the caller retry
		klog.Warningf("failed to create secret: %v", err)
		return nil, fmt.Errorf("failed to create secret")
	}

	return token, nil
}
