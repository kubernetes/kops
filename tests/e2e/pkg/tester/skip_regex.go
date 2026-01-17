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
)

const (
	skipRegexBase = "\\[Slow\\]|\\[Serial\\]|\\[Disruptive\\]|\\[Flaky\\]|\\[Feature:.+\\]|nfs|NFS"
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

	if k8sVersion.Minor < 35 {
		// cpu.weight value changed in runc 1.3.2
		// https://github.com/kubernetes/kubernetes/issues/135214
		// https://github.com/opencontainers/runc/issues/4896
		skipRegex += "|[Burstable|Guaranteed].QoS.pod"
		skipRegex += "|Pod.InPlace.Resize.Container"
	}

	// Skip broken test, see https://github.com/kubernetes/kubernetes/pull/133262
	skipRegex += "|blackbox.*should.not.be.able.to.pull.image.from.invalid.registry"
	skipRegex += "|blackbox.*should.be.able.to.pull.from.private.registry.with.secret"

	// K8s 1.28 promoted ProxyTerminatingEndpoints to GA, but it has limited CNI support
	// https://github.com/kubernetes/kubernetes/pull/117718
	// https://github.com/cilium/cilium/issues/27358
	skipRegex += "|fallback.to.local.terminating.endpoints.when.there.are.no.ready.endpoints.with.externalTrafficPolicy.Local"

	networking := cluster.Spec.LegacyNetworking
	switch {
	case networking.Kubenet != nil, networking.Cilium != nil:
		skipRegex += "|Services.*rejected.*endpoints"
	}
	if networking.Cilium != nil {
		// Cilium upstream skip references: https://github.com/cilium/cilium/blob/main/.github/workflows/k8s-kind-network-e2e.yaml#L210
		// https://github.com/cilium/cilium/issues/10002
		skipRegex += "|TCP.CLOSE_WAIT"
		// https://github.com/cilium/cilium/issues/15361
		skipRegex += "|external.IP.is.not.assigned.to.a.node"

		// https://github.com/cilium/cilium/issues/14287
		skipRegex += "|same.hostPort.but.different.hostIP.and.protocol"

		// https://github.com/kubernetes/kubernetes/blob/418ae605ec1b788d43bff7ac44af66d8b669b833/test/e2e/network/networking.go#L135
		skipRegex += "|should.check.kube-proxy.urls"

		if k8sVersion.Minor < 37 {
			// This seems to be specific to the kube-proxy replacement
			// < 36 so we look at this again
			skipRegex += "|Services.should.support.externalTrafficPolicy.Local.for.type.NodePort"
			// https://github.com/kubernetes/kubernetes/issues/129221
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
		// this test assumes the cluster runs COS but kOps uses Ubuntu by default
		// ref: https://github.com/kubernetes/test-infra/pull/22190
		skipRegex += "|should.be.mountable.when.non-attachable"
		// The in-tree driver and its E2E tests use `topology.kubernetes.io/zone` but the CSI driver uses `topology.gke.io/zone`
		skipRegex += "|In-tree.Volumes.\\[Driver:.gcepd\\].*topology.should.provision.a.volume.and.schedule.a.pod.with.AllowedTopologies"
	}

	// This test fails on RHEL-based distros because they return fully qualified hostnames yet the k8s node names are not fully qualified.
	// Dedicated job testing this: https://testgrid.k8s.io/kops-misc#kops-aws-k28-hostname-bug123255
	// ref: https://github.com/kubernetes/kops/issues/16349
	// ref: https://github.com/kubernetes/kubernetes/issues/123255
	// ref: https://github.com/kubernetes/kubernetes/issues/121018
	// ref: https://github.com/kubernetes/kubernetes/pull/126896
	// < 37 so we look at this again
	if k8sVersion.Minor < 37 {
		skipRegex += "|Services.should.function.for.service.endpoints.using.hostNetwork"
		skipRegex += "|Services.should.implement.NodePort.and.HealthCheckNodePort.correctly.when.ExternalTrafficPolicy.changes"
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
