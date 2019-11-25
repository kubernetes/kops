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

package fi

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"

	"k8s.io/klog"
	"k8s.io/kops/util/pkg/hashing"
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
	klog.Infof("Writing file %q", destPath)

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
	klog.Infof("Changing file mode for %q to %s", destPath, fileMode)

	err = os.Chmod(destPath, fileMode)
	if err != nil {
		return changed, fmt.Errorf("error setting file mode for %q: %v", destPath, err)
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
		klog.V(2).Infof("Hash matched for %q: %v", f, expected)
		return true, nil
	}
	klog.V(2).Infof("Hash did not match for %q: actual=%v vs expected=%v", f, actual, expected)
	return false, nil
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
		klog.Warningf("unexpected error closing stream: %v", err)
	}
}
