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
	"errors"
	"fmt"
	osexec "os/exec"
	"strings"

	"github.com/google/shlex"

	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/kubetest2-kops/aws"
	"k8s.io/kops/tests/e2e/kubetest2-kops/do"
	"k8s.io/kops/tests/e2e/kubetest2-kops/gce"
	"k8s.io/kops/tests/e2e/pkg/util"
	"k8s.io/kops/tests/e2e/pkg/version"
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

	adminAccess := d.AdminAccess
	if adminAccess == "" {
		adminAccess = publicIP
	}

	zones, err := d.zones()
	if err != nil {
		return err
	}

	if d.TemplatePath != "" {
		values, err := d.templateValues(zones, adminAccess)
		if err != nil {
			return err
		}
		if err := d.renderTemplate(values); err != nil {
			return err
		}
		if err := d.replace(); err != nil {
			return err
		}
	} else {
		err := d.createCluster(zones, adminAccess)
		if err != nil {
			return err
		}
	}
	isUp, err := d.IsUp()
	if err != nil {
		return err
	} else if isUp {
		klog.V(1).Infof("cluster reported as up")
	} else {
		klog.Errorf("cluster reported as down")
	}
	return nil
}

func (d *deployer) createCluster(zones []string, adminAccess string) error {

	args := []string{
		d.KopsBinaryPath, "create", "cluster",
		"--name", d.ClusterName,
		"--cloud", d.CloudProvider,
		"--kubernetes-version", d.KubernetesVersion,
		"--ssh-public-key", d.SSHPublicKeyPath,
		"--override", "cluster.spec.nodePortAccess=0.0.0.0/0",
		"--yes",
	}

	if d.CreateArgs != "" {
		createArgs, err := shlex.Split(d.CreateArgs)
		if err != nil {
			return err
		}
		args = append(args, createArgs...)
	}
	args = appendIfUnset(args, "--admin-access", adminAccess)
	args = appendIfUnset(args, "--master-count", "1")
	args = appendIfUnset(args, "--master-volume-size", "48")
	args = appendIfUnset(args, "--node-count", "4")
	args = appendIfUnset(args, "--node-volume-size", "48")
	args = appendIfUnset(args, "--override", adminAccess)
	args = appendIfUnset(args, "--admin-access", adminAccess)
	args = appendIfUnset(args, "--zones", strings.Join(zones, ","))

	switch d.CloudProvider {
	case "aws":
		args = appendIfUnset(args, "--master-size", "c5.large")
	case "gce":
		args = appendIfUnset(args, "--master-size", "e2-standard-2")
	case "digitalocean":
		args = appendIfUnset(args, "--master-size", "s-8vcpu-16gb")
	}

	if d.terraform != nil {
		args = append(args, "--target", "terraform", "--out", d.terraform.Dir())
	}

	klog.Info(strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)

	exec.InheritOutput(cmd)
	err := cmd.Run()
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

func (d *deployer) IsUp() (bool, error) {
	args := []string{
		d.KopsBinaryPath, "validate", "cluster",
		"--name", d.ClusterName,
		"--count", "10",
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
	if d.KubernetesVersion == "" {
		return errors.New("missing required --kubernetes-version flag")
	}

	v, err := version.ParseKubernetesVersion(d.KubernetesVersion)
	if err != nil {
		return err
	}
	d.KubernetesVersion = v

	return nil
}

func (d *deployer) zones() ([]string, error) {
	switch d.CloudProvider {
	case "aws":
		return aws.RandomZones(1)
	case "gce":
		return gce.RandomZones(1)
	case "digitalocean":
		return do.RandomZones(1)
	}
	return nil, fmt.Errorf("unsupported CloudProvider: %v", d.CloudProvider)
}

// appendIfUnset will append an argument and its value to args if the arg is not already present
// This shouldn't be used for arguments that can be specified multiple times like --override
func appendIfUnset(args []string, arg, value string) []string {
	for _, existingArg := range args {
		existingKey := strings.Split(existingArg, "=")
		if existingKey[0] == arg {
			return args
		}
	}
	args = append(args, arg, value)
	return args
}
