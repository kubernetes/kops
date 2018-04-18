/*
Copyright 2017 The Kubernetes Authors.

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

package build

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var GenerateForBuild bool = true
var goos string = "linux"
var goarch string = "amd64"
var outputdir string = "bin"
var Bazel bool
var Gazelle bool

var createBuildExecutablesCmd = &cobra.Command{
	Use:   "executables",
	Short: "Builds the source into executables to run on the local machine",
	Long:  `Builds the source into executables to run on the local machine`,
	Example: `# Generate code and build the apiserver and controller
# binaries in the bin directory so they can be run locally.
apiserver-boot build executables

# Build binaries into the linux/ directory using the cross compiler for linux:amd64
apiserver-boot build executables --goos linux --goarch amd64 --output linux/

# Regenerate Bazel BUILD files, and then build with bazel
# Must first install bazel and gazelle !!!
apiserver-boot build executables --bazel --gazelle

# Run Bazel without generating BUILD files
apiserver-boot build executables --bazel

# Run Bazel without generating BUILD files or generated code
apiserver-boot build executables --bazel --generate=false
`,
	Run: RunBuildExecutables,
}

func AddBuildExecutables(cmd *cobra.Command) {
	cmd.AddCommand(createBuildExecutablesCmd)

	createBuildExecutablesCmd.Flags().BoolVar(&GenUnversionedClient, "gen-unversioned-client", true, "If true, generate unversioned clients.")
	createBuildExecutablesCmd.Flags().StringVar(&vendorDir, "vendor-dir", "", "Location of directory containing vendor files.")
	createBuildExecutablesCmd.Flags().BoolVar(&GenerateForBuild, "generate", true, "if true, generate code before building")
	createBuildExecutablesCmd.Flags().StringVar(&goos, "goos", "", "if specified, set this GOOS")
	createBuildExecutablesCmd.Flags().StringVar(&goarch, "goarch", "", "if specified, set this GOARCH")
	createBuildExecutablesCmd.Flags().StringVar(&outputdir, "output", "bin", "if set, write the binaries to this directory")
	createBuildExecutablesCmd.Flags().BoolVar(&Bazel, "bazel", false, "if true, use bazel to build.  May require updating build rules with gazelle.")
	createBuildExecutablesCmd.Flags().BoolVar(&Gazelle, "gazelle", false, "if true, run gazelle before running bazel.")
}

func RunBuildExecutables(cmd *cobra.Command, args []string) {
	if Bazel {
		BazelBuild(cmd, args)
	} else {
		GoBuild(cmd, args)
	}
}

func BazelBuild(cmd *cobra.Command, args []string) {
	if GenerateForBuild {
		log.Printf("regenerating generated code.  To disable regeneration, run with --generate=false.")
		RunGenerate(cmd, args)
	}

	if Gazelle {
		c := exec.Command("bazel", "run", "//:gazelle")
		fmt.Printf("%s\n", strings.Join(c.Args, " "))
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		err := c.Run()
		if err != nil {
			log.Fatal(err)
		}
	}

	c := exec.Command("bazel", "build",
		filepath.Join("cmd", "apiserver"),
		filepath.Join("cmd", "controller-manager"))
	fmt.Printf("%s\n", strings.Join(c.Args, " "))
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	err := c.Run()
	if err != nil {
		log.Fatal(err)
	}

	os.RemoveAll(filepath.Join("bin", "apiserver"))
	os.RemoveAll(filepath.Join("bin", "controller-manager"))

	c = exec.Command("cp",
		filepath.Join("bazel-bin", "cmd", "apiserver", "apiserver"),
		filepath.Join("bin", "apiserver"))
	fmt.Printf("%s\n", strings.Join(c.Args, " "))
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	err = c.Run()
	if err != nil {
		log.Fatal(err)
	}

	c = exec.Command("cp",
		filepath.Join("bazel-bin", "cmd", "controller-manager", "controller-manager"),
		filepath.Join("bin", "controller-manager"))
	fmt.Printf("%s\n", strings.Join(c.Args, " "))
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	err = c.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func GoBuild(cmd *cobra.Command, args []string) {
	if GenerateForBuild {
		log.Printf("regenerating generated code.  To disable regeneration, run with --generate=false.")
		RunGenerate(cmd, args)
	}

	os.RemoveAll(filepath.Join("bin", "apiserver"))
	os.RemoveAll(filepath.Join("bin", "controller-manager"))

	// Build the apiserver
	path := filepath.Join("cmd", "apiserver", "main.go")
	c := exec.Command("go", "build", "-o", filepath.Join(outputdir, "apiserver"), path)
	c.Env = append(os.Environ(), "CGO_ENABLED=0")
	log.Printf("CGO_ENABLED=0")
	if len(goos) > 0 {
		c.Env = append(c.Env, fmt.Sprintf("GOOS=%s", goos))
		log.Printf(fmt.Sprintf("GOOS=%s", goos))
	}
	if len(goarch) > 0 {
		c.Env = append(c.Env, fmt.Sprintf("GOARCH=%s", goarch))
		log.Printf(fmt.Sprintf("GOARCH=%s", goarch))
	}

	fmt.Printf("%s\n", strings.Join(c.Args, " "))
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	err := c.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Build the controller manager
	path = filepath.Join("cmd", "controller-manager", "main.go")
	c = exec.Command("go", "build", "-o", filepath.Join(outputdir, "controller-manager"), path)
	c.Env = append(os.Environ(), "CGO_ENABLED=0")
	if len(goos) > 0 {
		c.Env = append(c.Env, fmt.Sprintf("GOOS=%s", goos))
	}
	if len(goarch) > 0 {
		c.Env = append(c.Env, fmt.Sprintf("GOARCH=%s", goarch))
	}

	fmt.Println(strings.Join(c.Args, " "))
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	err = c.Run()
	if err != nil {
		log.Fatal(err)
	}
}
