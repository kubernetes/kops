/*
Copyright 2020 The Kubernetes Authors.

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
	"fmt"
	"os"
	"path"
	"strings"

	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

func (d *deployer) DumpClusterLogs() error {
	args := []string{
		d.KopsBinaryPath, "toolbox", "dump",
		"--name", d.ClusterName,
		"--dir", d.ArtifactsDir,
		"--private-key", d.SSHPrivateKeyPath,
		"--ssh-user", d.SSHUser,
	}
	klog.Info(strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)
	if err := runWithOutput(cmd); err != nil {
		return err
	}

	if err := d.dumpClusterManifest(); err != nil {
		return err
	}

	if err := d.dumpClusterInfo(); err != nil {
		return err
	}

	return nil
}

func (d *deployer) dumpClusterManifest() error {
	resourceTypes := []string{"cluster", "instancegroups"}
	for _, rt := range resourceTypes {
		yamlFile, err := os.Create(path.Join(d.ArtifactsDir, fmt.Sprintf("%v.yaml", rt)))
		if err != nil {
			panic(err)
		}
		defer yamlFile.Close()

		args := []string{
			d.KopsBinaryPath, "get", rt,
			"--name", d.ClusterName,
			"-o", "yaml",
		}
		klog.Info(strings.Join(args, " "))

		cmd := exec.Command(args[0], args[1:]...)
		cmd.SetStdout(yamlFile)
		cmd.SetEnv(d.env()...)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (d *deployer) dumpClusterInfo() error {
	args := []string{
		"kubectl", "cluster-info", "dump",
		"--all-namespaces",
		"-o", "yaml",
		"--output-directory", path.Join(d.ArtifactsDir, "cluster-info"),
	}
	klog.Info(strings.Join(args, " "))

	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func runWithOutput(cmd exec.Cmd) error {
	exec.InheritOutput(cmd)
	return cmd.Run()
}
