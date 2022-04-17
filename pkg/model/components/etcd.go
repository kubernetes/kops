/*
Copyright 2019 The Kubernetes Authors.

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

package components

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// EtcdOptionsBuilder adds options for etcd to the model
type EtcdOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &EtcdOptionsBuilder{}

const (
	DefaultEtcd3Version_1_19 = "3.4.13"
	DefaultEtcd3Version_1_22 = "3.5.3"
)

// BuildOptions is responsible for filling in the defaults for the etcd cluster model
func (b *EtcdOptionsBuilder) BuildOptions(o interface{}) error {
	spec := o.(*kops.ClusterSpec)

	for i := range spec.EtcdClusters {
		c := &spec.EtcdClusters[i]
		// Ensure the version is set
		if c.Version == "" {
			// We run the k8s-recommended versions of etcd
			if b.IsKubernetesGTE("1.22") {
				c.Version = DefaultEtcd3Version_1_22
			} else {
				c.Version = DefaultEtcd3Version_1_19
			}
		}
	}

	return nil
}
