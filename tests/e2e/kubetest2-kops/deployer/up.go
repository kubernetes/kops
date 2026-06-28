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
	"context"
	"errors"
	"fmt"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/shlex"
	"golang.org/x/exp/slices"

	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/kubetest2-kops/aws"
	"k8s.io/kops/tests/e2e/kubetest2-kops/azure"
	"k8s.io/kops/tests/e2e/kubetest2-kops/do"
	"k8s.io/kops/tests/e2e/kubetest2-kops/gce"
	"k8s.io/kops/tests/e2e/pkg/kops"
	"k8s.io/kops/tests/e2e/pkg/util"
	"k8s.io/kops/tests/e2e/pkg/version"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

// Default AWS instance types that createCluster injects when the caller did
// not pin --control-plane-size / --node-size. awsInstanceTypes mirrors this
// for zone selection, so update both call sites by editing these constants.
const (
	awsDefaultControlPlaneSize    = "c5.large"
	awsDefaultArmControlPlaneSize = "c6g.large"
	awsDefaultArmNodeSize         = "c6g.large"
)

func (d *deployer) Up() error {
	ctx := context.TODO()

	if err := d.init(); err != nil {
		return err
	}

	// kops is fetched when --up is called instead of init to support a scenario where k/k is being built
	// and a kops build is not ready yet
	if d.KopsVersionMarker != "" || d.KopsVersion != "" {
		d.KopsBinaryPath = d.resolvedKopsBinaryPath()
		baseURL, err := kops.DownloadKops(d.KopsVersionMarker, d.KopsBinaryPath, d.KopsVersion)
		if err != nil {
			return fmt.Errorf("init failed to download kops from url: %v", err)
		}
		d.KopsBaseURL = baseURL
	}

	// PreTestCmd inherits the process environment rather than d.env().
	if d.KopsBaseURL != "" {
		os.Setenv("KOPS_BASE_URL", d.KopsBaseURL)
	}

	if d.terraform == nil {
		klog.Info("Cleaning up any leaked resources from previous cluster")
		// Intentionally ignore errors:
		// Either the cluster didn't exist or something failed that the next cluster creation will catch
		_ = d.Down()
	}

	switch d.CloudProvider {
	case "aws":
		ctx := context.Background()
		if d.createStateStore {
			if err := d.aws.EnsureS3Bucket(ctx, d.region, d.stateStore(), false); err != nil {
				return err
			}
		}
		if d.createDiscoveryStore {
			if err := d.aws.EnsureS3Bucket(ctx, d.region, d.discoveryStore(), true); err != nil {
				return err
			}
		}
	case "gce":
		if d.createStateStore {
			if err := gce.EnsureGCSBucket(d.stateStore(), d.region, d.GCPProject, false); err != nil {
				return err
			}
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

	// Write out the env file for kops
	if err := d.writeEnvFile(ctx); err != nil {
		return fmt.Errorf("error writing env file %q: %v", d.EnvFile, err)
	}

	if d.TemplatePath != "" {
		values, err := d.templateValues(d.zones, adminAccess)
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
			if err := d.createCluster(d.zones, adminAccess, true); err != nil {
				return err
			}
		} else {
			// For the non-terraform case, we want to see the preview output.
			// So run a create (which logs the output), then do an update
			if err := d.createCluster(d.zones, adminAccess, false); err != nil {
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

// writeEnvFile writes out the env file (if EnvFile is specified)
// This allows us to dynamically generate KOPS_STATE_STORE, but still call kops commands
func (d *deployer) writeEnvFile(ctx context.Context) error {
	log := klog.FromContext(ctx)

	if d.EnvFile == "" {
		log.V(2).Info("no env file specified, skipping write of env file")
		return nil
	}

	log.V(2).Info("writing env file", "path", d.EnvFile)
	env := d.env()

	// Also export cluster name, in case we generated it
	if d.ClusterName != "" {
		env = append(env, fmt.Sprintf("CLUSTER_NAME=%v", d.ClusterName))
	}

	data := strings.Join(env, "\n") + "\n"
	if err := os.WriteFile(d.EnvFile, []byte(data), 0o644); err != nil {
		return fmt.Errorf("error writing env file %q: %v", d.EnvFile, err)
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
		// Register a test-handler runtime under both containerd config schemas so that
		// RuntimeClass e2e tests work across the matrix: kops < 1.36 (always emits v2)
		// and kops 1.36+ with k8s < 1.32 (emits v2) need the v2 path; kops 1.36+ with
		// k8s >= 1.32 (emits v3) needs the v3 path. containerd reads only the plugin
		// namespace matching its config version, so the other entry is inert.
		// TODO(rifelpet): drop the v2 path once kops < 1.36 and k8s < 1.32 are no longer tested.
		// containerd config schema v2 (containerd < 2.0):
		"--set", `spec.containerd.configAdditions=plugins."io.containerd.grpc.v1.cri".containerd.runtimes.test-handler.runtime_type=io.containerd.runc.v2`,
		// containerd config schema v3 (containerd >= 2.0):
		"--set", `spec.containerd.configAdditions=plugins."io.containerd.cri.v1.runtime".containerd.runtimes.test-handler.runtime_type=io.containerd.runc.v2`,
	}

	if d.discoveryStore() != "" {
		args = append(args, "--discovery-store", d.discoveryStore())
	}

	if yes {
		args = append(args, "--yes")
	}

	tags := []string{
		"group=sig-cluster-lifecycle",
		"subproject=kops",
	}
	if d.CloudProvider == "azure" {
		// Ensure https://github.com/Azure/rg-cleanup deletes removes resources
		tags = append(tags, "creationTimestamp="+time.Now().Format(time.RFC3339))
	}
	if label, ok := prowJobLabel(d.CloudProvider, os.Getenv("JOB_NAME")); ok {
		tags = append(tags, label)
	}
	args = appendIfUnset(args, "--cloud-labels", strings.Join(tags, ","))

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
	// Use the PR's own channels only when kops was built from the checkout. Downloaded
	// --kops-version[-marker] binaries can land in <cwd>/_rundir under KopsRoot, where path
	// containment alone wouldn't exclude them.
	if d.KopsVersionMarker == "" && d.KopsVersion == "" && builtFromKopsRoot(d.KopsBinaryPath, d.KopsRoot) {
		args = localChannelArgs(args, d.KopsRoot)
	}
	args = appendIfUnset(args, "--admin-access", adminAccess)

	// Dont set --control-plane-count if either --control-plane-count or --master-count
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
		args = appendIfUnset(args, "--control-plane-count", fmt.Sprintf("%d", d.ControlPlaneCount))
	}

	switch d.CloudProvider {
	case "aws":
		if isArm {
			args = appendIfUnset(args, "--control-plane-size", awsDefaultArmControlPlaneSize)
			args = appendIfUnset(args, "--node-size", awsDefaultArmNodeSize)
		} else {
			args = appendIfUnset(args, "--control-plane-size", awsDefaultControlPlaneSize)
		}
	case "azure":
		// Use SKUs for which there is enough quota
		args = appendIfUnset(args, "--control-plane-size", "Standard_D4ls_v6")
		args = appendIfUnset(args, "--node-size", "Standard_D4ls_v6")
	case "gce":
		if isArm {
			args = appendIfUnset(args, "--control-plane-size", "n4a-standard-2")
			args = appendIfUnset(args, "--node-size", "n4a-standard-2")
		} else {
			args = appendIfUnset(args, "--control-plane-size", "e2-standard-2")
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
		args = appendIfUnset(args, "--control-plane-size", "c2-16vcpu-32gb")
		args = appendIfUnset(args, "--node-size", "c2-16vcpu-32gb")
	}

	args = appendIfUnset(args, "--control-plane-volume-size", "48")
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
	err := cmd.Run()
	if err != nil {
		return err
	}

	if err = d.setInstanceGroupOverrides(); err != nil {
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

func extractZones(args string) []string {
	// Zones are specified by --zones=zone1,zone2
	prefix := "--zones="
	startIdx := strings.Index(args, prefix)

	if startIdx == -1 {
		return []string{}
	}

	startIdx += len(prefix)

	endIdx := strings.Index(args[startIdx:], " ")
	var zonesValue string

	if endIdx == -1 {
		zonesValue = args[startIdx:]
	} else {
		zonesValue = args[startIdx : startIdx+endIdx]
	}
	return strings.Split(zonesValue, ",")
}

func (d *deployer) getZones() ([]string, error) {
	if d.CreateArgs != "" {
		zones := extractZones(d.CreateArgs)
		if len(zones) > 0 {
			return zones, nil
		}
	}
	switch d.CloudProvider {
	case "aws":
		return aws.RandomZones(d.ControlPlaneCount, d.awsInstanceTypes())
	case "azure":
		return azure.RandomZones(1)
	case "gce":
		return gce.RandomZones(1)
	case "digitalocean":
		return do.RandomZones(1)
	}
	return nil, fmt.Errorf("unsupported CloudProvider: %v", d.CloudProvider)
}

// awsInstanceTypes returns the AWS instance types that the cluster will use,
// so that zone selection can be restricted to AZs offering them. It collects
// the types from any --control-plane-size / --master-size / --node-size flags
// in CreateArgs and falls back to the same defaults createCluster would inject
// when those flags are absent. When a template is in use the instance types
// live in the template (not in CreateArgs), so zone filtering only covers the
// defaults and may pick zones incompatible with template-pinned types.
func (d *deployer) awsInstanceTypes() []string {
	if d.TemplatePath != "" {
		klog.V(2).Infof("template %q in use; AWS zone selection will only filter on default instance types, not types pinned inside the template", d.TemplatePath)
	}
	seen := make(map[string]bool)
	var types []string
	add := func(values ...string) {
		for _, v := range values {
			for _, t := range strings.Split(v, ",") {
				t = strings.TrimSpace(t)
				if t == "" || seen[t] {
					continue
				}
				seen[t] = true
				types = append(types, t)
			}
		}
	}

	cpSizes := extractFlagValues(d.CreateArgs, "--control-plane-size")
	cpSizes = append(cpSizes, extractFlagValues(d.CreateArgs, "--master-size")...)
	nodeSizes := extractFlagValues(d.CreateArgs, "--node-size")
	add(cpSizes...)
	add(nodeSizes...)

	// Mirror the defaults in createCluster when the user did not pin sizes.
	isArm := strings.Contains(d.CreateArgs, "arm64")
	if len(cpSizes) == 0 {
		if isArm {
			add(awsDefaultArmControlPlaneSize)
		} else {
			add(awsDefaultControlPlaneSize)
		}
	}
	if len(nodeSizes) == 0 && isArm {
		add(awsDefaultArmNodeSize)
	}
	return types
}

// extractFlagValues returns every value passed for the given flag in args,
// supporting both "--flag=value" and "--flag value" forms. StringSlice flags
// may be comma-separated; callers are expected to split on commas.
func extractFlagValues(args, flag string) []string {
	if args == "" {
		return nil
	}
	var values []string
	fields := strings.Fields(args)
	for i := 0; i < len(fields); i++ {
		f := fields[i]
		if f == flag && i+1 < len(fields) {
			values = append(values, fields[i+1])
			i++
		} else if strings.HasPrefix(f, flag+"=") {
			values = append(values, strings.TrimPrefix(f, flag+"="))
		}
	}
	return values
}

// builtFromKopsRoot reports whether kopsBinaryPath lies inside the repo checkout at kopsRoot.
// Scenario scripts always pass --kops-root, so its presence alone doesn't imply a source build.
func builtFromKopsRoot(kopsBinaryPath, kopsRoot string) bool {
	if kopsBinaryPath == "" || kopsRoot == "" {
		return false
	}
	rel, err := filepath.Rel(kopsRoot, kopsBinaryPath)
	if err != nil {
		return false
	}
	return filepath.IsLocal(rel)
}

// localChannelArgs rewrites a bare --channel shorthand (e.g. alpha) to a file:// URL under the
// checkout's channels/ dir, defaulting to alpha when absent. URLs and "none" are passed through.
func localChannelArgs(args []string, kopsRoot string) []string {
	for i, a := range args {
		if v, ok := strings.CutPrefix(a, "--channel="); ok {
			if isChannelShorthand(v) {
				args[i] = "--channel=" + localChannelURL(kopsRoot, v)
			}
			return args
		}
		if a == "--channel" && i+1 < len(args) {
			if isChannelShorthand(args[i+1]) {
				args[i+1] = localChannelURL(kopsRoot, args[i+1])
			}
			return args
		}
	}
	return append(args, "--channel="+localChannelURL(kopsRoot, "alpha"))
}

// isChannelShorthand reports whether v is a bare channel name, not a URL or the "none" sentinel.
func isChannelShorthand(v string) bool {
	return v != "" && v != "none" && !strings.Contains(v, "://")
}

// localChannelURL returns the file:// URL for channel name under the checkout's channels/ dir.
func localChannelURL(kopsRoot, name string) string {
	return "file://" + filepath.ToSlash(filepath.Join(kopsRoot, "channels", name))
}

// prowJobLabel returns a "key=value" cloud-label fragment recording the prow
// JOB_NAME, sanitized for the target cloud provider. Returns ok=false when
// jobName is empty or the cloud provider does not surface kops cloudLabels.
func prowJobLabel(cloudProvider, jobName string) (string, bool) {
	if jobName == "" {
		return "", false
	}
	switch cloudProvider {
	case "aws":
		// AWS tags allow '/' and '.' in keys, values up to 256 chars.
		return "prow.k8s.io/job=" + jobName, true
	case "azure":
		// Azure tag names cannot contain '<>%&\?/'.
		return "prow.k8s.io_job=" + jobName, true
	case "gce":
		// GCE labels: keys/values must be lowercase, <=63 chars,
		// match [a-z0-9_-], and keys must start with a letter.
		return "prow_k8s_io_job=" + sanitizeGCELabelValue(jobName), true
	default:
		// digitalocean: kops does not pass cloudLabels to DO resources, and DO
		// tags are flat single words rather than key=value pairs.
		return "", false
	}
}

func sanitizeGCELabelValue(v string) string {
	v = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		}
		return '-'
	}, strings.ToLower(v))
	if len(v) > 63 {
		v = v[:63]
	}
	return v
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
