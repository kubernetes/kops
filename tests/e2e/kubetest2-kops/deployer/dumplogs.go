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
	"k8s.io/kops/pkg/resources"
	"sigs.k8s.io/kubetest2/pkg/exec"
	"sigs.k8s.io/yaml"
)

func (d *deployer) DumpClusterLogs() error {
	yamlFile, err := os.Create(path.Join(d.ArtifactsDir, "toolbox-dump.yaml"))
	if err != nil {
		panic(err)
	}
	defer yamlFile.Close()

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
	cmd.SetStdout(yamlFile)
	if err := cmd.Run(); err != nil {
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
		if err = d.dumpClusterInfoSSH(); err != nil {
			return err
		}
	}

	resourceTypes := []string{
		"csinodes", "csidrivers", "storageclasses", "persistentvolumes",
		"mutatingwebhookconfigurations", "validatingwebhookconfigurations",
	}
	if err := os.MkdirAll(path.Join(d.ArtifactsDir, "cluster-info"), 0o755); err != nil {
		return err
	}
	for _, resType := range resourceTypes {
		yamlFile, err := os.Create(path.Join(d.ArtifactsDir, "cluster-info", fmt.Sprintf("%v.yaml", resType)))
		if err != nil {
			return err
		}
		defer yamlFile.Close()

		args = []string{
			"kubectl", "--request-timeout", "5s", "get", resType,
			"--all-namespaces",
			"--show-managed-fields",
			"-o", "yaml",
		}
		klog.Info(strings.Join(args, " "))

		cmd := exec.Command(args[0], args[1:]...)
		cmd.SetEnv(d.env()...)
		cmd.SetStdout(yamlFile)
		if err := cmd.Run(); err != nil {
			klog.Warningf("Failed to get %v: %v", resType, err)
		}
	}

	nsCmd := exec.Command(
		"kubectl", "--request-timeout", "5s", "get", "namespaces", "--no-headers", "-o", "custom-columns=name:.metadata.name",
	)
	namespaces, err := exec.OutputLines(nsCmd)
	if err != nil {
		return fmt.Errorf("failed to get namespaces: %s", err)
	}

	namespacedResourceTypes := []string{
		"configmaps",
		"endpoints",
		"endpointslices",
		"leases",
		"persistentvolumeclaims",
		"poddisruptionbudgets",
	}
	for _, namespace := range namespaces {
		namespace = strings.TrimSpace(namespace)
		if err := os.MkdirAll(path.Join(d.ArtifactsDir, "cluster-info", namespace), 0o755); err != nil {
			return err
		}
		for _, resType := range namespacedResourceTypes {
			yamlFile, err := os.Create(path.Join(d.ArtifactsDir, "cluster-info", namespace, fmt.Sprintf("%v.yaml", resType)))
			if err != nil {
				return err
			}
			defer yamlFile.Close()

			args = []string{
				"kubectl", "get", resType,
				"-n", namespace,
				"--show-managed-fields",
				"-o", "yaml",
			}
			klog.Info(strings.Join(args, " "))

			cmd := exec.Command(args[0], args[1:]...)
			cmd.SetEnv(d.env()...)
			cmd.SetStdout(yamlFile)
			if err := cmd.Run(); err != nil {
				if err = d.dumpClusterInfoSSH(); err != nil {
					klog.Warningf("Failed to get %v: %v", resType, err)
				}
			}
		}
	}
	return nil
}

// dumpClusterInfoSSH runs `kubectl cluster-info dump` on a control plane host via SSH
// and copies the output to the local artifacts directory.
// This can be useful when the k8s API is inaccessible from kubetest2-kops directly
func (d *deployer) dumpClusterInfoSSH() error {
	toolboxDumpArgs := []string{
		d.KopsBinaryPath, "toolbox", "dump",
		"--name", d.ClusterName,
		"--private-key", d.SSHPrivateKeyPath,
		"--ssh-user", d.SSHUser,
		"-o", "yaml",
	}
	klog.Info(strings.Join(toolboxDumpArgs, " "))

	cmd := exec.Command(toolboxDumpArgs[0], toolboxDumpArgs[1:]...)
	dumpOutput, err := exec.Output(cmd)
	if err != nil {
		return err
	}

	var dump resources.Dump
	err = yaml.Unmarshal(dumpOutput, &dump)
	if err != nil {
		return err
	}
	controlPlaneIP, controlPlaneUser, found := findControlPlaneIPUser(dump)
	if !found {
		return nil
	}

	sshURL := fmt.Sprintf("%v@%v", controlPlaneUser, controlPlaneIP)
	sshArgs := []string{
		"ssh", "-i", d.SSHPrivateKeyPath,
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		sshURL, "--",
		"kubectl", "cluster-info", "dump",
		"--all-namespaces",
		"-o", "yaml",
		"--output-directory", "/tmp/cluster-info",
	}
	klog.Info(strings.Join(sshArgs, " "))

	cmd = exec.Command(sshArgs[0], sshArgs[1:]...)
	exec.InheritOutput(cmd)

	if err := cmd.Run(); err != nil {
		return err
	}
	scpArgs := []string{
		"scp", "-i", d.SSHPrivateKeyPath,
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null", "-r",
		fmt.Sprintf("%v:/tmp/cluster-info", sshURL),
		path.Join(d.ArtifactsDir, "cluster-info"),
	}
	klog.Info(strings.Join(scpArgs, " "))

	cmd = exec.Command(scpArgs[0], scpArgs[1:]...)
	exec.InheritOutput(cmd)

	if err := cmd.Run(); err != nil {
		return err
	}

	rmArgs := []string{
		"ssh", "-i", d.SSHPrivateKeyPath,
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		sshURL, "--",
		"rm", "-rf", "/tmp/cluster-info",
	}
	klog.Info(strings.Join(rmArgs, " "))

	cmd = exec.Command(rmArgs[0], rmArgs[1:]...)
	exec.InheritOutput(cmd)

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func findControlPlaneIPUser(dump resources.Dump) (string, string, bool) {
	for _, instance := range dump.Instances {
		if len(instance.PublicAddresses) == 0 {
			continue
		}
		for _, role := range instance.Roles {
			if role == "control-plane" {
				return instance.PublicAddresses[0], instance.SSHUser, true
			}
		}
	}
	klog.Warning("ControlPlane instance not found from kops toolbox dump")
	return "", "", false
}

func runWithOutput(cmd exec.Cmd) error {
	exec.InheritOutput(cmd)
	return cmd.Run()
}
