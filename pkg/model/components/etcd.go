/*
Copyright 2016 The Kubernetes Authors.

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
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &EtcdOptionsBuilder{}

// BuildOptions is responsible for filling in the defaults for the etcd cluster model
func (b *EtcdOptionsBuilder) BuildOptions(o interface{}) error {
	spec := o.(*kops.ClusterSpec)

	// @check the version are set and if not preset the defaults
	for _, x := range spec.EtcdClusters {
		if x.StorageType == "" {
			x.StorageType = kops.EtcdStorageTypeV2
		}
		if x.Version == "" {
			switch UseEtcdV3(spec) {
			case true:
				x.Version = "3.0.17"
			default:
				x.Version = "2.2.1"
			}
		}
	}

	return nil
}
