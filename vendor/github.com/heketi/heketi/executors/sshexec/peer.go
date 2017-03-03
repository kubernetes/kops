//
// Copyright (c) 2015 The heketi Authors
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
	"fmt"
	"github.com/lpabon/godbc"
)

// :TODO: Rename this function to NodeInit or something
func (s *SshExecutor) PeerProbe(host, newnode string) error {

	godbc.Require(host != "")
	godbc.Require(newnode != "")

	logger.Info("Probing: %v -> %v", host, newnode)
	// create the commands
	commands := []string{
		fmt.Sprintf("gluster peer probe %v", newnode),
	}
	_, err := s.RemoteExecutor.RemoteCommandExecute(host, commands, 10)
	if err != nil {
		return err
	}

	// Determine if there is a snapshot limit configuration setting
	if s.RemoteExecutor.SnapShotLimit() > 0 {
		logger.Info("Setting snapshot limit")
		commands = []string{
			fmt.Sprintf("gluster --mode=script snapshot config snap-max-hard-limit %v",
				s.RemoteExecutor.SnapShotLimit()),
		}
		_, err := s.RemoteExecutor.RemoteCommandExecute(host, commands, 10)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SshExecutor) PeerDetach(host, detachnode string) error {
	godbc.Require(host != "")
	godbc.Require(detachnode != "")

	// create the commands
	logger.Info("Detaching node %v", detachnode)
	commands := []string{
		fmt.Sprintf("gluster peer detach %v", detachnode),
	}
	_, err := s.RemoteExecutor.RemoteCommandExecute(host, commands, 10)
	if err != nil {
		logger.Err(err)
	}

	return nil
}
