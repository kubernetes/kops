package executor

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"

	"github.com/golang/glog"
)

type Executor interface {
	Close() error

	Run(c *CommandExecution) error

	Put(dest string, length int, content io.Reader, mode os.FileMode) error
	Mkdir(dest string, mode os.FileMode) error
}

type runFunction func(cmd []string) ([]byte, error)

// runCommand is a helper function for executing a command
func runCommand(cmd *CommandExecution, x Executor, runner runFunction) error {
	// Warn if the caller is doing something dumb
	if cmd.Sudo && cmd.Command[0] == "sudo" {
		glog.Warningf("sudo used with command that includes sudo (%q)", cmd.Command)
	}

	var script bytes.Buffer

	needScript := false

	script.WriteString("#!/bin/bash -e\n")
	if cmd.Cwd != "" {
		script.WriteString("cd " + cmd.Cwd + "\n")
		needScript = true
	}
	if cmd.Env != nil && len(cmd.Env) != 0 {
		// Most SSH servers are configured not to accept arbitrary env vars
		for k, v := range cmd.Env {
			/*			err := session.Setenv(k, v)
						if err != nil {
							return fmt.Errorf("error setting env var in SSH session: %v", err)
						}
			*/
			script.WriteString("export " + k + "='" + v + "'\n")
			needScript = true
		}
	}
	script.WriteString(joinCommand(cmd.Command) + "\n")

	cmdToRun := cmd.Command
	if needScript {
		tmpScript := fmt.Sprintf("/tmp/ssh-exec-%d", rand.Int63())
		scriptBytes := script.Bytes()
		err := x.Put(tmpScript, len(scriptBytes), bytes.NewReader(scriptBytes), 0755)
		if err != nil {
			return fmt.Errorf("error uploading temporary script: %v", err)
		}
		defer runner([]string{"rm", "-rf", tmpScript})
		if cmd.Sudo {
			cmdToRun = []string{"sudo", tmpScript}
		} else {
			cmdToRun = []string{tmpScript}
		}
	} else {
		if cmd.Sudo {
			cmdToRun = append([]string{"sudo"}, cmdToRun...)
		}
	}

	// We "lie" about the command we're running when we're using a script
	glog.Infof("Executing command: %q", cmd.Command)
	output, err := runner(cmdToRun)
	if err != nil {
		glog.Infof("Error from SSH command %q: %v", cmd.Command, err)
		glog.Infof("Output was: %s", output)
		return fmt.Errorf("error executing SSH command %q: %v", cmd.Command, err)
	}

	glog.V(2).Infof("Output was: %s", output)
	return nil
}

func joinCommand(argv []string) string {
	// TODO: escaping
	return strings.Join(argv, " ")
}
