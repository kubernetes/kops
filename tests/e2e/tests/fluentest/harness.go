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
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Harness struct {
	*testing.T

	clusterName      string
	restConfig       *rest.Config
	kubernetesClient kubernetes.Interface
}

func (h *Harness) ClusterName() string {
	t := h.T

	if h.clusterName == "" {
		clusterName, err := KubectlCurrentContext()
		if err != nil {
			t.Fatalf("KubectlCurrentContext() => %v", err)
		}
		h.clusterName = clusterName
	}
	return h.clusterName
}

func (h *Harness) InstanceGroups() []*KopsInstanceGroup {
	t := h.T

	clusterName := h.ClusterName()

	igs, err := KopsGetInstanceGroups(clusterName)
	if err != nil {
		t.Fatalf("cluster.GetInstanceGroups() failed: %v", err)
	}

	return igs
}

func (h *Harness) RESTConfig() *rest.Config {
	t := h.T

	if h.restConfig == nil {
		c, err := RESTConfigFromKubeconfig()
		if err != nil {
			t.Fatalf("RESTConfigFromKubeconfig() failed: %v", err)
		}
		h.restConfig = c
	}

	return h.restConfig
}

func (h *Harness) KubernetesClient() kubernetes.Interface {
	t := h.T

	if h.kubernetesClient == nil {
		restConfig := h.RESTConfig()

		k, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			t.Fatalf("kubernetes.NewForConfig failed: %v", err)
		}

		h.kubernetesClient = k
	}

	return h.kubernetesClient
}

func (h *Harness) Nodes() *Nodes {
	restConfig := h.RESTConfig()
	k := h.KubernetesClient()

	ctx := context.Background()
	return newNodes(ctx, restConfig, k)
}
