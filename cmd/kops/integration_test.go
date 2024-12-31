/*
Copyright 2019 The Kubernetes Authors.

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

package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/pkg/testutils/golden"
	"k8s.io/kops/pkg/testutils/testcontext"
	"k8s.io/kops/pkg/truncate"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

// updateClusterTestBase is added automatically to the srcDir on all
// tests using runTest, including runTestTerraformAWS, runTestTerraformGCE
const updateClusterTestBase = "../../tests/integration/update_cluster/"

type integrationTest struct {
	clusterName    string
	srcDir         string
	version        string
	private        bool
	zones          int
	expectPolicies bool
	// expectServiceAccountRolePolicies is a list of per-ServiceAccount IAM roles (instead of just using the node roles)
	expectServiceAccountRolePolicies []string
	expectTerraformFilenames         []string
	kubeDNS                          bool
	discovery                        bool
	lifecycleOverrides               []string
	sshKey                           bool
	bastionUserData                  bool
	ciliumEtcd                       bool
	// nth is true if we should check for files created by nth queue processor add on
	nth          bool
	nthRebalance bool
	// enable GCE startup script
	startupScript bool
}

func newIntegrationTest(clusterName, srcDir string) *integrationTest {
	return &integrationTest{
		clusterName:    clusterName,
		srcDir:         srcDir,
		version:        "v1alpha2",
		zones:          1,
		expectPolicies: true,
		nth:            true,
		sshKey:         true,
	}
}

func (i *integrationTest) withStartupScript() *integrationTest {
	i.startupScript = true
	return i
}

func (i *integrationTest) withVersion(version string) *integrationTest {
	i.version = version
	return i
}

func (i *integrationTest) withZones(zones int) *integrationTest {
	i.zones = zones
	return i
}

func (i *integrationTest) withoutSSHKey() *integrationTest {
	i.sshKey = false
	return i
}

func (i *integrationTest) withoutPolicies() *integrationTest {
	i.expectPolicies = false
	return i
}

func (i *integrationTest) withLifecycleOverrides(lco []string) *integrationTest {
	i.lifecycleOverrides = lco
	return i
}

func (i *integrationTest) withPrivate() *integrationTest {
	i.private = true
	return i
}

// withServiceAccountRole indicates we expect to assign an IAM role for a ServiceAccount (instead of just using the node roles)
func (i *integrationTest) withServiceAccountRole(sa string, inlinePolicy bool) *integrationTest {
	role := truncate.TruncateString(sa+".sa."+i.clusterName, truncate.TruncateStringOptions{MaxLength: iam.MaxLengthIAMRoleName, AlwaysAddHash: false})
	i.expectServiceAccountRolePolicies = append(i.expectServiceAccountRolePolicies, fmt.Sprintf("aws_iam_role_%s_policy", role))
	if inlinePolicy {
		i.expectServiceAccountRolePolicies = append(i.expectServiceAccountRolePolicies, fmt.Sprintf("aws_iam_role_policy_%s_policy", role))
	}
	return i
}

func (i *integrationTest) withBastionUserData() *integrationTest {
	i.bastionUserData = true
	return i
}

func (i *integrationTest) withCiliumEtcd() *integrationTest {
	i.ciliumEtcd = true
	return i
}

func (i *integrationTest) withDedicatedAPIServer() *integrationTest {
	i.expectTerraformFilenames = append(i.expectTerraformFilenames,
		"aws_iam_role_apiservers."+i.clusterName+"_policy",
		"aws_iam_role_policy_apiservers."+i.clusterName+"_policy",
		"aws_launch_template_apiserver.apiservers."+i.clusterName+"_user_data",
		"aws_s3_object_nodeupconfig-apiserver_content",
	)
	return i
}

func (i *integrationTest) withoutNTH() *integrationTest {
	i.nth = false
	return i
}

func (i *integrationTest) withNTHRebalance() *integrationTest {
	i.nthRebalance = true
	return i
}

func (i *integrationTest) withOIDCDiscovery() *integrationTest {
	i.discovery = true
	return i
}

func (i *integrationTest) withManagedFiles(files ...string) *integrationTest {
	for _, file := range files {
		i.expectTerraformFilenames = append(i.expectTerraformFilenames,
			"aws_s3_object_"+file+"_content")
	}
	return i
}

func (i *integrationTest) withAddons(addons ...string) *integrationTest {
	for _, addon := range addons {
		i.expectTerraformFilenames = append(i.expectTerraformFilenames,
			"aws_s3_object_"+i.clusterName+"-addons-"+addon+"_content")
	}
	return i
}

func (i integrationTest) withDefaultServiceAccountRoles24() *integrationTest {
	return i.withServiceAccountRole("dns-controller.kube-system", true).
		withServiceAccountRole("aws-cloud-controller-manager.kube-system", true).
		withServiceAccountRole("ebs-csi-controller-sa.kube-system", true)
}

// withDefaultAddons24 adds the default addons for an AWS cluster running k8s 1.24
func (i integrationTest) withDefaultAddons24() *integrationTest {
	return i.withAddons(
		awsCCMAddon,
		awsEBSCSIAddon,
		dnsControllerAddon,
		leaderElectionAddon,
	)
}

// withDefaultAddons30 adds the default addons for an AWS cluster running k8s 1.30
func (i integrationTest) withDefaultAddons30() *integrationTest {
	return i.withAddons(
		awsCCMAddon,
		awsEBSCSIAddon,
		dnsControllerAddon,
	)
}

func (i integrationTest) withDefaults24() *integrationTest {
	return i.withDefaultAddons24().withDefaultServiceAccountRoles24()
}

const (
	awsAuthenticatorAddon = "authentication.aws-k8s-1.12"
	awsCCMAddon           = "aws-cloud-controller.addons.k8s.io-k8s-1.18"
	awsEBSCSIAddon        = "aws-ebs-csi-driver.addons.k8s.io-k8s-1.17"

	gcpCCMAddon   = "gcp-cloud-controller.addons.k8s.io-k8s-1.23"
	gcpPDCSIAddon = "gcp-pd-csi-driver.addons.k8s.io-k8s-1.23"

	scwCCMAddon = "scaleway-cloud-controller.addons.k8s.io-k8s-1.24"
	scwCSIAddon = "scaleway-csi-driver.addons.k8s.io-k8s-1.24"

	calicoAddon  = "networking.projectcalico.org-k8s-1.25"
	canalAddon   = "networking.projectcalico.org.canal-k8s-1.25"
	ciliumAddon  = "networking.cilium.io-k8s-1.16"
	flannelAddon = "networking.flannel-k8s-1.25"
	kindnetAddon = "networking.kindnet-k8s-1.32"

	certManagerAddon         = "certmanager.io-k8s-1.16"
	clusterAutoscalerAddon   = "cluster-autoscaler.addons.k8s.io-k8s-1.15"
	dnsControllerAddon       = "dns-controller.addons.k8s.io-k8s-1.12"
	leaderElectionAddon      = "leader-migration.rbac.addons.k8s.io-k8s-1.23"
	metricsServerAddon       = "metrics-server.addons.k8s.io-k8s-1.11"
	nodeProblemDetectorAddon = "node-problem-detector.addons.k8s.io-k8s-1.17"
)

// TestMinimalAWS runs the test on a minimum configuration, similar to kops create cluster minimal.example.com --zones us-west-1a
func TestMinimalAWS(t *testing.T) {
	newIntegrationTest("minimal-aws.example.com", "minimal-aws").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestMinimal runs the test on a minimum configuration
func TestMinimal_v1_25(t *testing.T) {
	newIntegrationTest("minimal.example.com", "minimal-1.25").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
			leaderElectionAddon,
		).
		runTestTerraformAWS(t)
}

// TestMinimal runs the test on a minimum configuration
func TestMinimal_v1_26(t *testing.T) {
	t.Setenv("KOPS_RUN_TOO_NEW_VERSION", "1")

	newIntegrationTest("minimal.example.com", "minimal-1.26").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestMinimal runs the test on a minimum configuration
func TestMinimal_v1_27(t *testing.T) {
	t.Setenv("KOPS_RUN_TOO_NEW_VERSION", "1")

	newIntegrationTest("minimal.example.com", "minimal-1.27").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestMinimal runs the test on a minimum configuration
func TestMinimal_v1_28(t *testing.T) {
	t.Setenv("KOPS_RUN_TOO_NEW_VERSION", "1")

	newIntegrationTest("minimal.example.com", "minimal-1.28").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestMinimal runs the test on a minimum configuration
func TestMinimal_v1_29(t *testing.T) {
	t.Setenv("KOPS_RUN_TOO_NEW_VERSION", "1")

	newIntegrationTest("minimal.example.com", "minimal-1.29").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestMinimal runs the test on a minimum configuration
func TestMinimal_v1_30(t *testing.T) {
	t.Setenv("KOPS_RUN_TOO_NEW_VERSION", "1")

	newIntegrationTest("minimal.example.com", "minimal-1.30").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestMinimal_NoneDNS runs the test on a minimum configuration with --dns=none
func TestMinimal_NoneDNS(t *testing.T) {
	t.Setenv("KOPS_RUN_TOO_NEW_VERSION", "1")

	newIntegrationTest("minimal.example.com", "minimal-dns-none").
		withAddons(
			awsEBSCSIAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestHetzner runs the test on a minimum configuration
func TestHetzner(t *testing.T) {
	t.Setenv("HCLOUD_TOKEN", "REDACTED")
	newIntegrationTest("minimal.example.com", "minimal_hetzner").
		runTestTerraformHetzner(t)
}

func TestNvidia(t *testing.T) {
	newIntegrationTest("minimal.example.com", "nvidia").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
			"nvidia.addons.k8s.io-k8s-1.16",
		).
		runTestTerraformAWS(t)
}

// TestMinimal runs the test on a minimum gossip configuration
func TestMinimalGossip(t *testing.T) {
	newIntegrationTest("minimal.k8s.local", "minimal_gossip").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestMinimal runs the test on a minimum gossip configuration with irsa enabled
func TestMinimalGossipIRSA(t *testing.T) {
	newIntegrationTest("minimal.k8s.local", "minimal_gossip_irsa").
		withOIDCDiscovery().
		withServiceAccountRole("aws-cloud-controller-manager.kube-system", true).
		withServiceAccountRole("ebs-csi-controller-sa.kube-system", true).
		withServiceAccountRole("aws-node-termination-handler.kube-system", true).
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestMinimalGCE runs tests on a minimal GCE configuration
func TestMinimalGCE(t *testing.T) {
	newIntegrationTest("minimal-gce.example.com", "minimal_gce").
		withAddons(
			dnsControllerAddon,
			gcpCCMAddon,
			gcpPDCSIAddon,
		).
		runTestTerraformGCE(t)
}

// TestMinimalGCEPrivate runs tests on a minimal GCE configuration with private topology.
func TestMinimalGCEPrivate(t *testing.T) {
	newIntegrationTest("minimal-gce-private.example.com", "minimal_gce_private").
		withAddons(
			dnsControllerAddon,
			gcpCCMAddon,
			gcpPDCSIAddon,
		).
		runTestTerraformGCE(t)
}

// TestMinimalGCEInternalLoadBalancer runs tests on a minimal GCE configuration with an internal load balancer.
func TestMinimalGCEInternalLoadBalancer(t *testing.T) {
	newIntegrationTest("minimal-gce-ilb.example.com", "minimal_gce_ilb").
		withAddons(
			dnsControllerAddon,
			gcpCCMAddon,
			gcpPDCSIAddon,
		).
		runTestTerraformGCE(t)
}

// TestMinimalGCEPublicLoadBalancer runs tests on a minimal GCE configuration with a public load balancer.
func TestMinimalGCEPublicLoadBalancer(t *testing.T) {
	newIntegrationTest("minimal-gce-plb.example.com", "minimal_gce_plb").
		withAddons(
			dnsControllerAddon,
			gcpCCMAddon,
			gcpPDCSIAddon,
		).
		runTestTerraformGCE(t)
}

// TestMinimalGCELongClusterName runs tests on a minimal GCE configuration with a very long cluster name
func TestMinimalGCELongClusterName(t *testing.T) {
	newIntegrationTest("minimal-gce-with-a-very-very-very-very-very-long-name.example.com", "minimal_gce_longclustername").
		withStartupScript().
		withAddons(
			dnsControllerAddon,
			gcpCCMAddon,
			gcpPDCSIAddon,
		).
		runTestTerraformGCE(t)
}

// TestMinimalGCEInternalLoadBalancerLongClusterName runs tests on a minimal GCE configuration with an internal load balancer and a very long cluster name
func TestMinimalGCEInternalLoadBalancerLongClusterName(t *testing.T) {
	newIntegrationTest("minimal-gce-with-a-very-very-very-very-very-long-name.example.com", "minimal_gce_ilb_longclustername").
		withAddons(
			dnsControllerAddon,
			gcpCCMAddon,
			gcpPDCSIAddon,
		).
		runTestTerraformGCE(t)
}

// TestMinimalGCEDNSNone runs tests on a minimal GCE configuration with --dns=none
func TestMinimalGCEDNSNone(t *testing.T) {
	newIntegrationTest("minimal-gce.example.com", "minimal_gce_dns-none").
		withAddons(
			gcpCCMAddon,
			gcpPDCSIAddon,
		).
		runTestTerraformGCE(t)
}

// TestMinimalScaleway runs tests on a minimal Scaleway cluster with gossip DNS
func TestMinimalScaleway(t *testing.T) {
	t.Setenv("SCW_PROFILE", "REDACTED")
	newIntegrationTest("scw-minimal.k8s.local", "minimal_scaleway").
		withAddons(
			scwCCMAddon,
			scwCSIAddon,
			dnsControllerAddon,
		).
		runTestTerraformScaleway(t)
}

// TestHA runs the test on a simple HA configuration, similar to kops create cluster minimal.example.com --zones us-west-1a,us-west-1b,us-west-1c --master-count=3
func TestHA(t *testing.T) {
	newIntegrationTest("ha.example.com", "ha").withZones(3).
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestHighAvailabilityGCE runs the test on a simple HA GCE configuration, similar to kops create cluster ha-gce.example.com
// --zones us-test1-a,us-test1-b,us-test1-c --master-count=3
func TestHighAvailabilityGCE(t *testing.T) {
	newIntegrationTest("ha-gce.example.com", "ha_gce").withZones(3).
		withAddons(
			dnsControllerAddon,
			gcpCCMAddon,
			gcpPDCSIAddon,
		).
		runTestTerraformGCE(t)
}

// TestComplex runs the test on a more complex configuration, intended to hit more of the edge cases
func TestComplex(t *testing.T) {
	newIntegrationTest("complex.example.com", "complex").withoutSSHKey().
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
			awsAuthenticatorAddon,
		).
		runTestTerraformAWS(t)
	newIntegrationTest("complex.example.com", "complex").withoutSSHKey().withVersion("legacy-v1alpha2").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
			awsAuthenticatorAddon,
		).
		runTestTerraformAWS(t)
}

// TestCompress runs a test on compressing structs in nodeus.sh user-data
func TestCompress(t *testing.T) {
	newIntegrationTest("compress.example.com", "compress").withoutSSHKey().
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestExternalPolicies tests external policies output
func TestExternalPolicies(t *testing.T) {
	newIntegrationTest("externalpolicies.example.com", "externalpolicies").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestMinimalIPv6 runs the test on a minimum IPv6 configuration
func TestMinimalIPv6(t *testing.T) {
	newIntegrationTest("minimal-ipv6.example.com", "minimal-ipv6").
		withDefaultAddons24().
		runTestTerraformAWS(t)
}

// TestMinimalIPv6Calico runs the test on a minimum IPv6 configuration with Calico
func TestMinimalIPv6Calico(t *testing.T) {
	newIntegrationTest("minimal-ipv6.example.com", "minimal-ipv6-calico").
		withDefaultAddons24().
		withAddons(calicoAddon).
		runTestTerraformAWS(t)
}

// TestMinimalIPv6Cilium runs the test on a minimum IPv6 configuration with Cilium
func TestMinimalIPv6Cilium(t *testing.T) {
	newIntegrationTest("minimal-ipv6.example.com", "minimal-ipv6-cilium").
		withDefaultAddons24().
		withAddons(ciliumAddon).
		runTestTerraformAWS(t)
}

// TestMinimalIPv6NoSubnetPrefix runs the test with "/64#N" subnet notation
func TestMinimalIPv6NoSubnetPrefix(t *testing.T) {
	newIntegrationTest("minimal-ipv6.example.com", "minimal-ipv6-no-subnet-prefix").
		withDefaultAddons24().
		runTestTerraformAWS(t)
}

// TestMinimalWarmPool runs the test on a minimum Warm Pool configuration
func TestMinimalWarmPool(t *testing.T) {
	newIntegrationTest("minimal-warmpool.example.com", "minimal-warmpool").
		withAddons(
			awsEBSCSIAddon,
			ciliumAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestMinimalEtcd runs the test on a minimum configuration using custom etcd config, similar to kops create cluster minimal.example.com --zones us-west-1a
func TestMinimalEtcd(t *testing.T) {
	newIntegrationTest("minimal-etcd.example.com", "minimal-etcd").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestMinimalGp3 runs the test on a minimum configuration using gp3 volumes, similar to kops create cluster minimal.example.com --zones us-west-1a
func TestMinimalGp3(t *testing.T) {
	newIntegrationTest("minimal.example.com", "minimal-gp3").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestMinimal runs the test on a minimum configuration, similar to kops create cluster minimal.example.com --zones us-west-1a
func TestMinimalLongClusterName(t *testing.T) {
	newIntegrationTest("this.is.truly.a.really.really.long.cluster-name.minimal.example.com", "minimal-longclustername").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestExistingSG runs the test with existing Security Group, similar to kops create cluster minimal.example.com --zones us-west-1a
func TestExistingSG(t *testing.T) {
	newIntegrationTest("existingsg.example.com", "existing_sg").withZones(3).
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestBastionAdditionalUserData runs the test on passing additional user-data to a bastion instance group
func TestBastionAdditionalUserData(t *testing.T) {
	newIntegrationTest("bastionuserdata.example.com", "bastionadditional_user-data").withPrivate().
		withBastionUserData().
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestPrivateFlannel runs the test on a configuration with private topology, flannel networking
func TestPrivateFlannel(t *testing.T) {
	newIntegrationTest("privateflannel.example.com", "privateflannel").
		withPrivate().
		withDefaultAddons24().
		withAddons(flannelAddon).
		runTestTerraformAWS(t)
}

// TestPrivateKindnet runs the test on a configuration with private topology, flannel networking
func TestPrivateKindnet(t *testing.T) {
	newIntegrationTest("privatekindnet.example.com", "privatekindnet").
		withPrivate().
		withDefaultAddons24().
		withAddons(kindnetAddon).
		runTestTerraformAWS(t)
}

// TestPrivateCalico runs the test on a configuration with private topology, calico networking
func TestPrivateCalico(t *testing.T) {
	newIntegrationTest("privatecalico.example.com", "privatecalico").
		withPrivate().
		withDefaultAddons24().
		withAddons(calicoAddon).
		runTestTerraformAWS(t)
}

func TestPrivateCilium(t *testing.T) {
	newIntegrationTest("privatecilium.example.com", "privatecilium").
		withPrivate().
		withAddons(
			awsEBSCSIAddon,
			ciliumAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

func TestPrivateCilium2(t *testing.T) {
	newIntegrationTest("privatecilium.example.com", "privatecilium2").
		withPrivate().
		withDefaultAddons30().
		withAddons("networking.cilium.io-k8s-1.16").
		withAddons(certManagerAddon).
		runTestTerraformAWS(t)
}

func TestPrivateCiliumAdvanced(t *testing.T) {
	newIntegrationTest("privateciliumadvanced.example.com", "privateciliumadvanced").
		withPrivate().
		withCiliumEtcd().
		withManagedFiles("etcd-cluster-spec-cilium", "manifests-etcdmanager-cilium-master-us-test-1a").
		withAddons(
			awsEBSCSIAddon,
			ciliumAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

func TestPrivateCiliumENI(t *testing.T) {
	newIntegrationTest("privatecilium.example.com", "privatecilium-eni").
		withPrivate().
		withAddons(
			awsEBSCSIAddon,
			ciliumAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestPrivateCanal runs the test on a configuration with private topology, canal networking
func TestPrivateCanal(t *testing.T) {
	newIntegrationTest("privatecanal.example.com", "privatecanal").
		withPrivate().
		withDefaultAddons24().
		withAddons(canalAddon).
		runTestTerraformAWS(t)
}

const kopeioNetworkingAddon = "networking.kope.io-k8s-1.12"

// TestPrivateKopeio runs the test on a configuration with private topology, kopeio networking
func TestPrivateKopeio(t *testing.T) {
	newIntegrationTest("privatekopeio.example.com", "privatekopeio").
		withPrivate().
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
			kopeioNetworkingAddon,
		).
		runTestTerraformAWS(t)
}

// TestUnmanaged is a test where all the subnets opt-out of route management
func TestUnmanaged(t *testing.T) {
	newIntegrationTest("unmanaged.example.com", "unmanaged").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		withPrivate().
		runTestTerraformAWS(t)
}

// TestPrivateSharedSubnet runs the test on a configuration with private topology & shared subnets
func TestPrivateSharedSubnet(t *testing.T) {
	newIntegrationTest("private-shared-subnet.example.com", "private-shared-subnet").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		withPrivate().
		runTestTerraformAWS(t)
}

// TestPrivateSharedIP runs the test on a configuration with private topology & shared subnets
func TestPrivateSharedIP(t *testing.T) {
	newIntegrationTest("private-shared-ip.example.com", "private-shared-ip").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		withPrivate().
		runTestTerraformAWS(t)
}

// TestPrivateDns1 runs the test on a configuration with private topology, private dns
func TestPrivateDns1(t *testing.T) {
	newIntegrationTest("privatedns1.example.com", "privatedns1").
		withPrivate().
		withAddons(
			awsCCMAddon,
			awsEBSCSIAddon,
			dnsControllerAddon,
		).
		runTestTerraformAWS(t)
}

// TestPrivateDns2 runs the test on a configuration with private topology, private dns, extant vpc
func TestPrivateDns2(t *testing.T) {
	newIntegrationTest("privatedns2.example.com", "privatedns2").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		withPrivate().
		runTestTerraformAWS(t)
}

// TestDiscoveryFeatureGate runs a simple configuration, but with UseServiceAccountExternalPermissions and the ServiceAccountIssuerDiscovery feature gate enabled
func TestDiscoveryFeatureGate(t *testing.T) {
	newIntegrationTest("minimal.example.com", "public-jwks-apiserver").
		withDefaultServiceAccountRoles24().
		withServiceAccountRole("aws-node-termination-handler.kube-system", true).
		withDefaultAddons30().
		withOIDCDiscovery().
		runTestTerraformAWS(t)
}

func TestVFSServiceAccountIssuerDiscovery(t *testing.T) {
	newIntegrationTest("minimal.example.com", "vfs-said").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		withOIDCDiscovery().
		runTestTerraformAWS(t)
}

// TestAWSLBController runs a simple configuration, but with AWS LB controller and UseServiceAccountExternalPermissions enabled
func TestAWSLBController(t *testing.T) {
	newIntegrationTest("minimal.example.com", "aws-lb-controller").
		withOIDCDiscovery().
		withServiceAccountRole("dns-controller.kube-system", true).
		withServiceAccountRole("aws-load-balancer-controller.kube-system", true).
		withServiceAccountRole("aws-cloud-controller-manager.kube-system", true).
		withServiceAccountRole("aws-node-termination-handler.kube-system", true).
		withServiceAccountRole("ebs-csi-controller-sa.kube-system", true).
		withAddons("aws-load-balancer-controller.addons.k8s.io-k8s-1.19",
			"certmanager.io-k8s-1.16",
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

func TestManyAddons(t *testing.T) {
	newIntegrationTest("many-addons.example.com", "many-addons").
		withAddons(
			"aws-load-balancer-controller.addons.k8s.io-k8s-1.19",
			"certmanager.io-k8s-1.16",
			"cluster-autoscaler.addons.k8s.io-k8s-1.15",
			"networking.amazon-vpc-routed-eni-k8s-1.16",
			"snapshot-controller.addons.k8s.io-k8s-1.20",
			metricsServerAddon,
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
			nodeProblemDetectorAddon,
		).
		runTestTerraformAWS(t)
}

func TestManyAddonsCCMIRSA(t *testing.T) {
	newIntegrationTest("minimal.example.com", "many-addons-ccm-irsa").
		withOIDCDiscovery().
		withServiceAccountRole("dns-controller.kube-system", true).
		withServiceAccountRole("aws-load-balancer-controller.kube-system", true).
		withServiceAccountRole("aws-cloud-controller-manager.kube-system", true).
		withServiceAccountRole("aws-node-termination-handler.kube-system", true).
		withServiceAccountRole("cluster-autoscaler.kube-system", true).
		withServiceAccountRole("ebs-csi-controller-sa.kube-system", true).
		withAddons(
			"aws-ebs-csi-driver.addons.k8s.io-k8s-1.17",
			"aws-load-balancer-controller.addons.k8s.io-k8s-1.19",
			"certmanager.io-k8s-1.16",
			"cluster-autoscaler.addons.k8s.io-k8s-1.15",
			"networking.amazon-vpc-routed-eni-k8s-1.16",
			"snapshot-controller.addons.k8s.io-k8s-1.20",
			"aws-cloud-controller.addons.k8s.io-k8s-1.18",
			metricsServerAddon,
			dnsControllerAddon,
		).
		runTestTerraformAWS(t)
}

func TestManyAddonsCCMIRSA25(t *testing.T) {
	newIntegrationTest("minimal.example.com", "many-addons-ccm-irsa25").
		withOIDCDiscovery().
		withServiceAccountRole("aws-load-balancer-controller.kube-system", true).
		withServiceAccountRole("dns-controller.kube-system", true).
		withServiceAccountRole("aws-cloud-controller-manager.kube-system", true).
		withServiceAccountRole("cluster-autoscaler.kube-system", true).
		withServiceAccountRole("ebs-csi-controller-sa.kube-system", true).
		withServiceAccountRole("aws-node-termination-handler.kube-system", true).
		withAddons(
			"aws-load-balancer-controller.addons.k8s.io-k8s-1.19",
			"aws-ebs-csi-driver.addons.k8s.io-k8s-1.17",
			"certmanager.io-k8s-1.16",
			"cluster-autoscaler.addons.k8s.io-k8s-1.15",
			"networking.amazon-vpc-routed-eni-k8s-1.16",
			"snapshot-controller.addons.k8s.io-k8s-1.20",
			"aws-cloud-controller.addons.k8s.io-k8s-1.18",
			leaderElectionAddon,
			metricsServerAddon,
			dnsControllerAddon,
		).
		runTestTerraformAWS(t)
}

func TestManyAddonsCCMIRSA26(t *testing.T) {
	t.Setenv("KOPS_RUN_TOO_NEW_VERSION", "1")
	newIntegrationTest("minimal.example.com", "many-addons-ccm-irsa26").
		withOIDCDiscovery().
		withServiceAccountRole("aws-load-balancer-controller.kube-system", true).
		withServiceAccountRole("dns-controller.kube-system", true).
		withServiceAccountRole("aws-cloud-controller-manager.kube-system", true).
		withServiceAccountRole("cluster-autoscaler.kube-system", true).
		withServiceAccountRole("ebs-csi-controller-sa.kube-system", true).
		withServiceAccountRole("aws-node-termination-handler.kube-system", true).
		withAddons(
			"aws-load-balancer-controller.addons.k8s.io-k8s-1.19",
			"aws-ebs-csi-driver.addons.k8s.io-k8s-1.17",
			"certmanager.io-k8s-1.16",
			"cluster-autoscaler.addons.k8s.io-k8s-1.15",
			"networking.amazon-vpc-routed-eni-k8s-1.16",
			"snapshot-controller.addons.k8s.io-k8s-1.20",
			"aws-cloud-controller.addons.k8s.io-k8s-1.18",
			"nodelocaldns.addons.k8s.io-k8s-1.12",
			metricsServerAddon,
			dnsControllerAddon,
		).
		runTestTerraformAWS(t)
}

func TestManyAddonsGCE(t *testing.T) {
	t.Setenv("KOPS_RUN_TOO_NEW_VERSION", "1")
	newIntegrationTest("minimal.example.com", "many-addons-gce").
		withAddons(
			certManagerAddon,
			clusterAutoscalerAddon,
			dnsControllerAddon,
			gcpCCMAddon,
			gcpPDCSIAddon,
			metricsServerAddon,
		).
		runTestTerraformGCE(t)
}

func TestCCM(t *testing.T) {
	newIntegrationTest("minimal.example.com", "many-addons-ccm").
		withAddons(
			"aws-ebs-csi-driver.addons.k8s.io-k8s-1.17",
			"aws-load-balancer-controller.addons.k8s.io-k8s-1.19",
			"certmanager.io-k8s-1.16",
			"cluster-autoscaler.addons.k8s.io-k8s-1.15",
			"networking.amazon-vpc-routed-eni-k8s-1.16",
			"snapshot-controller.addons.k8s.io-k8s-1.20",
			"aws-cloud-controller.addons.k8s.io-k8s-1.18",
			dnsControllerAddon,
			metricsServerAddon,
		).
		withNTHRebalance().
		runTestTerraformAWS(t)
}

func TestExternalDNS(t *testing.T) {
	newIntegrationTest("minimal.example.com", "external_dns").
		withAddons(
			awsEBSCSIAddon,
			awsCCMAddon,
			"external-dns.addons.k8s.io-k8s-1.19",
		).
		runTestTerraformAWS(t)
}

func TestExternalDNSIRSA(t *testing.T) {
	newIntegrationTest("minimal.example.com", "external_dns_irsa").
		withOIDCDiscovery().
		withAddons(
			awsEBSCSIAddon,
			awsCCMAddon,
			"external-dns.addons.k8s.io-k8s-1.19",
		).
		withServiceAccountRole("aws-cloud-controller-manager.kube-system", true).
		withServiceAccountRole("aws-node-termination-handler.kube-system", true).
		withServiceAccountRole("ebs-csi-controller-sa.kube-system", true).
		withServiceAccountRole("external-dns.kube-system", true).
		runTestTerraformAWS(t)
}

func TestKarpenter(t *testing.T) {
	test := newIntegrationTest("minimal.example.com", "karpenter").
		withOIDCDiscovery().
		withDefaults24().
		withAddons("karpenter.sh-k8s-1.19").
		withServiceAccountRole("aws-node-termination-handler.kube-system", true).
		withServiceAccountRole("karpenter.kube-system", true)
	test.expectTerraformFilenames = append(test.expectTerraformFilenames,
		"aws_launch_template_karpenter-nodes-single-machinetype.minimal.example.com_user_data",
		"aws_launch_template_karpenter-nodes-default.minimal.example.com_user_data",
		"aws_s3_object_nodeupconfig-karpenter-nodes-single-machinetype_content",
		"aws_s3_object_nodeupconfig-karpenter-nodes-default_content",
	)
	test.runTestTerraformAWS(t)
}

// TestSharedSubnet runs the test on a configuration with a shared subnet (and VPC)
func TestSharedSubnet(t *testing.T) {
	newIntegrationTest("sharedsubnet.example.com", "shared_subnet").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestSharedVPC runs the test on a configuration with a shared VPC
func TestSharedVPC(t *testing.T) {
	newIntegrationTest("sharedvpc.example.com", "shared_vpc").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestSharedVPCIPv6 runs the test on a configuration with a shared VPC using IPv6
func TestSharedVPCIPv6(t *testing.T) {
	newIntegrationTest("minimal-ipv6.example.com", "shared_vpc_ipv6").
		withDefaultAddons24().
		runTestTerraformAWS(t)
}

// TestExistingIAM runs the test on a configuration with existing IAM instance profiles
func TestExistingIAM(t *testing.T) {
	lifecycleOverrides := []string{"IAMRole=ExistsAndWarnIfChanges", "IAMRolePolicy=ExistsAndWarnIfChanges", "IAMInstanceProfileRole=ExistsAndWarnIfChanges"}
	newIntegrationTest("existing-iam.example.com", "existing_iam").
		withZones(3).
		withoutPolicies().
		withLifecycleOverrides(lifecycleOverrides).
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestPhaseNetwork tests the output of tf for the network phase
func TestPhaseNetwork(t *testing.T) {
	newIntegrationTest("lifecyclephases.example.com", "lifecycle_phases").
		runTestPhase(t, cloudup.PhaseNetwork)
}

func TestExternalLoadBalancer(t *testing.T) {
	newIntegrationTest("externallb.example.com", "externallb").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestPhaseIAM tests the output of tf for the iam phase
func TestPhaseIAM(t *testing.T) {
	t.Skip("unable to test w/o allowing failed validation")
	newIntegrationTest("lifecyclephases.example.com", "lifecycle_phases").
		runTestPhase(t, cloudup.PhaseSecurity)
}

// TestPhaseCluster tests the output of tf for the cluster phase
func TestPhaseCluster(t *testing.T) {
	// TODO fix tf for phase, and allow override on validation
	t.Skip("unable to test w/o allowing failed validation")
	newIntegrationTest("lifecyclephases.example.com", "lifecycle_phases").
		runTestPhase(t, cloudup.PhaseCluster)
}

// TestMixedInstancesASG tests ASGs using a mixed instance policy
func TestMixedInstancesASG(t *testing.T) {
	newIntegrationTest("mixedinstances.example.com", "mixed_instances").
		withZones(3).
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestMixedInstancesSpotASG tests ASGs using a mixed instance policy and spot instances
func TestMixedInstancesSpotASG(t *testing.T) {
	newIntegrationTest("mixedinstances.example.com", "mixed_instances_spot").
		withZones(3).
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestAdditionalObjects runs the test on a configuration that includes additional objects
func TestAdditionalObjects(t *testing.T) {
	newIntegrationTest("additionalobjects.example.com", "additionalobjects").
		withAddons(dnsControllerAddon, awsEBSCSIAddon, awsCCMAddon).
		runTestTerraformAWS(t)
}

// TestContainerd runs the test on a containerd configuration
func TestContainerd(t *testing.T) {
	newIntegrationTest("containerd.example.com", "containerd").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestContainerdCustom runs the test on a custom containerd URL configuration
func TestContainerdCustom(t *testing.T) {
	newIntegrationTest("containerd.example.com", "containerd-custom").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestAPIServerNodes runs a simple configuration with dedicated apiserver nodes
func TestAPIServerNodes(t *testing.T) {
	featureflag.ParseFlags("+APIServerNodes")
	unsetFeatureFlags := func() {
		featureflag.ParseFlags("-APIServerNodes")
	}
	defer unsetFeatureFlags()

	newIntegrationTest("minimal.example.com", "apiservernodes").
		withAddons(
			awsCCMAddon,
			awsEBSCSIAddon,
			dnsControllerAddon,
		).
		withDedicatedAPIServer().
		runTestTerraformAWS(t)
}

// TestNTHIMDSProcessor tests the output for resources required by NTH IMDS Processor mode
func TestNTHIMDSProcessor(t *testing.T) {
	newIntegrationTest("nthimdsprocessor.longclustername.example.com", "nth-imds-processor").
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
			"node-termination-handler.aws-k8s-1.11",
		).
		withoutNTH().
		runTestTerraformAWS(t)
}

// TestNTHIMDSProcessorIRSA tests the output for resources required by NTH IMDS Processor mode with IRSA
func TestNTHIMDSProcessorIRSA(t *testing.T) {
	newIntegrationTest("nthimdsprocessor.longclustername.example.com", "nth-imds-processor-irsa").
		withOIDCDiscovery().
		withServiceAccountRole("dns-controller.kube-system", true).
		withServiceAccountRole("aws-cloud-controller-manager.kube-system", true).
		withServiceAccountRole("aws-node-termination-handler.kube-system", true).
		withServiceAccountRole("ebs-csi-controller-sa.kube-system", true).
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
			"node-termination-handler.aws-k8s-1.11",
		).
		withoutNTH().
		runTestTerraformAWS(t)
}

// TestCustomIRSA runs a simple configuration, but with some additional IAM roles for ServiceAccounts
func TestCustomIRSA(t *testing.T) {
	newIntegrationTest("minimal.example.com", "irsa").
		withOIDCDiscovery().
		withServiceAccountRole("myserviceaccount.default", false).
		withServiceAccountRole("myserviceaccount.test-wildcard", false).
		withServiceAccountRole("myotherserviceaccount.myapp", true).
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
			certManagerAddon,
		).
		withAddons("eks-pod-identity-webhook.addons.k8s.io-k8s-1.16").
		runTestTerraformAWS(t)
}

// TestClusterNameDigit runs a configuration with a cluster name beginning with a digit
func TestClusterNameDigit(t *testing.T) {
	newIntegrationTest("123.example.com", "digit").
		withOIDCDiscovery().
		withServiceAccountRole("myserviceaccount.default", false).
		withServiceAccountRole("myotherserviceaccount.myapp", true).
		withAddons(
			awsEBSCSIAddon,
			dnsControllerAddon,
			awsCCMAddon,
		).
		runTestTerraformAWS(t)
}

// TestCASPriorityExpander tests cluster-autoscaler priority-expander configMap based on instance group autoscalePriority
func TestCASPriorityExpander(t *testing.T) {
	test := newIntegrationTest("cas-priority-expander.example.com", "cluster-autoscaler-priority-expander").
		withAddons(
			awsCCMAddon,
			awsEBSCSIAddon,
			dnsControllerAddon,
			"cluster-autoscaler.addons.k8s.io-k8s-1.15",
		)
	test.expectTerraformFilenames = append(test.expectTerraformFilenames,
		"aws_launch_template_nodes-high-priority.cas-priority-expander.example.com_user_data",
		"aws_launch_template_nodes-low-priority.cas-priority-expander.example.com_user_data",
		"aws_s3_object_nodeupconfig-nodes-high-priority_content",
		"aws_s3_object_nodeupconfig-nodes-low-priority_content",
	)
	test.runTestTerraformAWS(t)
}

// TestCASPriorityExpanderCustom tests cluster-autoscaler priority-expander configMap with custom priority config
func TestCASPriorityExpanderCustom(t *testing.T) {
	test := newIntegrationTest("cas-priority-expander-custom.example.com", "cluster-autoscaler-priority-expander-custom").
		withAddons(
			awsCCMAddon,
			awsEBSCSIAddon,
			dnsControllerAddon,
			"cluster-autoscaler.addons.k8s.io-k8s-1.15",
		)
	test.expectTerraformFilenames = append(test.expectTerraformFilenames,
		"aws_launch_template_nodes-high-priority.cas-priority-expander-custom.example.com_user_data",
		"aws_launch_template_nodes-low-priority.cas-priority-expander-custom.example.com_user_data",
		"aws_s3_object_nodeupconfig-nodes-high-priority_content",
		"aws_s3_object_nodeupconfig-nodes-low-priority_content",
	)
	test.runTestTerraformAWS(t)
}

func (i *integrationTest) runTest(t *testing.T, ctx context.Context, h *testutils.IntegrationTestHarness, expectedDataFilenames []string, tfFileName string, expectedTfFileName string, phase *cloudup.Phase) {
	var stdout bytes.Buffer

	i.srcDir = updateClusterTestBase + i.srcDir
	inputYAML := "in-" + i.version + ".yaml"
	testDataTFPath := "kubernetes.tf"
	actualTFPath := "kubernetes.tf"

	if tfFileName != "" {
		testDataTFPath = tfFileName
	}

	if expectedTfFileName != "" {
		actualTFPath = expectedTfFileName
	}

	factory := i.setupCluster(t, ctx, inputYAML, stdout)

	{
		options := &UpdateClusterOptions{}
		options.InitDefaults()
		options.Target = "terraform"
		options.OutDir = path.Join(h.TempDir, "out")
		options.RunTasksOptions.MaxTaskDuration = 30 * time.Second
		if phase != nil {
			options.Phase = string(*phase)
		}

		// We don't test it here, and it adds a dependency on kubectl
		options.CreateKubecfg = false
		options.ClusterName = i.clusterName
		options.LifecycleOverrides = i.lifecycleOverrides

		_, err := RunUpdateCluster(ctx, factory, &stdout, options)
		if err != nil {
			t.Fatalf("error running update cluster %q: %v", i.clusterName, err)
		}
	}

	// Compare main files
	{
		files, err := os.ReadDir(path.Join(h.TempDir, "out"))
		if err != nil {
			t.Fatalf("failed to read dir: %v", err)
		}

		var fileNames []string
		for _, f := range files {
			fileNames = append(fileNames, f.Name())
		}
		sort.Strings(fileNames)

		actualFilenames := strings.Join(fileNames, ",")
		expectedFilenames := actualTFPath

		if len(expectedDataFilenames) > 0 {
			expectedFilenames = "data," + actualTFPath
		}

		if actualFilenames != expectedFilenames {
			t.Fatalf("unexpected files.  actual=%q, expected=%q, test=%q", actualFilenames, expectedFilenames, testDataTFPath)
		}

		actualTF, err := os.ReadFile(path.Join(h.TempDir, "out", actualTFPath))
		if err != nil {
			t.Fatalf("unexpected error reading actual terraform output: %v", err)
		}

		golden.AssertMatchesFile(t, string(actualTF), path.Join(i.srcDir, testDataTFPath))
	}

	// Compare data files if they are provided
	if len(expectedDataFilenames) > 0 {
		actualDataDir := filepath.Join(h.TempDir, "out", "data")

		expectedDataDir := filepath.Join(i.srcDir, "data")
		for _, filename := range expectedDataFilenames {
			expectedPath := filepath.Join(expectedDataDir, filename)
			actualPath := filepath.Join(actualDataDir, filename)
			actualDataContent, err := os.ReadFile(actualPath)
			if err != nil {
				t.Errorf("failed to read actual data file %q: %v", actualPath, err)
				continue
			}
			golden.AssertMatchesFile(t, string(actualDataContent), expectedPath)
		}

		actualFiles, err := os.ReadDir(actualDataDir)
		if err != nil {
			t.Fatalf("failed to read data dir %q: %v", actualDataDir, err)
		}

		var actualDataFilenames []string
		for _, f := range actualFiles {
			actualDataFilenames = append(actualDataFilenames, f.Name())

			if golden.UpdateExpectedOutput() {
				filename := f.Name()
				expectedPath := filepath.Join(expectedDataDir, filename)
				actualPath := filepath.Join(actualDataDir, filename)
				actualDataContent, err := os.ReadFile(actualPath)
				if err != nil {
					t.Errorf("failed to read actual data file %q: %v", actualPath, err)
					continue
				}
				golden.AssertMatchesFile(t, string(actualDataContent), expectedPath)
			}
		}

		sort.Strings(expectedDataFilenames)
		if !reflect.DeepEqual(actualDataFilenames, expectedDataFilenames) {
			for j := 0; j < len(actualDataFilenames) && j < len(expectedDataFilenames); j++ {
				if actualDataFilenames[j] != expectedDataFilenames[j] {
					t.Errorf("diff @%d: %q vs %q", j, actualDataFilenames[j], expectedDataFilenames[j])
					break
				}
			}
			actual := strings.Join(actualDataFilenames, "\n")
			expected := strings.Join(expectedDataFilenames, "\n")
			diff := diff.FormatDiff(actual, expected)
			t.Log(diff)
			t.Error("unexpected data files.")
		}

		existingExpectedFiles, err := os.ReadDir(expectedDataDir)
		if err != nil {
			t.Fatalf("failed to read data dir %q: %v", expectedDataDir, err)
		}
		existingExpectedFilenames := make([]string, len(existingExpectedFiles))
		for i, f := range existingExpectedFiles {
			existingExpectedFilenames[i] = f.Name()
		}
		for j := 0; j < len(existingExpectedFilenames) && j < len(expectedDataFilenames); j++ {
			if existingExpectedFilenames[j] != expectedDataFilenames[j] {
				t.Errorf("diff with source directory @%d: %q vs %q", j, existingExpectedFilenames[j], expectedDataFilenames[j])
				break
			}
		}
	}
}

func (i *integrationTest) setupCluster(t *testing.T, ctx context.Context, inputYAML string, stdout bytes.Buffer) *util.Factory {
	factoryOptions := &util.FactoryOptions{}
	factoryOptions.RegistryPath = "memfs://tests"

	factory := util.NewFactory(factoryOptions)

	{
		options := &CreateOptions{}
		options.Filenames = []string{path.Join(i.srcDir, inputYAML)}

		err := RunCreate(ctx, factory, &stdout, options)
		if err != nil {
			t.Fatalf("error running %q create: %v", inputYAML, err)
		}
	}

	if i.sshKey {
		options := &CreateSSHPublicKeyOptions{}
		options.ClusterName = i.clusterName
		options.PublicKeyPath = path.Join(i.srcDir, "id_rsa.pub")

		err := RunCreateSSHPublicKey(ctx, factory, &stdout, options)
		if err != nil {
			t.Fatalf("error running %q create public key: %v", inputYAML, err)
		}
	}

	cluster, err := GetCluster(ctx, factory, i.clusterName)
	if err != nil {
		t.Fatalf("error getting cluster: %v", err)
	}

	clientSet, err := factory.KopsClient()
	if err != nil {
		t.Fatalf("error getting clientset: %v", err)
	}

	keyStore, err := clientSet.KeyStore(cluster)
	if err != nil {
		t.Fatalf("error getting keystore: %v", err)
	}

	storeKeyset(t, ctx, keyStore, fi.CertificateIDCA, &testingKeyset{
		primaryKey:           "-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJBANFI3zr0Tk8krsW8vwjfMpzJOlWQ8616vG3YPa2qAgI7V4oKwfV0\nyIg1jt+H6f4P/wkPAPTPTfRp9Iy8oHEEFw0CAwEAAQJATmTyoZ3D+6dtBErocEVT\nKyHBhS3P6YrRLIBU0kmdiQHN8BuzvENqm5PASTq1m6yAAJs7qu9S0kO8u4G+SILv\n7QIhAPNCeJoFHmNUwQ1kxuta1RqICGcNoA4Yx5LiHXd9dPM7AiEA3D7gq8WB8csD\nghBNu/zLy3RdFCkfJqWkX5FhdX29alcCIHw4A1HTL1NV4kcuoQ1qEsw7jt7g7EyG\nhtMQuC9eVywlAiA1Z12s6Og4S+Se3fsrUQHNZHrJT6tJALMZpTO/fGy4YwIhANlJ\nR6hkVKtJp9zhipu6WpvpiAtoIlsNnPMPyuDRwV/u\n-----END RSA PRIVATE KEY-----",
		primaryCertificate:   "-----BEGIN CERTIFICATE-----\nMIIBbjCCARigAwIBAgIMFpANqBD8NSD82AUSMA0GCSqGSIb3DQEBCwUAMBgxFjAU\nBgNVBAMTDWt1YmVybmV0ZXMtY2EwHhcNMjEwNzA3MDcwODAwWhcNMzEwNzA3MDcw\nODAwWjAYMRYwFAYDVQQDEw1rdWJlcm5ldGVzLWNhMFwwDQYJKoZIhvcNAQEBBQAD\nSwAwSAJBANFI3zr0Tk8krsW8vwjfMpzJOlWQ8616vG3YPa2qAgI7V4oKwfV0yIg1\njt+H6f4P/wkPAPTPTfRp9Iy8oHEEFw0CAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgEG\nMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFNG3zVjTcLlJwDsJ4/K9DV7KohUA\nMA0GCSqGSIb3DQEBCwUAA0EAB8d03fY2w7WKpfO29qI295pu2C4ca9AiVGOpgSc8\ntmQsq6rcxt3T+rb589PVtz0mw/cKTxOk6gH2CCC+yHfy2w==\n-----END CERTIFICATE-----",
		secondaryKey:         "-----BEGIN RSA PRIVATE KEY-----\nMIIBOQIBAAJBAMF6F4aZdpe0RUpyykaBpWwZCnwbffhYGOw+fs6RdLuUq7QCNmJm\n/Eq7WWOziMYDiI9SbclpD+6QiJ0N3EqppVUCAwEAAQJAV9YPAit/vKW542+zx0iq\niiXgLbHpgaq1PeOtfChrH5E4C/Bq4P/0MV6bSBm+Hfc9HKaGQE8HMQT7pdkbTECq\nQQIhANSEABWO1ycqVMUeqgnIkkQi/F/m3cZ9r2HIQPj8upcRAiEA6RDOOrrgvpka\nDoDK+eucjeDDKiR5uLFHvftz0PUNkgUCIDutpehn6HuTI6MHbXC55nlD6eN0jasD\n+JBZEAXb0vpBAiBy/qfCspJReJkyrrl3tpj4J/4jvPuR9WbAhmEOqNqZQQIgBrnt\n9mujgf4rNXZTuxAt0ljAzwKFjs+JcTtm4z59uZg=\n-----END RSA PRIVATE KEY-----",
		secondaryCertificate: "-----BEGIN CERTIFICATE-----\nMIIBbjCCARigAwIBAgIMFpANvmSa0OAlYmXKMA0GCSqGSIb3DQEBCwUAMBgxFjAU\nBgNVBAMTDWt1YmVybmV0ZXMtY2EwHhcNMjEwNzA3MDcwOTM2WhcNMzEwNzA3MDcw\nOTM2WjAYMRYwFAYDVQQDEw1rdWJlcm5ldGVzLWNhMFwwDQYJKoZIhvcNAQEBBQAD\nSwAwSAJBAMF6F4aZdpe0RUpyykaBpWwZCnwbffhYGOw+fs6RdLuUq7QCNmJm/Eq7\nWWOziMYDiI9SbclpD+6QiJ0N3EqppVUCAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgEG\nMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFLImp6ARjPDAH6nhI+scWVt3Q9bn\nMA0GCSqGSIb3DQEBCwUAA0EAVQVx5MUtuAIeePuP9o51xtpT2S6Fvfi8J4ICxnlA\n9B7UD2ushcVFPtaeoL9Gfu8aY4KJBeqqg5ojl4qmRnThjw==\n-----END CERTIFICATE-----",
	})
	storeKeyset(t, ctx, keyStore, "apiserver-aggregator-ca", &testingKeyset{
		primaryKey:           "-----BEGIN RSA PRIVATE KEY-----\nMIIBOwIBAAJBAMshO9QDlN4KOVxXoC0On4nSNC4YTMews6U84dsVinB1H2zSO4rY\nCbwv/hpchuVvgxeVe22tCCYkC7Bb3tKC3XsCAwEAAQJAe4xCLGjlQcvsKYsuZFlR\nle0hSawD/y0thuIp6SwH4O92AOsfrWDdiWIVCP6S47oBv351BOcoPbOjxfMTN+f6\naQIhAPIfBCHL/GecX1IVyitI1ueG1z0n5DDOKQAxmxTg82SnAiEA1sYK+vXMIV/e\nCl/CHxKwu7f+ufh1bV0OFyd+eI2+Vw0CICs6eG1kUzNYivhH5ammvp/lxkYn+ijw\nlgdv0+V9aFdfAiEAsTUytiK8zQTGthSQnQbU3+5OtK82ZIgVKjGh/mIlnLkCIQC1\neG3yBXM7/cxw1doWZ7AzMncufx9R8Q2Hblm80UrpaQ==\n-----END RSA PRIVATE KEY-----",
		primaryCertificate:   "-----BEGIN CERTIFICATE-----\nMIIBgjCCASygAwIBAgIMFo3gINaZLHjisEcbMA0GCSqGSIb3DQEBCwUAMCIxIDAe\nBgNVBAMTF2FwaXNlcnZlci1hZ2dyZWdhdG9yLWNhMB4XDTIxMDYzMDA0NTExMloX\nDTMxMDYzMDA0NTExMlowIjEgMB4GA1UEAxMXYXBpc2VydmVyLWFnZ3JlZ2F0b3It\nY2EwXDANBgkqhkiG9w0BAQEFAANLADBIAkEAyyE71AOU3go5XFegLQ6fidI0LhhM\nx7CzpTzh2xWKcHUfbNI7itgJvC/+GlyG5W+DF5V7ba0IJiQLsFve0oLdewIDAQAB\no0IwQDAOBgNVHQ8BAf8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU\nALfqF5ZmfqvqORuJIFilZYKF3d0wDQYJKoZIhvcNAQELBQADQQAHAomFKsF4jvYX\nWM/UzQXDj9nSAFTf8dBPCXyZZNotsOH7+P6W4mMiuVs8bAuGiXGUdbsQ2lpiT/Rk\nCzMeMdr4\n-----END CERTIFICATE-----",
		secondaryKey:         "-----BEGIN RSA PRIVATE KEY-----\nMIIBOwIBAAJBAMshO9QDlN4KOVxXoC0On4nSNC4YTMews6U84dsVinB1H2zSO4rY\nCbwv/hpchuVvgxeVe22tCCYkC7Bb3tKC3XsCAwEAAQJAe4xCLGjlQcvsKYsuZFlR\nle0hSawD/y0thuIp6SwH4O92AOsfrWDdiWIVCP6S47oBv351BOcoPbOjxfMTN+f6\naQIhAPIfBCHL/GecX1IVyitI1ueG1z0n5DDOKQAxmxTg82SnAiEA1sYK+vXMIV/e\nCl/CHxKwu7f+ufh1bV0OFyd+eI2+Vw0CICs6eG1kUzNYivhH5ammvp/lxkYn+ijw\nlgdv0+V9aFdfAiEAsTUytiK8zQTGthSQnQbU3+5OtK82ZIgVKjGh/mIlnLkCIQC1\neG3yBXM7/cxw1doWZ7AzMncufx9R8Q2Hblm80UrpaQ==\n-----END RSA PRIVATE KEY-----",
		secondaryCertificate: "-----BEGIN CERTIFICATE-----\nMIIBgjCCASygAwIBAgIMFo3gM0nxQpiX/agfMA0GCSqGSIb3DQEBCwUAMCIxIDAe\nBgNVBAMTF2FwaXNlcnZlci1hZ2dyZWdhdG9yLWNhMB4XDTIxMDYzMDA0NTIzMVoX\nDTMxMDYzMDA0NTIzMVowIjEgMB4GA1UEAxMXYXBpc2VydmVyLWFnZ3JlZ2F0b3It\nY2EwXDANBgkqhkiG9w0BAQEFAANLADBIAkEAyyE71AOU3go5XFegLQ6fidI0LhhM\nx7CzpTzh2xWKcHUfbNI7itgJvC/+GlyG5W+DF5V7ba0IJiQLsFve0oLdewIDAQAB\no0IwQDAOBgNVHQ8BAf8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU\nALfqF5ZmfqvqORuJIFilZYKF3d0wDQYJKoZIhvcNAQELBQADQQCXsoezoxXu2CEN\nQdlXZOfmBT6cqxIX/RMHXhpHwRiqPsTO8IO2bVA8CSzxNwMuSv/ZtrMHoh8+PcVW\nHLtkTXH8\n-----END CERTIFICATE-----",
	})
	storeKeyset(t, ctx, keyStore, "etcd-clients-ca", &testingKeyset{
		primaryKey:           "-----BEGIN RSA PRIVATE KEY-----\nMIIBPQIBAAJBANiW3hfHTcKnxCig+uWhpVbOfH1pANKmXVSysPKgE80QSU4tZ6m4\n9pAEeIMsvwvDMaLsb2v6JvXe0qvCmueU+/sCAwEAAQJBAKt/gmpHqP3qA3u8RA5R\n2W6L360Z2Mnza1FmkI/9StCCkJGjuE5yDhxU4JcVnFyX/nMxm2ockEEQDqRSu7Oo\nxTECIQD2QsUsgFL4FnXWzTclySJ6ajE4Cte3gSDOIvyMNMireQIhAOEnsV8UaSI+\nZyL7NMLzMPLCgtsrPnlamr8gdrEHf9ITAiEAxCCLbpTI/4LL2QZZrINTLVGT34Fr\nKl/yI5pjrrp/M2kCIQDfOktQyRuzJ8t5kzWsUxCkntS+FxHJn1rtQ3Jp8dV4oQIh\nAOyiVWDyLZJvg7Y24Ycmp86BZjM9Wk/BfWpBXKnl9iDY\n-----END RSA PRIVATE KEY-----",
		primaryCertificate:   "-----BEGIN CERTIFICATE-----\nMIIBcjCCARygAwIBAgIMFo1ogHnr26DL9YkqMA0GCSqGSIb3DQEBCwUAMBoxGDAW\nBgNVBAMTD2V0Y2QtY2xpZW50cy1jYTAeFw0yMTA2MjgxNjE5MDFaFw0zMTA2Mjgx\nNjE5MDFaMBoxGDAWBgNVBAMTD2V0Y2QtY2xpZW50cy1jYTBcMA0GCSqGSIb3DQEB\nAQUAA0sAMEgCQQDYlt4Xx03Cp8QooPrloaVWznx9aQDSpl1UsrDyoBPNEElOLWep\nuPaQBHiDLL8LwzGi7G9r+ib13tKrwprnlPv7AgMBAAGjQjBAMA4GA1UdDwEB/wQE\nAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBQjlt4Ue54AbJPWlDpRM51s\nx+PeBDANBgkqhkiG9w0BAQsFAANBAAZAdf8ROEVkr3Rf7I+s+CQOil2toadlKWOY\nqCeJ2XaEROfp9aUTEIU1MGM3g57MPyAPPU7mURskuOQz6B1UFaY=\n-----END CERTIFICATE-----",
		secondaryKey:         "-----BEGIN RSA PRIVATE KEY-----\nMIIBPQIBAAJBANiW3hfHTcKnxCig+uWhpVbOfH1pANKmXVSysPKgE80QSU4tZ6m4\n9pAEeIMsvwvDMaLsb2v6JvXe0qvCmueU+/sCAwEAAQJBAKt/gmpHqP3qA3u8RA5R\n2W6L360Z2Mnza1FmkI/9StCCkJGjuE5yDhxU4JcVnFyX/nMxm2ockEEQDqRSu7Oo\nxTECIQD2QsUsgFL4FnXWzTclySJ6ajE4Cte3gSDOIvyMNMireQIhAOEnsV8UaSI+\nZyL7NMLzMPLCgtsrPnlamr8gdrEHf9ITAiEAxCCLbpTI/4LL2QZZrINTLVGT34Fr\nKl/yI5pjrrp/M2kCIQDfOktQyRuzJ8t5kzWsUxCkntS+FxHJn1rtQ3Jp8dV4oQIh\nAOyiVWDyLZJvg7Y24Ycmp86BZjM9Wk/BfWpBXKnl9iDY\n-----END RSA PRIVATE KEY-----",
		secondaryCertificate: "-----BEGIN CERTIFICATE-----\nMIIBcjCCARygAwIBAgIMFo1olfBnC/CsT+dqMA0GCSqGSIb3DQEBCwUAMBoxGDAW\nBgNVBAMTD2V0Y2QtY2xpZW50cy1jYTAeFw0yMTA2MjgxNjIwMzNaFw0zMTA2Mjgx\nNjIwMzNaMBoxGDAWBgNVBAMTD2V0Y2QtY2xpZW50cy1jYTBcMA0GCSqGSIb3DQEB\nAQUAA0sAMEgCQQDYlt4Xx03Cp8QooPrloaVWznx9aQDSpl1UsrDyoBPNEElOLWep\nuPaQBHiDLL8LwzGi7G9r+ib13tKrwprnlPv7AgMBAAGjQjBAMA4GA1UdDwEB/wQE\nAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBQjlt4Ue54AbJPWlDpRM51s\nx+PeBDANBgkqhkiG9w0BAQsFAANBAF1xUz77PlUVUnd9duF8F7plou0TONC9R6/E\nYQ8C6vM1b+9NSDGjCW8YmwEU2fBgskb/BBX2lwVZ32/RUEju4Co=\n-----END CERTIFICATE-----",
	})
	storeKeyset(t, ctx, keyStore, "etcd-manager-ca-events", &testingKeyset{
		primaryKey:           "-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJBAKiC8tndMlEFZ7qzeKxeKqFVjaYpsh/Hg7RxWo15+1kgH3suO0lx\np9+RxSVv97hnsfbySTPZVhy2cIQj7eZtZt8CAwEAAQJASgIRBIw4YAseronKEvHc\niTTY3ERtvbVTa7lpCr+rG03g4l5xgZXCrP+TvZFr04OH4Ka0Qr4QwvT4qTzOx7He\n+QIhANWjbYUnZ73TC5HTlv9CKr7J34rtuG3soz75ihUbX3tlAiEAyezR8MWSqMkv\nN9Yul0a0YsTq7MuSw+iM+bhNxCeAzvMCIQCNANONOcff4sZVFjkn+ozp5aWUNXgv\nnSrVqq+3ZJytfQIgfZ2n1QL0A7B0gWXqwg0oNrGN/BWAjgNjgA5ZwodYqGUCIA+1\nTJZinwh9+JkPJ8CS3xnQBV7OG2b7C+e3kEkdTHFC\n-----END RSA PRIVATE KEY-----",
		primaryCertificate:   "-----BEGIN CERTIFICATE-----\nMIIBgDCCASqgAwIBAgIMFo+bKjm04vB4rNtaMA0GCSqGSIb3DQEBCwUAMCExHzAd\nBgNVBAMTFmV0Y2QtbWFuYWdlci1jYS1ldmVudHMwHhcNMjEwNzA1MjAwOTU2WhcN\nMzEwNzA1MjAwOTU2WjAhMR8wHQYDVQQDExZldGNkLW1hbmFnZXItY2EtZXZlbnRz\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAKiC8tndMlEFZ7qzeKxeKqFVjaYpsh/H\ng7RxWo15+1kgH3suO0lxp9+RxSVv97hnsfbySTPZVhy2cIQj7eZtZt8CAwEAAaNC\nMEAwDgYDVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFBg6\nCEZkQNnRkARBwFce03AEWa+sMA0GCSqGSIb3DQEBCwUAA0EAJMnBThok/uUe8q8O\nsS5q19KUuE8YCTUzMDj36EBKf6NX4NoakCa1h6kfQVtlMtEIMWQZCjbm8xGK5ffs\nGS/VUw==\n-----END CERTIFICATE-----",
		secondaryKey:         "-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJBAKFhHVVxxDGv8d1jBvtdSxz7KIVoBOjLDMxsmTsINiQkTQaFlb+X\nPlnY1ar4+RhE519AFUkqfhypk4Zxqf1YFXUCAwEAAQJAa2aWfycXy3mtHgmpu+B6\n/O6qKR7xJXz9J4+e6wqr/aCca7ArI3T5mOPl/Bud+mC991SEtkIXIGQMNPXgbr5s\ngQIhANKTO1E4/W2Yez/nGBrizWZRjo8NZClT4gxzxV5hFjD3AiEAxDEabVsGlMJR\nwkdX+zEniY1NoHcWE5iJqRwNRfLZffMCIQC5AWgNHV/zKROn+jZAcOF7Ms5oOqC0\neqFQxWozWGMx0wIgaTy1okcbZpw9YusGBJW/UYdcRmDalLRT00Ra0lSL2YUCIDUp\nz1z7kOIHbVyHalFZDv9t1t9wRhBRKPL0ZjSOQwj0\n-----END RSA PRIVATE KEY-----",
		secondaryCertificate: "-----BEGIN CERTIFICATE-----\nMIIBgDCCASqgAwIBAgIMFo+bQ+EgIiBmGghjMA0GCSqGSIb3DQEBCwUAMCExHzAd\nBgNVBAMTFmV0Y2QtbWFuYWdlci1jYS1ldmVudHMwHhcNMjEwNzA1MjAxMTQ2WhcN\nMzEwNzA1MjAxMTQ2WjAhMR8wHQYDVQQDExZldGNkLW1hbmFnZXItY2EtZXZlbnRz\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAKFhHVVxxDGv8d1jBvtdSxz7KIVoBOjL\nDMxsmTsINiQkTQaFlb+XPlnY1ar4+RhE519AFUkqfhypk4Zxqf1YFXUCAwEAAaNC\nMEAwDgYDVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFNuW\nLLH5c8kDubDbr6BHgedW0iJ9MA0GCSqGSIb3DQEBCwUAA0EAiKUoBoaGu7XzboFE\nhjfKlX0TujqWuW3qMxDEJwj4dVzlSLrAoB/G01MJ+xxYKh456n48aG6N827UPXhV\ncPfVNg==\n-----END CERTIFICATE-----",
	})
	storeKeyset(t, ctx, keyStore, "etcd-manager-ca-main", &testingKeyset{
		primaryKey:           "-----BEGIN RSA PRIVATE KEY-----\nMIIBOwIBAAJBAMW5A2xmJgkkoaURt6/pc0zhbo8rq7kX4zoWJmUV+MNVLXecut3V\nHPfLI3PRhlGDB3ftJNapf2uPLRoZyujeoycCAwEAAQJBALIOHMEfdB1DubW3MN3f\ns4+Ga1PPFgPHOT9z9vuNP8pWcRWGACXdln4T/VM5LQYrwTQ/i9EMZycl3ISbTUfy\nEPECIQD5RWUR1dF4S2VGFtxhttbZbP6m3Nk/eiOmT3wPv4TJDQIhAMsPY9YgTmfV\nuZwykVu/UopdjVY/vFAiFYwA2Km8b2gDAiB9jdiUnTA++SrvnMAwb5nUNjQl9ANx\nF6IxOMPyYrMNWQIhALb2wANRCrSeq+ak3bqockwALXi4ZwphG78RiCewhUVXAiA+\n4yljHjbbEGQje8VuxmA3ITMeCwAkIqjXY1Z5DUTnDA==\n-----END RSA PRIVATE KEY-----",
		primaryCertificate:   "-----BEGIN CERTIFICATE-----\nMIIBfDCCASagAwIBAgIMFo+bKjm1c3jfv6hIMA0GCSqGSIb3DQEBCwUAMB8xHTAb\nBgNVBAMTFGV0Y2QtbWFuYWdlci1jYS1tYWluMB4XDTIxMDcwNTIwMDk1NloXDTMx\nMDcwNTIwMDk1NlowHzEdMBsGA1UEAxMUZXRjZC1tYW5hZ2VyLWNhLW1haW4wXDAN\nBgkqhkiG9w0BAQEFAANLADBIAkEAxbkDbGYmCSShpRG3r+lzTOFujyuruRfjOhYm\nZRX4w1Utd5y63dUc98sjc9GGUYMHd+0k1ql/a48tGhnK6N6jJwIDAQABo0IwQDAO\nBgNVHQ8BAf8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUWZLkbBFx\nGAgPU4i62c52unSo7RswDQYJKoZIhvcNAQELBQADQQAj6Pgd0va/8FtkyMlnohLu\nGf4v8RJO6zk3Y6jJ4+cwWziipFM1ielMzSOZfFcCZgH3m5Io40is4hPSqyq2TOA6\n-----END CERTIFICATE-----",
		secondaryKey:         "-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJBAMN9483Hf4qLDdOG9Fl2w7ewdHN7Cd2mn3Biz7xt8UQfTeW2K/fq\nmQKt5swBZMbHJ+I9XHuW9fxikwxAApZmYHUCAwEAAQJAOOGfcBe1L52oRz0ESie5\naPBJ4fQR+dFqoOvPYBdpVRV4h8PcLGhH7H0RO0pJf9ni0MxWDMn2R8Nw6/I7zSgr\n/QIhAN432G6YOItNGj0wrNBgZerFIOVdnHe+higgAhJOtNFbAiEA4TXsL5ALyAYI\nVDS66EbriI15z5XxiauBk0zAbqun7m8CIQDUK+Ichn7GkpGRBx6ZvtDQvfNQzHaO\n5nzVZupTbI68rQIgLzkNU1PTBJgvOujroDTuwm1X820vfnyV6PsZBpu71MUCIAPQ\nTjwL4gGtCZtHXHqAUS9vgf4sQ40oBqNb3NhshheB\n-----END RSA PRIVATE KEY-----",
		secondaryCertificate: "-----BEGIN CERTIFICATE-----\nMIIBfDCCASagAwIBAgIMFo+bQ+Eg8Si30gr4MA0GCSqGSIb3DQEBCwUAMB8xHTAb\nBgNVBAMTFGV0Y2QtbWFuYWdlci1jYS1tYWluMB4XDTIxMDcwNTIwMTE0NloXDTMx\nMDcwNTIwMTE0NlowHzEdMBsGA1UEAxMUZXRjZC1tYW5hZ2VyLWNhLW1haW4wXDAN\nBgkqhkiG9w0BAQEFAANLADBIAkEAw33jzcd/iosN04b0WXbDt7B0c3sJ3aafcGLP\nvG3xRB9N5bYr9+qZAq3mzAFkxscn4j1ce5b1/GKTDEAClmZgdQIDAQABo0IwQDAO\nBgNVHQ8BAf8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUE/h+3gDP\nDvKwHRyiYlXM8voZ1wowDQYJKoZIhvcNAQELBQADQQBXuimeEoAOu5HN4hG7NqL9\nt40K3ZRhRZv3JQWnRVJCBDjg1rD0GQJR/n+DoWvbeijI5C9pNjr2pWSIYR1eYCvd\n-----END CERTIFICATE-----",
	})
	storeKeyset(t, ctx, keyStore, "etcd-peers-ca-events", &testingKeyset{
		primaryKey:           "-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJBAL+YOBxdsZq2MqLiX2PY18dTN4Dyw/6bqb8T2McoycOaTQsuTOVx\nkt4k6kQ+UQxNH1rnVRxWSiyHvFj3NOjQKV8CAwEAAQJATy6MugRq20LDaJffzncW\nrnUQ8kTihX41yBdetuh/gkuyMifMRLi1wVKjrtvIcjhj1vCoCoDLYnUJ/au2rFjO\neQIhAMwZbPwLshFZocs27a+9ngWlF67uHawBsWeC8rddc6u9AiEA8FDBJrDjckMh\ngPoFA29l4JmJTNT16wbBiIopKOwpTUsCIDXDvOHocs//PI+7uIFDAg2an9KFB2v4\nRjNuW2HSTFZBAiA7pD8bpCD+tax1/xcJcDc/k7tgpyXVS5rykR9/+YSSmwIhAIqA\nuHHsA+iviwxdgjDQR8Cc0jWzH9LOC3/AM0+WH4Pe\n-----END RSA PRIVATE KEY-----",
		primaryCertificate:   "-----BEGIN CERTIFICATE-----\nMIIBfDCCASagAwIBAgIMFo+bKjmxTPh3/lYJMA0GCSqGSIb3DQEBCwUAMB8xHTAb\nBgNVBAMTFGV0Y2QtcGVlcnMtY2EtZXZlbnRzMB4XDTIxMDcwNTIwMDk1NloXDTMx\nMDcwNTIwMDk1NlowHzEdMBsGA1UEAxMUZXRjZC1wZWVycy1jYS1ldmVudHMwXDAN\nBgkqhkiG9w0BAQEFAANLADBIAkEAv5g4HF2xmrYyouJfY9jXx1M3gPLD/pupvxPY\nxyjJw5pNCy5M5XGS3iTqRD5RDE0fWudVHFZKLIe8WPc06NApXwIDAQABo0IwQDAO\nBgNVHQ8BAf8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUf6xiDI+O\nYph1ziCGr2hZaQYt+fUwDQYJKoZIhvcNAQELBQADQQBBxj5hqEQstonTb8lnqeGB\nDEYtUeAk4eR/HzvUMjF52LVGuvN3XVt+JTrFeKNvb6/RDUbBNRj3azalcUkpPh6V\n-----END CERTIFICATE-----",
		secondaryKey:         "-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJBAKOTY9go19aqd5hD8NR+ZxwBVi6BjUi0pURSVtNzcWjTzBcy+T6w\nqMjl61/PzFnM7mWMNAq3/BDzjkFotvltFy8CAwEAAQJAUIYQEqsYhZ5pPVXEynZn\nP8wQptgzuuTirp1yDKm53IYNYkRMdPD1XPymeCOvS1lvkwIFCiyuo1EUMQzVowdU\nMQIhAMj9iSDnm2nSzXdv7lOA3hUsh5/sCZbmAHe8+Y3P8LtFAiEA0FhibI6FkmQC\n7/ifuhS90Y3Qmo/B9N8HiFIN84Gm9eMCIC9E2VxAvB8+MY5WZ7GBzDkkmNz2kSbI\n/vEqI3LDpbUVAiEAnhgTR5C2ZqkhWXrtqUQH7bWQ71fas7dxfc3V7EsbqEUCIEv+\nfsV/d2yUde2L5E6eYiL0lZ5DwhKkXOjZlZX7rT8c\n-----END RSA PRIVATE KEY-----",
		secondaryCertificate: "-----BEGIN CERTIFICATE-----\nMIIBfDCCASagAwIBAgIMFo+bQ+Eq69jgzpKwMA0GCSqGSIb3DQEBCwUAMB8xHTAb\nBgNVBAMTFGV0Y2QtcGVlcnMtY2EtZXZlbnRzMB4XDTIxMDcwNTIwMTE0NloXDTMx\nMDcwNTIwMTE0NlowHzEdMBsGA1UEAxMUZXRjZC1wZWVycy1jYS1ldmVudHMwXDAN\nBgkqhkiG9w0BAQEFAANLADBIAkEAo5Nj2CjX1qp3mEPw1H5nHAFWLoGNSLSlRFJW\n03NxaNPMFzL5PrCoyOXrX8/MWczuZYw0Crf8EPOOQWi2+W0XLwIDAQABo0IwQDAO\nBgNVHQ8BAf8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUxauhhKQh\ncvdZND78rHe0RQVTTiswDQYJKoZIhvcNAQELBQADQQB+cq4jIS9q0zXslaRa+ViI\nJ+dviA3sMygbmSJO0s4DxYmoazKJblux5q0ASSvS9iL1l9ShuZ1dWyp2tpZawHyb\n-----END CERTIFICATE-----",
	})
	storeKeyset(t, ctx, keyStore, "etcd-peers-ca-main", &testingKeyset{
		primaryKey:           "-----BEGIN RSA PRIVATE KEY-----\nMIIBOwIBAAJBALJFpdanCA3og1CrCz2n8G88SUm/ZGej11VMWGVCoMBpQld7swGa\nI7g0lxbvoSjN4GHnO1Hf/g0TUUzbHxOKxLcCAwEAAQJBAI418S1i4ZH2wYpAaB8v\nMSYLOYuTGk1y7fwlgv6EQCg8esJcMCeDsqT5V5sUicT6jT5m3KdpKA4v4kpZJzHo\nr8ECIQDRtEmpTSmTQ1FAVPu34j6ZU0W5zT8RMaoUFPCXPJ/M9QIhANmg7bTqNNBY\nd7TUxmgm2NW5GDn0yyg1WqoIL4wOJz97AiBvrCad9e1x8qNOMvNpVR4o4GN9MoOn\nUF9WGmCU6T/gEQIgdhnEBdK3eH0Z8TMqvKigMVNyFzmF6jsSCYXJr7qah/MCIQCy\npxPa6cKMC0n9t61B+1f7O2yCvwllormxaFYVm9J4xw==\n-----END RSA PRIVATE KEY-----",
		primaryCertificate:   "-----BEGIN CERTIFICATE-----\nMIIBeDCCASKgAwIBAgIMFo+bKjmuLDDLcDHsMA0GCSqGSIb3DQEBCwUAMB0xGzAZ\nBgNVBAMTEmV0Y2QtcGVlcnMtY2EtbWFpbjAeFw0yMTA3MDUyMDA5NTZaFw0zMTA3\nMDUyMDA5NTZaMB0xGzAZBgNVBAMTEmV0Y2QtcGVlcnMtY2EtbWFpbjBcMA0GCSqG\nSIb3DQEBAQUAA0sAMEgCQQCyRaXWpwgN6INQqws9p/BvPElJv2Rno9dVTFhlQqDA\naUJXe7MBmiO4NJcW76EozeBh5ztR3/4NE1FM2x8TisS3AgMBAAGjQjBAMA4GA1Ud\nDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBQtE1d49uSvpURf\nOQ25Vlu6liY20DANBgkqhkiG9w0BAQsFAANBAAgLVaetJZcfOA3OIMMvQbz2Ydrt\nuWF9BKkIad8jrcIrm3IkOtR8bKGmDIIaRKuG/ZUOL6NMe2fky3AAfKwleL4=\n-----END CERTIFICATE-----",
		secondaryKey:         "-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJBALE1vJwNk3HlXVk6JfFlK9oWkdHAp4cN9y4xSK12g+2dpUyUxMYN\nYAy4JWYUcUBaiEhjKd6YR6CZmRnXlLsASt8CAwEAAQJABeku812Yj3IBHRrNbTHc\ntpeOIZr1e5HBru7B59dOKzzKrI2SozD+wKmhi2r+8yPkdU1nq4DE1Pboc1BmPh9C\n0QIhAMiAQ+yZRuThl8qOCZ+D9Frmml102DIf5d1NjGGQD84FAiEA4kMJCM194VPV\n2W7QsLH+szbwRHXg1dOlR9WQHJ8rZpMCIF/F7SwyV0vzerdVu8EHngxhxPDJZJAk\n7n8UkO71iqclAiEAypza9z4E7oWDZ507Vi9edJ/K0pN4jiJjzIrq7SZ/1+8CID2K\nAMbqYsKhlMt8zM+hSUg+u8wcWs8CVBb4ozQY2Xyb\n-----END RSA PRIVATE KEY-----",
		secondaryCertificate: "-----BEGIN CERTIFICATE-----\nMIIBeDCCASKgAwIBAgIMFo+bQ+EuVthBfuZvMA0GCSqGSIb3DQEBCwUAMB0xGzAZ\nBgNVBAMTEmV0Y2QtcGVlcnMtY2EtbWFpbjAeFw0yMTA3MDUyMDExNDZaFw0zMTA3\nMDUyMDExNDZaMB0xGzAZBgNVBAMTEmV0Y2QtcGVlcnMtY2EtbWFpbjBcMA0GCSqG\nSIb3DQEBAQUAA0sAMEgCQQCxNbycDZNx5V1ZOiXxZSvaFpHRwKeHDfcuMUitdoPt\nnaVMlMTGDWAMuCVmFHFAWohIYynemEegmZkZ15S7AErfAgMBAAGjQjBAMA4GA1Ud\nDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBTAjQ8T4HclPIsC\nqipEfUIcLP6jqTANBgkqhkiG9w0BAQsFAANBAJdZ17TN3HlWrH7HQgfR12UBwz8K\nG9DurDznVaBVUYaHY8Sg5AvAXeb+yIF2JMmRR+bK+/G1QYY2D3/P31Ic2Oo=\n-----END CERTIFICATE-----",
	})
	storeKeyset(t, ctx, keyStore, "service-account", &testingKeyset{
		primaryKey:           "-----BEGIN RSA PRIVATE KEY-----\nMIIBPQIBAAJBANiW3hfHTcKnxCig+uWhpVbOfH1pANKmXVSysPKgE80QSU4tZ6m4\n9pAEeIMsvwvDMaLsb2v6JvXe0qvCmueU+/sCAwEAAQJBAKt/gmpHqP3qA3u8RA5R\n2W6L360Z2Mnza1FmkI/9StCCkJGjuE5yDhxU4JcVnFyX/nMxm2ockEEQDqRSu7Oo\nxTECIQD2QsUsgFL4FnXWzTclySJ6ajE4Cte3gSDOIvyMNMireQIhAOEnsV8UaSI+\nZyL7NMLzMPLCgtsrPnlamr8gdrEHf9ITAiEAxCCLbpTI/4LL2QZZrINTLVGT34Fr\nKl/yI5pjrrp/M2kCIQDfOktQyRuzJ8t5kzWsUxCkntS+FxHJn1rtQ3Jp8dV4oQIh\nAOyiVWDyLZJvg7Y24Ycmp86BZjM9Wk/BfWpBXKnl9iDY\n-----END RSA PRIVATE KEY-----",
		primaryCertificate:   "-----BEGIN CERTIFICATE-----\nMIIBZzCCARGgAwIBAgIBAjANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQDEw9zZXJ2\naWNlLWFjY291bnQwHhcNMjEwNTAyMjAzMDA2WhcNMzEwNTAyMjAzMDA2WjAaMRgw\nFgYDVQQDEw9zZXJ2aWNlLWFjY291bnQwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA\n2JbeF8dNwqfEKKD65aGlVs58fWkA0qZdVLKw8qATzRBJTi1nqbj2kAR4gyy/C8Mx\nouxva/om9d7Sq8Ka55T7+wIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAQYwDwYDVR0T\nAQH/BAUwAwEB/zAdBgNVHQ4EFgQUI5beFHueAGyT1pQ6UTOdbMfj3gQwDQYJKoZI\nhvcNAQELBQADQQBwPLO+Np8o6k3aNBGKE4JTCOs06X72OXNivkWWWP/9XGz6x4DI\nHPU65kbUn/pWXBUVVlpsKsdmWA2Bu8pd/vD+\n-----END CERTIFICATE-----\n",
		secondaryKey:         "-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJBAKOE64nZbH+GM91AIrqf7HEk4hvzqsZFFtxc+8xir1XC3mI/RhCC\nrs6AdVRZNZ26A6uHArhi33c2kHQkCjyLA7sCAwEAAQJAejInjmEzqmzQr0NxcIN4\nPukwK3FBKl+RAOZfqNIKcww14mfOn7Gc6lF2zEC4GnLiB3tthbSXoBGi54nkW4ki\nyQIhANZNne9UhQlwyjsd3WxDWWrl6OOZ3J8ppMOIQni9WRLlAiEAw1XEdxPOSOSO\nB6rucpTT1QivVvyEFIb/ukvPm769Mh8CIQDNQwKnHdlfNX0+KljPPaMD1LrAZbr/\naC+8aWLhqtsKUQIgF7gUcTkwdV17eabh6Xv09Qtm7zMefred2etWvFy+8JUCIECv\nFYOKQVWHX+Q7CHX2K1oTECVnZuW1UItdDYVlFYxQ\n-----END RSA PRIVATE KEY-----",
		secondaryCertificate: "-----BEGIN CERTIFICATE-----\nMIIBZzCCARGgAwIBAgIBBDANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQDEw9zZXJ2\naWNlLWFjY291bnQwHhcNMjEwNTAyMjAzMjE3WhcNMzEwNTAyMjAzMjE3WjAaMRgw\nFgYDVQQDEw9zZXJ2aWNlLWFjY291bnQwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA\no4Tridlsf4Yz3UAiup/scSTiG/OqxkUW3Fz7zGKvVcLeYj9GEIKuzoB1VFk1nboD\nq4cCuGLfdzaQdCQKPIsDuwIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAQYwDwYDVR0T\nAQH/BAUwAwEB/zAdBgNVHQ4EFgQUhPbxEmUbwVOCa+fZgxreFhf67UEwDQYJKoZI\nhvcNAQELBQADQQALMsyK2Q7C/bk27eCvXyZKUfrLvor10hEjwGhv14zsKWDeTj/J\nA1LPYp7U9VtFfgFOkVbkLE9Rstc0ltNrPqxA\n-----END CERTIFICATE-----\n",
	})
	if i.ciliumEtcd {
		storeKeyset(t, ctx, keyStore, "etcd-clients-ca-cilium", &testingKeyset{
			primaryKey:           "-----BEGIN RSA PRIVATE KEY-----\nMIIBPQIBAAJBANiW3hfHTcKnxCig+uWhpVbOfH1pANKmXVSysPKgE80QSU4tZ6m4\n9pAEeIMsvwvDMaLsb2v6JvXe0qvCmueU+/sCAwEAAQJBAKt/gmpHqP3qA3u8RA5R\n2W6L360Z2Mnza1FmkI/9StCCkJGjuE5yDhxU4JcVnFyX/nMxm2ockEEQDqRSu7Oo\nxTECIQD2QsUsgFL4FnXWzTclySJ6ajE4Cte3gSDOIvyMNMireQIhAOEnsV8UaSI+\nZyL7NMLzMPLCgtsrPnlamr8gdrEHf9ITAiEAxCCLbpTI/4LL2QZZrINTLVGT34Fr\nKl/yI5pjrrp/M2kCIQDfOktQyRuzJ8t5kzWsUxCkntS+FxHJn1rtQ3Jp8dV4oQIh\nAOyiVWDyLZJvg7Y24Ycmp86BZjM9Wk/BfWpBXKnl9iDY\n-----END RSA PRIVATE KEY-----",
			primaryCertificate:   "-----BEGIN CERTIFICATE-----\nMIIBgDCCASqgAwIBAgIMFotPsR9PsbCKkTJsMA0GCSqGSIb3DQEBCwUAMCExHzAd\nBgNVBAMTFmV0Y2QtY2xpZW50cy1jYS1jaWxpdW0wHhcNMjEwNjIxMjAyMTUyWhcN\nMzEwNjIxMjAyMTUyWjAhMR8wHQYDVQQDExZldGNkLWNsaWVudHMtY2EtY2lsaXVt\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBANiW3hfHTcKnxCig+uWhpVbOfH1pANKm\nXVSysPKgE80QSU4tZ6m49pAEeIMsvwvDMaLsb2v6JvXe0qvCmueU+/sCAwEAAaNC\nMEAwDgYDVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFCOW\n3hR7ngBsk9aUOlEznWzH494EMA0GCSqGSIb3DQEBCwUAA0EAR4UEW5ZK+NVtqm7s\nHF/JbSYPd+BhcNaJVOv8JP+/CGfCOXOmxjpZICSYQqe6UjjjP7fbJy8FANTpKTuJ\nUQC1kQ==\n-----END CERTIFICATE-----",
			secondaryKey:         "-----BEGIN RSA PRIVATE KEY-----\nMIIBPQIBAAJBANiW3hfHTcKnxCig+uWhpVbOfH1pANKmXVSysPKgE80QSU4tZ6m4\n9pAEeIMsvwvDMaLsb2v6JvXe0qvCmueU+/sCAwEAAQJBAKt/gmpHqP3qA3u8RA5R\n2W6L360Z2Mnza1FmkI/9StCCkJGjuE5yDhxU4JcVnFyX/nMxm2ockEEQDqRSu7Oo\nxTECIQD2QsUsgFL4FnXWzTclySJ6ajE4Cte3gSDOIvyMNMireQIhAOEnsV8UaSI+\nZyL7NMLzMPLCgtsrPnlamr8gdrEHf9ITAiEAxCCLbpTI/4LL2QZZrINTLVGT34Fr\nKl/yI5pjrrp/M2kCIQDfOktQyRuzJ8t5kzWsUxCkntS+FxHJn1rtQ3Jp8dV4oQIh\nAOyiVWDyLZJvg7Y24Ycmp86BZjM9Wk/BfWpBXKnl9iDY\n-----END RSA PRIVATE KEY-----",
			secondaryCertificate: "-----BEGIN CERTIFICATE-----\nMIIBgDCCASqgAwIBAgIMFotP940EXpD3N1D7MA0GCSqGSIb3DQEBCwUAMCExHzAd\nBgNVBAMTFmV0Y2QtY2xpZW50cy1jYS1jaWxpdW0wHhcNMjEwNjIxMjAyNjU1WhcN\nMzEwNjIxMjAyNjU1WjAhMR8wHQYDVQQDExZldGNkLWNsaWVudHMtY2EtY2lsaXVt\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBANiW3hfHTcKnxCig+uWhpVbOfH1pANKm\nXVSysPKgE80QSU4tZ6m49pAEeIMsvwvDMaLsb2v6JvXe0qvCmueU+/sCAwEAAaNC\nMEAwDgYDVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFCOW\n3hR7ngBsk9aUOlEznWzH494EMA0GCSqGSIb3DQEBCwUAA0EARXoKy6mExpD6tHFO\nCN3ZGNZ5BsHl5W5y+gwUuVskgC7xt/bgTuXm5hz8TLgnG5kYtG4uxjFg4yCvtNg2\nMQNfAQ==\n-----END CERTIFICATE-----",
		})
		storeKeyset(t, ctx, keyStore, "etcd-manager-ca-cilium", &testingKeyset{
			primaryKey:           "-----BEGIN RSA PRIVATE KEY-----\nMIIBOwIBAAJBAMHrFsj6jdcV2UZnTJmqNdbz7kQjh0NW0PrIWcRAD6Y1q9/Nvbnd\nWF8jGay206KXJk1r/qHXyDuwHCKgZkfbnS0CAwEAAQJAbmWl/RkXMwHPRlN8uma6\na/tHBCet09pS8tKouB84SYh61MmgKnd+IGVmoUA18zSSOVYkueiHxUjVNIx5Oe6b\nwQIhANfLXoFFoW2MHXEgTmZV3N8t/zcpWk24PfjuoutR1YSFAiEA5gxOtNgVfTv6\nUPb1zixknCLy/QRUyuA1UH4mlPMIiokCIQCZq7t692kDp/n3a3gpLBAD5q+OSqaC\nHigTs2zVgws4OQIgZ86j8X0UbVeUQ9a84pUrrT0kEsJSlN2JkVHrjQkCEKkCIQCs\ngOQHglDw6452+lc/qokpE4vGEyrm6uyMj07Uz4KY6A==\n-----END RSA PRIVATE KEY-----",
			primaryCertificate:   "-----BEGIN CERTIFICATE-----\nMIIBgDCCASqgAwIBAgIMFo+bv6kG/ijs2GJsMA0GCSqGSIb3DQEBCwUAMCExHzAd\nBgNVBAMTFmV0Y2QtbWFuYWdlci1jYS1jaWxpdW0wHhcNMjEwNzA1MjAyMDM3WhcN\nMzEwNzA1MjAyMDM3WjAhMR8wHQYDVQQDExZldGNkLW1hbmFnZXItY2EtY2lsaXVt\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAMHrFsj6jdcV2UZnTJmqNdbz7kQjh0NW\n0PrIWcRAD6Y1q9/NvbndWF8jGay206KXJk1r/qHXyDuwHCKgZkfbnS0CAwEAAaNC\nMEAwDgYDVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFDKE\nITER3OCn4C7w9YVi2YdHDUkJMA0GCSqGSIb3DQEBCwUAA0EAo2zLlhHTpYlTM7dh\netdG+8zu6GpzoNs6caeYT1F7LCUp5CX8T05QVHZNSwTU41wFFu3nRa5Fr8/2nB+M\nEcE5pA==\n-----END CERTIFICATE-----",
			secondaryKey:         "-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJBAKObYoPZoxsjbLbCy1tA2JyHFKEPHg3XgOPCmQLAYvnDOIxAewih\nwpdjjcuJP+xoz0vUA+fcJaBei/3lAFNV0MUCAwEAAQJASYREM20zfrlfW4ySppGw\nBD4qxeiuH5gr4ayK5xKeJw6bHCh/bdUn5SPFY3PWzqj/RsvegNSZyNU7rfOFWV1n\nbQIhAMP2awFys/VQeokXH4hIXX6lreLnNWaCX9gVvkUvbWJbAiEA1btHLJj+EZ5m\nQPZvLJ469ASs4F0yMbjKer+xPhnpw18CIG2tVWaSFDaQvIRN9NAJ8IoZoKEGVtTw\n00PVp5CBYu9RAiAeoSgiDArdG4Yr6SUlj8eDEOh1fuWimojp7m7IJ46IoQIhAIO0\nJpW2I4J+WHOqUKJVjugNtBSqNDF5mDXINHo7U/gO\n-----END RSA PRIVATE KEY-----",
			secondaryCertificate: "-----BEGIN CERTIFICATE-----\nMIIBgDCCASqgAwIBAgIMFo+b23aziPjha6o+MA0GCSqGSIb3DQEBCwUAMCExHzAd\nBgNVBAMTFmV0Y2QtbWFuYWdlci1jYS1jaWxpdW0wHhcNMjEwNzA1MjAyMjM3WhcN\nMzEwNzA1MjAyMjM3WjAhMR8wHQYDVQQDExZldGNkLW1hbmFnZXItY2EtY2lsaXVt\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAKObYoPZoxsjbLbCy1tA2JyHFKEPHg3X\ngOPCmQLAYvnDOIxAewihwpdjjcuJP+xoz0vUA+fcJaBei/3lAFNV0MUCAwEAAaNC\nMEAwDgYDVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFOBa\nmp4zlA4aPNrVCZgS+Ot9sG5BMA0GCSqGSIb3DQEBCwUAA0EABBJLTr+G+TxDLF3E\nJyV/pgEM/QggrBJozK1bWCvxIUKsnZHiX6E/WVeDeT1QlM1HaxumLGMsKAAyxPV4\nGY7LCw==\n-----END CERTIFICATE-----",
		})
		storeKeyset(t, ctx, keyStore, "etcd-peers-ca-cilium", &testingKeyset{
			primaryKey:           "-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJBANiACqgi/3txqkMV6kTSMA1ZR6M3ul4QiGthUuW7TPKkNHhnq5rR\nFdyhLcQJYsetmVR2TrgH0hQD9Nofn5H5yWkCAwEAAQJBAJEjbYGATOPVtH3a0D2o\n5vvb8XGTJ4Zt8PaDvU4zfYdfoAGpL/Pq3QijpESEKX9t4+sh4w94dG7oDpniGCvV\nO4ECIQDsUkKcDiNKH7TxZxYLx9MYEIXMQK/71ge+QHN9DSSQeQIhAOqHP0EhCqtZ\niYHYvPnO4gf4Du+eCqlfrb2u3z3FbSRxAiBPn1OkArtvIQm1ADeUVopQJFkAPZdN\nsYpAVrTSoFf+eQIgOCMNcgJ9skwpTOpbOZRaqDupH5P9y1L6nGeqSffiyxECIF2N\nrfTIH7lUlRexa0ExTFVRnblo9qawPxhWQkd2u3En\n-----END RSA PRIVATE KEY-----",
			primaryCertificate:   "-----BEGIN CERTIFICATE-----\nMIIBfDCCASagAwIBAgIMFo+bv6kGnIBWECkZMA0GCSqGSIb3DQEBCwUAMB8xHTAb\nBgNVBAMTFGV0Y2QtcGVlcnMtY2EtY2lsaXVtMB4XDTIxMDcwNTIwMjAzN1oXDTMx\nMDcwNTIwMjAzN1owHzEdMBsGA1UEAxMUZXRjZC1wZWVycy1jYS1jaWxpdW0wXDAN\nBgkqhkiG9w0BAQEFAANLADBIAkEA2IAKqCL/e3GqQxXqRNIwDVlHoze6XhCIa2FS\n5btM8qQ0eGermtEV3KEtxAlix62ZVHZOuAfSFAP02h+fkfnJaQIDAQABo0IwQDAO\nBgNVHQ8BAf8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUfr/92gfR\nqn/blYJEH3A38U51A8AwDQYJKoZIhvcNAQELBQADQQCC6qoc1PX3AXOtt+lqTtu0\noHrjU5/YXFbqDxEh/VdGYhqtpg3YuoHWAp3JDg1RVW1SRfUx30/375hoB5Nrw/5S\n-----END CERTIFICATE-----",
			secondaryKey:         "-----BEGIN RSA PRIVATE KEY-----\nMIIBOwIBAAJBAMN09qchDoATwSsKH7iCy6JD8QBaZVc3bueNH6ERCeIlaoq6FJbM\n9RvdJhMJqkfge/9JLe9L3vYuWehO0M9p0GkCAwEAAQJAIhzRx41/aF8KQaa8rok1\nXRaag0NDmJs2IfeBY60DmpI66uTtDHhpwxC9p6XDWdxcv0FJma0CHoTEksg8GDm5\nGQIhANFFU345K3Aezn6oeoT7vV0iAj0PRqEwiJ2f7l0lhtUHAiEA7xn76xIsJUCB\nAeshuO83KSsei6Traudg/+4G3H0Jww8CIQC8hLVIOfwVjsr6co+ciKL36REXLFG2\nF2Cajl5ObuXdtQIgCpoiW4gQwQ4dKlKcyjCBR6gL0LFdZv4fhPmvADPjLO0CIQCT\nNBQjZG61HYyhBYaexj+ZVleuheY6re75KkncxUYwNw==\n-----END RSA PRIVATE KEY-----",
			secondaryCertificate: "-----BEGIN CERTIFICATE-----\nMIIBfDCCASagAwIBAgIMFo+b23acX0hZEkbkMA0GCSqGSIb3DQEBCwUAMB8xHTAb\nBgNVBAMTFGV0Y2QtcGVlcnMtY2EtY2lsaXVtMB4XDTIxMDcwNTIwMjIzN1oXDTMx\nMDcwNTIwMjIzN1owHzEdMBsGA1UEAxMUZXRjZC1wZWVycy1jYS1jaWxpdW0wXDAN\nBgkqhkiG9w0BAQEFAANLADBIAkEAw3T2pyEOgBPBKwofuILLokPxAFplVzdu540f\noREJ4iVqiroUlsz1G90mEwmqR+B7/0kt70ve9i5Z6E7Qz2nQaQIDAQABo0IwQDAO\nBgNVHQ8BAf8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU0hyEvGir\n2ucsJrojyZaDBIb8JLAwDQYJKoZIhvcNAQELBQADQQA9vQylgkvgROIMspzOlbZr\nZwsTAzp9J2ZxZL06AQ9iWzpvIw/H3oClV63q6zN2aHtpBTkhUOSX3Q4L/X/0MOkj\n-----END CERTIFICATE-----",
		})
	}

	return factory
}

type testingKeyset struct {
	primaryKey           string
	primaryCertificate   string
	secondaryKey         string
	secondaryCertificate string
}

func storeKeyset(t *testing.T, ctx context.Context, keyStore fi.Keystore, name string, testingKeyset *testingKeyset) {
	{
		privateKey, err := pki.ParsePEMPrivateKey([]byte(testingKeyset.primaryKey))
		if err != nil {
			t.Fatalf("error loading private key %v", err)
		}

		cert, err := pki.ParsePEMCertificate([]byte(testingKeyset.primaryCertificate))
		if err != nil {
			t.Fatalf("error loading certificate %v", err)
		}

		keyset, err := fi.NewKeyset(cert, privateKey)
		if err != nil {
			t.Fatalf("error creating keyset: %v", err)
		}

		if testingKeyset.secondaryKey != "" {
			privateKey, err = pki.ParsePEMPrivateKey([]byte(testingKeyset.secondaryKey))
			if err != nil {
				t.Fatalf("error loading private key %v", err)
			}

			cert, err = pki.ParsePEMCertificate([]byte(testingKeyset.secondaryCertificate))
			if err != nil {
				t.Fatalf("error loading certificate %v", err)
			}

			_, _ = keyset.AddItem(cert, privateKey, false)
		}

		err = keyStore.StoreKeyset(ctx, name, keyset)
		if err != nil {
			t.Fatalf("error storing user provided keys: %v", err)
		}
	}
}

func (i *integrationTest) runTestTerraformAWS(t *testing.T) {
	ctx := testcontext.ForTest(t)
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.21.0-alpha.1")
	h.SetupMockAWS()

	expectedFilenames := i.expectTerraformFilenames
	expectedFilenames = append(expectedFilenames,
		"aws_launch_template_nodes."+i.clusterName+"_user_data",
		"aws_s3_object_cluster-completed.spec_content",
		"aws_s3_object_etcd-cluster-spec-events_content",
		"aws_s3_object_etcd-cluster-spec-main_content",
		"aws_s3_object_kops-version.txt_content",
		"aws_s3_object_manifests-static-kube-apiserver-healthcheck_content",
		"aws_s3_object_nodeupconfig-nodes_content",
		"aws_s3_object_"+i.clusterName+"-addons-bootstrap_content",
		"aws_s3_object_"+i.clusterName+"-addons-kops-controller.addons.k8s.io-k8s-1.16_content",
		"aws_s3_object_"+i.clusterName+"-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_content",
		"aws_s3_object_"+i.clusterName+"-addons-limit-range.addons.k8s.io_content",
		"aws_s3_object_"+i.clusterName+"-addons-storage-aws.addons.k8s.io-v1.15.0_content")

	if i.kubeDNS {
		expectedFilenames = append(expectedFilenames, "aws_s3_object_"+i.clusterName+"-addons-kube-dns.addons.k8s.io-k8s-1.12_content")
	} else {
		expectedFilenames = append(expectedFilenames, "aws_s3_object_"+i.clusterName+"-addons-coredns.addons.k8s.io-k8s-1.12_content")
	}

	if i.discovery {
		expectedFilenames = append(expectedFilenames,
			"aws_s3_object_discovery.json_content",
			"aws_s3_object_keys.json_content")
	}

	if i.sshKey {
		expectedFilenames = append(expectedFilenames, "aws_key_pair_kubernetes."+i.clusterName+"-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")
	}

	masterRole := truncate.TruncateString("masters."+i.clusterName, truncate.TruncateStringOptions{MaxLength: iam.MaxLengthIAMRoleName, AlwaysAddHash: false})
	nodeRole := truncate.TruncateString("nodes."+i.clusterName, truncate.TruncateStringOptions{MaxLength: iam.MaxLengthIAMRoleName, AlwaysAddHash: false})

	for j := 0; j < i.zones; j++ {
		zone := "us-test-1" + string([]byte{byte('a') + byte(j)})
		expectedFilenames = append(expectedFilenames,
			"aws_s3_object_manifests-etcdmanager-events-master-"+zone+"_content",
			"aws_s3_object_manifests-etcdmanager-main-master-"+zone+"_content",
			"aws_s3_object_nodeupconfig-master-"+zone+"_content",
			"aws_launch_template_master-"+zone+".masters."+i.clusterName+"_user_data")
	}

	if i.expectPolicies {
		expectedFilenames = append(expectedFilenames, []string{
			"aws_iam_role_" + masterRole + "_policy",
			"aws_iam_role_" + nodeRole + "_policy",
			"aws_iam_role_policy_" + masterRole + "_policy",
			"aws_iam_role_policy_" + nodeRole + "_policy",
		}...)
		if i.private {
			expectedFilenames = append(expectedFilenames, []string{
				"aws_iam_role_bastions." + i.clusterName + "_policy",
				"aws_iam_role_policy_bastions." + i.clusterName + "_policy",
			}...)
			if i.bastionUserData {
				expectedFilenames = append(expectedFilenames,
					"aws_s3_object_nodeupconfig-bastion_content",
					"aws_launch_template_bastion."+i.clusterName+"_user_data")
			}
		}
	}
	if i.nth {
		expectedFilenames = append(expectedFilenames, []string{
			"aws_s3_object_" + i.clusterName + "-addons-node-termination-handler.aws-k8s-1.11_content",
			"aws_cloudwatch_event_rule_" + awsup.GetClusterName40(i.clusterName) + "-ASGLifecycle_event_pattern",
			"aws_cloudwatch_event_rule_" + awsup.GetClusterName40(i.clusterName) + "-SpotInterruption_event_pattern",
			"aws_cloudwatch_event_rule_" + awsup.GetClusterName40(i.clusterName) + "-InstanceStateChange_event_pattern",
			"aws_cloudwatch_event_rule_" + awsup.GetClusterName40(i.clusterName) + "-InstanceScheduledChange_event_pattern",
			"aws_sqs_queue_" + strings.Replace(i.clusterName, ".", "-", -1) + "-nth_policy",
		}...)
	}
	if i.nthRebalance {
		expectedFilenames = append(expectedFilenames, "aws_cloudwatch_event_rule_"+awsup.GetClusterName40(i.clusterName)+"-RebalanceRecommendation_event_pattern")
	}
	expectedFilenames = append(expectedFilenames, i.expectServiceAccountRolePolicies...)

	i.runTest(t, ctx, h, expectedFilenames, "", "", nil)
}

func (i *integrationTest) runTestPhase(t *testing.T, phase cloudup.Phase) {
	ctx := testcontext.ForTest(t)
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.21.0-alpha.1")
	h.SetupMockAWS()
	phaseName := string(phase)
	if phaseName == "" {
		t.Fatalf("phase must be set")
	}
	tfFileName := phaseName + "-kubernetes.tf"

	expectedFilenames := i.expectTerraformFilenames

	if phase == cloudup.PhaseSecurity {
		expectedFilenames = []string{
			"aws_iam_role_masters." + i.clusterName + "_policy",
			"aws_iam_role_nodes." + i.clusterName + "_policy",
			"aws_iam_role_policy_masters." + i.clusterName + "_policy",
			"aws_iam_role_policy_nodes." + i.clusterName + "_policy",
			"aws_key_pair_kubernetes." + i.clusterName + "-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key",
		}
		if i.private {
			expectedFilenames = append(expectedFilenames, []string{
				"aws_iam_role_bastions." + i.clusterName + "_policy",
				"aws_iam_role_policy_bastions." + i.clusterName + "_policy",
				"aws_launch_template_bastion." + i.clusterName + "_user_data",
			}...)
		}
	} else if phase == cloudup.PhaseCluster {
		expectedFilenames = []string{
			"aws_launch_template_nodes." + i.clusterName + "_user_data",
		}

		for j := 0; j < i.zones; j++ {
			zone := "us-test-1" + string([]byte{byte('a') + byte(j)})
			s := "aws_launch_template_master-" + zone + ".masters." + i.clusterName + "_user_data"
			expectedFilenames = append(expectedFilenames, s)
		}
	}

	i.runTest(t, ctx, h, expectedFilenames, tfFileName, "", &phase)
}

func (i *integrationTest) runTestTerraformGCE(t *testing.T) {
	ctx := testcontext.ForTest(t)
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.21.0-alpha.1")
	h.SetupMockGCE()

	expectedFilenames := i.expectTerraformFilenames

	prefix := "google_compute_instance_template_nodes-" + gce.SafeClusterName(i.clusterName) + "_metadata_"
	if !i.startupScript {
		expectedFilenames = append(expectedFilenames, prefix+"user-data")
	} else {
		expectedFilenames = append(expectedFilenames, prefix+"startup-script")
	}

	expectedFilenames = append(expectedFilenames,
		"aws_s3_object_cluster-completed.spec_content",
		"aws_s3_object_etcd-cluster-spec-events_content",
		"aws_s3_object_etcd-cluster-spec-main_content",
		"aws_s3_object_kops-version.txt_content",
		"aws_s3_object_manifests-static-kube-apiserver-healthcheck_content",
		"aws_s3_object_nodeupconfig-nodes_content",
		"aws_s3_object_"+i.clusterName+"-addons-bootstrap_content",
		"aws_s3_object_"+i.clusterName+"-addons-coredns.addons.k8s.io-k8s-1.12_content",
		"aws_s3_object_"+i.clusterName+"-addons-kops-controller.addons.k8s.io-k8s-1.16_content",
		"aws_s3_object_"+i.clusterName+"-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_content",
		"aws_s3_object_"+i.clusterName+"-addons-limit-range.addons.k8s.io_content",
		"aws_s3_object_"+i.clusterName+"-addons-metadata-proxy.addons.k8s.io-v0.1.12_content",
		"aws_s3_object_"+i.clusterName+"-addons-storage-gce.addons.k8s.io-v1.7.0_content")

	for j := 0; j < i.zones; j++ {
		zone := "us-test1-" + string([]byte{byte('a') + byte(j)})

		expectedFilenames = append(expectedFilenames, "aws_s3_object_manifests-etcdmanager-events-master-"+zone+"_content")
		expectedFilenames = append(expectedFilenames, "aws_s3_object_manifests-etcdmanager-main-master-"+zone+"_content")
		expectedFilenames = append(expectedFilenames, "aws_s3_object_nodeupconfig-master-"+zone+"_content")

		prefix := "google_compute_instance_template_master-" + zone + "-" + gce.SafeClusterName(i.clusterName) + "_metadata_"
		if !i.startupScript {
			expectedFilenames = append(expectedFilenames, prefix+"user-data")
		} else {
			expectedFilenames = append(expectedFilenames, prefix+"startup-script")
		}
	}

	i.runTest(t, ctx, h, expectedFilenames, "", "", nil)
}

func (i *integrationTest) runTestTerraformHetzner(t *testing.T) {
	ctx := testcontext.ForTest(t)
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.21.0-alpha.1")

	expectedFilenames := i.expectTerraformFilenames

	expectedFilenames = append(expectedFilenames,
		"aws_s3_object_cluster-completed.spec_content",
		"aws_s3_object_etcd-cluster-spec-events_content",
		"aws_s3_object_etcd-cluster-spec-main_content",
		"aws_s3_object_kops-version.txt_content",
		"aws_s3_object_manifests-etcdmanager-events-master-fsn1_content",
		"aws_s3_object_manifests-etcdmanager-main-master-fsn1_content",
		"aws_s3_object_manifests-static-kube-apiserver-healthcheck_content",
		"aws_s3_object_nodeupconfig-master-fsn1_content",
		"aws_s3_object_nodeupconfig-nodes-fsn1_content",
		"aws_s3_object_"+i.clusterName+"-addons-bootstrap_content",
		"aws_s3_object_"+i.clusterName+"-addons-coredns.addons.k8s.io-k8s-1.12_content",
		"aws_s3_object_"+i.clusterName+"-addons-hcloud-cloud-controller.addons.k8s.io-k8s-1.22_content",
		"aws_s3_object_"+i.clusterName+"-addons-hcloud-csi-driver.addons.k8s.io-k8s-1.22_content",
		"aws_s3_object_"+i.clusterName+"-addons-kops-controller.addons.k8s.io-k8s-1.16_content",
		"aws_s3_object_"+i.clusterName+"-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_content",
		"aws_s3_object_"+i.clusterName+"-addons-limit-range.addons.k8s.io_content",
		"hcloud_server_master-fsn1_user_data",
		"hcloud_server_nodes-fsn1_user_data",
	)

	i.runTest(t, ctx, h, expectedFilenames, "", "", nil)
}

func (i *integrationTest) runTestTerraformScaleway(t *testing.T) {
	featureflag.ParseFlags("+Scaleway")
	unsetFeatureFlags := func() {
		featureflag.ParseFlags("-Scaleway")
	}
	defer unsetFeatureFlags()

	ctx := testcontext.ForTest(t)
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.21.0-alpha.1")

	expectedFilenames := i.expectTerraformFilenames

	expectedFilenames = append(expectedFilenames,
		"aws_s3_object_cluster-completed.spec_content",
		"aws_s3_object_etcd-cluster-spec-events_content",
		"aws_s3_object_etcd-cluster-spec-main_content",
		"aws_s3_object_kops-version.txt_content",
		"aws_s3_object_manifests-etcdmanager-events-control-plane-fr-par-1_content",
		"aws_s3_object_manifests-etcdmanager-main-control-plane-fr-par-1_content",
		"aws_s3_object_manifests-static-kube-apiserver-healthcheck_content",
		"aws_s3_object_nodeupconfig-control-plane-fr-par-1_content",
		"aws_s3_object_nodeupconfig-nodes-fr-par-1_content",
		"aws_s3_object_"+i.clusterName+"-addons-bootstrap_content",
		"aws_s3_object_"+i.clusterName+"-addons-coredns.addons.k8s.io-k8s-1.12_content",
		"aws_s3_object_"+i.clusterName+"-addons-kops-controller.addons.k8s.io-k8s-1.16_content",
		"aws_s3_object_"+i.clusterName+"-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_content",
		"aws_s3_object_"+i.clusterName+"-addons-limit-range.addons.k8s.io_content",
		"aws_s3_object_"+i.clusterName+"-addons-networking.cilium.io-k8s-1.16_content",
		"scaleway_instance_server_control-plane-fr-par-1-0_user_data",
		"scaleway_instance_server_nodes-fr-par-1-0_user_data",
	)

	i.runTest(t, ctx, h, expectedFilenames, "", "", nil)
}

func MakeSSHKeyPair(publicKeyPath string, privateKeyPath string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return err
	}

	var privateKeyBytes bytes.Buffer
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err := pem.Encode(&privateKeyBytes, privateKeyPEM); err != nil {
		return err
	}
	if err := os.WriteFile(privateKeyPath, privateKeyBytes.Bytes(), os.FileMode(0o700)); err != nil {
		return err
	}

	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)
	if err := os.WriteFile(publicKeyPath, publicKeyBytes, os.FileMode(0o744)); err != nil {
		return err
	}

	return nil
}
