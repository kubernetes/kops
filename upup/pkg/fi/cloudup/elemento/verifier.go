/*
Copyright 2025 The Kubernetes Authors.

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

package elemento

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"k8s.io/kops/pkg/bootstrap"
)

type ElementoVerifierOptions struct {
}

type elementoVerifier struct {
	opt ElementoVerifierOptions
}

var _ bootstrap.Verifier = &elementoVerifier{}

type staticBootstrapNode struct {
	NodeName          string
	InstanceGroupName string
}

var staticBootstrapNodesByIP = map[string]staticBootstrapNode{
	"192.168.100.10": {NodeName: "control-plane-europe-1", InstanceGroupName: "control-plane-europe"},
	"192.168.100.11": {NodeName: "nodes-europe-1", InstanceGroupName: "nodes-europe"},
	"192.168.100.12": {NodeName: "nodes-europe-2", InstanceGroupName: "nodes-europe"},
}

func NewElementoVerifier(opt *ElementoVerifierOptions) (bootstrap.Verifier, error) {
	return &elementoVerifier{
		opt: *opt,
	}, nil
}

func (e elementoVerifier) VerifyToken(ctx context.Context, rawRequest *http.Request, token string, body []byte) (*bootstrap.VerifyResult, error) {
	if !strings.HasPrefix(token, ElementoAuthenticationTokenPrefix) {
		return nil, bootstrap.ErrNotThisVerifier
	}
	token = strings.TrimSpace(strings.TrimPrefix(token, ElementoAuthenticationTokenPrefix))

	remoteHost, _, err := net.SplitHostPort(rawRequest.RemoteAddr)
	if err != nil {
		remoteHost = rawRequest.RemoteAddr
	}

	nodeName := token
	instanceGroupName := ""
	if node, ok := staticBootstrapNodesByIP[remoteHost]; ok {
		nodeName = node.NodeName
		instanceGroupName = node.InstanceGroupName
	}
	if instanceGroupName == "" {
		instanceGroupName = inferInstanceGroupName(nodeName)
	}
	if instanceGroupName == "" {
		return nil, fmt.Errorf("failed to determine instance group for node %q", nodeName)
	}
	certificateNames := []string{nodeName}
	if remoteHost != "" {
		certificateNames = append(certificateNames, remoteHost)
	}

	return &bootstrap.VerifyResult{
		NodeName:          nodeName,
		CertificateNames:  certificateNames,
		InstanceGroupName: instanceGroupName,
	}, nil
}

func inferInstanceGroupName(serverName string) string {
	i := strings.LastIndex(serverName, "-")
	if i == -1 || i == len(serverName)-1 {
		return serverName
	}
	for _, r := range serverName[i+1:] {
		if r < '0' || r > '9' {
			return serverName
		}
	}
	return serverName[:i]
}
