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

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/bundle"
	"k8s.io/kops/upup/pkg/kutil"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/kubectl/util/templates"
)

var (
	toolboxBundleLong = templates.LongDesc(i18n.T(`
	Creates a bundle for enrolling a bare metal machine.`))

	toolboxBundleExample = templates.Examples(i18n.T(`
	# Bundle
	kops toolbox bundle --name k8s-cluster.example.com
	`))

	toolboxBundleShort = i18n.T(`Bundle cluster information`)
)

type ToolboxBundleOptions struct {
	// Target is the machine we are enrolling in the cluster
	Target string
}

func (o *ToolboxBundleOptions) InitDefaults() {
}

func NewCmdToolboxBundle(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ToolboxBundleOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:     "bundle",
		Short:   toolboxBundleShort,
		Long:    toolboxBundleLong,
		Example: toolboxBundleExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunToolboxBundle(f, out, options, args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVar(&options.Target, "target", options.Target, "machine to target (IP address)")

	return cmd
}

func RunToolboxBundle(context Factory, out io.Writer, options *ToolboxBundleOptions, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("specify name of instance group for node")
	}
	if len(args) != 1 {
		return fmt.Errorf("can only specify one instance group")
	}

	if options.Target == "" {
		return fmt.Errorf("target is required")
	}
	groupName := args[0]

	cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	clientset, err := context.Clientset()
	if err != nil {
		return err
	}

	ig, err := clientset.InstanceGroupsFor(cluster).Get(groupName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error reading InstanceGroup %q: %v", groupName, err)
	}
	if ig == nil {
		return fmt.Errorf("InstanceGroup %q not found", groupName)
	}

	builder := bundle.Builder{
		Clientset: clientset,
	}
	bundleData, err := builder.Build(cluster, ig)
	if err != nil {
		return fmt.Errorf("error building bundle: %v", err)
	}

	sshUser := os.Getenv("USER")

	nodeSSH := &kutil.NodeSSH{
		Hostname: options.Target,
	}
	nodeSSH.SSHConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	nodeSSH.SSHConfig.User = sshUser
	sshIdentity := filepath.Join(homedir.HomeDir(), ".ssh", "id_rsa")
	if err := kutil.AddSSHIdentity(&nodeSSH.SSHConfig, sshIdentity); err != nil {
		return err
	}

	sshClient, err := nodeSSH.GetSSHClient()
	if err != nil {
		return fmt.Errorf("error getting SSH client: %v", err)
	}

	if err := runSshCommand(sshClient, "sudo mkdir -p /etc/kubernetes/bootstrap"); err != nil {
		return err
	}

	root, err := nodeSSH.Root()
	if err != nil {
		return fmt.Errorf("error connecting to nodeSSH: %v", err)
	}
	for _, file := range bundleData.Files {
		sshAcl := &vfs.SSHAcl{
			Mode: file.Header.FileInfo().Mode(),
		}
		p := root.Join("etc", "kubernetes", "bootstrap", file.Header.Name)
		klog.Infof("writing %s", p)
		if err := p.WriteFile(bytes.NewReader(file.Data), sshAcl); err != nil {
			return fmt.Errorf("error writing file %q: %v", file.Header.Name, err)
		}
	}

	if err := runSshCommand(sshClient, "sudo /etc/kubernetes/bootstrap/bootstrap.sh"); err != nil {
		return err
	}

	return nil
}

func runSshCommand(sshClient *ssh.Client, cmd string) error {
	s, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("error creating ssh session: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	s.Stdout = io.MultiWriter(&stdout, os.Stdout)
	s.Stderr = io.MultiWriter(&stderr, os.Stderr)

	klog.Infof("running %s", cmd)
	if err := s.Run(cmd); err != nil {
		return fmt.Errorf("error running %s: %v\nstdout: %s\nstderr: %s", cmd, err, stdout.String(), stderr.String())
	}

	klog.Infof("stdout: %s", stdout.String())
	klog.Infof("stderr: %s", stderr.String())
	return nil
}

// bazel build //cmd/kops && bazel-bin/cmd/kops/kops toolbox bundle --name ${CLUSTER} ${IGNAME} && scp output.tar.gz ${TARGET}:/tmp/output.tar.gz
// sudo apt-get install --yes ca-certificates
// sudo mkdir -p /etc/kubernetes/bootstrap
// sudo tar zx -C /etc/kubernetes/bootstrap -f /tmp/output.tar.gz
// sudo /etc/kubernetes/bootstrap/bootstrap.sh
