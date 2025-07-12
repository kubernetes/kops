/*
Copyright 2025 The Kubernetes Authors.

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
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/kops/cmd/kops/util"
	"testing"
)

func TestRunVersion(t *testing.T) {
	factoryOptions := &util.FactoryOptions{}
	factory := util.NewFactory(factoryOptions)

	tests := []struct {
		name           string
		opt            *VersionOptions
		expectedOutput string
		wantErr        error
	}{
		{
			name: "client Version",
			opt: &VersionOptions{
				Output: "",
			},
			expectedOutput: "Client Version",
		},
		{
			name: "output yaml format",
			opt: &VersionOptions{
				Output: OutputYaml,
			},
			expectedOutput: "clientVersion",
		},
		{
			name: "output json format",
			opt: &VersionOptions{
				Output: OutputJSON,
			},
			expectedOutput: "\"clientVersion\"",
		},
		{
			name: "unknown output format",
			opt: &VersionOptions{
				Output: "Xml",
			},
			wantErr: fmt.Errorf("VersionOptions were not validated: --output=%q should have been rejected", "Xml"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			err := RunVersion(factory, &stdout, tt.opt)
			require.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			assert.Containsf(
				t,
				stdout.String(),
				tt.expectedOutput,
				"%s : Unexpected output! Expected\n%s\ngot\n%s",
				tt.name,
				tt.expectedOutput,
				stdout,
			)
		})
	}
}
