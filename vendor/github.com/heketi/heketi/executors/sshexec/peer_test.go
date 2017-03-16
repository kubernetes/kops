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

package sshexec

import (
	"testing"

	"github.com/heketi/heketi/pkg/utils"
	"github.com/heketi/tests"
)

func TestSshExecPeerProbe(t *testing.T) {

	f := NewFakeSsh()
	defer tests.Patch(&sshNew,
		func(logger *utils.Logger, user string, file string) (Ssher, error) {
			return f, nil
		}).Restore()

	config := &SshConfig{
		PrivateKeyFile: "xkeyfile",
		User:           "xuser",
		CLICommandConfig: CLICommandConfig{
			Fstab: "/my/fstab",
		},
	}

	s, err := NewSshExecutor(config)
	tests.Assert(t, err == nil)
	tests.Assert(t, s != nil)

	// Mock ssh function
	f.FakeConnectAndExec = func(host string,
		commands []string,
		timeoutMinutes int,
		useSudo bool) ([]string, error) {

		tests.Assert(t, host == "host:22", host)
		tests.Assert(t, len(commands) == 1)
		tests.Assert(t, commands[0] == "gluster peer probe newnode", commands)

		return nil, nil
	}

	// Call function
	err = s.PeerProbe("host", "newnode")
	tests.Assert(t, err == nil, err)

	// Now set the snapshot limit
	config = &SshConfig{
		PrivateKeyFile: "xkeyfile",
		User:           "xuser",
		CLICommandConfig: CLICommandConfig{
			Fstab:         "/my/fstab",
			SnapShotLimit: 14,
		},
	}

	s, err = NewSshExecutor(config)
	tests.Assert(t, err == nil)
	tests.Assert(t, s != nil)

	// Mock ssh function
	count := 0
	f.FakeConnectAndExec = func(host string,
		commands []string,
		timeoutMinutes int,
		useSudo bool) ([]string, error) {

		switch count {
		case 0:
			tests.Assert(t, host == "host:22", host)
			tests.Assert(t, len(commands) == 1)
			tests.Assert(t, commands[0] == "gluster peer probe newnode", commands)

		case 1:
			tests.Assert(t, host == "host:22", host)
			tests.Assert(t, len(commands) == 1)
			tests.Assert(t, commands[0] == "gluster --mode=script snapshot config snap-max-hard-limit 14", commands)

		default:
			tests.Assert(t, false, "Should not be reached")
		}
		count++

		return nil, nil
	}

	// Call function
	err = s.PeerProbe("host", "newnode")
	tests.Assert(t, err == nil, err)
	tests.Assert(t, count == 2)

}
