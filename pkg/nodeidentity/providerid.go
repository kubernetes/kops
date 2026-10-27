/*
Copyright The Kubernetes Authors.

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

package nodeidentity

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// NewProviderIDDispatcher selects a configured identifier by provider ID scheme.
// The selected identifier remains responsible for verifying the node.
func NewProviderIDDispatcher(identifiers map[string]Identifier) Identifier {
	return &providerIDDispatcher{identifiers: identifiers}
}

type providerIDDispatcher struct {
	identifiers map[string]Identifier
}

func (d *providerIDDispatcher) IdentifyNode(ctx context.Context, node *corev1.Node) (*Info, error) {
	providerID := node.Spec.ProviderID
	if providerID == "" {
		return nil, fmt.Errorf("providerID was not set for node %s", node.Name)
	}

	scheme, _, found := strings.Cut(providerID, "://")
	if !found || scheme == "" {
		return nil, fmt.Errorf("providerID %q not recognized for node %s", providerID, node.Name)
	}

	identifier := d.identifiers[scheme]
	if identifier == nil {
		return nil, fmt.Errorf("no node identifier configured for provider %q on node %s", scheme, node.Name)
	}
	return identifier.IdentifyNode(ctx, node)
}
