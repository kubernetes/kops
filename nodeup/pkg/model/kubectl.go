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

package model

import (
	"fmt"

	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"

	"k8s.io/klog"
)

// KubectlBuilder install kubectl
type KubectlBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KubectlBuilder{}

// Build is responsible for managing the kubectl on the nodes
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

		c.AddTask(&nodetasks.File{
			Path:     b.KubectlPath() + "/" + assetName,
			Contents: asset,
			Type:     nodetasks.FileType_File,
			Mode:     s("0755"),
		})
	}

	{
		kubeconfig, err := b.BuildPKIKubeconfig("kubecfg")
		if err != nil {
			return err
		}

		c.AddTask(&nodetasks.File{
			Path:     "/var/lib/kubectl/kubeconfig",
			Contents: fi.NewStringResource(kubeconfig),
			Type:     nodetasks.FileType_File,
			Mode:     s("0400"),
		})

		adminUser, adminGroup, err := b.findKubeconfigUser()
		if err != nil {
			return err
		}

		if adminUser != nil && adminUser.Home != "" {
			c.AddTask(&nodetasks.File{
				Path:  adminUser.Home + "/.kube/",
				Type:  nodetasks.FileType_Directory,
				Mode:  s("0700"),
				Owner: s(adminUser.Name),
				Group: s(adminGroup.Name),
			})

			c.AddTask(&nodetasks.File{
				Path:     adminUser.Home + "/.kube/config",
				Contents: fi.NewStringResource(kubeconfig),
				Type:     nodetasks.FileType_File,
				Mode:     s("0400"),
				Owner:    s(adminUser.Name),
				Group:    s(adminGroup.Name),
			})
		}
	}

	return nil
}

// findKubeconfigUser finds the default user for whom we should create a kubeconfig
func (b *KubectlBuilder) findKubeconfigUser() (*fi.User, *fi.Group, error) {
	var users []string
	switch b.Distribution {
	case distros.DistributionDebian9, distros.DistributionDebian10:
		users = []string{"admin", "root"}
	case distros.DistributionXenial, distros.DistributionBionic, distros.DistributionFocal:
		users = []string{"ubuntu"}
	case distros.DistributionCentos7, distros.DistributionCentos8:
		users = []string{"centos"}
	case distros.DistributionAmazonLinux2, distros.DistributionRhel7, distros.DistributionRhel8:
		users = []string{"ec2-user"}
	case distros.DistributionFlatcar:
		users = []string{"core"}
	default:
		klog.Warningf("Unknown distro; won't write kubeconfig to homedir %s", b.Distribution)
		return nil, nil, nil
	}

	for _, s := range users {
		user, err := fi.LookupUser(s)
		if err != nil {
			klog.Warningf("error looking up user %q: %v", s, err)
			continue
		}
		if user == nil {
			continue
		}
		group, err := fi.LookupGroupById(user.Gid)
		if err != nil {
			klog.Warningf("unable to find group %d for user %q", user.Gid, s)
			continue
		}
		if group == nil {
			continue
		}
		return user, group, nil
	}

	return nil, nil, nil
}
