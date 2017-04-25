package net

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

var ErrLinkNotFound = errors.New("Link not found")

// NB: The following function is unsafe, because:
//     - It changes a network namespace (netns) of an OS thread which runs
//       the function. During execution, the Go runtime might clone a new OS thread
//       for scheduling other go-routines, thus they might end up running in
//       a "wrong" netns.
//     - runtime.LockOSThread does not guarantee that a spawned go-routine on
//       the locked thread will be run by it. Thus, the work function is
//       not allowed to spawn any go-routine which is dependent on the given netns.

//     Please see https://github.com/weaveworks/weave/issues/2388#issuecomment-228365069
//     for more details and make sure that you understand the implications before
//     using the function!
func WithNetNSUnsafe(ns netns.NsHandle, work func() error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	oldNs, err := netns.Get()
	if err == nil {
		defer oldNs.Close()

		err = netns.Set(ns)
		if err == nil {
			defer netns.Set(oldNs)

			err = work()
		}
	}

	return err
}

func WithNetNSLinkUnsafe(ns netns.NsHandle, ifName string, work func(link netlink.Link) error) error {
	return WithNetNSUnsafe(ns, func() error {
		link, err := netlink.LinkByName(ifName)
		if err != nil {
			if err.Error() == errors.New("Link not found").Error() {
				return ErrLinkNotFound
			}
			return err
		}
		return work(link)
	})
}

var WeaveUtilCmd = "weaveutil"

// A safe version of WithNetNS* which creates a process executing
// "nsenter --net=<ns-path> weaveutil <cmd> [args]".
func WithNetNS(nsPath string, cmd string, args ...string) ([]byte, error) {
	var stdout, stderr bytes.Buffer

	args = append([]string{"--net=" + nsPath, WeaveUtilCmd, cmd}, args...)
	c := exec.Command("nsenter", args...)
	c.Stdout = &stdout
	c.Stderr = &stderr
	if err := c.Run(); err != nil {
		return nil, fmt.Errorf("%s: %s", string(stderr.Bytes()), err)
	}

	return stdout.Bytes(), nil
}

func WithNetNSByPid(pid int, cmd string, args ...string) ([]byte, error) {
	return WithNetNS(NSPathByPid(pid), cmd, args...)
}

func NSPathByPid(pid int) string {
	return NSPathByPidWithRoot("/", pid)
}

func NSPathByPidWithRoot(root string, pid int) string {
	return filepath.Join(root, fmt.Sprintf("/proc/%d/ns/net", pid))
}
