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
	"time"

	"github.com/google/shlex"

	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/kubetest2-kops/aws"
	"k8s.io/kops/tests/e2e/kubetest2-kops/do"
	"k8s.io/kops/tests/e2e/kubetest2-kops/gce"
	"k8s.io/kops/tests/e2e/pkg/kops"
	"k8s.io/kops/tests/e2e/pkg/util"
	"k8s.io/kops/tests/e2e/pkg/version"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

func (d *deployer) Up() error {
	if err := d.init(); err != nil {
		return err
	}

	if d.terraform == nil {
		klog.Info("Cleaning up any leaked resources from previous cluster")
		// Intentionally ignore errors:
		// Either the cluster didn't exist or something failed that the next cluster creation will catch
		_ = d.Down()
	}

	if d.CloudProvider == "gce" && d.createBucket {
		if err := gce.EnsureGCSBucket(d.stateStore(), d.GCPProject); err != nil {
			return err
		}
	}

	adminAccess := d.AdminAccess
	if adminAccess == "" {
		publicIP, err := util.ExternalIPRange()
		if err != nil {
			return err
		}

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
		if d.terraform != nil {
			if err := d.createCluster(zones, adminAccess, true); err != nil {
				return err
			}
		} else {
			// For the non-terraform case, we want to see the preview output.
			// So run a create (which logs the output), then do an update
			if err := d.createCluster(zones, adminAccess, false); err != nil {
				return err
			}
			if err := d.updateCluster(true); err != nil {
				return err
			}
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

func (d *deployer) createCluster(zones []string, adminAccess string, yes bool) error {
	args := []string{
		d.KopsBinaryPath, "create", "cluster",
		"--name", d.ClusterName,
		"--cloud", d.CloudProvider,
		"--kubernetes-version", d.KubernetesVersion,
		"--ssh-public-key", d.SSHPublicKeyPath,
		"--override", "cluster.spec.nodePortAccess=0.0.0.0/0",
	}
	if yes {
		args = append(args, "--yes")
	}

	if d.CreateArgs != "" {
		createArgs, err := shlex.Split(d.CreateArgs)
		if err != nil {
			return err
		}
		args = append(args, createArgs...)
	}
	args = appendIfUnset(args, "--admin-access", adminAccess)
	args = appendIfUnset(args, "--master-count", fmt.Sprintf("%d", d.ControlPlaneSize))
	args = appendIfUnset(args, "--master-volume-size", "48")
	args = appendIfUnset(args, "--node-count", "4")
	args = appendIfUnset(args, "--node-volume-size", "48")
	args = appendIfUnset(args, "--override", adminAccess)
	args = appendIfUnset(args, "--zones", strings.Join(zones, ","))

	switch d.CloudProvider {
	case "aws":
		args = appendIfUnset(args, "--master-size", "c5.large")
	case "gce":
		args = appendIfUnset(args, "--master-size", "e2-standard-2")
		if d.GCPProject != "" {
			args = appendIfUnset(args, "--project", d.GCPProject)
		}
		// We used to set the --vpc flag to split clusters into different networks, this is now the default.
		// args = appendIfUnset(args, "--vpc", strings.Split(d.ClusterName, ".")[0])
	case "digitalocean":
		args = appendIfUnset(args, "--master-size", "c2-16vcpu-32gb")
		args = appendIfUnset(args, "--node-size", "c2-16vcpu-32gb")
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

	if d.setInstanceGroupOverrides(); err != nil {
		return err
	}

	if d.terraform != nil {
		if err := d.terraform.InitApply(); err != nil {
			return err
		}
	}

	return nil
}

func (d *deployer) setInstanceGroupOverrides() error {
	igs, err := kops.GetInstanceGroups(d.KopsBinaryPath, d.ClusterName, d.env())
	if err != nil {
		return err
	}
	for _, ig := range igs {
		if string(ig.Spec.Role) == "Master" && len(d.ControlPlaneIGOverrides) > 0 {
			if err := d.setIGOverrides(ig.ObjectMeta.Name, d.ControlPlaneIGOverrides); err != nil {
				return err
			}
		}
		if string(ig.Spec.Role) == "Node" && len(d.NodeIGOverrides) > 0 {
			if err := d.setIGOverrides(ig.ObjectMeta.Name, d.NodeIGOverrides); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *deployer) updateCluster(yes bool) error {
	args := []string{
		d.KopsBinaryPath, "update", "cluster",
		"--name", d.ClusterName,
		"--admin",
	}
	if yes {
		args = append(args, "--yes")
	}

	klog.Info(strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)

	exec.InheritOutput(cmd)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (d *deployer) IsUp() (bool, error) {
	wait := d.ValidationWait
	if wait == 0 {
		if d.TerraformVersion != "" || d.CloudProvider == "digitalocean" {
			// `--target terraform` doesn't precreate the API DNS records,
			// so kops is more likely to hit negative TTLs during validation.
			// Digital Ocean also occasionally takes longer to validate.
			wait = time.Duration(20) * time.Minute
		} else {
			wait = time.Duration(15) * time.Minute
		}
	}
	args := []string{
		d.KopsBinaryPath, "validate", "cluster",
		"--name", d.ClusterName,
		"--count", "10",
		"--wait", wait.String(),
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
	if err == nil && d.TerraformVersion != "" && d.commonOptions.ShouldTest() {
		klog.Info("Waiting 5 minutes for DNS TTLs before starting tests")
		time.Sleep(5 * time.Minute)
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
		return aws.RandomZones(d.ControlPlaneSize)
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

func (d *deployer) setIGOverrides(igName string, overrides []string) error {
	args := []string{
		d.KopsBinaryPath, "edit", "instancegroup",
		"--name", d.ClusterName,
		igName,
	}
	for _, override := range overrides {
		args = append(args, "--set", override)
	}
	klog.Info(strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)

	exec.InheritOutput(cmd)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
