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
	"regexp"

	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	skipRegexBase = "\\[Slow\\]|\\[Serial\\]|\\[Disruptive\\]|\\[Flaky\\]|\\[Feature:.+\\]|nfs|NFS|Gluster|NodeProblemDetector"
)

func (t *Tester) setSkipRegexFlag() error {
	if t.SkipRegex != "" {
		return nil
	}

	cluster, err := t.getKopsCluster()
	if err != nil {
		return err
	}
	k8sVersion, err := util.ParseKubernetesVersion(cluster.Spec.KubernetesVersion)
	if err != nil {
		return err
	}

	skipRegex := skipRegexBase

	if k8sVersion.Minor == 26 && cluster.Spec.LegacyCloudProvider == "aws" {
		// This test was introduced in k8s 1.26
		// and skipped automatically for AWS clusters as of k8s 1.27
		// because Classic Load Balancers dont support UDP
		// https://github.com/kubernetes/kubernetes/pull/113650
		// https://github.com/kubernetes/kubernetes/pull/115977
		skipRegex += "|LoadBalancers.should.be.able.to.preserve.UDP.traffic"
	}

	networking := cluster.Spec.LegacyNetworking
	if networking.Cilium != nil {
		if k8sVersion.Minor < 27 {
			// Partially implemented in Cilium 1.13 but kops doesn't enable it
			// Ref: https://github.com/cilium/cilium/pull/20033
			// K8s 1.27+ added [Serial] to the test case, which is skipped by default
			// Ref: https://github.com/kubernetes/kubernetes/pull/113335
			skipRegex += "|should.create.a.Pod.with.SCTP.HostPort"
		}
	}

	if cluster.Spec.LegacyCloudProvider == "digitalocean" {
		// https://github.com/kubernetes/kubernetes/issues/121018
		skipRegex += "|Services.should.respect.internalTrafficPolicy=Local.Pod.and.Node,.to.Pod"
	}

	if cluster.Spec.LegacyCloudProvider == "gce" {
		// The in-tree driver and its E2E tests use `topology.kubernetes.io/zone` but the CSI driver uses `topology.gke.io/zone`
		skipRegex += "|In-tree.Volumes.\\[Driver:.gcepd\\].*topology.should.provision.a.volume.and.schedule.a.pod.with.AllowedTopologies"
	}

	if cluster.Spec.LegacyCloudProvider == "gce" {
		// this tests assumes a custom config for containerd:
		// https://github.com/kubernetes/test-infra/blob/578d86a7be187214be6ccd60e6ea7317b51aeb15/jobs/e2e_node/containerd/config.toml#L19-L21
		// ref: https://github.com/kubernetes/kubernetes/pull/104803
		skipRegex += "|RuntimeClass.should.run"
		// https://github.com/kubernetes/kubernetes/pull/108694
		skipRegex += "|Metadata.Concealment"
	}

	if k8sVersion.Minor >= 22 {
		// this test was being skipped automatically because it isn't applicable with CSIMigration=true which is default
		// but skipping logic has been changed and now the test is planned for removal
		// Should be skipped on all versions we enable CSI drivers on
		// ref: https://github.com/kubernetes/kubernetes/pull/109649#issuecomment-1108574843
		skipRegex += "|should.verify.that.all.nodes.have.volume.limits"
	}

	if cluster.Spec.LegacyCloudProvider == "aws" {
		// This test fails on RHEL-based distros because they return fully qualified hostnames yet the k8s node names are not fully qualified.
		// Dedicated job testing this: https://testgrid.k8s.io/kops-misc#kops-aws-k28-hostname-bug123255
		// ref: https://github.com/kubernetes/kops/issues/16349
		// ref: https://github.com/kubernetes/kubernetes/issues/123255
		skipRegex += "|Services.should.function.for.service.endpoints.using.hostNetwork"
	}

	if cluster.Spec.CloudConfig != nil && cluster.Spec.CloudConfig.AWSEBSCSIDriver != nil && fi.ValueOf(cluster.Spec.CloudConfig.AWSEBSCSIDriver.Enabled) {
		skipRegex += "|In-tree.Volumes.\\[Driver:.aws\\]"
	}

	for _, subnet := range cluster.Spec.Subnets {
		if subnet.Type == v1alpha2.SubnetTypePrivate || subnet.Type == v1alpha2.SubnetTypeDualStack {
			skipRegex += "|SSH.should.SSH.to.all.nodes.and.run.commands"
			break
		}
	}

	// Ensure it is valid regex
	if _, err := regexp.Compile(skipRegex); err != nil {
		return err
	}
	t.SkipRegex = skipRegex
	return nil
}
