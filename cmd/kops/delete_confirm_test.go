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

package main

import (
	"bytes"
	"strings"
	"testing"

	"k8s.io/kops/util/pkg/ui"
)

// TestContainsString tests the ContainsString() function
func TestContainsString(t *testing.T) {
	testString := "my test string"
	answer := ui.ContainsString(strings.Split(testString, " "), "my")
	if !answer {
		t.Fatal("Failed to find string using ui.ContainsString()")
	}
	answer = ui.ContainsString(strings.Split(testString, " "), "string")
	if !answer {
		t.Fatal("Failed to find string using ui.ContainsString()")
	}
	answer = ui.ContainsString(strings.Split(testString, " "), "random")
	if answer {
		t.Fatal("Found string that does not exist using ui.ContainsString()")
	}
}

// TestConfirmation attempts to test the majority of the ui.GetConfirm function used in the 'kogs delete' commands
func TestConfirmation(t *testing.T) {
	var out bytes.Buffer
	c := &ui.ConfirmArgs{
		Message: "Are you sure you want to remove?",
		Out:     &out,
		TestVal: "no",
	}

	answer := ui.GetConfirm(c)
	if !strings.Contains(out.String(), "Are you sure") {
		t.Fatal("Confirmation not in output")
	}
	if answer == true {
		t.Fatal("Confirmation should have been denied.")
	}

	c.TestVal = "yes"
	answer = ui.GetConfirm(c)
	if answer != true {
		t.Fatal("Confirmation should have been approved.")
	}

}
