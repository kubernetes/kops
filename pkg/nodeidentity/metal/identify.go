/*
Copyright 2024 The Kubernetes Authors.

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

package metal

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/kops/pkg/nodeidentity"
)

// nodeIdentifier identifies a node in a bare-metal cloud
type nodeIdentifier struct {
}

// New creates and returns a nodeidentity.Identifier for Nodes running on bare metal
func New() (nodeidentity.Identifier, error) {
	return &nodeIdentifier{}, nil
}

// IdentifyNode returns the node identity information
func (i *nodeIdentifier) IdentifyNode(ctx context.Context, node *corev1.Node) (*nodeidentity.Info, error) {
	return nil, fmt.Errorf("metal::IdentifyNode not implemented")
}
