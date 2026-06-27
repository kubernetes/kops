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

package kops

import (
	"fmt"
	"testing"

	inf "gopkg.in/inf.v0"
	"k8s.io/kops/upup/pkg/fi/utils"
)

func Test_ParseInstanceGroupRole(t *testing.T) {
	grid := []struct {
		Input        string
		Lenient      bool
		ExpectedRole InstanceGroupRole
		ExpectedOK   bool
	}{
		{
			Input: "bastion", Lenient: false,
			ExpectedRole: canocicalBastion.Role(), ExpectedOK: true,
		},
		{
			Input: "bastions", Lenient: false,
			ExpectedRole: "", ExpectedOK: false,
		},
		{
			Input: "bastion", Lenient: true,
			ExpectedRole: canocicalBastion.Role(), ExpectedOK: true,
		},
		{
			Input: "bastions", Lenient: true,
			ExpectedRole: canocicalBastion.Role(), ExpectedOK: true,
		},
		{
			Input: "Nodes", Lenient: true,
			ExpectedRole: canocicalNode.Role(), ExpectedOK: true,
		},
		{
			Input: "Masters", Lenient: true,
			ExpectedRole: canocicalControlPlane.Role(), ExpectedOK: true,
		},
		{
			Input: "ControlPlanes", Lenient: true,
			ExpectedRole: canocicalControlPlane.Role(), ExpectedOK: true,
		},
		{
			Input: "Control-Planes", Lenient: true,
			ExpectedRole: canocicalControlPlane.Role(), ExpectedOK: true,
		},
		{
			Input: "NotARole", Lenient: true,
			ExpectedRole: "", ExpectedOK: false,
		},
	}
	for _, g := range grid {
		role, ok := ParseInstanceGroupRole(g.Input, g.Lenient)
		// Actual code should use role.HasX() or !role.HasX()
		// == and != do not generally give the correct answer.
		if ok != g.ExpectedOK || role != g.ExpectedRole {
			t.Errorf("unexpected result from %q, %v.  got %q, %v", g.Input, g.Lenient, role, ok)
		}
	}
}

func Test_ParseInstanceGroupRole_Equality(t *testing.T) {
	tests := []struct {
		Name     string
		InputOne string
		InputTwo string
		Expected bool
	}{
		{
			Name:     "Basic equality",
			InputOne: "ControlPlane",
			InputTwo: "Control-Plane",
			Expected: true,
		},
		{
			Name:     "Basic inequality",
			InputOne: "ControlPlane",
			InputTwo: "APIServer",
			Expected: false,
		},
		{
			Name:     "Ordering check",
			InputOne: "APIServer,Etcd",
			InputTwo: "Etcd,APIServer",
			Expected: true,
		},
		{
			Name:     "Partial Match",
			InputOne: "APIServer,Etcd",
			InputTwo: "APIServer",
			Expected: false,
		},
		{
			Name:     "Partial Match - Alternate order",
			InputOne: "APIServer",
			InputTwo: "APIServer,Etcd",
			Expected: false,
		},
	}
	for _, test := range tests {
		roleOne, ok := ParseInstanceGroupRole(test.InputOne, true)
		if ok != true {
			t.Errorf("unexpected result from test %q, role %q failed to parse", test.Name, test.InputOne)
		}
		roleTwo, ok := ParseInstanceGroupRole(test.InputTwo, true)
		if ok != true {
			t.Errorf("unexpected result from test %q, role %q failed to parse", test.Name, test.InputTwo)
		}
		if test.Expected {
			if roleOne != roleTwo {
				t.Errorf("Test %q Role %s != Role %s but should be", test.Name, test.InputOne, test.InputTwo)
			}
		} else {
			if roleOne == roleTwo {
				t.Errorf("Test %q Role %s == Role %s but should not be", test.Name, test.InputOne, test.InputTwo)
			}
		}
	}
}

func TestParseConfigYAML(t *testing.T) {
	pi := inf.NewDec(314, 2) // Or there abouts

	grid := []struct {
		Config        string
		ExpectedValue *inf.Dec
	}{
		{
			Config:        "kubeAPIServer: {  auditWebhookBatchThrottleQps: 3140m }",
			ExpectedValue: pi,
		},
		{
			Config:        "kubeAPIServer: {  auditWebhookBatchThrottleQps: 3.14 }",
			ExpectedValue: pi,
		},
		{
			Config:        "kubeAPIServer: {  auditWebhookBatchThrottleQps: 3.140 }",
			ExpectedValue: pi,
		},
		{
			Config:        "kubeAPIServer: {}",
			ExpectedValue: nil,
		},
	}

	for i := range grid {
		g := grid[i]
		t.Run(fmt.Sprintf("%q", g.Config), func(t *testing.T) {
			config := ClusterSpec{}
			err := utils.YamlUnmarshal([]byte(g.Config), &config)
			if err != nil {
				t.Errorf("error parsing configuration %q: %v", g.Config, err)
				return
			}

			actual := config.KubeAPIServer.AuditWebhookBatchThrottleQps
			if g.ExpectedValue == nil {
				if actual != nil {
					t.Errorf("expected null value for KubeAPIServer.AuditWebhookBatchThrottleQps, got %v", *actual)
					return
				}
			} else {
				if actual == nil {
					t.Errorf("expected %v value for KubeAPIServer.AuditWebhookBatchThrottleQps, got nil", *g.ExpectedValue)
					return
				} else if actual.AsDec().Cmp(g.ExpectedValue) != 0 {
					t.Errorf("expected %v value for KubeAPIServer.AuditWebhookBatchThrottleQps, got %v", g.ExpectedValue.String(), actual.AsDec().String())
					return
				}
			}
		})
	}
}
