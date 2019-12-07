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

package nodetasks

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
	"k8s.io/kops/upup/pkg/fi/utils"
	utilexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"
)

// MountDiskTask is responsible for mounting a device on a mountpoint
// It will wait for the device to show up, safe_format_and_mount it,
// and then mount it.
type MountDiskTask struct {
	Name string

	Device     string `json:"device"`
	Mountpoint string `json:"mountpoint"`
}

var _ fi.Task = &MountDiskTask{}

func (s *MountDiskTask) String() string {
	return fmt.Sprintf("MountDisk: %s %s->%s", s.Name, s.Device, s.Mountpoint)
}

var _ CreatesDir = &MountDiskTask{}

// Dir implements CreatesDir::Dir
func (e *MountDiskTask) Dir() string {
	return e.Mountpoint
}

var _ fi.HasDependencies = &MountDiskTask{}

// GetDependencies implements HasDependencies::GetDependencies
func (e *MountDiskTask) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task

	// Requires parent directories to be created
	deps = append(deps, findCreatesDirParents(e.Mountpoint, tasks)...)
	return deps
}

func NewMountDiskTask(name string, contents string, meta string) (fi.Task, error) {
	s := &MountDiskTask{Name: name}

	err := utils.YamlUnmarshal([]byte(contents), s)
	if err != nil {
		return nil, fmt.Errorf("error parsing json for disk %q: %v", name, err)
	}

	return s, nil
}

func (e *MountDiskTask) Find(c *fi.Context) (*MountDiskTask, error) {
	mounter := mount.New("")

	mps, err := mounter.List()
	if err != nil {
		return nil, fmt.Errorf("error finding existing mounts: %v", err)
	}

	// If device is a symlink, it will show up by its final name
	targetDevice, err := filepath.EvalSymlinks(e.Device)
	if err != nil {
		return nil, fmt.Errorf("error resolving device symlinks for %q: %v", e.Device, err)
	}

	for i := range mps {
		mp := &mps[i]
		if mp.Device == targetDevice {
			actual := &MountDiskTask{
				Name:       e.Name,
				Mountpoint: mp.Path,
				Device:     e.Device, // Use our alias, to keep change detection happy
			}
			return actual, nil
		}
	}

	return nil, nil
}

func (e *MountDiskTask) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *MountDiskTask) CheckChanges(a, e, changes *MountDiskTask) error {
	return nil
}

func (_ *MountDiskTask) RenderLocal(t *local.LocalTarget, a, e, changes *MountDiskTask) error {
	dirMode := os.FileMode(0755)

	// Create the mountpoint
	err := os.MkdirAll(e.Mountpoint, dirMode)
	if err != nil {
		return fmt.Errorf("error creating mountpoint %q: %v", e.Mountpoint, err)
	}

	// Wait for the device to show up
	for {
		_, err := os.Stat(e.Device)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("error checking for device %q: %v", e.Device, err)
		}
		klog.Infof("Waiting for device %q to be attached", e.Device)
		time.Sleep(1 * time.Second)
	}
	klog.Infof("Found device %q", e.Device)

	// Mount the device
	if changes.Mountpoint != "" {
		klog.Infof("Mounting device %q on %q", e.Device, e.Mountpoint)

		mounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: utilexec.New()}

		fstype := ""
		options := []string{}

		err := mounter.FormatAndMount(e.Device, e.Mountpoint, fstype, options)
		if err != nil {
			return fmt.Errorf("error formatting and mounting disk %q on %q: %v", e.Device, e.Mountpoint, err)
		}
	}

	// TODO: Should we add to /etc/fstab?
	// Mount the master PD as early as possible
	// echo "/dev/xvdb /mnt/master-pd ext4 noatime 0 0" >> /etc/fstab

	return nil
}

func (_ *MountDiskTask) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *MountDiskTask) error {
	// TODO: Run safe_format_and_mount
	// Download on aws (or bake into image)
	//	# TODO: Where to get safe_format_and_mount?
	//mkdir -p /usr/share/google
	//cd /usr/share/google
	//download-or-bust "dc96f40fdc9a0815f099a51738587ef5a976f1da" https://raw.githubusercontent.com/GoogleCloudPlatform/compute-image-packages/82b75f314528b90485d5239ab5d5495cc22d775f/google-startup-scripts/usr/share/google/safe_format_and_mount
	//chmod +x safe_format_and_mount

	return fmt.Errorf("Disk::RenderCloudInit not implemented")
}
