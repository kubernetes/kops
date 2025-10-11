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

package clusterapi

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Manager struct {
	kubeClient client.Client
}

func NewManager(kubeClient client.Client) *Manager {
	return &Manager{
		kubeClient: kubeClient,
	}
}

func (m *Manager) FindMachineByProviderID(ctx context.Context, providerID string) (*Machine, error) {
	// TODO: Can we build an index
	// selector := client.MatchingFieldsSelector{
	// 	Selector: fields.OneTermEqualSelector("spec.providerID", providerID),
	// }
	var machines unstructured.UnstructuredList
	machines.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "cluster.x-k8s.io",
		Kind:    "Machine",
		Version: "v1beta1",
	})
	if err := m.kubeClient.List(ctx, &machines); err != nil {
		return nil, fmt.Errorf("error listing machines: %w", err)
	}
	var matches []*unstructured.Unstructured
	for i := range machines.Items {
		machine := &machines.Items[i]
		machineSpecProviderID, _, _ := unstructured.NestedString(machine.Object, "spec", "providerID")
		if machineSpecProviderID != providerID {
			continue
		}
		matches = append(matches, machine)
	}
	if len(matches) > 0 {
		if len(matches) > 1 {
			return nil, fmt.Errorf("found multiple machines with providerID %q", providerID)
		}
		machine := matches[0]
		machine = machine.DeepCopy()
		return &Machine{u: machine}, nil
	}

	return nil, nil
}
