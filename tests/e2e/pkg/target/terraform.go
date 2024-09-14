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

package target

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/pkg/util"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

// Terraform represents a set of terraform commands to be ran against a directory
// containing a kops cluster's .tf output
type Terraform struct {
	dir           string
	terraformPath string
	artifactsDir  string
}

// NewTerraform creates a new Terraform object
func NewTerraform(version, artifactsDir string) (*Terraform, error) {
	var b bytes.Buffer
	url := fmt.Sprintf("https://releases.hashicorp.com/terraform/%v/terraform_%v_%v_%v.zip", version, version, runtime.GOOS, runtime.GOARCH)

	if err := util.HTTPGETWithHeaders(url, nil, &b); err != nil {
		return nil, err
	}
	binaryDir, err := util.UnzipToTempDir(b.Bytes())
	if err != nil {
		return nil, err
	}
	tfDir, err := os.MkdirTemp("", "kops-terraform")
	if err != nil {
		return nil, err
	}
	artifacts := filepath.Join(artifactsDir, "terraform")
	if err := os.MkdirAll(artifacts, 0644); err != nil {
		return nil, err
	}

	t := &Terraform{
		dir:           tfDir,
		terraformPath: filepath.Join(binaryDir, "terraform"),
		artifactsDir:  artifacts,
	}
	return t, nil
}

// Dir returns the directory containing the kops-generated .tf files
func (t *Terraform) Dir() string {
	return t.dir
}

// InitApply runs `terraform init` and `terraform apply` in the specified directory
func (t *Terraform) InitApply() error {
	args := []string{
		t.terraformPath, "init",
	}
	klog.Info(strings.Join(args, " "))

	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetDir(t.dir)

	exec.InheritOutput(cmd)
	err := cmd.Run()
	if err != nil {
		return err
	}

	err = t.Backup()
	if err != nil {
		return err
	}

	args = []string{
		t.terraformPath, "apply",
		"-auto-approve",
	}
	klog.Info(strings.Join(args, " "))

	cmd = exec.Command(args[0], args[1:]...)
	cmd.SetDir(t.dir)

	exec.InheritOutput(cmd)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// Destroy runs `terraform destroy` in the specified directory
func (t *Terraform) Destroy() error {
	args := []string{
		t.terraformPath, "destroy",
		"-auto-approve",
	}
	klog.Info(strings.Join(args, " "))

	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetDir(t.dir)

	exec.InheritOutput(cmd)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (t *Terraform) Backup() error {
	if t.artifactsDir == "" {
		return nil
	}

	files := []string{
		"kubernetes.tf",
		".terraform.lock.hcl",
	}
	for _, f := range files {
		klog.Infof("Copying %s to artifacts", f)
		contents, err := os.ReadFile(path.Join(t.Dir(), f))
		if err != nil {
			return fmt.Errorf("failed to read %s: %v", f, err)
		}
		err = os.WriteFile(path.Join(t.artifactsDir, f), contents, 0644)
		if err != nil {
			return fmt.Errorf("failed to write %s: %v", f, err)
		}
	}
	return nil
}
