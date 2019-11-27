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

package featureflag

import (
	"os"
	"testing"

	"k8s.io/klog"
)

func TestFlagToFalse(t *testing.T) {
	f := New("UnitTest1", Bool(true))
	if !f.Enabled() {
		t.Fatalf("Flag did not default true")
	}

	// Really just to force a dependency on glog, so that we can pass -v and -logtostderr to go test
	klog.Info("Created flag Unittest1")

	ParseFlags("-UnitTest1")
	if f.Enabled() {
		t.Fatalf("Flag did not default turn off")
	}

	ParseFlags("UnitTest1")
	if !f.Enabled() {
		t.Fatalf("Flag did not default turn on")
	}
}

func TestSetenv(t *testing.T) {
	f := New("UnitTest2", Bool(true))
	if !f.Enabled() {
		t.Fatalf("Flag did not default true")
	}

	os.Setenv("KOPS_FEATURE_FLAGS", "-UnitTest2")
	if !f.Enabled() {
		t.Fatalf("Flag was reparsed immediately after os.Setenv")
	}

	ParseFlags("-UnitTest2")
	if f.Enabled() {
		t.Fatalf("Flag was not updated by ParseFlags")
	}
}
