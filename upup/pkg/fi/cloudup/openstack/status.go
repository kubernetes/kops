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

package openstack

import (
	"fmt"
	"strings"

	cinderv3 "github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/protokube/pkg/etcd"
	"k8s.io/kops/upup/pkg/fi"
)

// FindClusterStatus discovers the status of the cluster, by looking for the tagged etcd volumes
func (c *openstackCloud) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	etcdStatus, err := findEtcdStatus(c, cluster)
	if err != nil {
		return nil, err
	}
	status := &kops.ClusterStatus{
		EtcdClusters: etcdStatus,
	}
	klog.V(2).Infof("Cluster status (from cloud): %v", fi.DebugAsJsonString(status))
	return status, nil
}

// findEtcdStatus discovers the status of etcd, by looking for the tagged etcd volumes
func findEtcdStatus(c *openstackCloud, cluster *kops.Cluster) ([]kops.EtcdClusterStatus, error) {
	statusMap := make(map[string]*kops.EtcdClusterStatus)
	klog.V(2).Infof("Querying Openstack for etcd volumes")
	opt := cinderv3.ListOpts{
		Metadata: c.tags,
	}
	volumes, err := c.ListVolumes(opt)
	if err != nil {
		return nil, fmt.Errorf("error describing volumes: %v", err)
	}

	for _, volume := range volumes {
		volumeID := volume.ID

		etcdClusterName := ""
		var etcdClusterSpec *etcd.EtcdClusterSpec

		master := false
		for k, v := range volume.Metadata {
			if strings.HasPrefix(k, TagNameEtcdClusterPrefix) {
				etcdClusterName := strings.TrimPrefix(k, TagNameEtcdClusterPrefix)
				etcdClusterSpec, err = etcd.ParseEtcdClusterSpec(etcdClusterName, v)
				if err != nil {
					return nil, fmt.Errorf("error parsing etcd cluster tag %q on volume %q: %v", v, volumeID, err)
				}
			} else if k == TagNameRolePrefix+TagRoleMaster {
				master = true
			}
		}
		if etcdClusterSpec == nil || !master {
			continue
		}

		etcdClusterName = etcdClusterSpec.ClusterKey
		status := statusMap[etcdClusterName]
		if status == nil {
			status = &kops.EtcdClusterStatus{
				Name: etcdClusterName,
			}
			statusMap[etcdClusterName] = status
		}

		memberName := etcdClusterSpec.NodeName
		status.Members = append(status.Members, &kops.EtcdMemberStatus{
			Name:     memberName,
			VolumeId: volume.ID,
		})
	}
	var status []kops.EtcdClusterStatus
	for _, v := range statusMap {
		status = append(status, *v)
	}
	return status, nil
}
