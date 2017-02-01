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
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
	"strings"
)

// KubeletOptionsBuilder adds options for kubelets
type KubeletOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &KubeletOptionsBuilder{}

func (b *KubeletOptionsBuilder) BuildOptions(o interface{}) error {
	options := o.(*kops.ClusterSpec)

	kubernetesVersion, err := b.Context.KubernetesVersion()
	if err != nil {
		return err
	}

	if options.Kubelet == nil {
		options.Kubelet = &kops.KubeletConfigSpec{}
	}
	if options.MasterKubelet == nil {
		options.MasterKubelet = &kops.KubeletConfigSpec{}
	}

	// In 1.5 we fixed this, but in 1.4 we need to set the PodCIDR on the master
	// so that hostNetwork pods can come up
	if kubernetesVersion.Major == 1 && kubernetesVersion.Minor <= 4 {
		// We bootstrap with a fake CIDR, but then this will be replaced (unless we're running with _isolated_master)
		options.MasterKubelet.PodCIDR = "10.123.45.0/28"
	}

	// 1.5 deprecates the reconcile cidr option (and 1.6 removes it)
	if kubernetesVersion.Major == 1 && kubernetesVersion.Minor <= 4 {
		options.MasterKubelet.ReconcileCIDR = fi.Bool(true)

		if fi.BoolValue(b.Context.Cluster.Spec.IsolateMasters) {
			options.MasterKubelet.ReconcileCIDR = fi.Bool(false)
		}

		usesKubenet, err := b.Context.UsesKubenet()
		if err != nil {
			return err
		}
		if usesKubenet {
			options.Kubelet.ReconcileCIDR = fi.Bool(true)
		}
	}

	if kubernetesVersion.Major == 1 && kubernetesVersion.Minor >= 4 {
		// For pod eviction in low memory or empty disk situations
		if options.Kubelet.EvictionHard == nil {
			evictionHard := []string{
				// TODO: Some people recommend 250Mi, but this would hurt small machines
				"memory.available<100Mi",

				// Disk eviction (evict old images)
				// We don't need to specify both, but it seems harmless / safer
				"nodefs.available<10%",
				"nodefs.inodesFree<5%",
				"imagefs.available<10%",
				"imagefs.inodesFree<5%",
			}
			options.Kubelet.EvictionHard = fi.String(strings.Join(evictionHard, ","))
		}
	}

	// IsolateMasters enables the legacy behaviour, where master pods on a separate network
	// In newer versions of kubernetes, most of that functionality has been removed though
	if fi.BoolValue(b.Context.Cluster.Spec.IsolateMasters) {
		options.MasterKubelet.EnableDebuggingHandlers = fi.Bool(false)
		options.MasterKubelet.HairpinMode = "none"
	}

	return nil
}
