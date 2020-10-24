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
	"os"
	osexec "os/exec"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/kubetest2-kops/aws"
	"k8s.io/kops/tests/e2e/kubetest2-kops/util"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

func (d *deployer) Up() error {
	if err := d.init(); err != nil {
		return err
	}

	publicIP, err := util.ExternalIPRange()
	if err != nil {
		return err
	}

	zones, err := aws.RandomZones(1)
	if err != nil {
		return err
	}

	args := []string{
		d.KopsBinaryPath, "create", "cluster",
		"--name", d.ClusterName,
		"--admin-access", publicIP,
		"--cloud", d.CloudProvider,
		"--master-count", "1",
		"--master-size", "c5.large",
		"--master-volume-size", "48",
		"--node-count", "4",
		"--node-volume-size", "48",
		"--override", "cluster.spec.nodePortAccess=0.0.0.0/0",
		"--ssh-public-key", d.SSHPublicKeyPath,
		"--zones", strings.Join(zones, ","),
		"--yes",
	}
	klog.Info(strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)

	exec.InheritOutput(cmd)
	return cmd.Run()
}

func (d *deployer) IsUp() (bool, error) {
	args := []string{
		d.KopsBinaryPath, "validate", "cluster",
		"--name", d.ClusterName,
		"--wait", "15m",
	}
	klog.Info(strings.Join(args, " "))

	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)

	exec.InheritOutput(cmd)
	err := cmd.Run()
	// `kops validate cluster` exits 2 if validation failed
	if exitErr, ok := err.(*osexec.ExitError); ok && exitErr.ExitCode() == 2 {
		return false, nil
	}
	return err == nil, err
}

// verifyUpFlags ensures fields are set for creation of the cluster
func (d *deployer) verifyUpFlags() error {
	// These environment variables are defined by the "preset-aws-ssh" prow preset
	// https://github.com/kubernetes/test-infra/blob/3d3b325c98b739b526ba5d93ce21c90a05e1f46d/config/prow/config.yaml#L653-L670
	if d.SSHPrivateKeyPath == "" {
		d.SSHPrivateKeyPath = os.Getenv("AWS_SSH_PRIVATE_KEY_FILE")
	}
	if d.SSHPublicKeyPath == "" {
		d.SSHPublicKeyPath = os.Getenv("AWS_SSH_PUBLIC_KEY_FILE")
	}

	return nil
}
