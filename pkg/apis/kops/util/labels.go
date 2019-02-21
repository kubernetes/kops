/*
Copyright 2017 The Kubernetes Authors.

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

package util

import (
	"strings"

	v1 "k8s.io/api/core/v1"
)

func GetNodeRole(node *v1.Node) string {
	role := ""
	// Newer labels
	for k := range node.Labels {
		if strings.HasPrefix(k, "node-role.kubernetes.io/") {
			role = strings.TrimPrefix(k, "node-role.kubernetes.io/")
		}
	}
	// Older label
	if role == "" {
		role = node.Labels["kubernetes.io/role"]
	}

	return role
}
