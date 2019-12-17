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

package gce

import (
	"context"
	"fmt"
	"strings"

	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/protokube/pkg/etcd"
	"k8s.io/kops/upup/pkg/fi"
)

func (c *gceCloudImplementation) allZones() ([]string, error) {
	var zones []string

	// TODO: use PageToken to list all not just the first 500
	ctx := context.Background()
	err := c.compute.Zones.List(c.project).Pages(ctx, func(page *compute.ZoneList) error {
		for _, zone := range page.Items {
			regionName := LastComponent(zone.Region)
			if regionName == c.region {
				zones = append(zones, zone.Name)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error listing zones: %v", err)
	}

	return zones, nil
}

// FindClusterStatus discovers the status of the cluster, by inspecting the cloud objects
func (c *gceCloudImplementation) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	etcdClusters, err := c.findEtcdStatus(cluster)
	if err != nil {
		return nil, err
	}

	status := &kops.ClusterStatus{
		EtcdClusters: etcdClusters,
	}
	klog.V(2).Infof("Cluster status (from cloud): %v", fi.DebugAsJsonString(status))
	return status, nil
}

// FindEtcdStatus discovers the status of etcd, by looking for the tagged etcd volumes
func (c *gceCloudImplementation) findEtcdStatus(cluster *kops.Cluster) ([]kops.EtcdClusterStatus, error) {
	statusMap := make(map[string]*kops.EtcdClusterStatus)

	labels := c.Labels()

	zones, err := c.allZones()
	if err != nil {
		return nil, err
	}

	var disks []*compute.Disk

	// TODO: Filter disks query by Label?
	ctx := context.Background()
	for _, zone := range zones {
		err := c.compute.Disks.List(c.project, zone).Pages(ctx, func(page *compute.DiskList) error {
			for _, d := range page.Items {
				klog.V(4).Infof("Found disk %q with labels %v", d.Name, d.Labels)

				match := true
				for k, v := range labels {
					if d.Labels[k] != v {
						match = false
					}
				}
				if match {
					disks = append(disks, d)
				}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error describing volumes: %v", err)
		}
	}

	for _, disk := range disks {
		etcdClusterName := ""
		var etcdClusterSpec *etcd.EtcdClusterSpec
		master := false
		for k, v := range disk.Labels {
			if strings.HasPrefix(k, GceLabelNameEtcdClusterPrefix) {
				etcdClusterName = strings.TrimPrefix(k, GceLabelNameEtcdClusterPrefix)
				value, err := DecodeGCELabel(v)
				if err != nil {
					return nil, fmt.Errorf("unexpected etcd label on volume %q: %s=%s", disk.Name, k, v)
				}
				spec, err := etcd.ParseEtcdClusterSpec(etcdClusterName, value)
				if err != nil {
					return nil, fmt.Errorf("error parsing etcd cluster label %q on volume %q: %v", value, disk.Name, err)
				}
				etcdClusterSpec = spec
			} else if strings.HasPrefix(k, GceLabelNameRolePrefix) {
				roleName := strings.TrimPrefix(k, GceLabelNameRolePrefix)
				if roleName == "master" {
					master = true
				}
			}
		}
		if etcdClusterName == "" || etcdClusterSpec == nil || !master {
			continue
		}

		status := statusMap[etcdClusterName]
		if status == nil {
			status = &kops.EtcdClusterStatus{
				Name: etcdClusterName,
			}
			statusMap[etcdClusterName] = status
		}

		status.Members = append(status.Members, &kops.EtcdMemberStatus{
			Name:     etcdClusterSpec.NodeName,
			VolumeId: disk.Name,
		})
	}

	var status []kops.EtcdClusterStatus
	for _, v := range statusMap {
		status = append(status, *v)
	}
	return status, nil
}
