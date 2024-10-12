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

package bare_metal

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func TestNodeAddresses(t *testing.T) {
	h := NewHarness(context.Background(), t)

	nodes := h.Nodes()

	// Quick check that we have some nodes
	if len(nodes) == 0 {
		t.Errorf("expected some nodes, got 0 nodes")
	}

	// Verify that node.status.addresses is populated
	for _, node := range nodes {
		if len(node.Status.Addresses) != 1 {
			t.Errorf("expected 1 status.addresses for node %s, got %d: [%v]", node.Name, len(node.Status.Addresses), node.Status.Addresses)
		}
	}
}

// Harness is a test harness for our bare-metal e2e tests
type Harness struct {
	*testing.T
	Ctx        context.Context
	RESTConfig *rest.Config
	Kube       *kubernetes.Clientset
}

// NewHarness creates a new harness
func NewHarness(ctx context.Context, t *testing.T) *Harness {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("error getting user home dir: %v", err)
		}
		kubeconfig = filepath.Join(homeDir, ".kube", "config")
	}
	// use the current context in kubeconfig
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatalf("error building rest config: %v", err)
	}

	httpClient, err := rest.HTTPClientFor(restConfig)
	if err != nil {
		t.Fatalf("error building http client: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfigAndClient(restConfig, httpClient)
	if err != nil {
		t.Fatalf("error building kube client: %v", err)
	}

	return &Harness{
		T:          t,
		Ctx:        ctx,
		RESTConfig: restConfig,
		Kube:       kubeClient,
	}
}

// Nodes returns all nodes in the cluster
func (h *Harness) Nodes() []corev1.Node {
	nodes, err := h.Kube.CoreV1().Nodes().List(h.Ctx, metav1.ListOptions{})
	if err != nil {
		h.Fatalf("error listing nodes: %v", err)
	}
	return nodes.Items
}
