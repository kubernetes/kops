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

package nodetasks

import (
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/kops/protokube/pkg/gossip/dns/hosts"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
)

// UpdateEtcHostsTask is responsible for updating /etc/hosts to set some DNS records, for gossip.
type UpdateEtcHostsTask struct {
	// Name is a reference for our task
	Name string

	// Records holds the records that should be updated
	Records []HostRecord
}

// HostRecord holds an individual host's addresses.
type HostRecord struct {
	fi.NotADependency

	// Hostname is the "DNS" name that we want to configure.
	Hostname string
	// Addresses holds the IP addresses to write.
	// Other IP addresses for the same Name will be removed.
	Addresses []string
}

var _ fi.Task = &UpdateEtcHostsTask{}

func (e *UpdateEtcHostsTask) String() string {
	return fmt.Sprintf("UpdateEtcHostsTask: %s", e.Name)
}

var _ fi.HasName = &UpdateEtcHostsTask{}

func (f *UpdateEtcHostsTask) GetName() *string {
	return &f.Name
}

func (e *UpdateEtcHostsTask) Find(c *fi.Context) (*UpdateEtcHostsTask, error) {
	// UpdateHostsFileWithRecords skips the update /etc/hosts if there are no changes,
	// so we don't check existing values here.
	return nil, nil
}

func (e *UpdateEtcHostsTask) Run(c *fi.NodeContext) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *UpdateEtcHostsTask) CheckChanges(a, e, changes *UpdateEtcHostsTask) error {
	return nil
}

func (_ *UpdateEtcHostsTask) RenderLocal(t *local.LocalTarget, a, e, changes *UpdateEtcHostsTask) error {
	etcHostsPath := "/etc/hosts"

	mutator := func(existing []string) (*hosts.HostMap, error) {
		hostMap := &hosts.HostMap{}
		badLines := hostMap.Parse(existing)
		if len(badLines) != 0 {
			klog.Warningf("ignoring unexpected lines in /etc/hosts: %v", badLines)
		}

		for _, record := range e.Records {
			hostMap.ReplaceRecords(record.Hostname, record.Addresses)
		}

		return hostMap, nil
	}

	if err := hosts.UpdateHostsFileWithRecords(etcHostsPath, mutator); err != nil {
		return fmt.Errorf("failed to update /etc/hosts: %w", err)
	}
	return nil
}

func (_ *UpdateEtcHostsTask) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *UpdateEtcHostsTask) error {
	return fmt.Errorf("UpdateEtcHostsTask::RenderCloudInit not supported")
}
