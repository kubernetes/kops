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
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// NetworkingOptionsBuilder adds options for the kubelet to the model
type NetworkingOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &NetworkingOptionsBuilder{}

func (b *NetworkingOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)

	k8sVersion, err := KubernetesVersion(clusterSpec)
	if err != nil {
		return err
	}

	options := o.(*kops.ClusterSpec)
	if options.Kubelet == nil {
		options.Kubelet = &kops.KubeletConfigSpec{}
	}

	networking := clusterSpec.Networking
	if networking == nil {
		return fmt.Errorf("networking not set")
	}

	if networking.CNI != nil || networking.Weave != nil || networking.Flannel != nil || networking.Calico != nil || networking.Canal != nil || networking.Kuberouter != nil || networking.Romana != nil || networking.AmazonVPC != nil || networking.Cilium != nil || networking.LyftVPC != nil {
		options.Kubelet.NetworkPluginName = "cni"

		if k8sVersion.Major == 1 && k8sVersion.Minor <= 4 {
			options.Kubelet.ConfigureCBR0 = fi.Bool(false)
		} else {
			// ConfigureCBR0 flag removed from 1.5
			options.Kubelet.ConfigureCBR0 = nil
		}
	}

	if networking.GCE != nil {
		// GCE IPAlias networking uses kubenet on the nodes
		options.Kubelet.NetworkPluginName = "kubenet"
	}

	if networking.Classic != nil {
		// The --configure-cbr0 option was deprecated in v1.4 and removed in v1.5.
		// Only use it when "classic" is allowed (which, by cluster validation,
		// is only allowed for clusters <v1.4).
		if k8sVersion.Major == 1 && k8sVersion.Minor <= 4 {
			options.Kubelet.ConfigureCBR0 = fi.Bool(true)
		} else {
			return fmt.Errorf("classic networking not supported after 1.4")
		}
	}

	if networking.Romana != nil {
		daemonIP, err := WellKnownServiceIP(clusterSpec, 99)
		if err != nil {
			return err
		}
		networking.Romana.DaemonServiceIP = daemonIP.String()
		etcdIP, err := WellKnownServiceIP(clusterSpec, 88)
		if err != nil {
			return err
		}
		networking.Romana.EtcdServiceIP = etcdIP.String()
	}

	return nil
}
