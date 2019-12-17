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

package systemd

import (
	"testing"
)

func TestRawManifest(t *testing.T) {
	expected := `[Unit]
Description=test
Requires=docker.service

[Service]
run the command
in this manifest`
	m := &Manifest{}
	m.Set("Unit", "Description", "test")
	m.Set("Unit", "Requires", "docker.service")
	m.SetSection("Service", `run the command
in this manifest`)

	rendered := m.Render()
	if rendered != expected {
		t.Errorf("the rendered manifest is not as expected: '%v', got: '%v'", expected, rendered)
	}
}

func TestRawMixedManifest(t *testing.T) {
	expected := `[Unit]
Description=test
Requires=docker.service

[Service]
run the command
in this manifest
key=pair
another=pair
`
	m := &Manifest{}
	m.Set("Unit", "Description", "test")
	m.Set("Unit", "Requires", "docker.service")
	m.SetSection("Service", `run the command
in this manifest
`)
	m.Set("Service", "key", "pair")
	m.Set("Service", "another", "pair")

	rendered := m.Render()
	if rendered != expected {
		t.Errorf("the rendered manifest is not as expected: '%v', got: '%v'", expected, rendered)
	}
}

func TestKeyPairOnlyManifest(t *testing.T) {
	expected := `[Unit]
Description=test
Requires=docker.service

[Service]
EnvironmentFile=/etc/somefile
EnvironmentFile=/etc/another_file
StartExecPre=some_command
Start=command
`
	m := &Manifest{}
	m.Set("Unit", "Description", "test")
	m.Set("Unit", "Requires", "docker.service")
	m.Set("Service", "EnvironmentFile", "/etc/somefile")
	m.Set("Service", "EnvironmentFile", "/etc/another_file")
	m.Set("Service", "StartExecPre", "some_command")
	m.Set("Service", "Start", "command")

	rendered := m.Render()
	if rendered != expected {
		t.Errorf("the rendered manifest is not as expected: '%v'\n, got: '%v'\n", expected, rendered)
	}
}
