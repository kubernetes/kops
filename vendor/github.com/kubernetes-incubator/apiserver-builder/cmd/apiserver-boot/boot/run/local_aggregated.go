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
	"k8s.io/client-go/util/homedir"
)

var localMinikubeCmd = &cobra.Command{
	Use:   "local-minikube",
	Short: "run the etcd, apiserver and controller-manager locally, but have them aggregated into a local minikube cluster",
	Long:  `run the etcd, apiserver and controller-manager locally, but have them aggregated into a local minikube cluster`,
	Example: `# Regenerate code and build binaries.  Then run them locally, but register them in a minikube
#cluster (default kube config must point to minikube cluster).
kubectl config current-context
apiserver-boot run local-minikube

# Check the api versions of the locally running server
kubectl api-versions

# Run locally without rebuilding
apiserver-boot run local-minikube --build=false

# Run locally using a different kubeconfig
apiserver-boot run local-minikube --kubeconfig <minikube-config>`,

	Run: RunLocalMinikube,
}

var certDir string
var minikubeconfig string
var minikubeport int32
var gazelle bool
var bazel bool
var generate bool

func AddLocalMinikube(cmd *cobra.Command) {
	localMinikubeCmd.Flags().StringSliceVar(&toRun, "run", []string{"etcd", "apiserver", "controller-manager"}, "path to apiserver binary to run")

	localMinikubeCmd.Flags().StringVar(&server, "apiserver", "", "path to apiserver binary to run")
	localMinikubeCmd.Flags().StringVar(&controllermanager, "controller-manager", "", "path to controller-manager binary to run")
	localMinikubeCmd.Flags().StringVar(&etcd, "etcd", "", "if non-empty, use this etcd instead of starting a new one")

	localMinikubeCmd.Flags().StringVar(&minikubeconfig, "config", filepath.Join(homedir.HomeDir(), ".kube", "config"), "path to the core apiserver kubeconfig")

	localMinikubeCmd.Flags().BoolVar(&printapiserver, "print-apiserver", true, "if true, pipe the apiserver stdout and stderr")
	localMinikubeCmd.Flags().BoolVar(&printcontrollermanager, "print-controller-manager", true, "if true, pipe the controller-manager stdout and stderr")
	localMinikubeCmd.Flags().BoolVar(&printetcd, "printetcd", false, "if true, pipe the etcd stdout and stderr")
	localMinikubeCmd.Flags().BoolVar(&buildBin, "build", true, "if true, build the binaries before running")

	localMinikubeCmd.Flags().Int32Var(&minikubeport, "secure-port", 443, "Secure port from apiserver to serve requests")
	localMinikubeCmd.Flags().StringVar(&certDir, "cert-dir", filepath.Join("config", "certificates"), "directory containing apiserver certificates")

	localMinikubeCmd.Flags().BoolVar(&bazel, "bazel", false, "if true, use bazel to build.  May require updating build rules with gazelle.")
	localMinikubeCmd.Flags().BoolVar(&gazelle, "gazelle", false, "if true, run gazelle before running bazel.")
	localMinikubeCmd.Flags().BoolVar(&generate, "generate", true, "if true, generate code before building")

	cmd.AddCommand(localMinikubeCmd)
}

func RunLocalMinikube(cmd *cobra.Command, args []string) {
	config = minikubeconfig
	securePort = minikubeport
	if buildBin {
		build.Bazel = bazel
		build.Gazelle = gazelle
		build.GenerateForBuild = generate
		build.RunBuildExecutables(cmd, args)
	}

	r := map[string]interface{}{}
	for _, s := range toRun {
		r[s] = nil
	}

	fmt.Printf("Checking sudo credentials for binding port 443\n")
	sudo := exec.Command("sudo", "-v")
	sudo.Stderr = os.Stderr
	sudo.Stdout = os.Stdout
	sudo.Stdin = os.Stdin
	err := sudo.Run()
	if err != nil {
		log.Fatalf("Failed to validate sudo credentials %v", err)
		os.Exit(-1)
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
		go RunApiserverMinikube()
		time.Sleep(time.Second * 2)
	}

	// Start controller manager
	if _, f := r["controller-manager"]; f {
		go RunControllerManager()
	}

	fmt.Printf("to test the server run `kubectl api-versions`, if you specified --kubeconfig you must also provide the flag `--kubeconfig %s`\n", config)
	select {} // wait forever
}

func RunApiserverMinikube() *exec.Cmd {
	if len(server) == 0 {
		server = "bin/apiserver"
	}

	flags := []string{
		"-n",
		server,
		fmt.Sprintf("--etcd-servers=%s", etcd),
		fmt.Sprintf("--secure-port=%v", securePort),
		fmt.Sprintf("--tls-cert-file=%s", filepath.Join(certDir, "apiserver.crt")),
		fmt.Sprintf("--tls-private-key-file=%s", filepath.Join(certDir, "apiserver.key")),
		"--delegated-auth=true",
		fmt.Sprintf("--kubeconfig=%s", config),
		fmt.Sprintf("--authentication-kubeconfig=%s", config),
		fmt.Sprintf("--authorization-kubeconfig=%s", config),
		"--authentication-skip-lookup",
		fmt.Sprintf("--audit-webhook-config-file=%s", config),
	}

	apiserverCmd := exec.Command("sudo", flags...)
	fmt.Printf("%s\n", strings.Join(apiserverCmd.Args, " "))
	fmt.Printf("Running apiserver with sudo to bind to port 443.  If this fails run `sudo -v`\n")
	if printapiserver {
		apiserverCmd.Stderr = os.Stderr
		apiserverCmd.Stdout = os.Stdout
		apiserverCmd.Stdin = os.Stdin
	}

	err := apiserverCmd.Run()
	if err != nil {
		defer apiserverCmd.Process.Kill()
		log.Fatalf("Failed to run apiserver %v", err)
		os.Exit(-1)
	}

	return apiserverCmd
}
