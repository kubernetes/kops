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
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/utils"
)

const (
	skipRegexBase = "\\[Slow\\]|\\[Serial\\]|\\[Disruptive\\]|\\[Flaky\\]|\\[Feature:.+\\]|nfs|NFS|Gluster"
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

	networking := cluster.Spec.Networking
	switch {
	case networking.Kubenet != nil, networking.Canal != nil, networking.Weave != nil, networking.Cilium != nil:
		skipRegex += "|Services.*rejected.*endpoints"
	}
	if networking.Cilium != nil {
		// https://github.com/cilium/cilium/issues/10002
		skipRegex += "|TCP.CLOSE_WAIT"
		// https://github.com/cilium/cilium/issues/15361
		skipRegex += "|external.IP.is.not.assigned.to.a.node"
		// https://github.com/cilium/cilium/issues/14287
		skipRegex += "|same.port.number.but.different.protocols|same.hostPort.but.different.hostIP.and.protocol"
		if k8sVersion.Minor >= 22 {
			// ref:
			// https://github.com/kubernetes/kubernetes/issues/96717
			// https://github.com/cilium/cilium/issues/5719
			skipRegex += "|should.create.a.Pod.with.SCTP.HostPort"
		}
		// https://github.com/cilium/cilium/issues/18241
		skipRegex += "|Services.should.create.endpoints.for.unready.pods"
		skipRegex += "|Services.should.be.able.to.connect.to.terminating.and.unready.endpoints.if.PublishNotReadyAddresses.is.true"
	} else if networking.Kuberouter != nil {
		skipRegex += "|load-balancer|hairpin|affinity\\stimeout|service\\.kubernetes\\.io|CLOSE_WAIT"
	} else if networking.Kubenet != nil {
		skipRegex += "|Services.*affinity"
	}

	if cluster.Spec.LegacyCloudProvider == "gce" {
		// Firewall tests expect a specific format for cluster and control plane host names
		// which kOps does not match
		// ref: https://github.com/kubernetes/kubernetes/blob/1bd00776b5d78828a065b5c21e7003accc308a06/test/e2e/framework/providers/gce/firewall.go#L92-L100
		skipRegex += "|Firewall"
		// kube-dns tests are not skipped automatically if a cluster uses CoreDNS instead
		skipRegex += "|kube-dns"
		// this test assumes the cluster runs COS but kOps uses Ubuntu by default
		// ref: https://github.com/kubernetes/test-infra/pull/22190
		skipRegex += "|should.be.mountable.when.non-attachable"
		// The in-tree driver and its E2E tests use `topology.kubernetes.io/zone` but the CSI driver uses `topology.gke.io/zone`
		skipRegex += "|In-tree.Volumes.\\[Driver:.gcepd\\].*topology.should.provision.a.volume.and.schedule.a.pod.with.AllowedTopologies"
	}

	if cluster.Spec.LegacyCloudProvider == "gce" || k8sVersion.Minor <= 23 {
		// this tests assumes a custom config for containerd:
		// https://github.com/kubernetes/test-infra/blob/578d86a7be187214be6ccd60e6ea7317b51aeb15/jobs/e2e_node/containerd/config.toml#L19-L21
		// ref: https://github.com/kubernetes/kubernetes/pull/104803
		skipRegex += "|RuntimeClass.should.run"
		// https://github.com/kubernetes/kubernetes/pull/108694
		skipRegex += "|Metadata.Concealment"
	}

	if k8sVersion.Minor == 23 && cluster.Spec.LegacyCloudProvider == "aws" && utils.IsIPv6CIDR(cluster.Spec.NonMasqueradeCIDR) {
		// ref: https://github.com/kubernetes/kubernetes/pull/106992
		skipRegex += "|should.not.disrupt.a.cloud.load-balancer.s.connectivity.during.rollout"
	}

	if k8sVersion.Minor == 23 {
		// beta feature not enabled by default
		skipRegex += "|Topology.Hints"
	}

	if k8sVersion.Minor >= 22 {
		// this test was being skipped automatically because it isn't applicable with CSIMigration=true which is default
		// but skipping logic has been changed and now the test is planned for removal
		// Should be skipped on all versions we enable CSI drivers on
		// ref: https://github.com/kubernetes/kubernetes/pull/109649#issuecomment-1108574843
		skipRegex += "|should.verify.that.all.nodes.have.volume.limits"
	}

	if cluster.Spec.CloudConfig != nil && cluster.Spec.CloudConfig.AWSEBSCSIDriver != nil && fi.BoolValue(cluster.Spec.CloudConfig.AWSEBSCSIDriver.Enabled) {
		skipRegex += "|In-tree.Volumes.\\[Driver:.aws\\]"
	}

	// Ensure it is valid regex
	if _, err := regexp.Compile(skipRegex); err != nil {
		return err
	}
	t.SkipRegex = skipRegex
	return nil
}
