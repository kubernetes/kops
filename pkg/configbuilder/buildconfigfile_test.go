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

package configbuilder

import (
	"testing"
)

// ClientConnectionConfig is used by kube-scheduler to talk to the api server
type DummyNestedStruct struct {
	Name *string  `yaml:"name,omitempty"`
	QPS  *float64 `yaml:"qps,omitempty"`
}

// SchedulerConfig is used to generate the config file
type DummyStruct struct {
	ClientConnection *DummyNestedStruct `yaml:"clientConnection,omitempty"`
}

func TestGetStructVal(t *testing.T) {
	str := "test"
	s := &DummyStruct{
		ClientConnection: &DummyNestedStruct{
			Name: &str,
		},
	}
	v, err := getValueFromStruct("ClientConnection.Name", s)
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
	s := &DummyStruct{
		ClientConnection: &DummyNestedStruct{
			Name: &str,
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
