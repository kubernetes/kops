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

package azure

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func TestUnmarshalMetadata(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/metadata.json")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	actual, err := unmarshalInstanceMetadata(data)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected := &instanceMetadata{
		Compute: &instanceComputeMetadata{
			ResourceGroupName: "macikgo-test-may-23",
			SubscriptionID:    "8d10da13-8125-4ba9-a717-bf7490507b3d",
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %+v, but got %+v", expected, actual)
	}
}
