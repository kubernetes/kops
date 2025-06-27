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

// Package deployer implements the kubetest2 Kops deployer
package deployer

import (
	"flag"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"github.com/urfave/sflags/gen/gpflag"
	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/kubetest2-kops/aws"
	"k8s.io/kops/tests/e2e/kubetest2-kops/builder"
	"k8s.io/kops/tests/e2e/pkg/target"

	"sigs.k8s.io/boskos/client"
	"sigs.k8s.io/kubetest2/pkg/types"
)

// Name is the name of the deployer
const Name = "kops"

type deployer struct {
	// generic parts
	commonOptions types.Options
	// doInit helps to make sure the initialization is performed only once
	doInit sync.Once

	KopsRoot string `flag:"kops-root" desc:"Path to root of the kops repo. Used with --build."`

	StageLocation string `flag:"stage-location" desc:"Storage location for kops artifacts. Only gs:// paths are supported."`

	KopsVersionMarker    string `flag:"kops-version-marker" desc:"The URL to the kops version marker. Conflicts with --build and --kops-binary-path"`
	KopsBaseURL          string `flag:"-"`
	KubernetesBaseURL    string `flag:"-"`
	PublishVersionMarker string `flag:"publish-version-marker" desc:"The GCS path to which the --kops-version-marker is uploaded if the tests pass"`

	ClusterName            string   `flag:"cluster-name" desc:"The FQDN to use for the cluster name"`
	CloudProvider          string   `flag:"cloud-provider" desc:"Which cloud provider to use"`
	GCPProject             string   `flag:"gcp-project" desc:"Which GCP Project to use when --cloud-provider=gce"`
	Env                    []string `flag:"env" desc:"Additional env vars to set for kops commands in NAME=VALUE format"`
	CreateArgs             string   `flag:"create-args" desc:"Extra space-separated arguments passed to 'kops create cluster'"`
	KopsBinaryPath         string   `flag:"kops-binary-path" desc:"The path to kops executable used for testing"`
	KubernetesFeatureGates string   `flag:"kubernetes-feature-gates" desc:"Feature Gates to enable on Kubernetes components"`

	// ControlPlaneCount specifies the number of VMs in the control-plane.
	ControlPlaneCount int `flag:"control-plane-count" desc:"Number of control-plane instances"`

	ControlPlaneIGOverrides []string `flag:"control-plane-instance-group-overrides" desc:"overrides for the control plane instance groups"`
	NodeIGOverrides         []string `flag:"node-instance-group-overrides" desc:"overrides for the node instance groups"`

	ValidationWait     time.Duration `flag:"validation-wait" desc:"time to wait for newly created cluster to pass validation"`
	ValidationCount    int           `flag:"validation-count" desc:"how many times should a validation pass"`
	ValidationInterval time.Duration `flag:"validation-interval" desc:"time in duration to wait between validation attempts"`
	MaxNodesToDump     string        `flag:"max-nodes-to-dump" desc:"max number of nodes to dump logs from, helpful to set when running scale tests"`

	TemplatePath string `flag:"template-path" desc:"The path to the manifest template used for cluster creation"`

	KubernetesVersion string `flag:"kubernetes-version" desc:"The kubernetes version to use in the cluster"`

	SSHPrivateKeyPath string `flag:"ssh-private-key" desc:"The path to the private key used for SSH access to instances"`
	SSHPublicKeyPath  string `flag:"ssh-public-key" desc:"The path to the public key passed to the cloud provider"`
	SSHUser           string `flag:"ssh-user" desc:"The SSH user to use for SSH access to instances"`

	TerraformVersion string `flag:"terraform-version" desc:"The version of terraform to use for applying the cluster"`

	ArtifactsDir string `flag:"-"`

	AdminAccess string `flag:"admin-access" desc:"The CIDR to restrict kubernetes API access"`

	BuildOptions *builder.BuildOptions

	// manifestPath is the location of the rendered manifest based on TemplatePath
	manifestPath string
	terraform    *target.Terraform

	aws *aws.Client

	createBucket       bool
	stateStoreName     string
	discoveryStoreName string
	stagingStoreName   string

	// boskos struct field will be non-nil when the deployer is
	// using boskos to acquire a GCP project
	boskos                  *client.Client
	BoskosLocation          string        `flag:"boskos-location" desc:"If set, manually specifies the location of the Boskos server."`
	BoskosAcquireTimeout    time.Duration `flag:"boskos-acquire-timeout" desc:"How long should boskos wait to acquire a resource before timing out"`
	BoskosHeartbeatInterval time.Duration `flag:"boskos-heartbeat-interval" desc:"How often should boskos send a heartbeat to Boskos to hold the acquired resource. 0 means no heartbeat."`
	BoskosResourceType      string        `flag:"boskos-resource-type" desc:"If set, manually specifies the resource type of GCP projects to acquire from Boskos."`
	// this channel serves as a signal channel for the hearbeat goroutine
	// so that it can be explicitly closed
	boskosHeartbeatClose chan struct{}
}

// assert that New implements types.NewDeployer
var _ types.NewDeployer = New

// assert that deployer implements types.Deployer
var (
	_ types.Deployer               = &deployer{}
	_ types.DeployerWithPostTester = &deployer{}
)

func (d *deployer) Provider() string {
	return Name
}

// New implements deployer.New for kops
func New(opts types.Options) (types.Deployer, *pflag.FlagSet) {
	// create a deployer object and set fields that are not flag controlled, and default values
	d := &deployer{
		commonOptions: opts,
		BuildOptions: &builder.BuildOptions{
			BuildKubernetes: false,
		},
		boskosHeartbeatClose:    make(chan struct{}),
		ValidationCount:         10,
		ValidationInterval:      10 * time.Second,
		BoskosLocation:          "http://boskos.test-pods.svc.cluster.local.",
		BoskosResourceType:      "gce-project",
		BoskosAcquireTimeout:    5 * time.Minute,
		BoskosHeartbeatInterval: 5 * time.Minute,
	}
	dir, err := defaultArtifactsDir()
	if err != nil {
		klog.Fatalf("unable to determine artifacts directory: %v", err)
	}
	d.ArtifactsDir = dir

	// register flags
	fs := bindFlags(d)

	fs.IntVar(&d.ControlPlaneCount, "control-plane-size", d.ControlPlaneCount, "Number of control-plane instances")
	fs.MarkDeprecated("control-plane-size", "use --control-plane-count instead")

	// register flags for klog
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)
	fs.AddGoFlagSet(klogFlags)

	return d, fs
}

func bindFlags(d *deployer) *pflag.FlagSet {
	flags, err := gpflag.Parse(d)
	if err != nil {
		klog.Fatalf("unable to generate flags from deployer: %v", err)
		return nil
	}
	return flags
}
