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
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"bytes"

	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate API reference docs from the openapi spec.",
	Long:  `Generate API reference docs from the openapi spec.`,
	Example: `# Edit docs examples
nano -w docs/examples/<kind>/<kind.yaml

# Start a new server, get the swagger.json, and generate docs from the swagger.json
apiserver-boot build executables
apiserver-boot build docs

# Build docs and include operations.
apiserver-boot build docs --operations=true

# Use the swagger.json at docs/openapi-spec/swagger.json instead
# of getting it from a server.
apiserver-boot build docs --build-openapi=false

# Use the server at my/bin/apiserver
apiserver-boot build docs --server my/bin/apiserver

# Instead of generating the table of contents, use the statically defined configuration
# from docs/config.yaml
# See an example config.yaml at in kubernetes-incubator/reference-docs
apiserver-boot build docs --generate-toc=false

# Add manual documentation to the generated docs
# Edit docs/static_includes/*.md
# e.g. docs/static_include/_overview.md

	# <strong>API OVERVIEW</strong>
	Add your markdown here

# Add examples in the right-most column
# Edit docs/examples/<type>/<type>.yaml
# e.g. docs/examples/pod/pod.yaml

	note: <Description of example>.
	sample: |
	  apiVersion: <version>
	  kind: <type>
	  metadata:
	    name: <name>
	  spec:
	    <spec-contents>`,
	Run: RunDocs,
}

var operations, buildOpenapi, generateToc bool
var server string
var disableDelegatedAuth bool
var cleanup bool
var outputDir string

func AddDocs(cmd *cobra.Command) {
	docsCmd.Flags().StringVar(&server, "server", "bin/apiserver", "path to apiserver binary to run to get swagger.json")
	docsCmd.Flags().BoolVar(&cleanup, "cleanup", true, "If true, cleanup intermediary files")
	docsCmd.Flags().BoolVar(&buildOpenapi, "build-openapi", true, "If true, run the server and get the new swagger.json")
	docsCmd.Flags().BoolVar(&operations, "operations", false, "if true, include operations in docs.")
	docsCmd.Flags().BoolVar(&generateToc, "generate-toc", true, "If true, generate the table of contents from the api groups instead of using a statically configured ToC.")
	docsCmd.Flags().BoolVar(&disableDelegatedAuth, "disable-delegated-auth", true, "If true, disable delegated auth in the apiserver with --delegated-auth=false.")
	docsCmd.Flags().StringVar(&outputDir, "output-dir", "docs", "Build docs into this directory")
	cmd.AddCommand(docsCmd)
	docsCmd.AddCommand(docsCleanCmd)
}

var docsCleanCmd = &cobra.Command{
	Use:     "clean",
	Short:   "Removes generated docs",
	Long:    `Removes generated docs`,
	Example: ``,
	Run:     RunCleanDocs,
}

func RunCleanDocs(cmd *cobra.Command, args []string) {
	os.RemoveAll(filepath.Join(outputDir, "build"))
	os.RemoveAll(filepath.Join(outputDir, "includes"))
	os.Remove(filepath.Join(outputDir, "manifest.json"))
}

func RunDocs(cmd *cobra.Command, args []string) {
	if len(server) == 0 && buildOpenapi {
		log.Fatal("Must specifiy --server or --build-openapi=false")
	}

	os.RemoveAll(filepath.Join(outputDir, "includes"))
	os.MkdirAll(filepath.Join(outputDir, "openapi-spec"), 0700)
	os.MkdirAll(filepath.Join(outputDir, "static_includes"), 0700)
	os.MkdirAll(filepath.Join(outputDir, "examples"), 0700)

	// Build the swagger.json
	if buildOpenapi {
		flags := []string{
			"--etcd-servers=http://localhost:2379",
			"--secure-port=9443",
			"--print-openapi",
		}

		if disableDelegatedAuth {
			flags = append(flags, "--delegated-auth=false")
		}

		c := exec.Command(server,
			flags...,
		)
		log.Printf("%s\n", strings.Join(c.Args, " "))

		var b bytes.Buffer
		c.Stdout = &b
		c.Stderr = os.Stderr

		err := c.Run()
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}

		err = ioutil.WriteFile(filepath.Join(outputDir, "openapi-spec", "swagger.json"), b.Bytes(), 0644)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
	}

	// Build the docs
	dir, err := os.Executable()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	dir = filepath.Dir(dir)
	c := exec.Command(filepath.Join(dir, "gen-apidocs"),
		fmt.Sprintf("--build-operations=%v", operations),
		fmt.Sprintf("--use-tags=%v", generateToc),
		"--allow-errors",
		"--config-dir="+outputDir)
	log.Printf("%s\n", strings.Join(c.Args, " "))
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	err = c.Run()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	// Run the docker command to build the docs
	c = exec.Command("docker", "run",
		"-v", fmt.Sprintf("%s:%s", filepath.Join(wd, outputDir, "includes"), "/source"),
		"-v", fmt.Sprintf("%s:%s", filepath.Join(wd, outputDir, "build"), "/build"),
		"-v", fmt.Sprintf("%s:%s", filepath.Join(wd, outputDir, "build"), "/build"),
		"-v", fmt.Sprintf("%s:%s", filepath.Join(wd, outputDir), "/manifest"),
		"pwittrock/brodocs",
	)
	log.Println(strings.Join(c.Args, " "))
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	err = c.Run()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	// Cleanup intermediate files
	if cleanup {
		os.RemoveAll(filepath.Join(wd, outputDir, "includes"))
		os.RemoveAll(filepath.Join(wd, outputDir, "manifest.json"))
		os.RemoveAll(filepath.Join(wd, outputDir, "openapi-spec"))
		os.RemoveAll(filepath.Join(wd, outputDir, "build", "documents"))
		os.RemoveAll(filepath.Join(wd, outputDir, "build", "documents"))
		os.RemoveAll(filepath.Join(wd, outputDir, "build", "runbrodocs.sh"))
		os.RemoveAll(filepath.Join(wd, outputDir, "build", "node_modules", "marked", "Makefile"))
	}
}
