package protokube

import (
	"github.com/golang/glog"
	"time"
)

type KubeBoot struct {
	volumes Volumes
}

func NewKubeBoot(volumes Volumes) *KubeBoot {
	k := &KubeBoot{
		volumes: volumes,
	}
	return k
}

func (k *KubeBoot) Bootstrap() error {
	for {
		done, err := k.tryBootstrap()
		if err != nil {
			glog.Warningf("error during attempt to acquire master volume (will sleep and retry): %v", err)
		} else if done {
			break
		} else {
			glog.Infof("unable to acquire master volume; will sleep and retry")
		}

		time.Sleep(1 * time.Minute)
	}

	return nil
}

func (k *KubeBoot) tryBootstrap() (bool, error) {
	mountpoint, err := k.mountMasterVolume()
	if err != nil {
		return false, err
	}

	glog.Infof("mounted master on %s", mountpoint)
	// TODO: Should we set up symlinks here?

	return true, nil
}
