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
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/upup/pkg/fi"
)

type NodeupModelContext struct {
	NodeupConfig *nodeup.NodeUpConfig

	Cluster       *kops.Cluster
	InstanceGroup *kops.InstanceGroup
	Architecture  Architecture
	Distribution  distros.Distribution

	IsMaster bool
	UsesCNI  bool

	Assets      *fi.AssetStore
	KeyStore    fi.CAStore
	SecretStore fi.SecretStore
}

func (c *NodeupModelContext) SSLHostPaths() []string {
	paths := []string{"/etc/ssl", "/etc/pki/tls", "/etc/pki/ca-trust"}

	switch c.Distribution {
	case distros.DistributionCoreOS:
		// Because /usr is read-only on CoreOS, we can't have any new directories; docker will try (and fail) to create them
		// TODO: Just check if the directories exist?

		paths = append(paths, "/usr/share/ca-certificates")

	default:
		paths = append(paths, "/usr/share/ssl", "/usr/ssl", "/usr/lib/ssl", "/usr/local/openssl", "/var/ssl", "/etc/openssl")
	}

	return paths
}
