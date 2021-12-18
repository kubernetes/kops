/*
Copyright 2021 The Kubernetes Authors.

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
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/model/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/mirrors"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

type ToolboxEnrollOptions struct {
	ClusterName   string
	InstanceGroup string

	Host string

	SSHUser string
}

func (o *ToolboxEnrollOptions) InitDefaults() {
	o.SSHUser = os.Getenv("USER")
}

func NewCmdToolboxEnroll(f commandutils.Factory, out io.Writer) *cobra.Command {
	options := &ToolboxEnrollOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:   "enroll [CLUSTER]",
		Short: i18n.T(`Add machine to cluster`),
		Long: templates.LongDesc(i18n.T(`
			Adds an individual machine to the cluster.`)),
		Example: templates.Examples(i18n.T(`
			kops toolbox enroll --name k8s-cluster.example.com
		`)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunToolboxEnroll(cmd.Context(), f, out, options)
		},
	}

	cmd.Flags().StringVar(&options.ClusterName, "cluster", options.ClusterName, "Name of cluster to join")
	cmd.Flags().StringVar(&options.InstanceGroup, "instance-group", options.InstanceGroup, "Name of instance-group to join")

	cmd.Flags().StringVar(&options.Host, "host", options.Host, "IP/hostname for machine to add")

	return cmd
}

func RunToolboxEnroll(ctx context.Context, f commandutils.Factory, out io.Writer, options *ToolboxEnrollOptions) error {
	if options.ClusterName == "" {
		return fmt.Errorf("cluster is required")
	}
	if options.InstanceGroup == "" {
		return fmt.Errorf("instance-group is required")
	}
	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	cluster, err := clientset.GetCluster(ctx, options.ClusterName)
	if err != nil {
		return err
	}

	if cluster == nil {
		return fmt.Errorf("cluster not found %q", options.ClusterName)
	}

	ig, err := clientset.InstanceGroupsFor(cluster).Get(ctx, options.InstanceGroup, v1.GetOptions{})
	if err != nil {
		return err
	}

	keyStore, err := clientset.KeyStore(cluster)
	if err != nil {
		return err
	}

	getAssets := false
	assetBuilder := assets.NewAssetBuilder(cluster, getAssets)

	assets := make(map[architectures.Architecture][]*mirrors.MirroredAsset)

	nodeupAssets := make(map[architectures.Architecture]*mirrors.MirroredAsset)
	for _, arch := range architectures.GetSupported() {
		asset, err := cloudup.NodeUpAsset(assetBuilder, arch)
		if err != nil {
			return err
		}
		nodeupAssets[arch] = asset
	}

	encryptionConfigSecretHash := ""

	nodeupConfigBuilder, err := cloudup.NewNodeUpConfigBuilder(cluster, assetBuilder, assets, encryptionConfigSecretHash)
	if err != nil {
		return err
	}

	var apiserverAdditionalIPs []string
	keysets := make(map[string]*fi.Keyset)

	// {
	// 	defaultCA := &fitasks.Keypair{
	// 		Name:      fi.String(fi.CertificateIDCA),
	// 		Lifecycle: fi.LifecycleExistsAndValidates,
	// 		Subject:   "cn=kubernetes-ca",
	// 		Type:      "ca",
	// 	}

	// 	keys[*defaultCA.Name] = defaultCA
	// }
	// 	modelBuilderContext.AddTask(defaultCA)
	// }

	{
		name := "kubernetes-ca"
		keyset, err := keyStore.FindKeyset(name)
		if err != nil {
			return fmt.Errorf("error finding key %q: %w", name, err)
		}
		keysets[name] = keyset
	}
	_, bootConfig, err := nodeupConfigBuilder.BuildConfig(ig, apiserverAdditionalIPs, keysets)
	if err != nil {
		return err
	}

	// configData, err := utils.YamlMarshal(nodeupConfig)
	// if err != nil {
	// 	return fmt.Errorf("error converting nodeup config to yaml: %w", err)
	// }
	// sum256 := sha256.Sum256(configData)

	// fmt.Printf("configData: %s\n", string(configData))

	bootConfig.CloudProvider = "metal"
	bootConfig.ConfigServer.Server = strings.ReplaceAll(bootConfig.ConfigServer.Server, ".internal.", ".")

	var script resources.NodeUpScript
	script.NodeUpAssets = nodeupAssets
	script.BootConfig = bootConfig

	resource, err := script.Build()
	if err != nil {
		return fmt.Errorf("error building script: %w", err)
	}

	scriptBytes, err := fi.ResourceAsBytes(resource)
	if err != nil {
		return fmt.Errorf("error generating script: %w", err)
	}

	if _, err := os.Stdout.Write(scriptBytes); err != nil {
		return err
	}

	if options.Host != "" {
		if err := enrollHost(ctx, options, string(scriptBytes)); err != nil {
			return err
		}
	}
	return nil
}

func enrollHost(ctx context.Context, options *ToolboxEnrollOptions, nodeupScript string) error {
	sudo := true

	host, err := NewSSHHost(ctx, options.Host, options.SSHUser, sudo)
	if err != nil {
		return err
	}
	defer host.Close()

	publicKeyPath := "/etc/kubernetes/kops/pki/machine/public.pem"

	publicKeyBytes, err := host.readFile(ctx, publicKeyPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			publicKeyBytes = nil
		} else {
			return fmt.Errorf("error reading public key %q: %w", publicKeyPath, err)
		}
	}

	publicKeyBytes = bytes.TrimSpace(publicKeyBytes)

	if len(publicKeyBytes) == 0 {
		if _, err := host.runScript(ctx, scriptCreateKey, ExecOptions{Sudo: sudo, Echo: true}); err != nil {
			return err
		}

		b, err := host.readFile(ctx, publicKeyPath)
		if err != nil {

			return fmt.Errorf("error reading public key %q (after creation): %w", publicKeyPath, err)
		}
		publicKeyBytes = b
	}
	klog.Infof("public key is %s", string(publicKeyBytes))

	if len(nodeupScript) != 0 {
		if _, err := host.runScript(ctx, nodeupScript, ExecOptions{Sudo: sudo, Echo: true}); err != nil {
			return err
		}
	}
	return nil

}

const scriptCreateKey = `
#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

set -x

DIR=/etc/kubernetes/kops/pki/machine/
mkdir -p ${DIR}

if [[ ! -f "${DIR}/private.pem" ]]; then
  openssl ecparam -name prime256v1 -genkey -noout -out "${DIR}/private.pem"
fi

if [[ ! -f "${DIR}/public.pem" ]]; then
  openssl ec -in "${DIR}/private.pem" -pubout -out "${DIR}/public.pem"
fi
`

type SSHHost struct {
	hostname  string
	sshClient *ssh.Client
	sudo      bool
}

func (s *SSHHost) Close() error {
	if s.sshClient != nil {
		if err := s.sshClient.Close(); err != nil {
			return err
		}
		s.sshClient = nil
	}
	return nil
}

func NewSSHHost(ctx context.Context, host string, sshUser string, sudo bool) (*SSHHost, error) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil, fmt.Errorf("cannot connect to SSH agent; SSH_AUTH_SOCK env variable not set")
	}
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH agent with SSH_AUTH_SOCK %q: %w", socket, err)
	}

	agentClient := agent.NewClient(conn)

	sshConfig := &ssh.ClientConfig{
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			klog.Warningf("accepting SSH key %v for %q", key, hostname)
			return nil
		},
		Auth: []ssh.AuthMethod{
			// Use a callback rather than PublicKeys so we only consult the
			// agent once the remote server wants it.
			ssh.PublicKeysCallback(agentClient.Signers),
		},
		User: sshUser,
	}
	sshClient, err := ssh.Dial("tcp", host+":22", sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to SSH to %q (with user %q): %w", host, sshUser, err)
	}
	return &SSHHost{
		hostname:  host,
		sshClient: sshClient,
		sudo:      sudo,
	}, nil
}

func (s *SSHHost) readFile(ctx context.Context, path string) ([]byte, error) {
	p := vfs.NewSSHPath(s.sshClient, s.hostname, path, s.sudo)

	return p.ReadFile()
}

func (s *SSHHost) runScript(ctx context.Context, script string, options ExecOptions) (*CommandOutput, error) {
	var tempDir string
	{
		b := make([]byte, 32)
		if _, err := cryptorand.Read(b); err != nil {
			return nil, fmt.Errorf("error getting random data: %w", err)
		}
		tempDir = path.Join("/tmp", hex.EncodeToString(b))
	}

	scriptPath := path.Join(tempDir, "script.sh")

	p := vfs.NewSSHPath(s.sshClient, s.hostname, scriptPath, s.sudo)

	defer func() {
		if _, err := s.runCommand(ctx, "rm -rf "+tempDir, ExecOptions{Sudo: s.sudo, Echo: false}); err != nil {
			klog.Warningf("error cleaning up temp directory %q: %v", tempDir, err)
		}
	}()

	if err := p.WriteFile(bytes.NewReader([]byte(script)), nil); err != nil {
		return nil, fmt.Errorf("error writing script to SSH target: %w", err)
	}

	// if _, err := s.runCommand(ctx, "chmod +x "+scriptPath); err != nil {
	// 	return nil, fmt.Errorf("error marking script as executable: %w", err)
	// }

	scriptCommand := "/bin/bash " + scriptPath
	return s.runCommand(ctx, scriptCommand, options)
}

type CommandOutput struct {
	Stdout bytes.Buffer
	Stderr bytes.Buffer
}

type ExecOptions struct {
	Sudo bool
	Echo bool
}

func (s *SSHHost) runCommand(ctx context.Context, command string, options ExecOptions) (*CommandOutput, error) {
	session, err := s.sshClient.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start SSH session: %w", err)
	}
	defer session.Close()

	output := &CommandOutput{}

	session.Stdout = &output.Stdout
	session.Stderr = &output.Stderr

	if options.Echo {
		session.Stdout = io.MultiWriter(os.Stdout, session.Stdout)
		session.Stderr = io.MultiWriter(os.Stderr, session.Stderr)
	}
	if options.Sudo {
		command = "sudo " + command
	}
	if err := session.Run(command); err != nil {
		return output, fmt.Errorf("error running command %q: %w", command, err)
	}
	return output, nil
}
