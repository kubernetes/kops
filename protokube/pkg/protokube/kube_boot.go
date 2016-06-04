package protokube

import (
	"github.com/golang/glog"
	"time"
)

type KubeBoot struct {
	master bool
	volumes Volumes
}

func NewKubeBoot(master bool, volumes Volumes) *KubeBoot {
	k := &KubeBoot{
		master: master,
		volumes: volumes,
	}
	return k
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
	if k.master {
		mountpoint, err := k.mountMasterVolume()
		if err != nil {
			return false, err
		}

		if mountpoint == "" {
			glog.Infof("unable to acquire master volume")
			return false, nil
		}

		glog.Infof("mounted master on %s", mountpoint)
		// TODO: Should we set up symlinks here?
	}

	return true, nil
}
