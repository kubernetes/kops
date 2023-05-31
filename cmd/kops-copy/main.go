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
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
)

func copyFile(source, target string) error {
	klog.Infof("Copying source file %q to target directory %q", source, target)

	sf, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("unable to open source file %q: %w", source, err)
	}
	defer sf.Close()

	fi, err := sf.Stat()
	if err != nil {
		return fmt.Errorf("unable to stat source file %q: %w", source, err)
	}

	fn := filepath.Join(target, filepath.Base(source))
	df, err := os.Create(fn)
	if err != nil {
		return fmt.Errorf("unable to create target file %q: %w", fn, err)
	}
	defer df.Close()

	_, err = io.Copy(df, sf)
	if err != nil {
		return fmt.Errorf("unable to copy source file %q contents to target file %q: %w", source, fn, err)
	}

	if err := df.Close(); err != nil {
		return fmt.Errorf("unable to close target file %q: %w", fn, err)
	}
	if err := os.Chmod(fn, fi.Mode()); err != nil {
		return fmt.Errorf("unable to change mode of target file %q: %w", fn, err)
	}

	return nil
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: kops-copy SOURCE ... TARGET")
	}

	target := os.Args[len(os.Args)-1]

	if err := os.MkdirAll(target, 0755); err != nil {
		klog.Exitf("unable to create target directory %q: %v", target, err)
	}

	for _, src := range os.Args[1 : len(os.Args)-1] {
		if err := copyFile(src, target); err != nil {
			klog.Exitf("unable to copy source file %q to target directory %q: %v", src, target, err)
		}
	}
}
