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

package try

import (
	"os"

	"k8s.io/klog/v2"
)

// RemoveFile will try to os.Remove the file, logging an error if it fails
func RemoveFile(fp string) {
	if err := os.Remove(fp); err != nil {
		klog.Warningf("unable to remove file %s: %v", fp, err)
	}
}

// CloseFile will try to call close on the file, logging an error if it fails
func CloseFile(f *os.File) {
	if err := f.Close(); err != nil {
		klog.Warningf("unable to close file %s: %v", f.Name(), err)
	}
}
