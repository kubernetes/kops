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

package main

import (
	"io"
	"os"

	"github.com/heketi/heketi/client/cli/go/cmds"
)

var (
	HEKETI_CLI_VERSION           = "(dev)"
	stdout             io.Writer = os.Stdout
	stderr             io.Writer = os.Stderr
	version            bool
)

func main() {
	cmd := cmds.NewHeketiCli(HEKETI_CLI_VERSION, stderr, stdout)
	if err := cmd.Execute(); err != nil {
		//fmt.Println(err) //Should be used for logging
		os.Exit(-1)
	}
}
