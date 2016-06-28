package protokube

import (
	"github.com/golang/glog"
	"net"
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

	DNS DNSProvider

	ModelDir string
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
	}

	return nil
}
