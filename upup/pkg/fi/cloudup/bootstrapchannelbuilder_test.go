/*
Copyright 2016 The Kubernetes Authors.

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

package cloudup

import (
	"testing"

	api "k8s.io/kops/pkg/apis/kops"
)

func TestBootstrapChanelBuilder_BuildTasks(t *testing.T) {
	// TODO need to figure out how to mock Loader
}

func TestBootstrapChanelBuilder_buildManifest(t *testing.T) {
	c := buildDefaultCluster(t)

	c.Spec.Networking.Weave = &api.WeaveNetworkingSpec{}

	bcb := BootstrapChannelBuilder{cluster: c}
	a, m := bcb.buildManifest()

	if a == nil {
		t.Fatal("Addons are nil")
	}

	if m == nil {
		t.Fatal("Manifests are nil")
	}

	var hasLimit, hasWeave bool

	for _, value := range a.Spec.Addons {
		if *value.Name == "networking.weave" {
			hasWeave = true
		}

		if *value.Name == "limit-range" {
			hasLimit = true
		}

	}

	if !hasWeave {
		t.Fatal("unable to find weave")
	}

	if !hasLimit {
		t.Fatal("unable to find limit-builder")
	}
}
