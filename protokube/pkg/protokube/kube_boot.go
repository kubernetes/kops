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

package protokube

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/dns-controller/pkg/dns"
	"net"
	"os/exec"
	"time"
)

type KubeBoot struct {
	Master            bool
	InternalDNSSuffix string
	InternalIP        net.IP
	//MasterID          int
	//EtcdClusters      []*EtcdClusterSpec

	volumeMounter   *VolumeMountController
	etcdControllers map[string]*EtcdController

	DNSScope dns.Scope

	ModelDir string

	Channels []string

	Kubernetes *KubernetesContext
}

func (k *KubeBoot) Init(volumesProvider Volumes) {
	k.volumeMounter = newVolumeMountController(volumesProvider)
	k.etcdControllers = make(map[string]*EtcdController)
}

var Containerized = false
var RootFS = "/"

func PathFor(hostPath string) string {
	if hostPath[0] != '/' {
		glog.Fatalf("path was not absolute: %q", hostPath)
	}
	return RootFS + hostPath[1:]
}

func (k *KubeBoot) String() string {
	return DebugString(k)
}

func (k *KubeBoot) RunSyncLoop() {
	for {
		err := k.syncOnce()
		if err != nil {
			glog.Warningf("error during attempt to bootstrap (will sleep and retry): %v", err)
		}

		time.Sleep(1 * time.Minute)
	}
}

func (k *KubeBoot) syncOnce() error {
	if k.Master {
		volumes, err := k.volumeMounter.mountMasterVolumes()
		if err != nil {
			return err
		}

		for _, v := range volumes {
			for _, etcdClusterSpec := range v.Info.EtcdClusters {
				key := etcdClusterSpec.ClusterKey + "::" + etcdClusterSpec.NodeName
				etcdController := k.etcdControllers[key]
				if etcdController == nil {
					glog.Infof("Found etcd cluster spec on volume %q: %v", v.ID, etcdClusterSpec)

					etcdController, err := newEtcdController(k, v, etcdClusterSpec)
					if err != nil {
						glog.Warningf("error building etcd controller: %v", err)
					} else {
						k.etcdControllers[key] = etcdController
						go etcdController.RunSyncLoop()
					}
				}
			}
		}

		//// Copy roles from volume
		//k.EtcdClusters = volumeInfo.EtcdClusters
		//for _, etcdClusterSpec := range volumeInfo.EtcdClusters {
		//	glog.Infof("Found etcd cluster spec on volume: %v", etcdClusterSpec)
		//}

		//k.MasterID = volumeInfo.MasterID

		// TODO: Should we set up symlinks here?
	} else {
		glog.V(4).Infof("Not in role master; won't scan for volumes")
	}

	if k.Master {
		if err := ApplyMasterTaints(k.Kubernetes); err != nil {
			glog.Warningf("error updating master taints: %v", err)
		}
	}

	// Ensure kubelet is running. We avoid doing this automatically so
	// that when kubelet comes up the first time, all volume mounts
	// and DNS are available, avoiding the scenario where
	// etcd/apiserver retry too many times and go into backoff.
	if err := enableKubelet(); err != nil {
		glog.Warningf("error ensuring kubelet started: %v", err)
	}

	for _, channel := range k.Channels {
		if err := ApplyChannel(channel); err != nil {
			glog.Warningf("error applying channel %q: %v", channel, err)
		}
	}

	return nil
}

// enableKubelet: Make sure kubelet is running.
func enableKubelet() error {
	// TODO: Check/log status of kubelet
	// (in particular, we want to avoid kubernetes/kubernetes#40123 )
	glog.V(2).Infof("ensuring that kubelet systemd service is running")
	cmd := exec.Command("systemctl", "start", "--no-block", "kubelet")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error starting kubelet: %v\nOutput: %s", err, output)
	}
	return nil
}
