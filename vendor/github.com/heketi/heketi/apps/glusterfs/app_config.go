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

package glusterfs

import (
	"encoding/json"
	"io"
	"os"

	"github.com/heketi/heketi/executors/kubeexec"
	"github.com/heketi/heketi/executors/sshexec"
)

type GlusterFSConfig struct {
	DBfile     string              `json:"db"`
	Executor   string              `json:"executor"`
	Allocator  string              `json:"allocator"`
	SshConfig  sshexec.SshConfig   `json:"sshexec"`
	KubeConfig kubeexec.KubeConfig `json:"kubeexec"`
	Loglevel   string              `json:"loglevel"`

	// advanced settings
	BrickMaxSize int `json:"brick_max_size_gb"`
	BrickMinSize int `json:"brick_min_size_gb"`
	BrickMaxNum  int `json:"max_bricks_per_volume"`
}

type ConfigFile struct {
	GlusterFS GlusterFSConfig `json:"glusterfs"`
}

func loadConfiguration(configIo io.Reader) *GlusterFSConfig {
	configParser := json.NewDecoder(configIo)

	var config ConfigFile
	if err := configParser.Decode(&config); err != nil {
		logger.LogError("Unable to parse config file: %v\n",
			err.Error())
		return nil
	}

	// Set environment variable to override configuration file
	env := os.Getenv("HEKETI_EXECUTOR")
	if env != "" {
		config.GlusterFS.Executor = env
	}

	return &config.GlusterFS
}
