package protokube

import (
	"github.com/golang/glog"
	"net"
	"time"
)

type KubeBoot struct {
	Containerized bool
	RootFS        string

	Master            bool
	InternalDNSSuffix string
	InternalIP        net.IP
	MasterID          int
	EtcdClusters      []*EtcdClusterSpec

	Volumes Volumes
	DNS     DNSProvider
}

func (k *KubeBoot) PathFor(hostPath string) string {
	if hostPath[0] != '/' {
		glog.Fatalf("path was not absolute: %q", hostPath)
	}
	return k.RootFS + hostPath[1:]
}

func (k *KubeBoot) String() string {
	return DebugString(k)
}

func (k *KubeBoot) Bootstrap() error {
	for {
		done, err := k.tryBootstrap()
		if err != nil {
			glog.Warningf("error during attempt to bootstrap (will sleep and retry): %v", err)
		} else if done {
			break
		} else {
			glog.Infof("unable to bootstrap; will sleep and retry")
		}

		time.Sleep(1 * time.Minute)
	}

	return nil
}

func (k *KubeBoot) tryBootstrap() (bool, error) {
	if k.Master {
		volumeInfo, mountpoint, err := k.mountMasterVolume()
		if err != nil {
			return false, err
		}

		if mountpoint == "" {
			glog.Infof("unable to acquire master volume")
			return false, nil
		}

		glog.Infof("mounted master volume %q on %s", volumeInfo.Name, mountpoint)

		// Copy roles from volume
		k.EtcdClusters = volumeInfo.EtcdClusters
		for _, etcdClusterSpec := range volumeInfo.EtcdClusters {
			glog.Infof("Found etcd cluster spec on volume: %v", etcdClusterSpec)
		}

		k.MasterID = volumeInfo.MasterID

		// TODO: Should we set up symlinks here?
	}

	return true, nil
}
