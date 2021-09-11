/*
Copyright 2021 The Kubernetes Authors.

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
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
)

type AptSource struct {
	Name    string
	Keyring string
	Sources []string
}

func (e *AptSource) Find(c *fi.Context) (*AptSource, error) {
	return nil, nil
}

func (f *AptSource) GetName() *string {
	return &f.Name
}

func (f *AptSource) String() string {
	return f.Name
}

func (f *AptSource) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(f, c)
}

func (*AptSource) CheckChanges(a, e, changes *AptSource) error {
	return nil
}

func (f *AptSource) RenderLocal(t *local.LocalTarget, a, e, changes *AptSource) error {

	tmpDir, err := ioutil.TempDir("", "aptsource")
	if err != nil {
		return fmt.Errorf("error creating temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			klog.Warningf("error deleting temp dir %q: %v", tmpDir, err)
		}
	}()
	filename := path.Join(tmpDir, f.Name+".gpg")

	if _, err := fi.DownloadURL(f.Keyring, filename, nil); err != nil {
		return err
	}

	args := []string{"apt-key", "add", filename}

	klog.Infof("running command %s", args)
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if exitCode := cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus(); err != nil && exitCode != 100 {
		return fmt.Errorf("error adding key: %v: %s", err, string(output))
	}

	debs := strings.Join(f.Sources, "\n")

	if err := os.WriteFile("/etc/apt/sources.list.d/"+f.Name+".list", []byte(debs), 0); err != nil {
		return err
	}

	return nil
}
