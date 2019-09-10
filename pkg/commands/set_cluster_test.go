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

package commands

import (
	"reflect"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func TestSetClusterFields(t *testing.T) {
	grid := []struct {
		Fields []string
		Input  kops.Cluster
		Output kops.Cluster
	}{
		{
			Fields: []string{
				"spec.kubernetesVersion=1.8.2",
				"spec.kubelet.authorizationMode=Webhook",
				"spec.kubelet.authenticationTokenWebhook=true",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					Kubelet: &kops.KubeletConfigSpec{},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "1.8.2",
					Kubelet: &kops.KubeletConfigSpec{
						AuthorizationMode:          "Webhook",
						AuthenticationTokenWebhook: fi.Bool(true),
					},
				},
			},
		},
	}

	for _, g := range grid {
		var igs []*kops.InstanceGroup
		c := g.Input

		err := SetClusterFields(g.Fields, &c, igs)
		if err != nil {
			t.Errorf("unexpected error from setClusterFields %v: %v", g.Fields, err)
			continue
		}

		if !reflect.DeepEqual(c, g.Output) {
			t.Errorf("unexpected output from setClusterFields %v.  expected=%v, actual=%v", g.Fields, g.Output, c)
			continue
		}

	}
}
