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

package protokube

import (
	"sort"
	"strings"
	"testing"

	"k8s.io/kops/protokube/pkg/etcd"
)

func getIDs(volumes []*Volume) string {
	var ids []string
	for _, v := range volumes {
		ids = append(ids, v.ID)
	}
	return strings.Join(ids, ",")
}

func Test_VolumeSort_ByEtcdClusterName(t *testing.T) {
	v1 := &Volume{}
	v1.ID = "1"
	v2 := &Volume{}
	v2.ID = "2"
	v3 := &Volume{}
	v3.ID = "3"

	volumes := []*Volume{v1, v2, v3}
	sort.Stable(ByEtcdClusterName(volumes))
	if getIDs(volumes) != "1,2,3" {
		t.Fatalf("Fail at sort 1: %v", getIDs(volumes))
	}

	v2.Info.EtcdClusters = append(v2.Info.EtcdClusters, &etcd.EtcdClusterSpec{ClusterKey: "events"})
	sort.Stable(ByEtcdClusterName(volumes))
	if getIDs(volumes) != "2,1,3" {
		t.Fatalf("Fail at sort 2: %v", getIDs(volumes))
	}

	v3.Info.EtcdClusters = append(v3.Info.EtcdClusters, &etcd.EtcdClusterSpec{ClusterKey: "main"})
	sort.Stable(ByEtcdClusterName(volumes))
	if getIDs(volumes) != "3,2,1" {
		t.Fatalf("Fail at sort 3: %v", getIDs(volumes))
	}

}

func Test_Mount_Volumes(t *testing.T) {
	grid := []struct {
		volume      *Volume
		doNotMount  bool
		description string
	}{
		{
			&Volume{
				LocalDevice: "/dev/xvda",
			},
			true,
			"xda without a etcd cluster, do not mount",
		},
		{
			&Volume{
				LocalDevice: "/dev/xvdb",
				Info: VolumeInfo{
					EtcdClusters: []*etcd.EtcdClusterSpec{
						{
							ClusterKey: "foo",
							NodeName:   "bar",
						},
					},
				},
			},
			true,
			"xdb with a etcd cluster, mount",
		},
	}

	for _, g := range grid {
		d := doNotMountVolume(g.volume)
		if d && !g.doNotMount {
			t.Fatalf("volume mount should not have mounted: %s, description: %s", g.volume.LocalDevice, g.description)
		}
	}

}
