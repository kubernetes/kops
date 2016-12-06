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
	"os"
	"os/exec"
	"strings"
)

// Will execute a command given a raw command string
func ExecOutput(c, args string, env []string) (string, error) {

	cmdSlice := strings.Split(c, " ")
	if len(cmdSlice) > 1 {
		return "", fmt.Errorf("Invalid command: %s", c)
	}

	args = strings.Replace(args, "\n", " ", -1)
	argsSlice := strings.Split(args, " ")

	cmd := exec.Command(c, argsSlice...)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if len(env) != 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	err := cmd.Run()
	if err != nil {

		edge("Error")
		fmt.Printf("Execution failed: %s %s\n", c, args)
		fmt.Printf("Execution Error: %s", err.Error())
		edge("Error")
		fmt.Printf("Execution stderr: \n%s\n", stderr.String())
		edge("Error")
		fmt.Printf("Execution stdout: \n%s\n", stdout.String())
		edge("Error")

		return "", fmt.Errorf("Execution Error: \n %s %s %s\n", stdout.String(), err.Error(), stderr.String())
	}

	edge("Success")
	fmt.Printf("Command Succeded: %s %s\n", c, args)
	edge("Success")

	return stdout.String(), nil
}

func edge(msg string) {
	fmt.Printf("%s: ==================================================================================================\n", msg)
}
