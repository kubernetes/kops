/*
Copyright 2022 The Kubernetes Authors.

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

package hetznermodel

import (
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetznertasks"
)

// ServerGroupModelBuilder configures server objects
type ServerGroupModelBuilder struct {
	*HetznerModelContext
	Lifecycle              fi.Lifecycle
	BootstrapScriptBuilder *model.BootstrapScriptBuilder
}

var _ fi.ModelBuilder = &ServerGroupModelBuilder{}

func (b *ServerGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	var sshkeyTasks []*hetznertasks.SSHKey
	for _, sshkey := range b.SSHPublicKeys {
		fingerprint, err := pki.ComputeOpenSSHKeyFingerprint(string(sshkey))
		if err != nil {
			return err
		}
		t := &hetznertasks.SSHKey{
			Name:      fi.String(b.ClusterName() + "-" + fingerprint),
			Lifecycle: b.Lifecycle,
			PublicKey: string(sshkey),
			Labels: map[string]string{
				hetzner.TagKubernetesClusterName: b.ClusterName(),
			},
		}
		c.AddTask(t)
		sshkeyTasks = append(sshkeyTasks, t)
	}

	for _, ig := range b.InstanceGroups {
		igSize := fi.Int32Value(ig.Spec.MinSize)

		labels := make(map[string]string)
		labels[hetzner.TagKubernetesClusterName] = b.ClusterName()
		labels[hetzner.TagKubernetesInstanceGroup] = ig.Name
		labels[hetzner.TagKubernetesInstanceRole] = string(ig.Spec.Role)

		userData, err := b.BootstrapScriptBuilder.ResourceNodeUp(c, ig)
		if err != nil {
			return err
		}

		serverGroup := hetznertasks.ServerGroup{
			Name:       fi.String(ig.Name),
			Lifecycle:  b.Lifecycle,
			SSHKeys:    sshkeyTasks,
			Network:    b.LinkToNetwork(),
			Count:      int(igSize),
			Location:   ig.Spec.Subnets[0],
			Size:       ig.Spec.MachineType,
			Image:      ig.Spec.Image,
			EnableIPv4: true,
			EnableIPv6: false,
			UserData:   userData,
			Labels:     labels,
		}

		c.AddTask(&serverGroup)
	}

	return nil
}
