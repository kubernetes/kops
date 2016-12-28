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
	"fmt"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// KubeletOptionsBuilder adds options for kubelets
type KubeletOptionsBuilder struct {
	Cluster *kops.Cluster
}

var _ loader.OptionsBuilder = &KubeletOptionsBuilder{}

func (b *KubeletOptionsBuilder) BuildOptions(o interface{}) error {
	options := o.(*kops.ClusterSpec)

	if options.KubernetesVersion == "" {
		return fmt.Errorf("KubernetesVersion is required")
	}

	sv, err := kops.ParseKubernetesVersion(options.KubernetesVersion)
	if err != nil {
		return fmt.Errorf("unable to determine kubernetes version from %q", options.KubernetesVersion)
	}

	if options.Kubelet == nil {
		options.Kubelet = &kops.KubeletConfigSpec{}
	}
	if options.MasterKubelet == nil {
		options.MasterKubelet = &kops.KubeletConfigSpec{}
	}

	// In 1.5 we fixed this, but in 1.4 we need to set the PodCIDR on the master
	// so that hostNetwork pods can come up
	if sv.Major == 1 && sv.Minor <= 4 {
		// We bootstrap with a fake CIDR, but then this will be replaced (unless we're running with _isolated_master)
		options.MasterKubelet.PodCIDR = "10.123.45.0/28"
	}

	return nil
}
