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
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/shlex"
	"golang.org/x/exp/slices"

	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/kubetest2-kops/aws"
	"k8s.io/kops/tests/e2e/kubetest2-kops/do"
	"k8s.io/kops/tests/e2e/kubetest2-kops/gce"
	"k8s.io/kops/tests/e2e/kubetest2-kops/scaleway"
	"k8s.io/kops/tests/e2e/pkg/kops"
	"k8s.io/kops/tests/e2e/pkg/util"
	"k8s.io/kops/tests/e2e/pkg/version"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

func (d *deployer) Up() error {
	if err := d.init(); err != nil {
		return err
	}

	// kops is fetched when --up is called instead of init to support a scenario where k/k is being built
	// and a kops build is not ready yet
	if d.KopsVersionMarker != "" {
		d.KopsBinaryPath = path.Join(d.commonOptions.RunDir(), "kops")
		baseURL, err := kops.DownloadKops(d.KopsVersionMarker, d.KopsBinaryPath)
		if err != nil {
			return fmt.Errorf("init failed to download kops from url: %v", err)
		}
		d.KopsBaseURL = baseURL
	}

	if d.terraform == nil {
		klog.Info("Cleaning up any leaked resources from previous cluster")
		// Intentionally ignore errors:
		// Either the cluster didn't exist or something failed that the next cluster creation will catch
		_ = d.Down()
	}

	if d.CloudProvider == "gce" && d.createBucket {
		if err := gce.EnsureGCSBucket(d.stateStore(), d.GCPProject, false); err != nil {
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

	time.Sleep(10 * time.Second)

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
		"--set", "cluster.spec.nodePortAccess=0.0.0.0/0",
	}

	version, err := kops.GetVersion(d.KopsBinaryPath)
	if err != nil {
		return err
	}
	if version > "1.29" {
		// Requires https://github.com/kubernetes/kops/pull/16128
		args = append(args, "--set", `spec.containerd.configAdditions=plugins."io.containerd.grpc.v1.cri".containerd.runtimes.test-handler.runtime_type=io.containerd.runc.v2`)
	}

	if yes {
		args = append(args, "--yes")
	}

	isArm := false
	if d.CreateArgs != "" {
		if strings.Contains(d.CreateArgs, "arm64") {
			isArm = true
		}
		createArgs, err := shlex.Split(d.CreateArgs)
		if err != nil {
			return err
		}
		args = append(args, createArgs...)
	}
	args = appendIfUnset(args, "--admin-access", adminAccess)

	// Dont set --master-count if either --control-plane-count or --master-count
	// has been provided in --create-args
	foundCPCount := false
	for _, existingArg := range args {
		existingKey := strings.Split(existingArg, "=")
		if existingKey[0] == "--control-plane-count" || existingKey[0] == "--master-count" {
			foundCPCount = true
			break
		}
	}
	if !foundCPCount {
		args = appendIfUnset(args, "--master-count", fmt.Sprintf("%d", d.ControlPlaneCount))
	}

	switch d.CloudProvider {
	case "aws":
		if isArm {
			args = appendIfUnset(args, "--master-size", "c6g.large")
			args = appendIfUnset(args, "--node-size", "c6g.large")
		} else {
			args = appendIfUnset(args, "--master-size", "c5.large")
		}
	case "gce":
		if isArm {
			args = appendIfUnset(args, "--master-size", "t2a-standard-2")
			args = appendIfUnset(args, "--node-size", "t2a-standard-2")
		} else {
			args = appendIfUnset(args, "--master-size", "e2-standard-2")
			args = appendIfUnset(args, "--node-size", "e2-standard-2")
		}
		if d.GCPProject != "" {
			args = appendIfUnset(args, "--project", d.GCPProject)
		}
		// set some sane default e2e testing behaviour on gce
		args = appendIfUnset(args, "--networking", "kubenet")
		args = appendIfUnset(args, "--node-volume-size", "100")

		// We used to set the --vpc flag to split clusters into different networks, this is now the default.
		// args = appendIfUnset(args, "--vpc", strings.Split(d.ClusterName, ".")[0])
	case "digitalocean":
		args = appendIfUnset(args, "--master-size", "c2-16vcpu-32gb")
		args = appendIfUnset(args, "--node-size", "c2-16vcpu-32gb")
	case "scaleway":
		args = appendIfUnset(args, "--master-size", "PRO2-S")
		args = appendIfUnset(args, "--node-size", "PRO2-S")
	}

	args = appendIfUnset(args, "--master-volume-size", "48")
	args = appendIfUnset(args, "--node-count", "4")
	args = appendIfUnset(args, "--node-volume-size", "48")
	args = appendIfUnset(args, "--zones", strings.Join(zones, ","))

	if d.terraform != nil {
		args = append(args, "--target", "terraform", "--out", d.terraform.Dir())
	}

	if d.KubernetesFeatureGates != "" {
		args = appendIfUnset(args, "--kubernetes-feature-gates", d.KubernetesFeatureGates)
	}

	klog.Info(strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)

	exec.InheritOutput(cmd)
	err = cmd.Run()
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
		// kOps is more likely to hit negative TTLs for API DNS during validation.
		wait = time.Duration(20) * time.Minute
	}
	args := []string{
		d.KopsBinaryPath, "validate", "cluster",
		"--name", d.ClusterName,
		"--count", strconv.Itoa(d.ValidationCount),
		"--wait", wait.String(),
	}
	if d.ValidationInterval > 10*time.Second {
		args = append(args, "--interval", d.ValidationInterval.String())
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
	if d.BuildOptions.BuildKubernetes {
		return nil
	}
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
		return aws.RandomZones(d.ControlPlaneCount)
	case "gce":
		return gce.RandomZones(1)
	case "digitalocean":
		return do.RandomZones(1)
	case "scaleway":
		return scaleway.RandomZones(1)
	}
	return nil, fmt.Errorf("unsupported CloudProvider: %v", d.CloudProvider)
}

// appendIfUnset will append an argument and its value to args if the arg is not already present
// This shouldn't be used for arguments that can be specified multiple times except --set
func appendIfUnset(args []string, arg, value string) []string {
	setFlags := []string{}
	for _, existingArg := range args {
		existingKey := strings.SplitN(existingArg, "=", 2)
		if existingKey[0] == "--set" {
			if len(existingKey) == 3 {
				setFlags = append(setFlags, existingKey[1])
			}
			if slices.Contains(setFlags, arg) {
				return args
			}
		} else if existingKey[0] == arg {
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
