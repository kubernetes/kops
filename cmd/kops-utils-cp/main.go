/*
Copyright 2023 The Kubernetes Authors.

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

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
)

func copyFile(source, targetDir string, force bool) error {
	klog.Infof("Copying source file %q to target directory %q", source, targetDir)

	sf, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("unable to open source file %q: %w", source, err)
	}
	defer sf.Close()

	fi, err := sf.Stat()
	if err != nil {
		return fmt.Errorf("unable to stat source file %q: %w", source, err)
	}

	destPath := filepath.Join(targetDir, filepath.Base(source))

	if force {
		if err := os.Remove(destPath); err != nil {
			if os.IsNotExist(err) {
				// ignore
			} else {
				return fmt.Errorf("error removing file %q (for force): %w", destPath, err)
			}
		} else {
			klog.Infof("removed existing file %q (for force)", destPath)
		}
	}

	df, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("unable to create target file %q: %w", destPath, err)
	}
	defer df.Close()

	_, err = io.Copy(df, sf)
	if err != nil {
		return fmt.Errorf("unable to copy source file %q contents to target file %q: %w", source, destPath, err)
	}

	if err := df.Close(); err != nil {
		return fmt.Errorf("unable to close target file %q: %w", destPath, err)
	}
	if err := os.Chmod(destPath, fi.Mode()); err != nil {
		return fmt.Errorf("unable to change mode of target file %q: %w", destPath, err)
	}

	return nil
}

func symlinkFile(oldPath, targetDir string, force bool) error {
	klog.Infof("symlinking source file %q to target directory %q", oldPath, targetDir)

	newPath := filepath.Join(targetDir, filepath.Base(oldPath))
	if force {
		if err := os.Remove(newPath); err != nil {
			if os.IsNotExist(err) {
				// ignore
			} else {
				return fmt.Errorf("error removing file %q (for force): %w", newPath, err)
			}
		} else {
			klog.Infof("removed existing file %q (for force)", newPath)
		}
	}
	if err := os.Symlink(oldPath, newPath); err != nil {
		return fmt.Errorf("unable to create symlink from %q -> %q: %w", newPath, oldPath, err)
	}

	return nil
}

type stringSliceFlags []string

func (f *stringSliceFlags) String() string {
	return strings.Join(*f, ",")
}

func (f *stringSliceFlags) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func main() {
	// We force (overwrite existing files), so we can be idempotent in case of restart
	force := true

	var symlink bool
	flag.BoolVar(&symlink, "symlink", symlink, "make symbolic link")
	var targetDirs stringSliceFlags
	flag.Var(&targetDirs, "target-dir", "copy to directory")
	var sources stringSliceFlags
	flag.Var(&sources, "src", "source files to copy")

	flag.Parse()

	if len(sources) == 0 || len(targetDirs) == 0 || len(flag.Args()) != 0 {
		flag.Usage()
		os.Exit(1)
	}

	for _, targetDir := range targetDirs {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			klog.Exitf("unable to create target directory %q: %v", targetDir, err)
		}

		for _, src := range sources {
			if symlink {
				if err := symlinkFile(src, targetDir, force); err != nil {
					klog.Exitf("unable to copy source file %q to target directory %q: %v", src, targetDir, err)
				}
			} else {
				if err := copyFile(src, targetDir, force); err != nil {
					klog.Exitf("unable to copy source file %q to target directory %q: %v", src, targetDir, err)
				}
			}
		}
	}
}
