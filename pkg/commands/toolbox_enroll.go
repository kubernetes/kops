/*
Copyright 2023 The Kubernetes Authors.

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

package commands

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
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/resources"
	"k8s.io/kops/pkg/nodemodel"
	"k8s.io/kops/pkg/wellknownservices"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/vfs"
)

type ToolboxEnrollOptions struct {
	ClusterName   string
	InstanceGroup string

	Host string

	SSHUser string
	SSHPort int
}

func (o *ToolboxEnrollOptions) InitDefaults() {
	o.SSHUser = "root"
	o.SSHPort = 22
}

func RunToolboxEnroll(ctx context.Context, f commandutils.Factory, out io.Writer, options *ToolboxEnrollOptions) error {
	if !featureflag.Metal.Enabled() {
		return fmt.Errorf("bare-metal support requires the Metal feature flag to be enabled")
	}
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

	channel, err := cloudup.ChannelForCluster(clientset.VFSContext(), cluster)
	if err != nil {
		return fmt.Errorf("getting channel for cluster %q: %w", options.ClusterName, err)
	}

	instanceGroupList, err := clientset.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return err
	}

	// The assetBuilder is used primarily to remap images.
	var assetBuilder *assets.AssetBuilder
	{
		// ApplyClusterCmd is get the assets.
		// We use DryRun and GetAssets to do this without applying any changes.
		apply := &cloudup.ApplyClusterCmd{
			Cloud:      cloud,
			Cluster:    cluster,
			Clientset:  clientset,
			DryRun:     true,
			GetAssets:  true,
			TargetName: cloudup.TargetDryRun,
		}
		applyResults, err := apply.Run(ctx)
		if err != nil {
			return fmt.Errorf("error during apply: %w", err)
		}
		assetBuilder = applyResults.AssetBuilder
	}

	// Populate the full cluster and instanceGroup specs.
	var fullInstanceGroup *kops.InstanceGroup
	var fullCluster *kops.Cluster
	{
		var instanceGroups []*kops.InstanceGroup
		for i := range instanceGroupList.Items {
			instanceGroup := &instanceGroupList.Items[i]
			instanceGroups = append(instanceGroups, instanceGroup)
		}

		populatedCluster, err := cloudup.PopulateClusterSpec(ctx, clientset, cluster, instanceGroups, cloud, assetBuilder)
		if err != nil {
			return fmt.Errorf("building full cluster spec: %w", err)
		}
		fullCluster = populatedCluster

		// Build full IG spec to ensure we end up with a valid IG
		for _, ig := range instanceGroups {
			if ig.Name != options.InstanceGroup {
				continue
			}
			populated, err := cloudup.PopulateInstanceGroupSpec(fullCluster, ig, cloud, channel)
			if err != nil {
				return err
			}
			fullInstanceGroup = populated
		}
	}
	if fullInstanceGroup == nil {
		return fmt.Errorf("instance group %q not found", options.InstanceGroup)
	}

	// Determine the well-known addresses for the cluster.
	wellKnownAddresses := make(model.WellKnownAddresses)
	{
		ingresses, err := cloud.GetApiIngressStatus(fullCluster)
		if err != nil {
			return fmt.Errorf("error getting ingress status: %v", err)
		}

		for _, ingress := range ingresses {
			// TODO: Do we need to support hostnames?
			// if ingress.Hostname != "" {
			// 	apiserverAdditionalIPs = append(apiserverAdditionalIPs, ingress.Hostname)
			// }
			if ingress.IP != "" {
				wellKnownAddresses[wellknownservices.KubeAPIServer] = append(wellKnownAddresses[wellknownservices.KubeAPIServer], ingress.IP)
			}
		}
	}
	if len(wellKnownAddresses[wellknownservices.KubeAPIServer]) == 0 {
		// TODO: Should we support DNS?
		return fmt.Errorf("unable to determine IP address for kube-apiserver")
	}
	for k := range wellKnownAddresses {
		sort.Strings(wellKnownAddresses[k])
	}

	// Build the bootstrap data for this node.
	bootstrapData, err := buildBootstrapData(ctx, clientset, fullCluster, fullInstanceGroup, wellKnownAddresses)
	if err != nil {
		return fmt.Errorf("building bootstrap data: %w", err)
	}

	// Enroll the node over SSH.
	if options.Host != "" {
		// TODO: This is the pattern we use a lot, but should we try to access it directly?
		contextName := fullCluster.ObjectMeta.Name
		clientGetter := genericclioptions.NewConfigFlags(true)
		clientGetter.Context = &contextName

		restConfig, err := clientGetter.ToRESTConfig()
		if err != nil {
			return fmt.Errorf("cannot load kubecfg settings for %q: %w", contextName, err)
		}

		if err := enrollHost(ctx, fullInstanceGroup, options, bootstrapData, restConfig); err != nil {
			return err
		}
	}

	return nil
}

func enrollHost(ctx context.Context, ig *kops.InstanceGroup, options *ToolboxEnrollOptions, bootstrapData *bootstrapData, restConfig *rest.Config) error {
	scheme := runtime.NewScheme()
	if err := v1alpha2.AddToScheme(scheme); err != nil {
		return fmt.Errorf("building kubernetes scheme: %w", err)
	}
	kubeClient, err := client.New(restConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return fmt.Errorf("building kubernetes client: %w", err)
	}

	sudo := true
	if options.SSHUser == "root" {
		sudo = false
	}

	host, err := NewSSHHost(ctx, options.Host, options.SSHPort, options.SSHUser, sudo)
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

	hostname, err := host.getHostname(ctx)
	if err != nil {
		return err
	}

	// We can't create the host resource in the API server for control-plane nodes,
	// because the API server (likely) isn't running yet.
	if !ig.IsControlPlane() {
		if err := createHostResourceInAPIServer(ctx, options, hostname, publicKeyBytes, kubeClient); err != nil {
			return err
		}
	}

	for k, v := range bootstrapData.configFiles {
		if err := host.writeFile(ctx, k, bytes.NewReader(v)); err != nil {
			return fmt.Errorf("writing file %q over SSH: %w", k, err)
		}
	}

	if len(bootstrapData.nodeupScript) != 0 {
		if _, err := host.runScript(ctx, string(bootstrapData.nodeupScript), ExecOptions{Sudo: sudo, Echo: true}); err != nil {
			return err
		}
	}
	return nil
}

func createHostResourceInAPIServer(ctx context.Context, options *ToolboxEnrollOptions, nodeName string, publicKey []byte, client client.Client) error {
	host := &v1alpha2.Host{}
	host.Namespace = "kops-system"
	host.Name = nodeName
	host.Spec.InstanceGroup = options.InstanceGroup
	host.Spec.PublicKey = string(publicKey)

	if err := client.Create(ctx, host); err != nil {
		return fmt.Errorf("failed to create host %s/%s: %w", host.Namespace, host.Name, err)
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

// SSHHost is a wrapper around an SSH connection to a host machine.
type SSHHost struct {
	hostname  string
	sshClient *ssh.Client
	sudo      bool
}

// Close closes the connection.
func (s *SSHHost) Close() error {
	if s.sshClient != nil {
		if err := s.sshClient.Close(); err != nil {
			return err
		}
		s.sshClient = nil
	}
	return nil
}

// NewSSHHost creates a new SSHHost.
func NewSSHHost(ctx context.Context, host string, sshPort int, sshUser string, sudo bool) (*SSHHost, error) {
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
	sshClient, err := ssh.Dial("tcp", host+":"+strconv.Itoa(sshPort), sshConfig)
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

	return p.ReadFile(ctx)
}

func (s *SSHHost) writeFile(ctx context.Context, path string, data io.ReadSeeker) error {
	p := vfs.NewSSHPath(s.sshClient, s.hostname, path, s.sudo)
	return p.WriteFile(ctx, data, nil)
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

	if err := p.WriteFile(ctx, bytes.NewReader([]byte(script)), nil); err != nil {
		return nil, fmt.Errorf("error writing script to SSH target: %w", err)
	}

	scriptCommand := "/bin/bash " + scriptPath
	return s.runCommand(ctx, scriptCommand, options)
}

// CommandOutput holds the results of running a command.
type CommandOutput struct {
	Stdout bytes.Buffer
	Stderr bytes.Buffer
}

// ExecOptions holds options for running a command remotely.
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

// getHostname gets the hostname of the SSH target.
// This is used as the node name when registering the node.
func (s *SSHHost) getHostname(ctx context.Context) (string, error) {
	output, err := s.runCommand(ctx, "hostname", ExecOptions{Sudo: false, Echo: true})
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}

	hostname := output.Stdout.String()
	hostname = strings.TrimSpace(hostname)
	if len(hostname) == 0 {
		return "", fmt.Errorf("hostname was empty")
	}
	return hostname, nil
}

type bootstrapData struct {
	nodeupScript []byte
	configFiles  map[string][]byte
}

func buildBootstrapData(ctx context.Context, clientset simple.Clientset, cluster *kops.Cluster, ig *kops.InstanceGroup, wellknownAddresses model.WellKnownAddresses) (*bootstrapData, error) {
	bootstrapData := &bootstrapData{}
	bootstrapData.configFiles = make(map[string][]byte)

	getAssets := false
	assetBuilder := assets.NewAssetBuilder(clientset.VFSContext(), cluster.Spec.Assets, cluster.Spec.KubernetesVersion, getAssets)

	encryptionConfigSecretHash := ""
	// TODO: Support encryption config?
	// if fi.ValueOf(c.Cluster.Spec.EncryptionConfig) {
	// 	secret, err := secretStore.FindSecret("encryptionconfig")
	// 	if err != nil {
	// 		return fmt.Errorf("could not load encryptionconfig secret: %v", err)
	// 	}
	// 	if secret == nil {
	// 		fmt.Println("")
	// 		fmt.Println("You have encryptionConfig enabled, but no encryptionconfig secret has been set.")
	// 		fmt.Println("See `kops create secret encryptionconfig -h` and https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/")
	// 		return fmt.Errorf("could not find encryptionconfig secret")
	// 	}
	// 	hashBytes := sha256.Sum256(secret.Data)
	// 	encryptionConfigSecretHash = base64.URLEncoding.EncodeToString(hashBytes[:])
	// }

	fileAssets := &nodemodel.FileAssets{Cluster: cluster}
	if err := fileAssets.AddFileAssets(assetBuilder); err != nil {
		return nil, err
	}

	configBuilder, err := nodemodel.NewNodeUpConfigBuilder(cluster, assetBuilder, fileAssets.Assets, encryptionConfigSecretHash)
	if err != nil {
		return nil, err
	}

	keysets := make(map[string]*fi.Keyset)

	keystore, err := clientset.KeyStore(cluster)
	if err != nil {
		return nil, err
	}

	keyNames := model.KeypairNamesForInstanceGroup(cluster, ig)
	for _, keyName := range keyNames {
		keyset, err := keystore.FindKeyset(ctx, keyName)
		if err != nil {
			return nil, fmt.Errorf("getting keyset %q: %w", keyName, err)
		}

		if keyset == nil {
			return nil, fmt.Errorf("failed to find keyset %q", keyName)
		}

		keysets[keyName] = keyset
	}

	nodeupConfig, bootConfig, err := configBuilder.BuildConfig(ig, wellknownAddresses, keysets)
	if err != nil {
		return nil, err
	}

	var nodeupScript resources.NodeUpScript
	nodeupScript.NodeUpAssets = fileAssets.NodeUpAssets
	nodeupScript.BootConfig = bootConfig

	nodeupScript.WithEnvironmentVariables(cluster, ig)
	nodeupScript.WithProxyEnv(cluster)
	nodeupScript.WithSysctls()

	nodeupScript.CloudProvider = string(cluster.GetCloudProvider())

	bootConfig.ConfigBase = fi.PtrTo("file:///etc/kubernetes/kops/config")

	nodeupScriptResource, err := nodeupScript.Build()
	if err != nil {
		return nil, err
	}

	if bootConfig.InstanceGroupRole == kops.InstanceGroupRoleControlPlane {
		nodeupConfigBytes, err := yaml.Marshal(nodeupConfig)
		if err != nil {
			return nil, fmt.Errorf("error converting nodeup config to yaml: %w", err)
		}
		// Not much reason to hash this, since we're reading it from the local file system
		// sum256 := sha256.Sum256(nodeupConfigBytes)
		// bootConfig.NodeupConfigHash = base64.StdEncoding.EncodeToString(sum256[:])

		p := filepath.Join("/etc/kubernetes/kops/config", "igconfig", bootConfig.InstanceGroupRole.ToLowerString(), ig.Name, "nodeupconfig.yaml")
		bootstrapData.configFiles[p] = nodeupConfigBytes
	}

	nodeupScriptBytes, err := fi.ResourceAsBytes(nodeupScriptResource)
	if err != nil {
		return nil, err
	}
	bootstrapData.nodeupScript = nodeupScriptBytes

	return bootstrapData, nil
}
