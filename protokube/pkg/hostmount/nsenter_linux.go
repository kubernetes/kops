// +build linux

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

package hostmount

import (
	"fmt"

	"k8s.io/klog"
	"k8s.io/utils/mount"
	"k8s.io/utils/nsenter"
)

// Based on code from kubernetes/kubernetes: https://github.com/kubernetes/kubernetes/blob/release-1.15/pkg/volume/util/nsenter/nsenter_mount.go

const (
	// hostProcMountsPath is the default mount path for rootfs
	hostProcMountsPath = "/rootfs/proc/1/mounts"
)

func New(ne *nsenter.Nsenter) *Mounter {
	return &Mounter{ne: ne}
}

type Mounter struct {
	ne *nsenter.Nsenter
}

var _ mount.Interface = &Mounter{}

// List returns a list of all mounted filesystems in the host's mount namespace.
func (*Mounter) List() ([]mount.MountPoint, error) {
	return mount.ListProcMounts(hostProcMountsPath)
}

// Mount runs mount(8) in the host's root mount namespace.  Aside from this
// aspect, Mount has the same semantics as the mounter returned by mount.New()
func (n *Mounter) Mount(source string, target string, fstype string, options []string) error {
	bind, bindOpts, bindRemountOpts := mount.MakeBindOpts(options)

	if bind {
		err := n.doNsenterMount(source, target, fstype, bindOpts)
		if err != nil {
			return err
		}
		return n.doNsenterMount(source, target, fstype, bindRemountOpts)
	}

	return n.doNsenterMount(source, target, fstype, options)
}

// doNsenterMount nsenters the host's mount namespace and performs the
// requested mount.
func (n *Mounter) doNsenterMount(source, target, fstype string, options []string) error {
	klog.V(5).Infof("nsenter mount %s %s %s %v", source, target, fstype, options)
	cmd, args := n.makeNsenterArgs(source, target, fstype, options)
	outputBytes, err := n.ne.Exec(cmd, args).CombinedOutput()
	if len(outputBytes) != 0 {
		klog.V(5).Infof("Output of mounting %s to %s: %v", source, target, string(outputBytes))
	}
	return err
}

// makeNsenterArgs makes a list of argument to nsenter in order to do the
// requested mount.
func (n *Mounter) makeNsenterArgs(source, target, fstype string, options []string) (string, []string) {
	mountCmd := n.ne.AbsHostPath("mount")
	mountArgs := mount.MakeMountArgs(source, target, fstype, options)

	if systemdRunPath, hasSystemd := n.ne.SupportsSystemd(); hasSystemd {
		// Complete command line:
		// nsenter --mount=/rootfs/proc/1/ns/mnt -- /bin/systemd-run --description=... --scope -- /bin/mount -t <type> <what> <where>
		// Expected flow is:
		// * nsenter breaks out of container's mount namespace and executes
		//   host's systemd-run.
		// * systemd-run creates a transient scope (=~ cgroup) and executes its
		//   argument (/bin/mount) there.
		// * mount does its job, forks a fuse daemon if necessary and finishes.
		//   (systemd-run --scope finishes at this point, returning mount's exit
		//   code and stdout/stderr - thats one of --scope benefits).
		// * systemd keeps the fuse daemon running in the scope (i.e. in its own
		//   cgroup) until the fuse daemon dies (another --scope benefit).
		//   Kubelet container can be restarted and the fuse daemon survives.
		// * When the daemon dies (e.g. during unmount) systemd removes the
		//   scope automatically.
		mountCmd, mountArgs = mount.AddSystemdScope(systemdRunPath, target, mountCmd, mountArgs)
	}

	// Fall back to simple mount when the host has no systemd:
	// Complete command line:
	// nsenter --mount=/rootfs/proc/1/ns/mnt -- /bin/mount -t <type> <what> <where>
	// Expected flow is:
	// * nsenter breaks out of container's mount namespace and executes host's /bin/mount.
	// * mount does its job, forks a fuse daemon if necessary and finishes.
	// * Any fuse daemon runs in cgroup of kubelet docker container,
	//   restart of kubelet container will kill it!
	// No code here, mountCmd and mountArgs use /bin/mount

	return mountCmd, mountArgs
}

// We deliberately implement only the functions we need, so we don't have to maintain them...

func (n *Mounter) GetMountRefs(pathname string) ([]string, error) {
	return nil, fmt.Errorf("GetMountRefs not implemented for containerized mounter")
}

func (mounter *Mounter) IsLikelyNotMountPoint(file string) (bool, error) {
	return false, fmt.Errorf("IsLikelyNotMountPoint not implemented for containerized mounter")
}

func (n *Mounter) Unmount(target string) error {
	return fmt.Errorf("Unmount not implemented for containerized mounter")
}
