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

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// Machine wraps a clusterapi Machine object
type Machine struct {
	u *unstructured.Unstructured
}

func (m *Machine) GetDeploymentName() string {
	return m.u.GetLabels()["cluster.x-k8s.io/deployment-name"]
}

func (m *Machine) GetFailureDomain() string {
	failureDomain, _, _ := unstructured.NestedString(m.u.Object, "spec", "failureDomain")
	return failureDomain
}
