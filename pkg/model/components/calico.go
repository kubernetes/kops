/*
Copyright 2020 The Kubernetes Authors.

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
	"context"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// CalicoOptionsBuilder prepares settings related to the Calico CNI implementation.
type CalicoOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &CalicoOptionsBuilder{}

func (b *CalicoOptionsBuilder) BuildOptions(ctx context.Context, o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)
	c := clusterSpec.Networking.Calico
	if c == nil {
		return nil
	}

	c.EncapsulationMode = "ipip"
	if clusterSpec.IsIPv6Only() {
		c.EncapsulationMode = "none"
	}

	return nil
}
