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

		if networking.Cilium.Version < "v1.17" {
			// https://github.com/cilium/cilium/issues/14287
			skipRegex += "|same.port.number.but.different.protocols"
			skipRegex += "|same.hostPort.but.different.hostIP.and.protocol"
			// https://github.com/cilium/cilium/issues/9207
			skipRegex += "|serve.endpoints.on.same.port.and.different.protocols"
		}

		// https://github.com/kubernetes/kubernetes/blob/418ae605ec1b788d43bff7ac44af66d8b669b833/test/e2e/network/networking.go#L135
		skipRegex += "|should.check.kube-proxy.urls"

		if k8sVersion.Minor < 33 {
			// This seems to be specific to the kube-proxy replacement
			// < 33 so we look at this again
			skipRegex += "|Services.should.support.externalTrafficPolicy.Local.for.type.NodePort"
		}

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

		if k8sVersion.Minor < 34 {
			// < 34 so we revisit this in future
			// This test checks for kube-proxy on port 10249 (`127.0.0.1:10249/proxyMode`)
			// It appears that the cilium kube-proxy replacement does not implement this.
			// Ref: https://github.com/kubernetes/kubernetes/issues/126903
			skipRegex += "|KubeProxy.should.update.metric.for.tracking.accepted.packets.destined.for.localhost.nodeports"
		}
	} else if networking.KubeRouter != nil {
		skipRegex += "|should set TCP CLOSE_WAIT timeout|should check kube-proxy urls"
	} else if networking.Kubenet != nil {
		skipRegex += "|Services.*affinity"
	}

	if cluster.Spec.LegacyCloudProvider == "digitalocean" {
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

		// this tests assumes a custom config for containerd:
		// https://github.com/kubernetes/test-infra/blob/578d86a7be187214be6ccd60e6ea7317b51aeb15/jobs/e2e_node/containerd/config.toml#L19-L21
		// ref: https://github.com/kubernetes/kubernetes/pull/104803
		skipRegex += "|RuntimeClass.should.run"
		// https://github.com/kubernetes/kubernetes/pull/108694
		skipRegex += "|Metadata.Concealment"

		if k8sVersion.Minor >= 31 {
			// Most e2e framework code for the in-tree provider has been removed but some test cases remain
			// https://github.com/kubernetes/kubernetes/pull/124519
			// https://github.com/kubernetes/test-infra/pull/33222
			skipRegex += "\\[sig-cloud-provider-gcp\\]"
		}
	}

	if k8sVersion.Minor >= 22 {
		// this test was being skipped automatically because it isn't applicable with CSIMigration=true which is default
		// but skipping logic has been changed and now the test is planned for removal
		// Should be skipped on all versions we enable CSI drivers on
		// ref: https://github.com/kubernetes/kubernetes/pull/109649#issuecomment-1108574843
		skipRegex += "|should.verify.that.all.nodes.have.volume.limits"
	}

	if cluster.Spec.LegacyCloudProvider == "aws" {
		if k8sVersion.Minor <= 26 {
			// Prow jobs are being migrated to community-owned EKS clusters.
			// The e2e.test binaries from older k/k builds dont have new enough aws-sdk-go versions to authenticate from EKS pods.
			// This disables tests that depend on e2e.test's aws-sdk-go.
			//
			// > Couldn't create a new PD in zone "ap-northeast-1c", sleeping 5 seconds: NoCredentialProviders: no valid providers in chain. Deprecated.
			//
			// We can remove this once we remove the old upgrade jobs.
			// Example: https://prow.k8s.io/view/gs/kubernetes-jenkins/logs/e2e-kops-aws-upgrade-k125-ko128-to-k126-kolatest/1808210907088556032
			skipRegex += "|\\[Driver:.aws\\].\\[Testpattern:.Pre-provisioned.PV|\\[Driver:.aws\\].\\[Testpattern:.Inline-volume"
		}
	}

	// This test fails on RHEL-based distros because they return fully qualified hostnames yet the k8s node names are not fully qualified.
	// Dedicated job testing this: https://testgrid.k8s.io/kops-misc#kops-aws-k28-hostname-bug123255
	// ref: https://github.com/kubernetes/kops/issues/16349
	// ref: https://github.com/kubernetes/kubernetes/issues/123255
	// ref: https://github.com/kubernetes/kubernetes/issues/121018
	// < 33 so we look at this again
	if k8sVersion.Minor < 33 {
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
