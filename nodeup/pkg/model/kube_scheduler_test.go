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

package model

import (
	"bytes"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/configbuilder"
)

func TestParseDefault(t *testing.T) {
	expect := []byte(
		`apiVersion: kubescheduler.config.k8s.io/v1alpha2
kind: KubeSchedulerConfiguration
clientConnection:
  kubeconfig: /var/lib/kube-scheduler/kubeconfig
`)

	s := &kops.KubeSchedulerConfig{}

	yaml, err := configbuilder.BuildConfigYaml(s, NewSchedulerConfig("kubescheduler.config.k8s.io/v1alpha2"))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if !bytes.Equal(yaml, expect) {
		t.Errorf("unexpected result: \n%s, expected: \n%s", yaml, expect)
	}
}

func TestParse(t *testing.T) {
	expect := []byte(
		`apiVersion: kubescheduler.config.k8s.io/v1alpha2
kind: KubeSchedulerConfiguration
clientConnection:
  burst: 100
  kubeconfig: /var/lib/kube-scheduler/kubeconfig
  qps: 3.1
`)
	qps, _ := resource.ParseQuantity("3.1")

	s := &kops.KubeSchedulerConfig{Qps: &qps, Burst: 100}

	yaml, err := configbuilder.BuildConfigYaml(s, NewSchedulerConfig("kubescheduler.config.k8s.io/v1alpha2"))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if !bytes.Equal(yaml, expect) {
		t.Errorf("unexpected result: \n%s, expected: \n%s", yaml, expect)
	}
}
