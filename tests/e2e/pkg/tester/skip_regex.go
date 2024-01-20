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
	"k8s.io/kops/upup/pkg/fi/utils"
)

const (
	skipRegexBase = "\\[Slow\\]|\\[Serial\\]|\\[Disruptive\\]|\\[Flaky\\]|\\[Feature:.+\\]|nfs|NFS|Gluster|NodeProblemDetector"
)

func (t *Tester) setSkipRegexFlag() error {
	if t.SkipRegex != "" {
		return nil
	}

	kopsVersion, err := t.getKopsVersion()
	if err != nil {
		return err
	}
	isPre28 := kopsVersion < "1.28"

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
	if !isPre28 {
		// K8s 1.28 promoted ProxyTerminatingEndpoints to GA, but it has limited CNI support
		// https://github.com/kubernetes/kubernetes/pull/117718
		// https://github.com/cilium/cilium/issues/27358
		skipRegex += "|fallback.to.local.terminating.endpoints.when.there.are.no.ready.endpoints.with.externalTrafficPolicy.Local"
	}

	networking := cluster.Spec.LegacyNetworking
	switch {
	case networking.Kubenet != nil, networking.Canal != nil, networking.Cilium != nil:
		skipRegex += "|Services.*rejected.*endpoints"
	}
	if networking.Cilium != nil {
		// https://github.com/cilium/cilium/issues/10002
		skipRegex += "|TCP.CLOSE_WAIT"
		// https://github.com/cilium/cilium/issues/15361
		skipRegex += "|external.IP.is.not.assigned.to.a.node"
		// https://github.com/cilium/cilium/issues/14287
		skipRegex += "|same.port.number.but.different.protocols"
		skipRegex += "|same.hostPort.but.different.hostIP.and.protocol"
		// https://github.com/cilium/cilium/issues/9207
		skipRegex += "|serve.endpoints.on.same.port.and.different.protocols"
		// https://github.com/kubernetes/kubernetes/blob/418ae605ec1b788d43bff7ac44af66d8b669b833/test/e2e/network/networking.go#L135
		skipRegex += "|should.check.kube-proxy.urls"

		if isPre28 {
			// These may be fixed in Cilium 1.13 but skipping for now
			skipRegex += "|Service.with.multiple.ports.specified.in.multiple.EndpointSlices"
			// https://github.com/cilium/cilium/issues/18241
			skipRegex += "|Services.should.create.endpoints.for.unready.pods"
			skipRegex += "|Services.should.be.able.to.connect.to.terminating.and.unready.endpoints.if.PublishNotReadyAddresses.is.true"
		}
		if k8sVersion.Minor < 27 {
			// Partially implemented in Cilium 1.13 but kops doesn't enable it
			// Ref: https://github.com/cilium/cilium/pull/20033
			// K8s 1.27+ added [Serial] to the test case, which is skipped by default
			// Ref: https://github.com/kubernetes/kubernetes/pull/113335
			skipRegex += "|should.create.a.Pod.with.SCTP.HostPort"
		}
	} else if networking.KubeRouter != nil {
		skipRegex += "|load-balancer|hairpin|affinity\\stimeout|service\\.kubernetes\\.io|CLOSE_WAIT"
		skipRegex += "|EndpointSlice.should.support.a.Service.with.multiple"
		skipRegex += "|internalTrafficPolicy|externallTrafficPolicy|only.terminating.endpoints"
	} else if networking.Kubenet != nil {
		skipRegex += "|Services.*affinity"
	}

	if cluster.Spec.LegacyCloudProvider == "digitalocean" {
		// https://github.com/kubernetes/kubernetes/issues/121018
		skipRegex += "|Services.should.respect.internalTrafficPolicy=Local.Pod.and.Node,.to.Pod"
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
