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

package configbuilder

import (
	"bytes"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/kops/pkg/apis/kops"
	"testing"
)

func resourceValue(s string) *resource.Quantity {
	q := resource.MustParse(s)
	return &q
}

func TestParseBasic(t *testing.T) {
	expect := []byte(
		`apiVersion: kubescheduler.config.k8s.io/v1alpha1
Kind: KubeSchedulerConfiguration
clientConnection:
  kubeconfig: null
  qps: 3
`)
	qps := float32(3.0)
	s := &kops.KubeSchedulerConfig{Qps: &qps}

	yaml, err := BuildConfigYaml(s)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if !bytes.Equal(yaml, expect) {
		t.Errorf("unexpected result: %v, expected: %v", expect, yaml)
	}
}

func TestGetStructVal(t *testing.T) {
	str := "test"
	s := &SchedulerConfig{
		ClientConnection: &ClientConnectionConfig{
			Kubeconfig: &str,
		},
	}
	v, err := getValueFromStruct("ClientConnection.Kubeconfig", s)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	inStruct := v.Elem().String()
	if inStruct != str {
		t.Errorf("unexpected value: %s, %s, expected: %s", inStruct, err, str)
	}

}

func TestWrongStructField(t *testing.T) {
	str := "test"
	s := &SchedulerConfig{
		ClientConnection: &ClientConnectionConfig{
			Kubeconfig: &str,
		},
	}
	v, err := getValueFromStruct("ClientConnection.NotExistent", s)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if v.IsValid() {
		t.Errorf("unexpected Valid value from non-existent field lookup")
	}

}
