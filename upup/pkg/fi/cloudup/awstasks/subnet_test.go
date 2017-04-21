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

package awstasks

import (
	"fmt"
	"k8s.io/kops/upup/pkg/fi"
	"testing"
)

func Test_Subnet_ValidateRequired(t *testing.T) {
	var a *Subnet
	e := &Subnet{}

	changes := &Subnet{}
	fi.BuildChanges(a, e, changes)

	err := e.CheckChanges(a, e, changes)
	if err == nil {
		t.Errorf("validation error was expected")
	}
	if fmt.Sprintf("%v", err) != "Subnet.VPC: Required value: must specify a VPC" {
		t.Errorf("unexpected error: %v", err)
	}
}

func Test_Subnet_CannotChangeSubnet(t *testing.T) {
	a := &Subnet{VPC: &VPC{Name: s("defaultvpc")}, CIDR: s("192.168.0.0/16")}
	e := &Subnet{}
	*e = *a

	e.CIDR = s("192.168.0.1/16")

	changes := &Subnet{}
	fi.BuildChanges(a, e, changes)

	err := e.CheckChanges(a, e, changes)
	if err == nil {
		t.Errorf("validation error was expected")
	}
	if fmt.Sprintf("%v", err) != "Subnet.CIDR: Invalid value: \"192.168.0.0/16\": field is immutable: old=\"192.168.0.1/16\" new=\"192.168.0.0/16\"" {
		t.Errorf("unexpected error: %v", err)
	}
}
