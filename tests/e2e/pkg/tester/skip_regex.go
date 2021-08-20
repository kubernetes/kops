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
	"strings"
)

const (
	skipRegexBase = "\\[Slow\\]|\\[Serial\\]|\\[Disruptive\\]|\\[Flaky\\]|\\[Feature:.+\\]|\\[HPA\\]|\\[Driver:.nfs\\]|Dashboard|Gluster|RuntimeClass|RuntimeHandler"
)

func (t *Tester) setSkipRegexFlag() error {
	if t.SkipRegex != "" {
		return nil
	}

	cluster, err := t.getKopsCluster()
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
	} else if networking.Calico != nil {
		skipRegex += "|Services.*functioning.*NodePort"
	} else if networking.Kuberouter != nil {
		skipRegex += "|load-balancer|hairpin|affinity\\stimeout|service\\.kubernetes\\.io|CLOSE_WAIT"
	} else if networking.Kubenet != nil {
		skipRegex += "|Services.*affinity"
	}

	if strings.HasSuffix(cluster.Spec.KubernetesVersion, "v1.23.0-alpha.0") {
		// This matches `k8s_version='latest'` in build_jobs.py
		// TODO(rifelpet): Remove once the next 1.23 pre-release tag has been created
		// ref: https://github.com/kubernetes/kubernetes/pull/104061
		skipRegex += "|MetricsGrabber.should.grab.all.metrics.from.a.ControllerManager"
	}

	// Ensure it is valid regex
	if _, err := regexp.Compile(skipRegex); err != nil {
		return err
	}
	t.SkipRegex = skipRegex
	return nil
}
