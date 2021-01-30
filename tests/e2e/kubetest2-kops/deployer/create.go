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

package deployer

import (
	"strings"

	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

// create performs a `kops create -f` followed by `kops update cluster --yes`
func (d *deployer) replace() error {
	args := []string{
		d.KopsBinaryPath, "create",
		"--filename", d.manifestPath,
		"--name", d.ClusterName,
	}
	klog.Info(strings.Join(args, " "))

	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)

	exec.InheritOutput(cmd)
	err := cmd.Run()
	if err != nil {
		return err
	}

	args = []string{
		d.KopsBinaryPath, "update", "cluster", "--yes",
		"--admin",
		"--name", d.ClusterName,
	}
	if d.terraform != nil {
		args = append(args, "--target", "terraform", "--out", d.terraform.Dir())
	}

	klog.Info(strings.Join(args, " "))

	cmd = exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)

	exec.InheritOutput(cmd)
	err = cmd.Run()
	if err != nil {
		return err
	}
	if d.terraform != nil {
		if err := d.terraform.InitApply(); err != nil {
			return err
		}
	}
	return nil
}
