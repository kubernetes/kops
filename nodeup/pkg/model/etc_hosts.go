/*
Copyright 2021 The Kubernetes Authors.

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

// EtcHostsBuilder seeds some hostnames into /etc/hosts, avoiding some circular dependencies.
type EtcHostsBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &EtcHostsBuilder{}

// Build is responsible for configuring the gossip DNS tasks.
func (b *EtcHostsBuilder) Build(c *fi.ModelBuilderContext) error {

	task := &nodetasks.UpdateEtcHostsTask{
		Name: "control-plane-address",
	}

	if b.IsMaster && b.Cluster.IsGossip() {
		task.Records = append(task.Records, nodetasks.HostRecord{
			Hostname:  b.Cluster.Spec.MasterInternalName,
			Addresses: []string{"127.0.0.1"},
		})
		if b.Cluster.Spec.MasterPublicName != "" {
			task.Records = append(task.Records, nodetasks.HostRecord{
				Hostname:  b.Cluster.Spec.MasterPublicName,
				Addresses: []string{"127.0.0.1"},
			})
		}
	} else if b.Cluster.UsesNoneDNS() {
		task.Records = append(task.Records, nodetasks.HostRecord{
			Hostname:  b.Cluster.Spec.MasterInternalName,
			Addresses: []string{b.BootConfig.APIServer},
		})
		if b.UseKopsControllerForNodeBootstrap() {
			task.Records = append(task.Records, nodetasks.HostRecord{
				Hostname:  "kops-controller.internal." + b.Cluster.Name,
				Addresses: []string{b.BootConfig.APIServer},
			})
		}
	}

	if len(task.Records) != 0 {
		c.AddTask(task)
	}

	return nil
}
