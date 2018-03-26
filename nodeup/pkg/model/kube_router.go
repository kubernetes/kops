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

package model

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// KubeProxyBuilder installs kube-proxy
type KubeRouterBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KubeRouterBuilder{}

func (b *KubeRouterBuilder) Build(c *fi.ModelBuilderContext) error {

	// Add kubeconfig
	{
		kubeconfig, err := b.buildPKIKubeconfig("kube-router")
		if err != nil {
			return err
		}
		t := &nodetasks.File{
			Path:     "/var/lib/kube-router/kubeconfig",
			Contents: fi.NewStringResource(kubeconfig),
			Type:     nodetasks.FileType_File,
			Mode:     s("0400"),
		}
		c.AddTask(t)
	}

	return nil
}
