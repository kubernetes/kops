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

// DefaultsOptionsBuilder adds default options.  This should come first!
type DefaultsOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &DefaultsOptionsBuilder{}

// BuildOptions is responsible for cluster options
func (b *DefaultsOptionsBuilder) BuildOptions(o interface{}) error {
	options := o.(*kops.ClusterSpec)

	if options.ClusterDNSDomain == "" {
		options.ClusterDNSDomain = "cluster.local"
	}

	if options.ContainerRuntime == "" {
		options.ContainerRuntime = "docker"
	}

	return nil
}
