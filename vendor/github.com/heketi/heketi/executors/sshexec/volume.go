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
	"encoding/xml"
	"fmt"

	"github.com/heketi/heketi/executors"
	"github.com/lpabon/godbc"
)

func (s *SshExecutor) VolumeCreate(host string,
	volume *executors.VolumeRequest) (*executors.VolumeInfo, error) {

	godbc.Require(volume != nil)
	godbc.Require(host != "")
	godbc.Require(len(volume.Bricks) > 0)
	godbc.Require(volume.Name != "")

	// Create volume command
	cmd := fmt.Sprintf("gluster --mode=script volume create %v ", volume.Name)

	// Add durability settings to the volume command
	var (
		inSet     int
		maxPerSet int
	)
	switch volume.Type {
	case executors.DurabilityNone:
		logger.Info("Creating volume %v with no durability", volume.Name)
		inSet = 1
		maxPerSet = 15
	case executors.DurabilityReplica:
		logger.Info("Creating volume %v replica %v", volume.Name, volume.Replica)
		cmd += fmt.Sprintf("replica %v ", volume.Replica)
		inSet = volume.Replica
		maxPerSet = 5
	case executors.DurabilityDispersion:
		logger.Info("Creating volume %v dispersion %v+%v",
			volume.Name, volume.Data, volume.Redundancy)
		cmd += fmt.Sprintf("disperse-data %v redundancy %v ", volume.Data, volume.Redundancy)
		inSet = volume.Data + volume.Redundancy
		maxPerSet = 1
	}

	// Setup volume create command
	// There could many, many bricks which could make the command line
	// too long.  Instead, create the volume first, then add each brick set.
	for _, brick := range volume.Bricks[:inSet] {
		cmd += fmt.Sprintf("%v:%v ", brick.Host, brick.Path)
	}

	// Initialize the commands with the create command
	commands := []string{cmd}

	// Now add all the commands to add the bricks
	commands = append(commands, s.createAddBrickCommands(volume, inSet, inSet, maxPerSet)...)

	// Add command to start the volume
	commands = append(commands, fmt.Sprintf("gluster --mode=script volume start %v", volume.Name))

	// Execute command
	_, err := s.RemoteExecutor.RemoteCommandExecute(host, commands, 10)
	if err != nil {
		s.VolumeDestroy(host, volume.Name)
		return nil, err
	}

	return &executors.VolumeInfo{}, nil
}

func (s *SshExecutor) VolumeExpand(host string,
	volume *executors.VolumeRequest) (*executors.VolumeInfo, error) {

	godbc.Require(volume != nil)
	godbc.Require(host != "")
	godbc.Require(len(volume.Bricks) > 0)
	godbc.Require(volume.Name != "")

	// Add durability settings to the volume command
	var (
		inSet     int
		maxPerSet int
	)
	switch volume.Type {
	case executors.DurabilityNone:
		inSet = 1
		maxPerSet = 15
	case executors.DurabilityReplica:
		inSet = volume.Replica
		maxPerSet = 5
	case executors.DurabilityDispersion:
		inSet = volume.Data + volume.Redundancy
		maxPerSet = 1
	}

	// Setup volume create command
	commands := s.createAddBrickCommands(volume,
		0, // start at the beginning of the brick list
		inSet,
		maxPerSet)

	// Rebalance if configured
	if s.RemoteExecutor.RebalanceOnExpansion() {
		commands = append(commands,
			fmt.Sprintf("gluster --mode=script volume rebalance %v start", volume.Name))
	}

	// Execute command
	_, err := s.RemoteExecutor.RemoteCommandExecute(host, commands, 10)
	if err != nil {
		return nil, err
	}

	return &executors.VolumeInfo{}, nil
}

func (s *SshExecutor) VolumeDestroy(host string, volume string) error {
	godbc.Require(host != "")
	godbc.Require(volume != "")

	// Shutdown volume
	commands := []string{
		// stop gluster volume
		fmt.Sprintf("gluster --mode=script volume stop %v force", volume),
	}

	// Execute command
	_, err := s.RemoteExecutor.RemoteCommandExecute(host, commands, 10)
	if err != nil {
		logger.LogError("Unable to stop volume %v: %v", volume, err)
	}

	// Shutdown volume
	commands = []string{
		// stop gluster volume
		fmt.Sprintf("gluster --mode=script volume delete %v", volume),
	}

	// Execute command
	_, err = s.RemoteExecutor.RemoteCommandExecute(host, commands, 10)
	if err != nil {
		return logger.Err(fmt.Errorf("Unable to delete volume %v: %v", volume, err))
	}

	return nil
}

func (s *SshExecutor) VolumeDestroyCheck(host, volume string) error {
	godbc.Require(host != "")
	godbc.Require(volume != "")

	// Determine if the volume is able to be deleted
	err := s.checkForSnapshots(host, volume)
	if err != nil {
		return err
	}

	return nil
}

func (s *SshExecutor) createAddBrickCommands(volume *executors.VolumeRequest,
	start, inSet, maxPerSet int) []string {

	commands := []string{}
	var cmd string

	// Go through all the bricks and create add-brick commands
	for index, brick := range volume.Bricks[start:] {
		if index%(inSet*maxPerSet) == 0 {
			if cmd != "" {
				// Add add-brick command to the command list
				commands = append(commands, cmd)
			}

			// Create a new add-brick command
			cmd = fmt.Sprintf("gluster --mode=script volume add-brick %v ", volume.Name)
		}

		// Add this brick to the add-brick command
		cmd += fmt.Sprintf("%v:%v ", brick.Host, brick.Path)
	}

	// Add the last add-brick command to the command list
	commands = append(commands, cmd)

	return commands
}

func (s *SshExecutor) checkForSnapshots(host, volume string) error {

	// Structure used to unmarshal XML from snapshot gluster cli
	type CliOutput struct {
		SnapList struct {
			Count int `xml:"count"`
		} `xml:"snapList"`
	}

	// Get snapshot information for the specified volume
	commands := []string{
		fmt.Sprintf("gluster --mode=script snapshot list %v --xml", volume),
	}

	// Execute command
	output, err := s.RemoteExecutor.RemoteCommandExecute(host, commands, 10)
	if err != nil {
		return fmt.Errorf("Unable to get snapshot information from volume %v: %v", volume, err)
	}

	var snapInfo CliOutput
	err = xml.Unmarshal([]byte(output[0]), &snapInfo)
	if err != nil {
		return fmt.Errorf("Unable to determine snapshot information from volume %v: %v", volume, err)
	}

	if snapInfo.SnapList.Count > 0 {
		return fmt.Errorf("Unable to delete volume %v because it contains %v snapshots",
			volume, snapInfo.SnapList.Count)
	}

	return nil
}
