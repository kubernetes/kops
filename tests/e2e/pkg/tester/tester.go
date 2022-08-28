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

package tester

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/octago/sflags/gen/gpflag"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	unversioned "k8s.io/kops/pkg/apis/kops"
	api "k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/tests/e2e/pkg/kops"
	"sigs.k8s.io/kubetest2/pkg/testers/ginkgo"
)

// Tester wraps kubetest2's ginkgo tester with additional functionality
type Tester struct {
	*ginkgo.Tester

	kopsCluster        *api.Cluster
	kopsInstanceGroups []*api.InstanceGroup
}

func (t *Tester) pretestSetup() error {
	err := t.AcquireKubectl()
	if err != nil {
		return fmt.Errorf("failed to get kubectl package from published releases: %s", err)
	}
	return nil
}

// parseKubeconfig will get the current kubeconfig, and extract the specified field by jsonpath.
func parseKubeconfig(jsonPath string) (string, error) {
	args := []string{
		"kubectl", "config", "view", "--minify", "-o", "jsonpath={" + jsonPath + "}",
	}
	c := exec.Command(args[0], args[1:]...)
	var stdout bytes.Buffer
	c.Stdout = &stdout
	var stderr bytes.Buffer
	c.Stderr = &stderr
	if err := c.Run(); err != nil {
		klog.Warningf("failed to run %s; stderr=%s", strings.Join(args, " "), stderr.String())
		return "", fmt.Errorf("error querying current config from kubectl: %w", err)
	}

	s := strings.TrimSpace(stdout.String())
	if s == "" {
		return "", fmt.Errorf("kubeconfig did not contain " + jsonPath)
	}
	return s, nil
}

// The --host flag was required in the kubernetes e2e tests, until https://github.com/kubernetes/kubernetes/pull/87030
// We can likely drop this when we drop support / testing for k8s 1.17
func (t *Tester) addHostFlag() error {
	server, err := parseKubeconfig(".clusters[0].cluster.server")
	if err != nil {
		return err
	}
	klog.Infof("Adding --host=%s", server)
	t.TestArgs += " --host=" + server
	return nil
}

// hasFlag detects if the specified flag has been passed in the args
func hasFlag(args string, flag string) bool {
	for _, arg := range strings.Split(args, " ") {
		if !strings.HasPrefix(arg, "-") {
			continue
		}

		arg = strings.TrimLeft(arg, "-")
		if arg == flag || strings.HasPrefix(arg, flag+"=") {
			return true
		}
	}
	return false
}

func (t *Tester) getKopsCluster() (*api.Cluster, error) {
	if t.kopsCluster != nil {
		return t.kopsCluster, nil
	}

	currentContext, err := parseKubeconfig(".current-context")
	if err != nil {
		return nil, err
	}

	kopsClusterName := currentContext

	cluster, err := kops.GetCluster("kops", kopsClusterName, nil)
	if err != nil {
		return nil, err
	}
	t.kopsCluster = cluster

	return cluster, nil
}

func (t *Tester) getKopsInstanceGroups() ([]*api.InstanceGroup, error) {
	if t.kopsInstanceGroups != nil {
		return t.kopsInstanceGroups, nil
	}

	cluster, err := t.getKopsCluster()
	if err != nil {
		return nil, err
	}

	igs, err := kops.GetInstanceGroups("kops", cluster.Name, nil)
	if err != nil {
		return nil, err
	}
	t.kopsInstanceGroups = igs

	return igs, nil
}

func (t *Tester) addProviderFlag() error {
	if hasFlag(t.TestArgs, "provider") {
		return nil
	}

	cluster, err := t.getKopsCluster()
	if err != nil {
		return err
	}

	provider := ""
	switch cluster.Spec.LegacyCloudProvider {
	case "aws", "gce":
		provider = cluster.Spec.LegacyCloudProvider
	case "digitalocean":
	default:
		klog.Warningf("unhandled cluster.spec.cloudProvider %q for determining ginkgo Provider", cluster.Spec.CloudProvider)
	}

	klog.Infof("Setting --provider=%s", provider)
	t.TestArgs += " --provider=" + provider
	return nil
}

func (t *Tester) addZoneFlag() error {
	// gce-zone is indeed used for AWS as well!
	if hasFlag(t.TestArgs, "gce-zone") {
		return nil
	}

	// The zone flag is used to provision volumes, and we try to attach that volume to a (normal) pod
	zoneNames, err := t.getSchedulableZones()
	if err != nil {
		return err
	}

	// gce-zone only expects one zone, we just pass the first one
	zone := zoneNames[0]
	klog.Infof("Setting --gce-zone=%s", zone)
	t.TestArgs += " --gce-zone=" + zone

	// TODO: Pass the new gce-zones flag for 1.21 with all zones?

	return nil
}

func (t *Tester) addMultiZoneFlag() error {
	if hasFlag(t.TestArgs, "gce-multizone") {
		return nil
	}

	zoneNames, err := t.getAllZones()
	if err != nil {
		return err
	}

	klog.Infof("Setting --gce-multizone=%t", len(zoneNames) > 1)
	t.TestArgs += fmt.Sprintf(" --gce-multizone=%t", len(zoneNames) > 1)

	return nil
}

func (t *Tester) addRegionFlag() error {
	// gce-zone is used for other cloud providers as well
	if hasFlag(t.TestArgs, "gce-region") {
		return nil
	}

	cluster, err := t.getKopsCluster()
	if err != nil {
		return err
	}

	// We don't explicitly set the provider's region in the spec so we need to extract it from vairous fields
	var region string
	switch cluster.Spec.LegacyCloudProvider {
	case "aws":
		zone := cluster.Spec.Subnets[0].Zone
		region = zone[:len(zone)-1]
	case "gce":
		region = cluster.Spec.Subnets[0].Region
	default:
		klog.Warningf("unhandled region detection for cloud provider: %v", cluster.Spec.CloudProvider)
	}

	klog.Infof("Setting --gce-region=%s", region)
	t.TestArgs += " --gce-region=" + region
	return nil
}

func (t *Tester) addClusterTagFlag() error {
	if hasFlag(t.TestArgs, "cluster-tag") {
		return nil
	}

	cluster, err := t.getKopsCluster()
	if err != nil {
		return err
	}

	clusterName := cluster.ObjectMeta.Name
	klog.Infof("Setting --cluster-tag=%s", clusterName)
	t.TestArgs += " --cluster-tag=" + clusterName

	return nil
}

func (t *Tester) addProjectFlag() error {
	if hasFlag(t.TestArgs, "gce-project") {
		return nil
	}

	cluster, err := t.getKopsCluster()
	if err != nil {
		return err
	}

	projectID := cluster.Spec.Project
	if projectID == "" {
		return nil
	}
	klog.Infof("Setting --gce-project=%s", projectID)
	t.TestArgs += " --gce-project=" + projectID

	return nil
}

func (t *Tester) getZonesForInstanceGroups(igs []*api.InstanceGroup) ([]string, error) {
	cluster, err := t.getKopsCluster()
	if err != nil {
		return nil, err
	}

	clusterSubnets := make(map[string]*api.ClusterSubnetSpec)
	for i := range cluster.Spec.Subnets {
		subnet := &cluster.Spec.Subnets[i]
		clusterSubnets[subnet.Name] = subnet
	}

	zones := sets.NewString()
	for _, ig := range igs {
		// Gather zones on GCE
		for _, zone := range ig.Spec.Zones {
			zones.Insert(zone)
		}

		// Gather zones on AWS
		for _, subnetName := range ig.Spec.Subnets {
			subnet := clusterSubnets[subnetName]
			if subnet == nil {
				return nil, fmt.Errorf("instanceGroup %q specified subnet %q, but was not found in cluster", ig.Name, subnetName)
			}
			if subnet.Zone != "" {
				zones.Insert(subnet.Zone)
			}
		}
	}

	zoneNames := zones.List()
	if len(zoneNames) == 0 {
		return nil, nil
	}
	return zoneNames, nil
}

func (t *Tester) getAllZones() ([]string, error) {
	igs, err := t.getKopsInstanceGroups()
	if err != nil {
		return nil, err
	}

	zoneNames, err := t.getZonesForInstanceGroups(igs)
	if err != nil {
		return nil, err
	}

	if len(zoneNames) == 0 {
		klog.Warningf("no zones found in instance groups")
	}
	return zoneNames, nil
}

func (t *Tester) getSchedulableZones() ([]string, error) {
	igs, err := t.getKopsInstanceGroups()
	if err != nil {
		return nil, err
	}

	var schedulable []*api.InstanceGroup
	for _, ig := range igs {
		if unversioned.InstanceGroupRole(ig.Spec.Role) == unversioned.InstanceGroupRoleMaster {
			continue
		}
		if unversioned.InstanceGroupRole(ig.Spec.Role) == unversioned.InstanceGroupRoleAPIServer {
			continue
		}

		schedulable = append(schedulable, ig)
	}

	zoneNames, err := t.getZonesForInstanceGroups(schedulable)
	if err != nil {
		return nil, err
	}

	if len(zoneNames) == 0 {
		klog.Warningf("no zones found in schedulable instance groups")
	}
	return zoneNames, nil
}

func (t *Tester) addNodeOSArchFlag() error {
	igs, err := t.getKopsInstanceGroups()
	if err != nil {
		return err
	}
	for _, ig := range igs {
		if strings.Contains(ig.Spec.Image, "arm64") {
			klog.Info("Setting --node-os-arch=arm64")
			t.TestArgs += " --node-os-arch=arm64"
			break
		}
	}
	return nil
}

func (t *Tester) addNonBlockingTaintsFlag() {
	if hasFlag(t.TestArgs, "non-blocking-taints") {
		return
	}
	nbt := "node-role.kubernetes.io/master,"
	nbt += "node-role.kubernetes.io/api-server,"
	nbt += "node-role.kubernetes.io/control-plane"
	klog.Infof("Setting --non-blocking-taints=%s", nbt)
	t.TestArgs += fmt.Sprintf(" --non-blocking-taints=%v", nbt)
}

func (t *Tester) addCSIDriverFlags() error {
	cluster, err := t.getKopsCluster()
	if err != nil {
		return err
	}

	if cluster.Spec.CloudConfig != nil &&
		cluster.Spec.CloudConfig.AWSEBSCSIDriver != nil &&
		cluster.Spec.CloudConfig.AWSEBSCSIDriver.Enabled != nil &&
		*cluster.Spec.CloudConfig.AWSEBSCSIDriver.Enabled {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		klog.Infof("Setting --storage.testdriver=%s/tests/e2e/csi-manifests/ebs.yaml --storage.migratedPlugins=kubernetes.io/aws-ebs", cwd)
		t.TestArgs += fmt.Sprintf(" --storage.testdriver=%s/tests/e2e/csi-manifests/aws-ebs/driver.yaml --storage.migratedPlugins=kubernetes.io/aws-ebs", cwd)
	} else {
		klog.Info("EBS CSI driver not enabled. Skipping tests")
	}
	return nil

}

func (t *Tester) execute() error {
	fs, err := gpflag.Parse(t)
	if err != nil {
		return fmt.Errorf("failed to initialize tester: %v", err)
	}

	help := fs.BoolP("help", "h", false, "")
	if err := fs.Parse(os.Args); err != nil {
		return fmt.Errorf("failed to parse flags: %v", err)
	}

	if *help {
		fs.SetOutput(os.Stdout)
		fs.PrintDefaults()
		return nil
	}

	if err := t.pretestSetup(); err != nil {
		return err
	}

	if err := t.addHostFlag(); err != nil {
		return err
	}

	if err := t.addProviderFlag(); err != nil {
		return err
	}

	if err := t.addZoneFlag(); err != nil {
		return err
	}

	if err := t.addClusterTagFlag(); err != nil {
		return err
	}

	if err := t.addRegionFlag(); err != nil {
		return err
	}

	if err := t.addMultiZoneFlag(); err != nil {
		return err
	}

	if err := t.addProjectFlag(); err != nil {
		return err
	}

	if err := t.setSkipRegexFlag(); err != nil {
		return err
	}

	if err := t.addNodeOSArchFlag(); err != nil {
		return err
	}

	t.addNonBlockingTaintsFlag()

	if err := t.addCSIDriverFlags(); err != nil {
		return err
	}

	t.TestArgs += " --disable-log-dump"

	return t.Test()
}

func NewDefaultTester() *Tester {
	t := &Tester{}
	t.Tester = ginkgo.NewDefaultTester()
	return t
}

func Main() {
	t := NewDefaultTester()
	if err := t.execute(); err != nil {
		klog.Fatalf("failed to run ginkgo tester: %v", err)
	}
}
