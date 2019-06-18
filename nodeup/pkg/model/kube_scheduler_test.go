/*
Copyright 2017 The Kubernetes Authors.

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

package model

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func Test_KubeSchedulerConfig(t *testing.T) {
	expected := `algorithmSource:
  provider: DefaultProvider
apiVersion: kubescheduler.config.k8s.io/v1alpha1
bindTimeoutSeconds: 600
clientConnection:
  acceptContentTypes: ""
  burst: 1000
  contentType: application/vnd.kubernetes.protobuf
  kubeconfig: /var/lib/kube-scheduler/kubeconfig
  qps: 100
disablePreemption: false
enableContentionProfiling: false
enableProfiling: false
failureDomains: kubernetes.io/hostname,failure-domain.beta.kubernetes.io/zone,failure-domain.beta.kubernetes.io/region
hardPodAffinitySymmetricWeight: 1
healthzBindAddress: 0.0.0.0:10251
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: true
  leaseDuration: 15s
  lockObjectName: kube-scheduler
  lockObjectNamespace: kube-system
  renewDeadline: 10s
  resourceLock: endpoints
  retryPeriod: 2s
metricsBindAddress: 0.0.0.0:10251
percentageOfNodesToScore: 50
schedulerName: default-scheduler
`
	qps := float32(100.0)
	burst := int32(1000)
	var cluster = &kops.Cluster{}
	trueVal := true
	cluster.Spec.KubernetesVersion = "1.6.0"
	cluster.Spec.KubeScheduler = &kops.KubeSchedulerConfig{}

	cluster.Spec.KubeScheduler.LeaderElection = &kops.LeaderElectionConfiguration{}
	cluster.Spec.KubeScheduler.LeaderElection.LeaderElect = &trueVal
	cluster.Spec.KubeScheduler.QPS = &qps
	cluster.Spec.KubeScheduler.Burst = &burst

	b := &KubeSchedulerBuilder{
		&NodeupModelContext{
			Cluster: cluster,
		},
	}
	if err := b.Init(); err != nil {
		t.Error(err)
	}

	var got, err = b.buildConfigFile()
	if err != nil {
		t.Error(err)
	}

	if got != expected {
		t.Error("Expected:\n" + expected)
		t.Error("Got:\n" + got)
	}

}
