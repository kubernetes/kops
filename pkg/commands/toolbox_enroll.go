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

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/mirrors"
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
	o.SSHUser = os.Getenv("USER")
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

	ig, err := clientset.InstanceGroupsFor(cluster).Get(ctx, options.InstanceGroup, metav1.GetOptions{})
	if err != nil {
		return err
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return err
	}

	apiserverAdditionalIPs := []string{}
	{
		ingresses, err := cloud.GetApiIngressStatus(cluster)
		if err != nil {
			return fmt.Errorf("error getting ingress status: %v", err)
		}

		for _, ingress := range ingresses {
			// if ingress.Hostname != "" {
			// 	apiserverAdditionalIPs = append(apiserverAdditionalIPs, ingress.Hostname)
			// }
			if ingress.IP != "" {
				apiserverAdditionalIPs = append(apiserverAdditionalIPs, ingress.IP)
			}
		}
	}

	if len(apiserverAdditionalIPs) == 0 {
		// TODO: Should we use DNS?
		return fmt.Errorf("unable to determine IP address for kops-controller")
	}

	scriptBytes, err := buildBootstrapData(ctx, clientset, cluster, ig, apiserverAdditionalIPs)
	if err != nil {
		return err
	}

	if _, err := os.Stdout.Write(scriptBytes); err != nil {
		return err
	}

	if options.Host != "" {
		// TODO: This is the pattern we use a lot, but should we try to access it directly?
		contextName := cluster.ObjectMeta.Name
		clientGetter := genericclioptions.NewConfigFlags(true)
		clientGetter.Context = &contextName

		restConfig, err := clientGetter.ToRESTConfig()
		if err != nil {
			return fmt.Errorf("cannot load kubecfg settings for %q: %w", contextName, err)
		}

		if err := enrollHost(ctx, options, string(scriptBytes), restConfig); err != nil {
			return err
		}
	}
	return nil
}

func enrollHost(ctx context.Context, options *ToolboxEnrollOptions, nodeupScript string, restConfig *rest.Config) error {
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

	if err := createHost(ctx, options, hostname, publicKeyBytes, kubeClient); err != nil {
		return err
	}

	if len(nodeupScript) != 0 {
		if _, err := host.runScript(ctx, nodeupScript, ExecOptions{Sudo: sudo, Echo: true}); err != nil {
			return err
		}
	}
	return nil
}

func createHost(ctx context.Context, options *ToolboxEnrollOptions, nodeName string, publicKey []byte, client client.Client) error {
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

func buildBootstrapData(ctx context.Context, clientset simple.Clientset, cluster *kops.Cluster, ig *kops.InstanceGroup, apiserverAdditionalIPs []string) ([]byte, error) {
	if cluster.Spec.KubeAPIServer == nil {
		cluster.Spec.KubeAPIServer = &kops.KubeAPIServerConfig{}
	}

	getAssets := false
	assetBuilder := assets.NewAssetBuilder(clientset.VFSContext(), cluster.Spec.Assets, cluster.Spec.KubernetesVersion, getAssets)

	encryptionConfigSecretHash := ""
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

	nodeUpAssets := make(map[architectures.Architecture]*mirrors.MirroredAsset)
	for _, arch := range architectures.GetSupported() {

		asset, err := cloudup.NodeUpAsset(assetBuilder, arch)
		if err != nil {
			return nil, err
		}
		nodeUpAssets[arch] = asset
	}

	assets := make(map[architectures.Architecture][]*mirrors.MirroredAsset)
	configBuilder, err := cloudup.NewNodeUpConfigBuilder(cluster, assetBuilder, assets, encryptionConfigSecretHash)
	if err != nil {
		return nil, err
	}

	// bootstrapScript := &model.BootstrapScript{
	// 	// KopsModelContext:    modelContext,
	// 	Lifecycle: fi.LifecycleSync,
	// 	// NodeUpConfigBuilder: configBuilder,
	// 	// NodeUpAssets:        c.NodeUpAssets,
	// }

	keysets := make(map[string]*fi.Keyset)

	keystore, err := clientset.KeyStore(cluster)
	if err != nil {
		return nil, err
	}

	for _, keyName := range []string{"kubernetes-ca"} {
		keyset, err := keystore.FindKeyset(ctx, keyName)
		if err != nil {
			return nil, fmt.Errorf("getting keyset %q: %w", keyName, err)
		}

		if keyset == nil {
			return nil, fmt.Errorf("failed to find keyset %q", keyName)
		}

		keysets[keyName] = keyset
	}

	_, bootConfig, err := configBuilder.BuildConfig(ig, apiserverAdditionalIPs, keysets)
	if err != nil {
		return nil, err
	}

	bootConfig.CloudProvider = "metal"
	// bootConfig.ConfigServer.Server = strings.ReplaceAll(bootConfig.ConfigServer.Server, ".internal.", ".")

	// configData, err := utils.YamlMarshal(config)
	// if err != nil {
	// 	return nil, fmt.Errorf("error converting nodeup config to yaml: %v", err)
	// }
	// sum256 := sha256.Sum256(configData)
	// bootConfig.NodeupConfigHash = base64.StdEncoding.EncodeToString(sum256[:])
	// b.nodeupConfig.Resource = fi.NewBytesResource(configData)

	var nodeupScript resources.NodeUpScript
	nodeupScript.NodeUpAssets = nodeUpAssets
	nodeupScript.BootConfig = bootConfig

	{
		nodeupScript.EnvironmentVariables = func() (string, error) {
			env := make(map[string]string)

			// env, err := b.buildEnvironmentVariables()
			// if err != nil {
			// 	return "", err
			// }

			// Sort keys to have a stable sequence of "export xx=xxx"" statements
			var keys []string
			for k := range env {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			var b bytes.Buffer
			for _, k := range keys {
				b.WriteString(fmt.Sprintf("export %s=%s\n", k, env[k]))
			}
			return b.String(), nil
		}

		nodeupScript.ProxyEnv = func() (string, error) {
			return "", nil
			// return b.createProxyEnv(cluster.Spec.Networking.EgressProxy)
		}
	}

	// TODO: nodeupScript.CompressUserData = fi.ValueOf(b.ig.Spec.CompressUserData)

	// By setting some sysctls early, we avoid broken configurations that prevent nodeup download.
	// See https://github.com/kubernetes/kops/issues/10206 for details.
	// TODO: nodeupScript.SetSysctls = setSysctls()

	nodeupScript.CloudProvider = string(cluster.Spec.GetCloudProvider())

	nodeupScriptResource, err := nodeupScript.Build()
	if err != nil {
		return nil, err
	}

	b, err := fi.ResourceAsBytes(nodeupScriptResource)
	if err != nil {
		return nil, err
	}

	return b, nil
}
