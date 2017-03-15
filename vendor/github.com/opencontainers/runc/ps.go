// +build linux

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli"
)

var psCommand = cli.Command{
	Name:      "ps",
	Usage:     "ps displays the processes running inside a container",
	ArgsUsage: `<container-id> [ps options]`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "format, f",
			Value: "",
			Usage: `select one of: ` + formatOptions,
		},
	},
	Action: func(context *cli.Context) error {
		container, err := getContainer(context)
		if err != nil {
			return err
		}

		pids, err := container.Processes()
		if err != nil {
			return err
		}

		if context.String("format") == "json" {
			if err := json.NewEncoder(os.Stdout).Encode(pids); err != nil {
				return err
			}
			return nil
		}

		pidlist := []string{}
		for _, pid := range pids {
			pidlist = append(pidlist, fmt.Sprintf("%d", pid))
		}

		// [1:] is to remove command name, ex:
		// context.Args(): [containet_id ps_arg1 ps_arg2 ...]
		// psArgs:         [ps_arg1 ps_arg2 ...]
		//
		psArgs := context.Args()[1:]
		if len(psArgs) == 0 {
			psArgs = []string{"-f"}
		}

		psArgs = append(psArgs, "-p", strings.Join(pidlist, ","))
		output, err := exec.Command("ps", psArgs...).Output()
		if err != nil {
			return err
		}

		fmt.Printf(string(output))
		return nil
	},
	SkipArgReorder: true,
}
