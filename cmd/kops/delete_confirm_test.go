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

package main

import (
	"bytes"
	"strings"
	"testing"

	"k8s.io/kops/util/pkg/ui"
)

// TestConfirmation attempts to test the majority of the ui.GetConfirm function used in the 'kops delete' commands
func TestConfirmation(t *testing.T) {
	var out bytes.Buffer
	c := &ui.ConfirmArgs{
		Message: "Are you sure you want to remove?",
		Out:     &out,
		TestVal: "no",
		Default: "no",
	}

	answer, err := ui.GetConfirm(c)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Are you sure") {
		t.Fatal("Confirmation not in output")
	}
	if !strings.Contains(out.String(), "y/N") {
		t.Fatal("Default 'No' was not set")
	}
	if answer == true {
		t.Fatal("Confirmation should have been denied.")
	}

	c.Default = "yes"
	_, err = ui.GetConfirm(c)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Y/n") {
		t.Fatal("Default 'Yes' was not set")
	}

	c.TestVal = "yes"
	answer, err = ui.GetConfirm(c)
	if err != nil {
		t.Fatal(err)
	}
	if answer != true {
		t.Fatal("Confirmation should have been approved.")
	}

}
