/*
Copyright 2021 The Kubernetes Authors.

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

package fluentest

import (
	"context"
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Nodes presents a fluent interface for querying nodes, convenient for tests
type Nodes struct {
	ctx        context.Context
	client     kubernetes.Interface
	restConfig *rest.Config
}

// MustItems queries and returns matching nodes.
func (n *Nodes) MustItems(t *testing.T) []*Node {
	items, err := n.Items()
	if err != nil {
		t.Fatalf("failed to fetch nodes: %v", err)
	}
	return items
}

// Items queries and returns matching nodes.
func (n *Nodes) Items() ([]*Node, error) {
	var opt metav1.ListOptions
	nodes, err := n.client.CoreV1().Nodes().List(n.ctx, opt)
	if err != nil {
		return nil, fmt.Errorf("CoreV1().Nodes().List failed: %w", err)
	}

	var ret []*Node
	for i := range nodes.Items {
		node := &nodes.Items[i]
		ret = append(ret, n.wrapNode(node))
	}

	return ret, nil
}

func (n *Nodes) wrapNode(node *v1.Node) *Node {
	node.Kind = "Node"
	node.APIVersion = "v1"

	return &Node{
		fluentBase: fluentBase{
			fluentOptions: fluentOptions{
				client:     n.client,
				restConfig: n.restConfig,
				ctx:        n.ctx,
			},
			obj:  node,
			meta: node,
		},
		obj: node,
	}
}

// newNodes returns the fluent interface for querying nodes.
func newNodes(ctx context.Context, restConfig *rest.Config, client kubernetes.Interface) *Nodes {
	return &Nodes{ctx: ctx, client: client, restConfig: restConfig}
}
