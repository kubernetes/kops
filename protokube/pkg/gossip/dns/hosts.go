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

package dns

import (
	"k8s.io/klog/v2"
	"k8s.io/kops/protokube/pkg/gossip/dns/hosts"
)

// HostsFile stores DNS records into /etc/hosts
type HostsFile struct {
	Path string
}

var _ DNSTarget = &HostsFile{}

func (h *HostsFile) Update(snapshot *DNSViewSnapshot) error {
	klog.V(2).Infof("Updating hosts file with snapshot version %v", snapshot.version)

	mutator := func(existing []string) (*hosts.HostMap, error) {
		hostMap := &hosts.HostMap{}
		badLines := hostMap.Parse(existing)
		if len(badLines) != 0 {
			klog.Warningf("ignoring unexpected lines in /etc/hosts: %v", badLines)
		}

		zones := snapshot.ListZones()
		for _, zone := range zones {
			records := snapshot.RecordsForZone(zone)

			for _, record := range records {
				if record.RrsType != "A" {
					klog.Warningf("skipping record of unhandled type: %v", record)
					continue
				}

				hostMap.ReplaceRecords(record.Name, record.Rrdatas)
			}
		}

		return hostMap, nil
	}

	return hosts.UpdateHostsFileWithRecords(h.Path, mutator)
}
