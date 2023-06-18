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

	"k8s.io/klog/v2"
)

func copyFile(source, dest string) error {
	klog.Infof("Copying source file %q to %q", source, dest)

	sf, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("unable to open source file %q: %w", source, err)
	}
	defer sf.Close()

	fi, err := sf.Stat()
	if err != nil {
		return fmt.Errorf("unable to stat source file %q: %w", source, err)
	}

	df, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("unable to create dest file %q: %w", dest, err)
	}
	defer df.Close()

	_, err = io.Copy(df, sf)
	if err != nil {
		return fmt.Errorf("unable to copy source file %q contents to dest file %q: %w", source, dest, err)
	}

	if err := df.Close(); err != nil {
		return fmt.Errorf("unable to close dest file %q: %w", dest, err)
	}
	if err := os.Chmod(dest, fi.Mode()); err != nil {
		return fmt.Errorf("unable to change mode of dest file %q: %w", dest, err)
	}

	return nil
}

// main is the entrypoint, and performs some simple busybox-style all-in-one binary dispatching.
func main() {
	cmdName := filepath.Base(os.Args[0])

	var cmd func() error
	switch cmdName {
	case "ln", "kops-utils-ln":
		cmd = commandLn
	default:
		cmd = commandCp
	}

	if err := cmd(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// commandCp is a lightweight substitute for the `cp` command.
func commandCp() error {
	var symlink bool
	flag.BoolVar(&symlink, "s", symlink, "make symbolic link")
	var targetDirectory bool
	flag.BoolVar(&targetDirectory, "t", targetDirectory, "copy to directory")

	flag.Parse()

	args := flag.Args()

	if targetDirectory {
		if len(args) < 2 {
			return fmt.Errorf("usage: kops-utils-cp -t DIRECTORY SOURCE...")
		}

		if symlink {
			return fmt.Errorf("symlink not supported with directory copying")
		}

		targetDir := args[0]
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("unable to create target directory %q: %v", targetDir, err)
		}

		for _, src := range args[1:] {
			name := filepath.Base(src)
			target := filepath.Join(targetDir, name)
			if err := copyFile(src, target); err != nil {
				return fmt.Errorf("unable to copy source file %q to target directory %q: %w", src, target, err)
			}
		}
	} else {
		if len(args) != 2 {
			return fmt.Errorf("usage: kops-utils-cp [-s] SOURCE DEST")
		}

		src := args[0]
		dest := args[1]

		if symlink {
			if err := os.Symlink(src, dest); err != nil {
				return fmt.Errorf("unable to symlink %q -> %q: %w", dest, src, err)
			}
		} else {
			if err := copyFile(src, dest); err != nil {
				return fmt.Errorf("unable to copy source file %q to target %q: %w", src, dest, err)
			}
		}
	}

	return nil
}

// commandLn is a lightweight substitute for the `ln` command.
func commandLn() error {
	var symlink bool
	flag.BoolVar(&symlink, "s", symlink, "make symbolic link")

	flag.Parse()

	args := flag.Args()

	if len(args) != 2 {
		return fmt.Errorf("usage: kops-utils-ln [-s] SOURCE TARGET")
	}

	target := args[0]
	linkName := args[1]

	if symlink {
		if err := os.Symlink(target, linkName); err != nil {
			return fmt.Errorf("unable to create symlink from %q -> %q: %w", linkName, target, err)
		}
	} else {
		if err := os.Link(target, linkName); err != nil {
			return fmt.Errorf("unable to create hard link from %q -> %q: %w", linkName, target, err)
		}
	}

	return nil
}
