package protokube

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/util/exec"
	"k8s.io/kubernetes/pkg/util/mount"
	"os"
	"time"
)

const MasterMountpoint = "/master-pd"

func (k *KubeBoot) mountMasterVolume() (string, error) {
	// TODO: mount ephemeral volumes (particular on AWS)?

	// Mount a master volume
	device, err := k.attachMasterVolume()
	if err != nil {
		return "", fmt.Errorf("unable to attach master volume: %q", err)
	}

	if device == "" {
		return "", nil
	}

	glog.V(2).Infof("Master volume is attached at %q", device)

	fstype := ""
	err = k.safeFormatAndMount(device, MasterMountpoint, fstype)
	if err != nil {
		return "", fmt.Errorf("unable to mount master volume: %q", err)
	}

	return MasterMountpoint, nil
}

func (k *KubeBoot) safeFormatAndMount(device string, mountpoint string, fstype string) error {
	// Wait for the device to show up
	for {
		_, err := os.Stat(device)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("error checking for device %q: %v", device, err)
		}
		glog.Infof("Waiting for device %q to be attached", device)
		time.Sleep(1 * time.Second)
	}
	glog.Infof("Found device %q", device)

	// Mount the device

	mounter := &mount.SafeFormatAndMount{Interface: mount.New(), Runner: exec.New()}

	// Only mount the PD globally once.
	notMnt, err := mounter.IsLikelyNotMountPoint(mountpoint)
	if err != nil {
		if os.IsNotExist(err) {
			glog.Infof("Creating mount directory %q", mountpoint)
			if err := os.MkdirAll(mountpoint, 0750); err != nil {
				return err
			}
			notMnt = true
		} else {
			return err
		}
	}

	options := []string{}
	//if readOnly {
	//	options = append(options, "ro")
	//}
	if notMnt {
		glog.Infof("Mounting device %q on %q", device, mountpoint)

		err = mounter.FormatAndMount(device, mountpoint, fstype, options)
		if err != nil {
			//os.Remove(mountpoint)
			return fmt.Errorf("error formatting and mounting disk %q on %q: %v", device, mountpoint, err)
		}
	} else {
		glog.Infof("Device already mounted on : %q, verifying it is our device", mountpoint)

		mounts, err := mounter.List()
		if err != nil {
			return fmt.Errorf("error listing existing mounts: %v", err)
		}

		var existing []*mount.MountPoint
		for i := range mounts {
			m := &mounts[i]
			if m.Path == mountpoint {
				existing = append(existing, m)
			}
		}

		if len(existing) != 1 {
			glog.Infof("Existing mounts unexpected")

			for i := range mounts {
				m := &mounts[i]
				glog.Infof("%s\t%s", m.Device, m.Path)
			}
		}

		if len(existing) == 0 {
			return fmt.Errorf("Unable to find existing mount of %q at %q", device, mountpoint)
		} else if len(existing) != 1 {
			return fmt.Errorf("Found multiple existing mounts of %q at %q", device, mountpoint)
		} else {
			glog.Infof("Found existing mount of %q and %q", device, mountpoint)
		}

	}
	return nil
}

func (k *KubeBoot) attachMasterVolume() (string, error) {
	volumes, err := k.volumes.FindMountedVolumes()
	if err != nil {
		return "", err
	}

	if len(volumes) != 0 {
		if len(volumes) != 1 {
			// TODO: unmount?
			glog.Warningf("Found multiple master volumes: %v", volumes)
		}

		glog.V(2).Infof("Found master volume already attached: %q", volumes[0].Name)

		device, err := k.volumes.AttachVolume(volumes[0])
		if err != nil {
			return "", fmt.Errorf("Error attaching volume %q: %v", volumes[0].Name, err)
		}
		return device, nil
	}

	volumes, err = k.volumes.FindMountableVolumes()
	if err != nil {
		return "", err
	}

	if len(volumes) == 0 {
		glog.Infof("No available master volumes")
		return "", nil
	}

	for _, volume := range volumes {
		if !volume.Available {
			continue
		}

		glog.V(2).Infof("Trying to mount master volume: %q", volume.Name)

		device, err := k.volumes.AttachVolume(volume)
		if err != nil {
			return "", fmt.Errorf("Error attaching volume %q: %v", volume.Name, err)
		}
		return device, nil
	}

	return "", nil
}
