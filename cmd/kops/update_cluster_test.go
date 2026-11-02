/*
Copyright 2026 The Kubernetes Authors.

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
	"io"
	"testing"
)

func TestUpdateClusterTargetValidation(t *testing.T) {
	const partialTerraformError = "cannot use --target=terraform with --instance-group or --instance-group-roles; generate the full configuration and use terraform apply -target instead"
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "terraform with instance groups",
			args:    []string{"--target=terraform", "--instance-group=nodes-a,nodes-b"},
			wantErr: partialTerraformError,
		},
		{
			name:    "terraform with instance group roles",
			args:    []string{"--target=terraform", "--instance-group-roles=control-plane,apiserver"},
			wantErr: partialTerraformError,
		},
		{
			name:    "terraform with both filters",
			args:    []string{"--target=terraform", "--instance-group=nodes", "--instance-group-roles=node"},
			wantErr: partialTerraformError,
		},
		{
			name: "terraform without filters",
			args: []string{"--target=terraform"},
		},
		{
			name: "terraform with empty filters",
			args: []string{"--target=terraform", "--instance-group=", "--instance-group-roles="},
		},
		{
			name: "direct with instance groups",
			args: []string{"--target=direct", "--instance-group=nodes", "--yes"},
		},
		{
			name: "direct with instance group roles",
			args: []string{"--target=direct", "--instance-group-roles=node", "--yes"},
		},
		{
			name: "dry run with instance groups",
			args: []string{"--target=dryrun", "--instance-group=nodes"},
		},
		{
			name: "default target with instance group roles",
			args: []string{"--instance-group-roles=node"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCmdUpdateCluster(nil, io.Discard)
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatalf("parsing flags: %v", err)
			}
			var err error
			if cmd.PreRunE != nil {
				err = cmd.PreRunE(cmd, nil)
			}
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected validation error: %v", err)
				}
			} else if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("validation error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}
