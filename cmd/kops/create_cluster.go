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
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	kopsbase "k8s.io/kops"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	kopsutil "k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/clouds"
	"k8s.io/kops/pkg/clusteraddons"
	"k8s.io/kops/pkg/commands"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/pkg/wellknownoperators"
	"k8s.io/kops/pkg/zones"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/utils"
)

type CreateClusterOptions struct {
	cloudup.NewClusterOptions
	Yes                        bool
	Target                     string
	ControlPlaneVolumeSize     int32
	NodeVolumeSize             int32
	ContainerRuntime           string
	OutDir                     string
	DisableSubnetTags          bool
	NodeSecurityGroups         []string
	ControlPlaneSecurityGroups []string
	AssociatePublicIP          *bool

	// SSHPublicKeys is a map of the SSH public keys we should configure; required on AWS, not required on GCE
	SSHPublicKeys map[string][]byte

	// Sets allows setting values directly in the spec.
	Sets []string
	// Unsets allows unsetting values directly in the spec.
	Unsets []string

	// CloudLabels are cloud-provider-level tags for instance groups and volumes.
	CloudLabels string

	// Specify tenancy (default or dedicated) for control-plane and worker nodes
	ControlPlaneTenancy string
	NodeTenancy         string

	// Allow custom public Kubernetes API name.
	APIPublicName string

	OpenstackNetworkID string

	// DryRun mode output a cluster manifest of Output type.
	DryRun bool
	// Output type during a DryRun
	Output string

	// AddonPaths specify paths to additional components that we can add to a cluster
	AddonPaths []string
}

func (o *CreateClusterOptions) InitDefaults() {
	o.NewClusterOptions.InitDefaults()

	o.Yes = false
	o.Target = cloudup.TargetDirect
}

var (
	createClusterLong = templates.LongDesc(i18n.T(`
	Create a Kubernetes cluster using command line flags.
	This command creates cloud based resources such as networks and virtual machines. Once
	the infrastructure is in place Kubernetes is installed on the virtual machines.

	These operations are done in parallel and rely on eventual consistency.
	`))

	createClusterExample = templates.Examples(i18n.T(`
	# Create a cluster in AWS in a single zone.
	kops create cluster --name=k8s-cluster.example.com \
		--state=s3://my-state-store \
		--zones=us-east-1a \
		--node-count=2

	# Create a cluster in AWS with a High Availability control plane. This cluster
	# has also been configured for private networking in a kops-managed VPC.
	# The bastion flag is set to create an entrypoint for admins to SSH.
	export KOPS_STATE_STORE="s3://my-state-store"
	export CONTROL_PLANE_SIZE="c5.large"
	export NODE_SIZE="m5.large"
	export ZONES="us-east-1a,us-east-1b,us-east-1c"
	kops create cluster k8s-cluster.example.com \
	    --node-count 3 \
		--zones $ZONES \
		--node-size $NODE_SIZE \
		--control-plane-size $CONTROL_PLANE_SIZE \
		--control-plane-zones $ZONES \
		--networking cilium \
		--topology private \
		--bastion="true" \
		--yes

	# Create a cluster in Digital Ocean.
	export KOPS_STATE_STORE="do://my-state-store"
	export ZONES="NYC1"
	kops create cluster k8s-cluster.example.com \
		--cloud digitalocean \
		--zones $ZONES \
		--control-plane-zones $ZONES \
		--node-count 3 \
		--yes

	# Generate a cluster spec to apply later.
	# Run the following, then: kops create -f filename.yaml
	kops create cluster --name=k8s-cluster.example.com \
		--state=s3://my-state-store \
		--zones=us-east-1a \
		--node-count=2 \
		--dry-run \
		-oyaml > filename.yaml
	`))

	createClusterShort = i18n.T("Create a Kubernetes cluster.")
)

func NewCmdCreateCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateClusterOptions{}
	options.InitDefaults()

	sshPublicKey := ""
	associatePublicIP := false
	encryptEtcdStorage := false

	cmd := &cobra.Command{
		Use:               "cluster [CLUSTER]",
		Short:             createClusterShort,
		Long:              createClusterLong,
		Example:           createClusterExample,
		Args:              rootCommand.clusterNameArgsNoKubeconfig(&options.ClusterName),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if cmd.Flag("associate-public-ip").Changed {
				options.AssociatePublicIP = &associatePublicIP
			}

			if cmd.Flag("encrypt-etcd-storage").Changed {
				options.EncryptEtcdStorage = &encryptEtcdStorage
			}

			if sshPublicKey != "" {
				options.SSHPublicKeys, err = loadSSHPublicKeys(sshPublicKey)
				if err != nil {
					return fmt.Errorf("error reading SSH key file %q: %v", sshPublicKey, err)
				}
			}

			return RunCreateCluster(cmd.Context(), f, out, options)
		},
	}

	cmd.Flags().BoolVarP(&options.Yes, "yes", "y", options.Yes, "Specify --yes to immediately create the cluster")
	cmd.Flags().StringVar(&options.Target, "target", options.Target, fmt.Sprintf("Valid targets: %s, %s. Set this flag to %s if you want kOps to generate terraform", cloudup.TargetDirect, cloudup.TargetTerraform, cloudup.TargetTerraform))
	cmd.RegisterFlagCompletionFunc("target", completeCreateClusterTarget(options))

	// Configuration / state location
	if featureflag.EnableSeparateConfigBase.Enabled() {
		cmd.Flags().StringVar(&options.ConfigBase, "config-base", options.ConfigBase, "A cluster-readable location where we mirror configuration information, separate from the state store.  Allows for a state store that is not accessible from the cluster.")
		cmd.RegisterFlagCompletionFunc("config-base", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// TODO complete vfs paths
			return nil, cobra.ShellCompDirectiveNoFileComp
		})
	}
	cmd.Flags().StringVar(&options.DiscoveryStore, "discovery-store", options.DiscoveryStore, "A public location where we publish OIDC-compatible discovery information under a cluster-specific directory. Enables IRSA in AWS.")
	cmd.RegisterFlagCompletionFunc("discovery-store", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// TODO complete vfs paths
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	var validClouds []string
	{
		allClouds := clouds.SupportedClouds()
		for _, c := range allClouds {
			validClouds = append(validClouds, string(c))
		}
		sort.Strings(validClouds)
	}
	cmd.Flags().StringVar(&options.CloudProvider, "cloud", options.CloudProvider, fmt.Sprintf("Cloud provider to use - %s", strings.Join(validClouds, ", ")))
	cmd.RegisterFlagCompletionFunc("cloud", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validClouds, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.Flags().StringSliceVar(&options.Zones, "zones", options.Zones, "Zones in which to run the cluster")
	cmd.RegisterFlagCompletionFunc("zones", completeZone(options, &rootCommand))
	cmd.Flags().StringSliceVar(&options.ControlPlaneZones, "master-zones", options.ControlPlaneZones, "Zones in which to run control-plane nodes. (must be an odd number)")
	cmd.Flags().MarkDeprecated("master-zones", "use --control-plane-zones instead")
	cmd.Flags().StringSliceVar(&options.ControlPlaneZones, "control-plane-zones", options.ControlPlaneZones, "Zones in which to run control-plane nodes. (must be an odd number)")
	cmd.RegisterFlagCompletionFunc("control-plane-zones", completeZone(options, &rootCommand))

	if featureflag.ClusterAddons.Enabled() {
		cmd.Flags().StringSliceVar(&options.AddonPaths, "add", options.AddonPaths, "Paths to addons we should add to the cluster")
		// TODO complete VFS paths
	}

	cmd.Flags().StringVar(&options.KubernetesVersion, "kubernetes-version", options.KubernetesVersion, "Version of Kubernetes to run (defaults to version in channel)")
	cmd.RegisterFlagCompletionFunc("kubernetes-version", completeKubernetesVersion)

	cmd.Flags().StringSliceVar(&options.KubernetesFeatureGates, "kubernetes-feature-gates", options.KubernetesFeatureGates, "List of Kubernetes feature gates to enable/disable")
	cmd.RegisterFlagCompletionFunc("kubernetes-version", completeKubernetesFeatureGates)

	cmd.Flags().StringVar(&options.ContainerRuntime, "container-runtime", options.ContainerRuntime, "Container runtime to use: containerd")
	cmd.Flags().MarkDeprecated("container-runtime", "containerd is the only supported value")
	cmd.RegisterFlagCompletionFunc("container-runtime", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"containerd"}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.Flags().StringVar(&sshPublicKey, "ssh-public-key", sshPublicKey, "SSH public key to use")
	cmd.RegisterFlagCompletionFunc("ssh-public-key", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"pub"}, cobra.ShellCompDirectiveFilterFileExt
	})

	cmd.Flags().Int32Var(&options.ControlPlaneCount, "master-count", options.ControlPlaneCount, "Number of control-plane nodes. Defaults to one control-plane node per control-plane-zone")
	cmd.Flags().MarkDeprecated("master-count", "use --control-plane-count instead")
	cmd.Flags().Int32Var(&options.ControlPlaneCount, "control-plane-count", options.ControlPlaneCount, "Number of control-plane nodes. Defaults to one control-plane node per control-plane-zone")
	cmd.Flags().Int32Var(&options.NodeCount, "node-count", options.NodeCount, "Total number of worker nodes. Defaults to one node per zone")

	cmd.Flags().StringVar(&options.Image, "image", options.Image, "Machine image for all instances")
	cmd.RegisterFlagCompletionFunc("image", completeInstanceImage)
	cmd.Flags().StringVar(&options.NodeImage, "node-image", options.NodeImage, "Machine image for worker nodes. Takes precedence over --image")
	cmd.RegisterFlagCompletionFunc("node-image", completeInstanceImage)
	cmd.Flags().StringVar(&options.ControlPlaneImage, "master-image", options.ControlPlaneImage, "Machine image for control-plane nodes. Takes precedence over --image")
	cmd.Flags().MarkDeprecated("master-image", "use --control-plane-image instead")
	cmd.Flags().StringVar(&options.ControlPlaneImage, "control-plane-image", options.ControlPlaneImage, "Machine image for control-plane nodes. Takes precedence over --image")
	cmd.RegisterFlagCompletionFunc("control-plane-image", completeInstanceImage)
	cmd.Flags().StringVar(&options.BastionImage, "bastion-image", options.BastionImage, "Machine image for bastions. Takes precedence over --image")
	cmd.RegisterFlagCompletionFunc("bastion-image", completeInstanceImage)

	cmd.Flags().StringSliceVar(&options.NodeSizes, "node-size", options.NodeSizes, "Machine type(s) for worker nodes")
	cmd.RegisterFlagCompletionFunc("node-size", completeMachineType)
	cmd.Flags().StringSliceVar(&options.ControlPlaneSizes, "master-size", options.ControlPlaneSizes, "Machine type(s) for control-plane nodes")
	cmd.Flags().MarkDeprecated("master-size", "use --control-plane-size instead")
	cmd.Flags().StringSliceVar(&options.ControlPlaneSizes, "control-plane-size", options.ControlPlaneSizes, "Machine type(s) for control-plane nodes")
	cmd.RegisterFlagCompletionFunc("control-plane-size", completeMachineType)

	cmd.Flags().Int32Var(&options.ControlPlaneVolumeSize, "master-volume-size", options.ControlPlaneVolumeSize, "Instance volume size (in GB) for control-plane nodes")
	cmd.Flags().MarkDeprecated("master-volume-size", "use --control-plane-volume-size instead")
	cmd.Flags().Int32Var(&options.ControlPlaneVolumeSize, "control-plane-volume-size", options.ControlPlaneVolumeSize, "Instance volume size (in GB) for control-plane nodes")
	cmd.Flags().Int32Var(&options.NodeVolumeSize, "node-volume-size", options.NodeVolumeSize, "Instance volume size (in GB) for worker nodes")

	cmd.Flags().StringVar(&options.NetworkID, "vpc", options.NetworkID, "Shared Network or VPC to use")
	cmd.Flags().MarkDeprecated("vpc", "use --network-id instead")
	cmd.RegisterFlagCompletionFunc("vpc", completeNetworkID)
	cmd.Flags().StringVar(&options.NetworkID, "network-id", options.NetworkID, "Shared Network or VPC to use")
	cmd.RegisterFlagCompletionFunc("network-id", completeNetworkID)
	cmd.Flags().StringSliceVar(&options.SubnetIDs, "subnets", options.SubnetIDs, "Shared subnets to use")
	cmd.RegisterFlagCompletionFunc("subnets", completeSubnetID(options))
	cmd.Flags().StringSliceVar(&options.UtilitySubnetIDs, "utility-subnets", options.UtilitySubnetIDs, "Shared utility subnets to use")
	cmd.RegisterFlagCompletionFunc("utility-subnets", completeSubnetID(options))
	cmd.Flags().StringSliceVar(&options.NetworkCIDRs, "network-cidr", options.NetworkCIDRs, "Network CIDR(s) to use")
	cmd.RegisterFlagCompletionFunc("network-cidr", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})
	cmd.Flags().BoolVar(&options.DisableSubnetTags, "disable-subnet-tags", options.DisableSubnetTags, "Disable automatic subnet tagging")

	cmd.Flags().StringSliceVar(&options.EtcdClusters, "etcd-clusters", options.EtcdClusters, "Names of the etcd clusters: main, events")
	cmd.RegisterFlagCompletionFunc("etcd-clusters", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"main", "events"}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.Flags().BoolVar(&encryptEtcdStorage, "encrypt-etcd-storage", false, "Generate key in AWS KMS and use it for encrypt etcd volumes")
	cmd.Flags().StringVar(&options.EtcdStorageType, "etcd-storage-type", options.EtcdStorageType, "The default storage type for etcd members")
	cmd.RegisterFlagCompletionFunc("etcd-storage-type", completeStorageType)

	cmd.Flags().StringVar(&options.Networking, "networking", options.Networking, "Networking mode.  kubenet, external, flannel-vxlan (or flannel), flannel-udp, calico, canal, kube-router, amazonvpc, cilium, cilium-etcd, cni.")
	cmd.RegisterFlagCompletionFunc("networking", completeNetworking(options))

	cmd.Flags().StringVar(&options.DNSZone, "dns-zone", options.DNSZone, "DNS hosted zone (defaults to longest matching zone)")
	cmd.RegisterFlagCompletionFunc("dns-zone", completeDNSZone(options))
	cmd.Flags().StringVar(&options.OutDir, "out", options.OutDir, "Path to write any local output")
	cmd.MarkFlagDirname("out")
	cmd.Flags().StringSliceVar(&options.AdminAccess, "admin-access", options.AdminAccess, "Restrict API access to this CIDR.  If not set, access will not be restricted by IP.")
	cmd.RegisterFlagCompletionFunc("admin-access", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})
	cmd.Flags().StringSliceVar(&options.SSHAccess, "ssh-access", options.SSHAccess, "Restrict SSH access to this CIDR.  If not set, uses the value of the admin-access flag.")
	cmd.RegisterFlagCompletionFunc("ssh-access", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	// TODO: Can we deprecate this flag - it is awkward?
	cmd.Flags().BoolVar(&associatePublicIP, "associate-public-ip", false, "Specify --associate-public-ip=[true|false] to enable/disable association of public IP for control-plane ASG and nodes. Default is 'true'.")

	cmd.Flags().BoolVar(&options.IPv6, "ipv6", false, "Use IPv6 for the pod network (AWS only)")

	cmd.Flags().StringSliceVar(&options.NodeSecurityGroups, "node-security-groups", options.NodeSecurityGroups, "Additional pre-created security groups to add to worker nodes.")
	cmd.RegisterFlagCompletionFunc("node-security-groups", completeSecurityGroup)
	cmd.Flags().StringSliceVar(&options.ControlPlaneSecurityGroups, "master-security-groups", options.ControlPlaneSecurityGroups, "Additional pre-created security groups to add to control-plane nodes.")
	cmd.Flags().MarkDeprecated("master-security-groups", "use --control-plane-security-groups instead")
	cmd.Flags().StringSliceVar(&options.ControlPlaneSecurityGroups, "control-plane-security-groups", options.ControlPlaneSecurityGroups, "Additional pre-created security groups to add to control-plane nodes.")
	cmd.RegisterFlagCompletionFunc("control-plane-security-groups", completeSecurityGroup)

	cmd.Flags().StringVar(&options.Channel, "channel", options.Channel, "Channel for default versions and configuration to use")
	cmd.RegisterFlagCompletionFunc("channel", completeChannel)

	// Network topology
	cmd.Flags().StringVarP(&options.Topology, "topology", "t", options.Topology, "Network topology for the cluster: 'public' or 'private'. Defaults to 'public' for IPv4 clusters and 'private' for IPv6 clusters.")
	cmd.RegisterFlagCompletionFunc("topology", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{api.TopologyPublic, api.TopologyPrivate}, cobra.ShellCompDirectiveNoFileComp
	})

	// Authorization
	cmd.Flags().StringVar(&options.Authorization, "authorization", options.Authorization, "Authorization mode: "+cloudup.AuthorizationFlagAlwaysAllow+" or "+cloudup.AuthorizationFlagRBAC)
	cmd.RegisterFlagCompletionFunc("authorization", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{cloudup.AuthorizationFlagAlwaysAllow, cloudup.AuthorizationFlagRBAC}, cobra.ShellCompDirectiveNoFileComp
	})

	// DNS
	supportedDnsTypes := []string{"public", "private", "none"}
	cmd.Flags().StringVar(&options.DNSType, "dns", options.DNSType, "DNS type to use: "+strings.Join(supportedDnsTypes, ", "))
	cmd.RegisterFlagCompletionFunc("dns", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return supportedDnsTypes, cobra.ShellCompDirectiveNoFileComp
	})

	// Bastion
	cmd.Flags().BoolVar(&options.Bastion, "bastion", options.Bastion, "Enable a bastion instance group. Only applies to private topology.")

	// Allow custom tags from the CLI
	cmd.Flags().StringVar(&options.CloudLabels, "cloud-labels", options.CloudLabels, "A list of key/value pairs used to tag all instance groups (for example \"Owner=John Doe,Team=Some Team\").")
	cmd.RegisterFlagCompletionFunc("cloud-labels", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	// Control Plane and Worker Node Tenancy
	cmd.Flags().StringVar(&options.ControlPlaneTenancy, "master-tenancy", options.ControlPlaneTenancy, "Tenancy of the control-plane group (AWS only): default or dedicated")
	cmd.Flags().MarkDeprecated("master-tenancy", "use --control-plane-tenancy instead")
	cmd.Flags().StringVar(&options.ControlPlaneTenancy, "control-plane-tenancy", options.ControlPlaneTenancy, "Tenancy of the control-plane group (AWS only): default or dedicated")
	cmd.RegisterFlagCompletionFunc("control-plane-tenancy", completeTenancy)
	cmd.Flags().StringVar(&options.NodeTenancy, "node-tenancy", options.NodeTenancy, "Tenancy of the node group (AWS only): default or dedicated")
	cmd.RegisterFlagCompletionFunc("node-tenancy", completeTenancy)

	cmd.Flags().StringVar(&options.APILoadBalancerClass, "api-loadbalancer-class", options.APILoadBalancerClass, "Class of load balancer for the Kubernetes API (AWS only): classic or network")
	cmd.Flags().MarkDeprecated("api-loadbalancer-class", "network load balancer should be used for all newly created clusters")
	cmd.RegisterFlagCompletionFunc("api-loadbalancer-class", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"classic", "network", "regional", "global"}, cobra.ShellCompDirectiveNoFileComp
	})
	cmd.Flags().StringVar(&options.APILoadBalancerType, "api-loadbalancer-type", options.APILoadBalancerType, "Type of load balancer for the Kubernetes API: public or internal")
	cmd.RegisterFlagCompletionFunc("api-loadbalancer-type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"public", "internal"}, cobra.ShellCompDirectiveNoFileComp
	})
	cmd.Flags().StringVar(&options.APISSLCertificate, "api-ssl-certificate", options.APISSLCertificate, "ARN of the SSL Certificate to use for the Kubernetes API load balancer (AWS only)")
	cmd.RegisterFlagCompletionFunc("api-ssl-certificate", completeSSLCertificate)

	// Allow custom public Kuberneters API name.
	cmd.Flags().StringVar(&options.APIPublicName, "master-public-name", options.APIPublicName, "Domain name of the public Kubernetes API")
	cmd.Flags().MarkDeprecated("master-public-name", "use --api-public-name instead")
	cmd.Flags().StringVar(&options.APIPublicName, "api-public-name", options.APIPublicName, "Domain name of the public Kubernetes API")
	cmd.RegisterFlagCompletionFunc("api-public-name", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	// DryRun mode that will print YAML or JSON
	cmd.Flags().BoolVar(&options.DryRun, "dry-run", options.DryRun, "If true, only print the object that would be sent, without sending it. This flag can be used to create a cluster YAML or JSON manifest.")
	cmd.Flags().StringVarP(&options.Output, "output", "o", options.Output, "Output format. One of json or yaml. Used with the --dry-run flag.")
	cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "yaml"}, cobra.ShellCompDirectiveNoFileComp
	})

	LazyQuoteStringSliceVar(cmd.Flags(), &options.Sets, "override", options.Sets, "Directly set values in the spec")
	cmd.Flags().MarkDeprecated("override", "use --set instead")
	LazyQuoteStringSliceVar(cmd.Flags(), &options.Sets, "set", options.Sets, "Directly set values in the spec")
	cmd.RegisterFlagCompletionFunc("set", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})
	cmd.Flags().StringSliceVar(&options.Unsets, "unset", options.Unsets, "Directly unset values in the spec")
	cmd.RegisterFlagCompletionFunc("unset", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	// GCE flags
	cmd.Flags().StringVar(&options.Project, "project", options.Project, "Project to use (must be set on GCE)")
	cmd.RegisterFlagCompletionFunc("project", completeProject)
	cmd.Flags().StringVar(&options.GCEServiceAccount, "gce-service-account", options.GCEServiceAccount, "Service account with which the GCE VM runs. Warning: if not set, VMs will run as default compute service account.")
	cmd.RegisterFlagCompletionFunc("gce-service-account", completeGCEServiceAccount)

	if featureflag.Azure.Enabled() {
		cmd.Flags().StringVar(&options.AzureSubscriptionID, "azure-subscription-id", options.AzureSubscriptionID, "Azure subscription where the cluster is created")
		cmd.RegisterFlagCompletionFunc("azure-subscription-id", completeAzureSubscriptionID)
		cmd.Flags().StringVar(&options.AzureTenantID, "azure-tenant-id", options.AzureTenantID, "Azure tenant where the cluster is created.")
		cmd.RegisterFlagCompletionFunc("azure-tenant-id", completeAzureTenantID)
		cmd.Flags().StringVar(&options.AzureResourceGroupName, "azure-resource-group-name", options.AzureResourceGroupName, "Azure resource group name where the cluster is created. The resource group will be created if it doesn't already exist. Defaults to the cluster name.")
		cmd.RegisterFlagCompletionFunc("azure-resource-group-name", completeAzureResourceGroupName)
		cmd.Flags().StringVar(&options.AzureRouteTableName, "azure-route-table-name", options.AzureRouteTableName, "Azure route table name where the cluster is created.")
		cmd.RegisterFlagCompletionFunc("azure-route-table-name", completeAzureRouteTableName)
		cmd.Flags().StringVar(&options.AzureAdminUser, "azure-admin-user", options.AzureAdminUser, "Azure admin user of VM ScaleSet.")
		cmd.RegisterFlagCompletionFunc("azure-admin-user", completeAzureAdminUsers)
	}

	if featureflag.Spotinst.Enabled() {
		// Spotinst flags
		cmd.Flags().StringVar(&options.SpotinstProduct, "spotinst-product", options.SpotinstProduct, "Product description (valid values: Linux/UNIX, Linux/UNIX (Amazon VPC), Windows and Windows (Amazon VPC))")
		cmd.RegisterFlagCompletionFunc("spotinst-product", completeSpotinstProduct)
		cmd.Flags().StringVar(&options.SpotinstOrientation, "spotinst-orientation", options.SpotinstOrientation, "Prediction strategy (valid values: balanced, cost, equal-distribution and availability)")
		cmd.RegisterFlagCompletionFunc("spotinst-orientation", completeSpotinstOrientation)
	}

	if featureflag.APIServerNodes.Enabled() {
		cmd.Flags().Int32Var(&options.APIServerCount, "api-server-count", options.APIServerCount, "Number of API server nodes. Defaults to 0.")
	}

	// Openstack flags
	cmd.Flags().StringVar(&options.OpenstackExternalNet, "os-ext-net", options.OpenstackExternalNet, "External network to use with the openstack router")
	cmd.RegisterFlagCompletionFunc("os-ext-net", completeOpenstackExternalNet)
	cmd.Flags().StringVar(&options.OpenstackExternalSubnet, "os-ext-subnet", options.OpenstackExternalSubnet, "External floating subnet to use with the openstack router")
	cmd.RegisterFlagCompletionFunc("os-ext-subnet", completeOpenstackExternalSubnet)
	cmd.Flags().StringVar(&options.OpenstackLBSubnet, "os-lb-floating-subnet", options.OpenstackLBSubnet, "External subnet to use with the Kubernetes API")
	cmd.RegisterFlagCompletionFunc("os-lb-floating-subnet", completeOpenstackLBSubnet)
	cmd.Flags().BoolVar(&options.OpenstackStorageIgnoreAZ, "os-kubelet-ignore-az", options.OpenstackStorageIgnoreAZ, "Attach volumes across availability zones")
	cmd.Flags().BoolVar(&options.OpenstackLBOctavia, "os-octavia", options.OpenstackLBOctavia, "Use octavia load balancer API")
	cmd.Flags().StringVar(&options.OpenstackOctaviaProvider, "os-octavia-provider", options.OpenstackOctaviaProvider, "Octavia provider to use")
	cmd.RegisterFlagCompletionFunc("os-octavia-provider", completeOpenstackOctaviaProvider)
	cmd.Flags().StringVar(&options.OpenstackDNSServers, "os-dns-servers", options.OpenstackDNSServers, "comma separated list of DNS Servers which is used in network")
	cmd.RegisterFlagCompletionFunc("os-dns-servers", completeOpenstackDNSServers)
	cmd.Flags().StringVar(&options.OpenstackNetworkID, "os-network", options.OpenstackNetworkID, "ID of the existing OpenStack network to use")
	cmd.RegisterFlagCompletionFunc("os-network", completeOpenstackNetworkID)

	cmd.Flags().StringVar(&options.InstanceManager, "instance-manager", options.InstanceManager, "Instance manager to use (cloudgroups or karpenter. Default: cloudgroups)")
	cmd.RegisterFlagCompletionFunc("instance-manager", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"cloudgroups", "karpenter"}, cobra.ShellCompDirectiveNoFileComp
	})
	return cmd
}

func RunCreateCluster(ctx context.Context, f *util.Factory, out io.Writer, c *CreateClusterOptions) error {
	ctx, span := tracer.Start(ctx, "RunCreateCluster")
	defer span.End()

	isDryrun := false
	// direct requires --yes (others do not, because they don't make changes)
	targetName := c.Target
	if c.Target == cloudup.TargetDirect {
		if !c.Yes {
			isDryrun = true
			targetName = cloudup.TargetDryRun
		}
	}
	if c.Target == cloudup.TargetDryRun {
		isDryrun = true
		targetName = cloudup.TargetDryRun
	}

	if c.DryRun && c.Output == "" {
		return fmt.Errorf("unable to execute --dry-run without setting --output")
	}

	// TODO: Reuse rootCommand stateStore logic?

	if c.OutDir == "" {
		if c.Target == cloudup.TargetTerraform {
			c.OutDir = "out/terraform"
		} else {
			c.OutDir = "out"
		}
	}

	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	if c.ClusterName == "" {
		return fmt.Errorf("--name is required")
	}

	{
		cluster, err := clientset.GetCluster(ctx, c.ClusterName)
		if err != nil {
			if apierrors.IsNotFound(err) {
				cluster = nil
			} else {
				return err
			}
		}

		if cluster != nil {
			return fmt.Errorf("cluster %q already exists; use 'kops update cluster' to apply changes", c.ClusterName)
		}
	}

	if c.OpenstackNetworkID != "" {
		c.NetworkID = c.OpenstackNetworkID
	}

	clusterResult, err := cloudup.NewCluster(&c.NewClusterOptions, clientset)
	if err != nil {
		return err
	}

	cluster := clusterResult.Cluster
	instanceGroups := clusterResult.InstanceGroups

	var controlPlanes []*api.InstanceGroup
	var nodes []*api.InstanceGroup
	for _, ig := range instanceGroups {
		switch ig.Spec.Role {
		case api.InstanceGroupRoleControlPlane:
			controlPlanes = append(controlPlanes, ig)
		case api.InstanceGroupRoleNode:
			nodes = append(nodes, ig)
		}
	}

	cloudLabels, err := parseCloudLabels(c.CloudLabels)
	if err != nil {
		return fmt.Errorf("error parsing global cloud labels: %v", err)
	}
	if len(cloudLabels) != 0 {
		cluster.Spec.CloudLabels = cloudLabels
	}

	if c.AssociatePublicIP != nil {
		for _, group := range instanceGroups {
			group.Spec.AssociatePublicIP = c.AssociatePublicIP
		}
	}

	if c.ControlPlaneTenancy != "" {
		for _, group := range controlPlanes {
			group.Spec.Tenancy = c.ControlPlaneTenancy
		}
	}

	if c.NodeTenancy != "" {
		for _, group := range nodes {
			group.Spec.Tenancy = c.NodeTenancy
		}
	}

	if len(c.NodeSecurityGroups) > 0 {
		for _, group := range nodes {
			group.Spec.AdditionalSecurityGroups = c.NodeSecurityGroups
		}
	}

	if len(c.ControlPlaneSecurityGroups) > 0 {
		for _, group := range controlPlanes {
			group.Spec.AdditionalSecurityGroups = c.ControlPlaneSecurityGroups
		}
	}

	if c.ControlPlaneVolumeSize != 0 {
		for _, group := range controlPlanes {
			if group.Spec.RootVolume == nil {
				group.Spec.RootVolume = &api.InstanceRootVolumeSpec{}
			}
			group.Spec.RootVolume.Size = fi.PtrTo(c.ControlPlaneVolumeSize)
		}
	}

	if c.NodeVolumeSize != 0 {
		for _, group := range nodes {
			if group.Spec.RootVolume == nil {
				group.Spec.RootVolume = &api.InstanceRootVolumeSpec{}
			}
			group.Spec.RootVolume.Size = fi.PtrTo(c.NodeVolumeSize)
		}
	}

	if c.DNSZone != "" {
		cluster.Spec.DNSZone = c.DNSZone
	}

	for i, cidr := range c.NetworkCIDRs {
		if i == 0 {
			cluster.Spec.Networking.NetworkCIDR = cidr
		} else {
			cluster.Spec.Networking.AdditionalNetworkCIDRs = append(cluster.Spec.Networking.AdditionalNetworkCIDRs, cidr)
		}
	}

	if c.DisableSubnetTags {
		cluster.Spec.Networking.TagSubnets = fi.PtrTo(false)
	}

	if c.APIPublicName != "" {
		cluster.Spec.API.PublicName = c.APIPublicName
	}

	if err := commands.UnsetClusterFields(c.Unsets, cluster); err != nil {
		return err
	}
	if err := commands.SetClusterFields(c.Sets, cluster); err != nil {
		return err
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return err
	}

	err = cloudup.PerformAssignments(cluster, clientset.VFSContext(), cloud)
	if err != nil {
		return fmt.Errorf("error populating configuration: %v", err)
	}

	strict := false
	err = validation.DeepValidate(cluster, instanceGroups, strict, clientset.VFSContext(), nil)
	if err != nil {
		return err
	}

	assetBuilder := assets.NewAssetBuilder(clientset.VFSContext(), cluster.Spec.Assets, cluster.Spec.KubernetesVersion, false)
	fullCluster, err := cloudup.PopulateClusterSpec(ctx, clientset, cluster, instanceGroups, cloud, assetBuilder)
	if err != nil {
		return err
	}

	kubernetesVersion, err := kopsutil.ParseKubernetesVersion(clusterResult.Cluster.Spec.KubernetesVersion)
	if err != nil {
		return fmt.Errorf("cannot parse KubernetesVersion %q in cluster: %w", clusterResult.Cluster.Spec.KubernetesVersion, err)
	}

	addons, err := wellknownoperators.CreateAddons(clusterResult.Channel, kubernetesVersion, fullCluster)
	if err != nil {
		return err
	}

	for _, p := range c.AddonPaths {
		addon, err := clusteraddons.LoadClusterAddon(clientset.VFSContext(), p)
		if err != nil {
			return fmt.Errorf("error loading cluster addon %s: %v", p, err)
		}
		addons = append(addons, addon.Objects...)
	}

	{
		// Build full IG spec to ensure we end up with a valid IG
		fullInstanceGroups := []*api.InstanceGroup{}
		for _, group := range instanceGroups {
			fullGroup, err := cloudup.PopulateInstanceGroupSpec(cluster, group, cloud, clusterResult.Channel)
			if err != nil {
				return err
			}
			fullInstanceGroups = append(fullInstanceGroups, fullGroup)
		}

		err = validation.DeepValidate(fullCluster, fullInstanceGroups, true, clientset.VFSContext(), nil)
		if err != nil {
			return fmt.Errorf("validation of the full cluster and instance group specs failed: %w", err)
		}
	}

	if c.DryRun {
		var obj []runtime.Object
		obj = append(obj, cluster)

		for _, group := range instanceGroups {
			// Cluster name is not populated, and we need it
			group.ObjectMeta.Labels = make(map[string]string)
			group.ObjectMeta.Labels[api.LabelClusterName] = cluster.ObjectMeta.Name
			obj = append(obj, group)
		}

		for name, key := range c.SSHPublicKeys {
			obj = append(obj, &api.SSHCredential{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					Labels: map[string]string{
						api.LabelClusterName: cluster.Name,
					},
				},
				Spec: api.SSHCredentialSpec{
					PublicKey: strings.TrimSpace(string(key)),
				},
			})
		}

		for _, o := range addons {
			obj = append(obj, o.ToUnstructured())
		}

		switch c.Output {
		case OutputYaml:
			if err := fullOutputYAML(out, obj...); err != nil {
				return fmt.Errorf("error writing cluster yaml to stdout: %v", err)
			}
			return nil
		case OutputJSON:
			if err := fullOutputJSON(out, true, obj...); err != nil {
				return fmt.Errorf("error writing cluster json to stdout: %v", err)
			}
			return nil
		default:
			return fmt.Errorf("unsupported output type %q", c.Output)
		}
	}

	// Note we perform as much validation as we can, before writing a bad config
	err = registry.CreateClusterConfig(ctx, clientset, cluster, instanceGroups, addons)
	if err != nil {
		return fmt.Errorf("error writing updated configuration: %v", err)
	}

	if len(c.SSHPublicKeys) == 0 {
		autoloadSSHPublicKeys := true
		switch c.CloudProvider {
		case "gce", "aws":
			autoloadSSHPublicKeys = false
		}

		if autoloadSSHPublicKeys {
			// Load from default locations, if found
			sshPublicKeyPaths := []string{
				"~/.ssh/id_rsa.pub",
				"~/.ssh/id_ed25519.pub",
			}
			var merr error
			for _, sshPublicKeyPath := range sshPublicKeyPaths {
				c.SSHPublicKeys, err = loadSSHPublicKeys(sshPublicKeyPath)
				if err == nil {
					break
				}
				// Don't wrap file-not-found
				if os.IsNotExist(err) {
					klog.V(2).Infof("ssh key not found at %s", sshPublicKeyPath)
				} else {
					merr = multierr.Append(merr, err)
				}
			}
			if merr != nil && len(c.SSHPublicKeys) == 0 {
				return fmt.Errorf("error reading SSH public key files %q: %v", sshPublicKeyPaths, merr)
			}
		}
	}

	if len(c.SSHPublicKeys) != 0 {
		sshCredentialStore, err := clientset.SSHCredentialStore(cluster)
		if err != nil {
			return err
		}

		for _, data := range c.SSHPublicKeys {
			err = sshCredentialStore.AddSSHPublicKey(ctx, data)
			if err != nil {
				return fmt.Errorf("error adding SSH public key: %v", err)
			}
		}
	}

	// Can we actually get to this if??
	if targetName != "" {
		if isDryrun {
			fmt.Fprintf(out, "Previewing changes that will be made:\n\n")
		}

		// TODO: Maybe just embed UpdateClusterOptions in CreateClusterOptions?
		updateClusterOptions := &UpdateClusterOptions{}
		updateClusterOptions.InitDefaults()

		updateClusterOptions.Yes = c.Yes
		updateClusterOptions.Target = c.Target
		updateClusterOptions.OutDir = c.OutDir
		updateClusterOptions.admin = kubeconfig.DefaultKubecfgAdminLifetime
		updateClusterOptions.ClusterName = cluster.Name
		updateClusterOptions.CreateKubecfg = true

		// SSHPublicKey has already been mapped
		updateClusterOptions.SSHPublicKey = ""

		_, err := RunUpdateCluster(ctx, f, out, updateClusterOptions)
		if err != nil {
			return err
		}

		if isDryrun {
			var sb bytes.Buffer
			fmt.Fprintf(&sb, "\n")
			fmt.Fprintf(&sb, "Cluster configuration has been created.\n")
			fmt.Fprintf(&sb, "\n")
			fmt.Fprintf(&sb, "Suggestions:\n")
			fmt.Fprintf(&sb, " * list clusters with: kops get cluster\n")
			fmt.Fprintf(&sb, " * edit this cluster with: kops edit cluster %s\n", cluster.Name)
			if len(nodes) > 0 {
				fmt.Fprintf(&sb, " * edit your node instance group: kops edit ig --name=%s %s\n", cluster.Name, nodes[0].ObjectMeta.Name)
			}
			if len(controlPlanes) > 0 {
				fmt.Fprintf(&sb, " * edit your control-plane instance group: kops edit ig --name=%s %s\n", cluster.Name, controlPlanes[0].ObjectMeta.Name)
			}
			fmt.Fprintf(&sb, "\n")
			fmt.Fprintf(&sb, "Finally configure your cluster with: kops update cluster --name %s --yes --admin\n", cluster.Name)
			fmt.Fprintf(&sb, "\n")

			_, err := out.Write(sb.Bytes())
			if err != nil {
				return fmt.Errorf("error writing to output: %v", err)
			}
		}
	}

	return nil
}

// parseCloudLabels takes a CSV list of key=value records and parses them into a map. Nested '='s are supported via
// quoted strings (eg `foo="bar=baz"` parses to map[string]string{"foo":"bar=baz"}. Nested commas are not supported.
func parseCloudLabels(s string) (map[string]string, error) {
	// Replace commas with newlines to allow a single pass with csv.Reader.
	// We can't use csv.Reader for the initial split because it would see each key=value record as a single field
	// and significantly complicates using quoted fields as keys or values.
	records := strings.Replace(s, ",", "\n", -1)

	// Let the CSV library do the heavy-lifting in handling nested ='s
	r := csv.NewReader(strings.NewReader(records))
	r.Comma = '='
	r.FieldsPerRecord = 2
	r.LazyQuotes = false
	r.TrimLeadingSpace = true
	kvPairs, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("one or more key=value pairs are malformed:\n%s\n:%w", records, err)
	}

	m := make(map[string]string, len(kvPairs))
	for _, pair := range kvPairs {
		m[pair[0]] = pair[1]
	}
	return m, nil
}

func loadSSHPublicKeys(sshPublicKey string) (map[string][]byte, error) {
	sshPublicKeys := make(map[string][]byte)
	if sshPublicKey != "" {
		sshPublicKey = utils.ExpandPath(sshPublicKey)
		authorized, err := os.ReadFile(sshPublicKey)
		if err != nil {
			return nil, err
		}
		sshPublicKeys[fi.SecretNameSSHPrimary] = authorized
		klog.Infof("Using SSH public key: %v\n", sshPublicKey)
	}
	return sshPublicKeys, nil
}

func completeZone(options *CreateClusterOptions, rootCommand *RootCmd) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if options.CloudProvider != "" {
			return zones.WellKnownZonesForCloud(api.CloudProviderID(options.CloudProvider), toComplete), cobra.ShellCompDirectiveNoFileComp
		}

		cloud, err := clouds.GuessCloudForPath(rootCommand.RegistryPath)
		if err != nil {
			return commandutils.CompletionError("listing cloud zones", err)
		}
		return zones.WellKnownZonesForCloud(cloud, toComplete), cobra.ShellCompDirectiveNoFileComp
	}
}

func completeKubernetesVersion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	kopsVersion, err := kopsutil.ParseKubernetesVersion(kopsbase.KOPS_RELEASE_VERSION)
	if err != nil {
		commandutils.CompletionError("parsing kops version", err)
	}
	tooNewVersion := kopsVersion
	tooNewVersion.Minor++
	tooNewVersion.Pre = nil
	tooNewVersion.Build = nil

	repo, err := name.NewRepository("registry.k8s.io/kube-apiserver")
	if err != nil {
		return commandutils.CompletionError("parsing kube-apiserver repo", err)
	}
	tags, err := remote.List(repo)
	if err != nil {
		return commandutils.CompletionError("listing kube-apiserver tags", err)
	}
	versions := sets.NewString()
	for _, tag := range tags {
		parsed, err := kopsutil.ParseKubernetesVersion(tag)
		if err != nil {
			continue
		}
		if kopsutil.IsKubernetesGTE(cloudup.OldestSupportedKubernetesVersion, *parsed) &&
			!kopsutil.IsKubernetesGTE(tooNewVersion.String(), *parsed) {
			versions.Insert(parsed.String())
		}
	}

	// Remove pre-release versions that have a subsequent stable version.
	// Also remove the non-useful -rc.0 versions.
	for _, version := range versions.UnsortedList() {
		split := strings.Split(version, "-")
		if len(split) > 1 && versions.Has(split[0]) {
			versions.Delete(version)
		}
		if strings.HasSuffix(version, "-rc.0") {
			versions.Delete(version)
		}
	}

	return versions.List(), cobra.ShellCompDirectiveNoFileComp
}

func completeKubernetesFeatureGates(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO check if there's a way to get the full list of feature gates from k8s libs
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeInstanceImage(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider(s) to get list of valid images
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeMachineType(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider(s) to get list of valid machine types
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeNetworkID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider(s) to get list of valid VPCs
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeSubnetID(options *CreateClusterOptions) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// TODO call into cloud provider(s) to get list of valid Subnet IDs
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

func completeStorageType(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider(s) to get list of valid storage types
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeNetworking(options *CreateClusterOptions) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		completions := []string{
			"external",
			"cni",
			"calico",
			"cilium",
			"cilium-eni",
			"cilium-etcd",
		}

		if !options.IPv6 {
			completions = append(completions,
				"kubenet",
				"kopeio",
				"flannel",
				"canal",
				"kube-router",
			)

			if options.CloudProvider == "aws" || options.CloudProvider == "" {
				completions = append(completions, "amazonvpc")
			}

			if options.CloudProvider == "gce" || options.CloudProvider == "" {
				completions = append(completions, "gcp")
			}
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

func completeCreateClusterTarget(options *CreateClusterOptions) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		completions := []string{
			cloudup.TargetDirect,
			cloudup.TargetDryRun,
		}
		for _, cp := range cloudup.TerraformCloudProviders {
			if options.CloudProvider == string(cp) {
				completions = append(completions, cloudup.TargetTerraform)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

func completeDNSZone(options *CreateClusterOptions) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		commandutils.ConfigureKlogForCompletion()

		clusterName, completions, directive := GetClusterNameForCompletionNoKubeconfig(args)
		if clusterName == "" {
			return completions, directive
		}

		zone := clusterName
		completions = nil
		for {
			split := strings.SplitN(zone, ".", 2)
			if len(split) != 2 || !strings.Contains(split[1], ".") {
				break
			}
			zone = split[1]
			// TODO Verify the zone against the cloud's DNS provider?
			completions = append(completions, zone)
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

func completeSecurityGroup(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider(s) to get list of valid Security groups
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeTenancy(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return ec2.Tenancy_Values(), cobra.ShellCompDirectiveNoFileComp
}

func completeSSLCertificate(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider(s) to get list of certificates
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeProject(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider(s) to get list of projects
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeGCEServiceAccount(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider to get list of service accounts
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeAzureSubscriptionID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider to get list of subscription IDs
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeAzureTenantID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider to get list of tenant IDs
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeAzureResourceGroupName(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider to get list of resource group names
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeAzureRouteTableName(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider to get list of route table names
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeAzureAdminUsers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider to get list of admin users
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeSpotinstProduct(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider to get list of products
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeSpotinstOrientation(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider to get list of orientations
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeOpenstackExternalNet(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider to get list of external networks
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeOpenstackExternalSubnet(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider to get list of external floating subnets
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeOpenstackLBSubnet(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider to get list of external subnets
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeOpenstackOctaviaProvider(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	providers := []string{
		"a10",
		"amphora",
		"amphorav2",
		"f5",
		"octavia",
		"ovn",
		"radware",
		"vmwareedge",
	}
	return providers, cobra.ShellCompDirectiveNoFileComp
}

func completeOpenstackDNSServers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider to get list of DNS servers
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeOpenstackNetworkID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO call into cloud provider to get list of network IDs
	return nil, cobra.ShellCompDirectiveNoFileComp
}
