// docker_tls_args: find the docker daemon's tls args
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/weaveworks/weave/common"
)

func dockerTLSArgs(args []string) error {
	if len(args) > 0 {
		cmdUsage("docker-tls-args", "")
	}
	procRoot := "/proc"
	pids, err := common.AllPids(procRoot)
	if err != nil {
		return err
	}

	for _, pid := range pids {
		isDaemon := false
		dirName := strconv.Itoa(pid)
		if comm, err := ioutil.ReadFile(filepath.Join(procRoot, dirName, "comm")); err != nil {
			continue
		} else if string(comm) == "dockerd\n" {
			isDaemon = true
		} else if string(comm) != "docker\n" {
			continue
		}

		cmdline, err := ioutil.ReadFile(filepath.Join(procRoot, dirName, "cmdline"))
		if err != nil {
			continue
		}

		tlsArgs := []string{}
		args := bytes.Split(cmdline, []byte{'\000'})
		for i := 0; i < len(args); i++ {
			arg := string(args[i])
			switch {
			case arg == "-d" || arg == "daemon":
				isDaemon = true
			case arg == "--tls", arg == "--tlsverify":
				tlsArgs = append(tlsArgs, arg)
			case strings.HasPrefix(arg, "--tls"):
				tlsArgs = append(tlsArgs, arg)
				if len(args) > i+1 &&
					!strings.Contains(arg, "=") &&
					!strings.HasPrefix(string(args[i+1]), "-") {
					tlsArgs = append(tlsArgs, string(args[i+1]))
					i++
				}
			}
		}
		if !isDaemon {
			continue
		}

		fmt.Println(strings.Join(tlsArgs, " "))
		return nil
	}

	return fmt.Errorf("cannot locate running docker daemon")
}
