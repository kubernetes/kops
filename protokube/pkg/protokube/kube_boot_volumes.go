package protokube

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/util/exec"
	"k8s.io/kubernetes/pkg/util/mount"
	"os"
	"time"
)

const MasterMountpoint = "/mnt/master-pd"

func (k *KubeBoot) mountMasterVolume() (*VolumeInfo, string, error) {
	// TODO: mount ephemeral volumes (particular on AWS)?

	// Mount a master volume
	volume, device, err := k.attachMasterVolume()
	if err != nil {
		return nil, "", fmt.Errorf("unable to attach master volume: %q", err)
	}

	if device == "" {
		return nil, "", nil
	}

	glog.V(2).Infof("Master volume %q is attached at %q", volume.Name, device)

	glog.Infof("Doing safe-format-and-mount of %s to %s", device, MasterMountpoint)
	fstype := ""
	err = k.safeFormatAndMount(device, MasterMountpoint, fstype)
	if err != nil {
		return nil, "", fmt.Errorf("unable to mount master volume: %q", err)
	}

	return volume, MasterMountpoint, nil
}

func (k *KubeBoot) safeFormatAndMount(device string, mountpoint string, fstype string) error {
	// Wait for the device to show up

	for {
		_, err := os.Stat(k.PathFor(device))
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

	//// Mount the device
	//var mounter mount.Interface
	//runner := exec.New()
	//if k.Containerized {
	//	mounter = mount.NewNsenterMounter()
	//	runner = NewChrootRunner(runner, "/rootfs")
	//} else {
	//	mounter = mount.New()
	//}

	// If we are containerized, we still first SafeFormatAndMount in our namespace
	// This is because SafeFormatAndMount doesn't seem to work in a container
	safeFormatAndMount := &mount.SafeFormatAndMount{Interface: mount.New(), Runner: exec.New()}

	// Check if it is already mounted
	mounts, err := safeFormatAndMount.List()
	if err != nil {
		return fmt.Errorf("error listing existing mounts: %v", err)
	}

	// Note: IsLikelyNotMountPoint is not containerized

	findMountpoint := k.PathFor(mountpoint)
	var existing []*mount.MountPoint
	for i := range mounts {
		m := &mounts[i]
		glog.V(2).Infof("found existing mount: %v", m)
		if m.Path == findMountpoint {
			existing = append(existing, m)
		}
	}

	options := []string{}
	//if readOnly {
	//	options = append(options, "ro")
	//}
	if len(existing) == 0 {
		glog.Infof("Creating mount directory %q", k.PathFor(mountpoint))
		if err := os.MkdirAll(k.PathFor(mountpoint), 0750); err != nil {
			return err
		}

		glog.Infof("Mounting device %q on %q", k.PathFor(device), k.PathFor(mountpoint))

		err = safeFormatAndMount.FormatAndMount(k.PathFor(device), k.PathFor(mountpoint), fstype, options)
		if err != nil {
			//os.Remove(mountpoint)
			return fmt.Errorf("error formatting and mounting disk %q on %q: %v", k.PathFor(device), k.PathFor(mountpoint), err)
		}

		// If we are containerized, we then also mount it into the host
		if k.Containerized {
			hostMounter := mount.NewNsenterMounter()
			err = hostMounter.Mount(device, mountpoint, fstype, options)
			if err != nil {
				//os.Remove(mountpoint)
				return fmt.Errorf("error formatting and mounting disk %q on %q in host: %v", device, mountpoint, err)
			}
		}
	} else {
		glog.Infof("Device already mounted on : %q, verifying it is our device", mountpoint)

		if len(existing) != 1 {
			glog.Infof("Existing mounts unexpected")

			for i := range mounts {
				m := &mounts[i]
				glog.Infof("%s\t%s", m.Device, m.Path)
			}

			return fmt.Errorf("Found multiple existing mounts of %q at %q", device, mountpoint)
		} else {
			glog.Infof("Found existing mount of %q and %q", device, mountpoint)
		}

	}
	return nil
}

func (k *KubeBoot) attachMasterVolume() (*VolumeInfo, string, error) {
	volumes, err := k.Volumes.FindMountedVolumes()
	if err != nil {
		return nil, "", err
	}

	if len(volumes) != 0 {
		if len(volumes) != 1 {
			// TODO: unmount?
			glog.Warningf("Found multiple master volumes: %v", volumes)
		}

		volume := volumes[0]

		glog.V(2).Infof("Found master volume already attached: %q", volume.Name)

		device, err := k.Volumes.AttachVolume(volume)
		if err != nil {
			return nil, "", fmt.Errorf("Error attaching volume %q: %v", volume.Name, err)
		}
		return &volume.Info, device, nil
	}

	volumes, err = k.Volumes.FindMountableVolumes()
	if err != nil {
		return nil, "", err
	}

	if len(volumes) == 0 {
		glog.Infof("No available master volumes")
		return nil, "", nil
	}

	for _, volume := range volumes {
		if !volume.Available {
			continue
		}

		glog.V(2).Infof("Trying to mount master volume: %q", volume.Name)

		device, err := k.Volumes.AttachVolume(volume)
		if err != nil {
			return nil, "", fmt.Errorf("Error attaching volume %q: %v", volume.Name, err)
		}
		return &volume.Info, device, nil
	}

	return nil, "", nil
}
