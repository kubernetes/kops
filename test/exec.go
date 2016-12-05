/*
Copyright 2016 The Kubernetes Authors.

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

package test

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Will execute a command given a raw command string
func ExecOuput(c, args string, env []string) (string, error) {
	cmdSlice := strings.Split(c, " ")
	if len(cmdSlice) > 1 {
		return "", fmt.Errorf("Invalid command: %s", c)
	}
	argsSlice := strings.Split(args, " ")

	cmd := exec.Command(c, argsSlice...)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if len(env) != 0 {
		cmd.Env = env
	}

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Execution Error: \n %s %s %s", stdout.String(), err.Error(), stderr.String())
	}
	if stderr.String() != "" {
		// Edge case, we have stdout AND stderr
		return stdout.String(), fmt.Errorf("%s", stderr.String())
	}
	return stdout.String(), nil
}
