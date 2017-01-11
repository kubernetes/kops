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

package kops

import (
	"testing"
)

func Test_ParseInstanceGroupRole(t *testing.T) {
	grid := []struct {
		Input        string
		Lenient      bool
		ExpectedRole InstanceGroupRole
		ExpectedOK   bool
	}{
		{
			"bastion", false,
			InstanceGroupRoleBastion, true,
		},
		{
			"bastions", false,
			"", false,
		},
		{
			"bastion", true,
			InstanceGroupRoleBastion, true,
		},
		{
			"bastions", true,
			InstanceGroupRoleBastion, true,
		},
		{
			"Nodes", true,
			InstanceGroupRoleNode, true,
		},
		{
			"Masters", true,
			InstanceGroupRoleMaster, true,
		},
		{
			"NotARole", true,
			"", false,
		},
	}
	for _, g := range grid {
		role, ok := ParseInstanceGroupRole(g.Input, g.Lenient)
		if ok != g.ExpectedOK || role != g.ExpectedRole {
			t.Errorf("unexpected result from %q, %v.  got %q, %v", g.Input, g.Lenient, role, ok)
		}
	}
}
