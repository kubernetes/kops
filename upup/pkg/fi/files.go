package fi

import (
	"fmt"
	"github.com/golang/glog"
	"io"
	"k8s.io/kube-deploy/upup/pkg/fi/hashing"
	"os"
	"path"
	"strconv"
	"syscall"
)

func WriteFile(destPath string, contents Resource, fileMode os.FileMode, dirMode os.FileMode) error {
	err := os.MkdirAll(path.Dir(destPath), dirMode)
	if err != nil {
		return fmt.Errorf("error creating directories for destination file %q: %v", destPath, err)
	}

	err = writeFileContents(destPath, contents, fileMode)
	if err != nil {
		return err
	}

	_, err = EnsureFileMode(destPath, fileMode)
	if err != nil {
		return err
	}

	return nil
}

func writeFileContents(destPath string, src Resource, fileMode os.FileMode) error {
	glog.Infof("Writing file %q", destPath)

	out, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return fmt.Errorf("error opening destination file %q: %v", destPath, err)
	}
	defer out.Close()

	in, err := src.Open()
	if err != nil {
		return fmt.Errorf("error opening source resource for file %q: %v", destPath, err)
	}
	defer SafeClose(in)

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("error writing file %q: %v", destPath, err)
	}
	return nil
}

func EnsureFileMode(destPath string, fileMode os.FileMode) (bool, error) {
	changed := false
	stat, err := os.Lstat(destPath)
	if err != nil {
		return changed, fmt.Errorf("error getting file mode for %q: %v", destPath, err)
	}
	if (stat.Mode() & os.ModePerm) == fileMode {
		return changed, nil
	}
	glog.Infof("Changing file mode for %q to %s", destPath, fileMode)

	err = os.Chmod(destPath, fileMode)
	if err != nil {
		return changed, fmt.Errorf("error setting file mode for %q: %v", destPath, err)
	}
	changed = true
	return changed, nil
}

func EnsureFileOwner(destPath string, owner string, groupName string) (bool, error) {
	changed := false
	stat, err := os.Lstat(destPath)
	if err != nil {
		return changed, fmt.Errorf("error getting file stat for %q: %v", destPath, err)
	}

	user, err := LookupUser(owner) //user.Lookup(owner)
	if err != nil {
		return changed, fmt.Errorf("error looking up user %q: %v", owner, err)
	}

	group, err := LookupGroup(groupName)
	if err != nil {
		return changed, fmt.Errorf("error looking up group %q: %v", groupName, err)
	}

	if int(stat.Sys().(*syscall.Stat_t).Uid) == user.Uid && int(stat.Sys().(*syscall.Stat_t).Gid) == group.Gid {
		return changed, nil
	}

	glog.Infof("Changing file owner/group for %q to %s:%s", destPath, owner, group)
	err = os.Lchown(destPath, user.Uid, group.Gid)
	if err != nil {
		return changed, fmt.Errorf("error setting file owner/group for %q: %v", destPath, err)
	}
	changed = true

	return changed, nil
}

func fileHasHash(f string, expected *hashing.Hash) (bool, error) {
	actual, err := expected.Algorithm.HashFile(f)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if actual.Equal(expected) {
		glog.V(2).Infof("Hash matched for %q: %v", f, expected)
		return true, nil
	} else {
		glog.V(2).Infof("Hash did not match for %q: actual=%v vs expected=%v", f, actual, expected)
		return false, nil
	}
}

func ParseFileMode(s string, defaultMode os.FileMode) (os.FileMode, error) {
	fileMode := defaultMode
	if s != "" {
		v, err := strconv.ParseUint(s, 8, 32)
		if err != nil {
			return fileMode, fmt.Errorf("cannot parse file mode %q", s)
		}
		fileMode = os.FileMode(v)
	}
	return fileMode, nil
}

func FileModeToString(mode os.FileMode) string {
	return "0" + strconv.FormatUint(uint64(mode), 8)
}

func SafeClose(r io.Reader) {
	if r == nil {
		return
	}
	closer, ok := r.(io.Closer)
	if !ok {
		return
	}
	err := closer.Close()
	if err != nil {
		glog.Warningf("unexpected error closing stream: %v", err)
	}
}
