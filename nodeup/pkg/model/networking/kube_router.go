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

package networking

import (
	"k8s.io/kops/nodeup/pkg/model"
	"k8s.io/kops/pkg/rbac"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/distributions"
)

// KuberouterBuilder installs kube-router
type KuberouterBuilder struct {
	*model.NodeupModelContext
}

var _ fi.NodeupModelBuilder = &KuberouterBuilder{}

// Build is responsible for configuring the kube-router
func (b *KuberouterBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if b.NodeupConfig.Networking.KubeRouter == nil {
		return nil
	}

	var kubeconfig fi.Resource
	var err error

	if b.HasAPIServer {
		kubeconfig = b.BuildIssuedKubeconfig("kube-router", nodetasks.PKIXName{CommonName: rbac.KubeRouter}, c)
	} else {
		kubeconfig, err = b.BuildBootstrapKubeconfig("kube-router", c)
		if err != nil {
			return err
		}
	}

	c.AddTask(&nodetasks.File{
		Path:           "/var/lib/kube-router/kubeconfig",
		Contents:       kubeconfig,
		Type:           nodetasks.FileType_File,
		Mode:           fi.PtrTo("0400"),
		BeforeServices: []string{"kubelet.service"},
	})

	// On older Debian/Ubuntu versions, iproute2 config lives in /etc/iproute2/ rather than /usr/share/iproute2/.
	// Create a symlink so the kube-router DaemonSet can mount /usr/share/iproute2/rt_tables.
	// Ref: https://github.com/kubernetes/kops/issues/17914
	switch b.Distribution {
	case distributions.DistributionDebian11,
		distributions.DistributionDebian12,
		distributions.DistributionUbuntu2204,
		distributions.DistributionUbuntu2404:
		c.AddTask(&nodetasks.File{
			Path:    "/usr/share/iproute2",
			Type:    nodetasks.FileType_Symlink,
			Symlink: fi.PtrTo("/etc/iproute2"),
			Owner:   fi.PtrTo("root"),
			Group:   fi.PtrTo("root"),
			Mode:    fi.PtrTo("0755"),
		})
	}

	return nil
}
