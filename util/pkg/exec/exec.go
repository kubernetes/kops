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

package exec

import "strings"

// WithTee returns the command to run the command while piping output to both the log file and stdout/stderr
func WithTee(cmd string, args []string, logfile string) []string {
	// exec so we don't have a shell that doesn't pass signals to the real process
	execCmd := "exec " + cmd + " " + strings.Join(args, " ")

	// NOTE: tee & mkfifo is in /usr/bin in the kube-proxy image, but /bin in other images

	// Bash supports something like this, but dash and other limited shells don't
	//shCmd := "exec &> >(/usr/bin/tee -a " + logfile + "); " + execCmd
	// Instead we create the pipe manually and wire up the tee:
	shCmd := "mkfifo /tmp/pipe; (tee -a " + logfile + " < /tmp/pipe & ) ; " + execCmd + " > /tmp/pipe 2>&1"

	// Execute shell command
	return []string{"/bin/sh", "-c", shCmd}
}
