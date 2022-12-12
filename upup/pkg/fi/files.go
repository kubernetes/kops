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
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"k8s.io/klog/v2"
	"k8s.io/kops/util/pkg/hashing"
)

// WriteFile writes a file to the specified path, setting the mode, owner & group.
func WriteFile(ctx context.Context, destPath string, contents Resource, fileMode os.FileMode, dirMode os.FileMode, owner string, group string) error {
	err := os.MkdirAll(path.Dir(destPath), dirMode)
	if err != nil {
		return fmt.Errorf("error creating directories for destination file %q: %v", destPath, err)
	}

	err = writeFileContents(ctx, destPath, contents, fileMode)
	if err != nil {
		return err
	}

	_, err = EnsureFileMode(destPath, fileMode)
	if err != nil {
		return err
	}

	_, err = EnsureFileOwner(destPath, owner, group)
	if err != nil {
		return err
	}

	return nil
}

func writeFileContents(ctx context.Context, destPath string, src Resource, fileMode os.FileMode) error {
	klog.Infof("Writing file %q", destPath)

	in, err := src.Open(ctx)
	if err != nil {
		return fmt.Errorf("error opening source resource for file %q: %v", destPath, err)
	}
	defer SafeClose(in)

	dir := filepath.Dir(destPath)

	tempFile, err := os.CreateTemp(dir, ".writefile")
	if err != nil {
		return fmt.Errorf("error creating temp file in %q: %w", dir, err)
	}

	closeTempFile := true
	deleteTempFile := true
	defer func() {
		if closeTempFile {
			if err := tempFile.Close(); err != nil {
				klog.Warningf("error closing tempfile %q: %v", tempFile.Name(), err)
			}
		}
		if deleteTempFile {
			if err := os.Remove(tempFile.Name()); err != nil {
				klog.Warningf("error removing tempfile %q: %v", tempFile.Name(), err)
			}
		}
	}()

	if _, err := io.Copy(tempFile, in); err != nil {
		return fmt.Errorf("error writing file %q: %v", tempFile.Name(), err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("error closing temp file %q: %w", tempFile.Name(), err)
	}
	closeTempFile = false

	if err := os.Rename(tempFile.Name(), destPath); err != nil {
		return fmt.Errorf("error renaming temp file %q -> %q: %w", tempFile.Name(), destPath, err)
	}
	deleteTempFile = false

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
