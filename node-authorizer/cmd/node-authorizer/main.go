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

package main

import (
	"fmt"
	"os"

	"k8s.io/kops/node-authorizer/pkg/server"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Authors = []cli.Author{
		{
			Name:  "Rohith Jayawardene",
			Email: "gambol99@gmail.com",
		},
	}
	app.Commands = []cli.Command{addServerCommand(), addClientCommand()}
	app.Usage = "used to provision the bootstrap tokens for node registration"
	app.Version = server.Version

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "[error] %s\n", err)
		os.Exit(1)
	}
}
