//
// Copyright (c) 2016 The heketi Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package kubeexec

import (
	"os"
	"testing"

	"github.com/heketi/heketi/executors/sshexec"
	"github.com/heketi/tests"
)

func TestNewKubeExecutor(t *testing.T) {
	config := &KubeConfig{
		Host: "myhost",
		CLICommandConfig: sshexec.CLICommandConfig{
			Fstab: "myfstab",
		},
		Namespace: "mynamespace",
	}

	k, err := NewKubeExecutor(config)
	tests.Assert(t, err == nil)
	tests.Assert(t, k.Fstab == "myfstab")
	tests.Assert(t, k.Throttlemap != nil)
	tests.Assert(t, k.config != nil)
}

func TestNewKubeExecutorNoNamespace(t *testing.T) {
	config := &KubeConfig{
		Host: "myhost",
		CLICommandConfig: sshexec.CLICommandConfig{
			Fstab: "myfstab",
		},
	}

	k, err := NewKubeExecutor(config)
	tests.Assert(t, err != nil)
	tests.Assert(t, k == nil)
}

func TestNewKubeExecutorRebalanceOnExpansion(t *testing.T) {

	// This tests access to configurations
	// from the sshconfig exector

	config := &KubeConfig{
		Host: "myhost",
		CLICommandConfig: sshexec.CLICommandConfig{
			Fstab: "myfstab",
		},
		Namespace: "mynamespace",
	}

	k, err := NewKubeExecutor(config)
	tests.Assert(t, err == nil)
	tests.Assert(t, k.Fstab == "myfstab")
	tests.Assert(t, k.Throttlemap != nil)
	tests.Assert(t, k.config != nil)
	tests.Assert(t, k.RebalanceOnExpansion() == false)

	config = &KubeConfig{
		Host: "myhost",
		CLICommandConfig: sshexec.CLICommandConfig{
			Fstab:                "myfstab",
			RebalanceOnExpansion: true,
		},
		Namespace: "mynamespace",
	}

	k, err = NewKubeExecutor(config)
	tests.Assert(t, err == nil)
	tests.Assert(t, k.Fstab == "myfstab")
	tests.Assert(t, k.Throttlemap != nil)
	tests.Assert(t, k.config != nil)
	tests.Assert(t, k.RebalanceOnExpansion() == true)
}

func TestKubeExecutorEnvVariables(t *testing.T) {

	// set environment
	err := os.Setenv("HEKETI_SNAPSHOT_LIMIT", "999")
	tests.Assert(t, err == nil)
	defer os.Unsetenv("HEKETI_SNAPSHOT_LIMIT")

	err = os.Setenv("HEKETI_FSTAB", "anotherfstab")
	tests.Assert(t, err == nil)
	defer os.Unsetenv("HEKETI_FSTAB")

	config := &KubeConfig{
		Host: "myhost",
		CLICommandConfig: sshexec.CLICommandConfig{
			Fstab: "myfstab",
		},
		Namespace: "mynamespace",
	}

	k, err := NewKubeExecutor(config)
	tests.Assert(t, err == nil)
	tests.Assert(t, k.Throttlemap != nil)
	tests.Assert(t, k.config != nil)
	tests.Assert(t, k.Fstab == "anotherfstab")
	tests.Assert(t, k.SnapShotLimit() == 999)

}
