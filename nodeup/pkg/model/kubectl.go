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

package model

import (
	"fmt"

	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"

	"github.com/golang/glog"
)

// KubectlBuilder install kubectl
type KubectlBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KubectlBuilder{}

// Build is responsible for mananging the kubectl on the nodes
func (b *KubectlBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.IsMaster {
		return nil
	}

	{
		// TODO: Extract to common function?
		assetName := "kubectl"
		assetPath := ""
		asset, err := b.Assets.Find(assetName, assetPath)
		if err != nil {
			return fmt.Errorf("error trying to locate asset %q: %v", assetName, err)
		}
		if asset == nil {
			return fmt.Errorf("unable to locate asset %q", assetName)
		}

		t := &nodetasks.File{
			Path:     b.kubectlPath(),
			Contents: asset,
			Type:     nodetasks.FileType_File,
			Mode:     s("0755"),
		}
		c.AddTask(t)
	}

	{
		kubeconfig, err := b.buildPKIKubeconfig("kubecfg")
		if err != nil {
			return err
		}

		t := &nodetasks.File{
			Path:     "/var/lib/kubectl/kubeconfig",
			Contents: fi.NewStringResource(kubeconfig),
			Type:     nodetasks.FileType_File,
			Mode:     s("0400"),
		}
		c.AddTask(t)

		switch b.Distribution {
		case distros.DistributionJessie:
			c.AddTask(&nodetasks.File{
				Path:  "/home/admin/.kube/",
				Type:  nodetasks.FileType_Directory,
				Mode:  s("0700"),
				Owner: s("admin"),
				Group: s("admin"),
			})

			c.AddTask(&nodetasks.File{
				Path:     "/home/admin/.kube/config",
				Contents: fi.NewStringResource(kubeconfig),
				Type:     nodetasks.FileType_File,
				Mode:     s("0400"),
				Owner:    s("admin"),
				Group:    s("admin"),
			})

		default:
			glog.Warningf("Unknown distro; won't write kubeconfig to homedir %s", b.Distribution)
		}
	}

	return nil
}

func (b *KubectlBuilder) kubectlPath() string {
	kubeletCommand := "/usr/local/bin/kubectl"
	if b.Distribution == distros.DistributionCoreOS {
		kubeletCommand = "/opt/bin/kubectl"
	}
	if b.Distribution == distros.DistributionContainerOS {
		kubeletCommand = "/home/kubernetes/bin/kubectl"
	}
	return kubeletCommand
}
