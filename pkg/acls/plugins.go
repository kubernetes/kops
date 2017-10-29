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

package acls

import (
	"fmt"
	"sync"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/vfs"
)

var strategies map[string]ACLStrategy
var strategiesMutex sync.Mutex

// GetACL returns the ACL for the vfs.Path, by consulting all registered strategies
func GetACL(p vfs.Path, cluster *kops.Cluster) (vfs.ACL, error) {
	strategiesMutex.Lock()
	defer strategiesMutex.Unlock()

	for k, strategy := range strategies {
		acl, err := strategy.GetACL(p, cluster)
		if err != nil {
			return nil, fmt.Errorf("error from acl provider %q: %v", k, err)
		}
		if acl != nil {
			return acl, nil
		}
	}
	return nil, nil
}

// RegisterPlugin adds the strategy to the registered strategies
func RegisterPlugin(key string, strategy ACLStrategy) {
	strategiesMutex.Lock()
	defer strategiesMutex.Unlock()

	if strategies == nil {
		strategies = make(map[string]ACLStrategy)
	}

	strategies[key] = strategy
}
