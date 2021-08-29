/*
Copyright 2021 The Kubernetes Authors.

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

package resources

import (
	"fmt"
	"k8s.io/kops/pkg/apis/kops"
	"strings"
	"testing"
)

func Test_NodeUpTabs(t *testing.T) {
	for i, line := range strings.Split(NodeUpTemplate, "\n") {
		if strings.Contains(line, "\t") {
			t.Errorf("NodeUpTemplate contains unexpected character %q on line %d: %q", "\t", i, line)
		}
	}
}

func Test_AWSNodeUpTemplate(t *testing.T) {
	beforeScriptName := "some_script_required_to_run_before_nodeup.sh"
	afterScriptName := "some_script.sh"

	ig := &kops.InstanceGroup{
		Spec: kops.InstanceGroupSpec{
			AdditionalUserData: []kops.UserData{
				{
					Name: "some_script.sh",
					Content: `#!/bin/bash

echo "I run after nodeup.sh to setup some other stuff"
`,
					Type: "text/x-shellscript",
				},
				{
					Name: beforeScriptName,
					Content: `#!/bin/bash

echo "I run before nodeup.sh to fix some stuff"
`,
					Type:   "text/x-shellscript",
					Before: true,
				},
			},
		},
	}

	actual, err := AWSNodeUpTemplate(ig)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	indexOfNodeupScript := strings.Index(actual, contentDispositionHeaderOf("nodeup.sh"))

	if strings.Index(actual, contentDispositionHeaderOf(beforeScriptName)) > indexOfNodeupScript {
		t.Fatalf("script %q should have been placed before 'nodeup.sh'", beforeScriptName)
	}

	if strings.Index(actual, contentDispositionHeaderOf(afterScriptName)) < indexOfNodeupScript {
		t.Fatalf("script %q should have been placed after 'nodeup.sh'", afterScriptName)
	}
}

func contentDispositionHeaderOf(scriptName string) string {
	return fmt.Sprintf("Content-Disposition: attachment; filename=%q", scriptName)
}
