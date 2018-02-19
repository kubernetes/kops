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

package etcd

import (
	"fmt"
	"strings"
)

// EtcdClusterSpec is configuration for the etcd cluster
type EtcdClusterSpec struct {
	// ClusterKey is a key that identifies the etcd cluster (main or events)
	ClusterKey string `json:"clusterKey,omitempty"`
	// NodeName is my nodename in the cluster
	NodeName string `json:"nodeName,omitempty"`
	// NodeNames is a collection of node members in the cluster
	NodeNames []string `json:"nodeNames,omitempty"`
}

// String returns a string representation of the EtcdClusterSpec
func (e *EtcdClusterSpec) String() string {
	return DebugString(e)
}

// ParseEtcdClusterSpec parses a tag on a volume that encodes an etcd cluster role
// The format is "<myname>/<allnames>", e.g. "node1/node1,node2,node3"
func ParseEtcdClusterSpec(clusterKey, v string) (*EtcdClusterSpec, error) {
	v = strings.TrimSpace(v)

	tokens := strings.Split(v, "/")
	if len(tokens) != 2 {
		return nil, fmt.Errorf("invalid EtcdClusterSpec (expected two tokens): %q", v)
	}

	nodeName := tokens[0]
	nodeNames := strings.Split(tokens[1], ",")

	found := false
	for _, s := range nodeNames {
		if s == nodeName {
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("invalid EtcdClusterSpec (member not found in all nodes): %q", v)
	}

	c := &EtcdClusterSpec{
		ClusterKey: clusterKey,
		NodeName:   nodeName,
		NodeNames:  nodeNames,
	}
	return c, nil
}
