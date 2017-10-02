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

package run

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot/boot/build"
	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot/boot/util"
)

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "run the etcd, apiserver and controller-manager",
	Long:  `run the etcd, apiserver and controller-manager`,
	Example: `# Regenerate code and build binaries.  Then run them locally.
apiserver-boot run local

# Check the api versions of the locally running server
kubectl --kubeconfig kubeconfig api-versions

# Run locally without rebuilding
apiserver-boot run local --build=false

# Create an instance and fetch it
nano -w samples/<type>.yaml
kubectl --kubeconfig kubeconfig apply -f samples/<type>.yaml
kubectl --kubeconfig kubeconfig get <type>`,
	Run: RunLocal,
}

var etcd string
var config string
var printapiserver bool
var printcontrollermanager bool
var printetcd bool
var buildBin bool

var server string
var controllermanager string
var toRun []string
var disableDelegatedAuth bool
var securePort int32

func AddLocal(cmd *cobra.Command) {
	localCmd.Flags().StringSliceVar(&toRun, "run", []string{"etcd", "apiserver", "controller-manager"}, "path to apiserver binary to run")
	localCmd.Flags().BoolVar(&disableDelegatedAuth, "disable-delegated-auth", true, "If true, disable delegated auth in the apiserver with --delegated-auth=false.")

	localCmd.Flags().StringVar(&server, "apiserver", "", "path to apiserver binary to run")
	localCmd.Flags().StringVar(&controllermanager, "controller-manager", "", "path to controller-manager binary to run")
	localCmd.Flags().StringVar(&etcd, "etcd", "", "if non-empty, use this etcd instead of starting a new one")

	localCmd.Flags().StringVar(&config, "config", "kubeconfig", "path to the kubeconfig to write for using kubectl")

	localCmd.Flags().BoolVar(&printapiserver, "print-apiserver", true, "if true, pipe the apiserver stdout and stderr")
	localCmd.Flags().BoolVar(&printcontrollermanager, "print-controller-manager", true, "if true, pipe the controller-manager stdout and stderr")
	localCmd.Flags().BoolVar(&printetcd, "printetcd", false, "if true, pipe the etcd stdout and stderr")
	localCmd.Flags().BoolVar(&buildBin, "build", true, "if true, build the binaries before running")

	localCmd.Flags().Int32Var(&securePort, "secure-port", 9443, "Secure port from apiserver to serve requests")

	localCmd.Flags().BoolVar(&bazel, "bazel", false, "if true, use bazel to build.  May require updating build rules with gazelle.")
	localCmd.Flags().BoolVar(&gazelle, "gazelle", false, "if true, run gazelle before running bazel.")
	localCmd.Flags().BoolVar(&generate, "generate", true, "if true, generate code before building")

	cmd.AddCommand(localCmd)
}

func RunLocal(cmd *cobra.Command, args []string) {
	if buildBin {
		build.Bazel = bazel
		build.Gazelle = gazelle
		build.GenerateForBuild = generate
		build.RunBuildExecutables(cmd, args)
	}

	WriteKubeConfig()

	r := map[string]interface{}{}
	for _, s := range toRun {
		r[s] = nil
	}

	// Start etcd
	if _, f := r["etcd"]; f {
		etcd = "http://localhost:2379"
		etcdCmd := RunEtcd()
		defer etcdCmd.Process.Kill()
		time.Sleep(time.Second * 2)
	}

	// Start apiserver
	if _, f := r["apiserver"]; f {
		go RunApiserver()
		time.Sleep(time.Second * 2)
	}

	// Start controller manager
	if _, f := r["controller-manager"]; f {
		go RunControllerManager()
	}

	fmt.Printf("to test the server run `kubectl --kubeconfig %s api-versions`\n", config)
	select {} // wait forever
}

func RunEtcd() *exec.Cmd {
	etcdCmd := exec.Command("etcd")
	if printetcd {
		etcdCmd.Stderr = os.Stderr
		etcdCmd.Stdout = os.Stdout
	}

	fmt.Printf("%s\n", strings.Join(etcdCmd.Args, " "))
	go func() {
		err := etcdCmd.Run()
		defer etcdCmd.Process.Kill()
		if err != nil {
			log.Fatalf("Failed to run etcd %v", err)
			os.Exit(-1)
		}
	}()
	return etcdCmd
}

func RunApiserver() *exec.Cmd {
	if len(server) == 0 {
		server = "bin/apiserver"
	}

	flags := []string{
		fmt.Sprintf("--etcd-servers=%s", etcd),
		fmt.Sprintf("--secure-port=%v", securePort),
	}

	if disableDelegatedAuth {
		flags = append(flags, "--delegated-auth=false")
	}

	apiserverCmd := exec.Command(server,
		flags...,
	)
	fmt.Printf("%s\n", strings.Join(apiserverCmd.Args, " "))
	if printapiserver {
		apiserverCmd.Stderr = os.Stderr
		apiserverCmd.Stdout = os.Stdout
	}

	err := apiserverCmd.Run()
	if err != nil {
		defer apiserverCmd.Process.Kill()
		log.Fatalf("Failed to run apiserver %v", err)
		os.Exit(-1)
	}

	return apiserverCmd
}

func RunControllerManager() *exec.Cmd {
	if len(controllermanager) == 0 {
		controllermanager = "bin/controller-manager"
	}
	controllerManagerCmd := exec.Command(controllermanager,
		fmt.Sprintf("--kubeconfig=%s", config),
	)
	fmt.Printf("%s\n", strings.Join(controllerManagerCmd.Args, " "))
	if printcontrollermanager {
		controllerManagerCmd.Stderr = os.Stderr
		controllerManagerCmd.Stdout = os.Stdout
	}

	err := controllerManagerCmd.Run()
	if err != nil {
		defer controllerManagerCmd.Process.Kill()
		log.Fatalf("Failed to run controller-manager %v", err)
		os.Exit(-1)
	}

	return controllerManagerCmd
}

func WriteKubeConfig() {
	// Write a kubeconfig
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Cannot get working directory %v", err)
		os.Exit(-1)
	}
	path := filepath.Join(dir, "apiserver.local.config", "certificates", "apiserver")
	util.WriteIfNotFound(config, "kubeconfig-template", configTemplate, ConfigArgs{Path: path, Port: fmt.Sprintf("%v", securePort)})
}

type ConfigArgs struct {
	Path string
	Port string
}

var configTemplate = `
apiVersion: v1
clusters:
- cluster:
    certificate-authority: {{ .Path }}.crt
    server: https://localhost:{{ .Port }}
  name: apiserver
contexts:
- context:
    cluster: apiserver
    user: apiserver
  name: apiserver
current-context: apiserver
kind: Config
preferences: {}
users:
- name: apiserver
  user:
    client-certificate: {{ .Path }}.crt
    client-key: {{ .Path }}.key
`
