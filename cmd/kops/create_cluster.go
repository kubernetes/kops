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
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog"
	"k8s.io/kops"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/commands"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/k8sversion"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

const (
	AuthorizationFlagAlwaysAllow = "AlwaysAllow"
	AuthorizationFlagRBAC        = "RBAC"
)

type CreateClusterOptions struct {
	ClusterName          string
	Yes                  bool
	Target               string
	Models               string
	Cloud                string
	Zones                []string
	MasterZones          []string
	NodeSize             string
	MasterSize           string
	MasterCount          int32
	NodeCount            int32
	MasterVolumeSize     int32
	NodeVolumeSize       int32
	EncryptEtcdStorage   bool
	EtcdStorageType      string
	Project              string
	KubernetesVersion    string
	ContainerRuntime     string
	OutDir               string
	Image                string
	VPCID                string
	SubnetIDs            []string
	UtilitySubnetIDs     []string
	DisableSubnetTags    bool
	NetworkCIDR          string
	DNSZone              string
	AdminAccess          []string
	SSHAccess            []string
	Networking           string
	NodeSecurityGroups   []string
	MasterSecurityGroups []string
	AssociatePublicIP    *bool

	// SSHPublicKeys is a map of the SSH public keys we should configure; required on AWS, not required on GCE
	SSHPublicKeys map[string][]byte

	// Overrides allows settings values direct in the spec
	Overrides []string

	// Channel is the location of the api.Channel to use for our defaults
	Channel string

	// The network topology to use
	Topology string

	// The authorization approach to use (RBAC, AlwaysAllow)
	Authorization string

	// The DNS type to use (public/private)
	DNSType string

	// Enable/Disable Bastion Host complete setup
	Bastion bool

	// Specify tags for AWS instance groups
	CloudLabels string

	// Egress configuration - FOR TESTING ONLY
	Egress string

	// Specify tenancy (default or dedicated) for masters and nodes
	MasterTenancy string
	NodeTenancy   string

	// Specify API loadbalancer as public or internal
	APILoadBalancerType string

	// Specify the SSL certificate to use for the API loadbalancer. Currently only supported in AWS.
	APISSLCertificate string

	// Allow custom public master name
	MasterPublicName string

	// vSphere options
	VSphereServer        string
	VSphereDatacenter    string
	VSphereResourcePool  string
	VSphereCoreDNSServer string
	// Note: We need open-vm-tools to be installed for vSphere Cloud Provider to work
	// We need VSphereDatastore to support Kubernetes vSphere Cloud Provider (v1.5.3)
	// We can remove this once we support higher versions.
	VSphereDatastore string

	// Spotinst options
	SpotinstProduct     string
	SpotinstOrientation string

	// OpenstackExternalNet is the name of the external network for the openstack router
	OpenstackExternalNet     string
	OpenstackExternalSubnet  string
	OpenstackStorageIgnoreAZ bool
	OpenstackDNSServers      string
	OpenstackLbSubnet        string
	OpenstackNetworkID       string
	// OpenstackLBOctavia is boolean value should we use octavia or old loadbalancer api
	OpenstackLBOctavia bool

	// GCEServiceAccount specifies the service account with which the GCE VM runs
	GCEServiceAccount string

	// ConfigBase is the location where we will store the configuration, it defaults to the state store
	ConfigBase string

	// DryRun mode output a cluster manifest of Output type.
	DryRun bool
	// Output type during a DryRun
	Output string
}

func (o *CreateClusterOptions) InitDefaults() {
	o.Yes = false
	o.Target = cloudup.TargetDirect
	o.Models = strings.Join(cloudup.CloudupModels, ",")
	o.Networking = "kubenet"
	o.Channel = api.DefaultChannel
	o.Topology = api.TopologyPublic
	o.DNSType = string(api.DNSTypePublic)
	o.Bastion = false

	// Default to open API & SSH access
	o.AdminAccess = []string{"0.0.0.0/0"}

	o.Authorization = AuthorizationFlagRBAC

	o.ContainerRuntime = "docker"
}

var (
	createClusterLong = templates.LongDesc(i18n.T(`
	Create a kubernetes cluster using command line flags.
	This command creates cloud based resources such as networks and virtual machines. Once
	the infrastructure is in place Kubernetes is installed on the virtual machines.

	These operations are done in parallel and rely on eventual consistency.
	`))

	createClusterExample = templates.Examples(i18n.T(`
	# Create a cluster in AWS in a single zone.
	kops create cluster --name=kubernetes-cluster.example.com \
		--state=s3://kops-state-1234 \
		--zones=eu-west-1a \
		--node-count=2

	# Create a cluster in AWS with HA masters. This cluster
	# has also been configured for private networking in a kops-managed VPC.
	# The bastion flag is set to create an entrypoint for admins to SSH.
	export NODE_SIZE=${NODE_SIZE:-m4.large}
	export MASTER_SIZE=${MASTER_SIZE:-m4.large}
	export ZONES=${ZONES:-"us-east-1d,us-east-1b,us-east-1c"}
	export KOPS_STATE_STORE="s3://my-state-store"
	kops create cluster k8s-clusters.example.com \
	    --node-count 3 \
		--zones $ZONES \
		--node-size $NODE_SIZE \
		--master-size $MASTER_SIZE \
		--master-zones $ZONES \
		--networking weave \
		--topology private \
		--bastion="true" \
		--yes

	# Create a cluster in GCE.
	export KOPS_STATE_STORE="gs://mybucket-kops"
	export ZONES=${MASTER_ZONES:-"us-east1-b,us-east1-c,us-east1-d"}
	export KOPS_FEATURE_FLAGS=AlphaAllowGCE # Note: GCE support is not GA.
	kops create cluster kubernetes-k8s-gce.example.com \
		--zones $ZONES \
		--master-zones $ZONES \
		--node-count 3 \
		--yes

	# Generate a cluster spec to apply later.
	# Run the following, then: kops create -f filename.yamlh
	kops create cluster --name=kubernetes-cluster.example.com \
		--state=s3://kops-state-1234 \
		--zones=eu-west-1a \
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

	cmd := &cobra.Command{
		Use:     "cluster",
		Short:   createClusterShort,
		Long:    createClusterLong,
		Example: createClusterExample,
		Run: func(cmd *cobra.Command, args []string) {
			if cmd.Flag("associate-public-ip").Changed {
				options.AssociatePublicIP = &associatePublicIP
			}

			err := rootCommand.ProcessArgs(args)
			if err != nil {
				exitWithError(err)
				return
			}

			options.ClusterName = rootCommand.clusterName

			if sshPublicKey != "" {
				options.SSHPublicKeys, err = loadSSHPublicKeys(sshPublicKey)
				if err != nil {
					exitWithError(fmt.Errorf("error reading SSH key file %q: %v", sshPublicKey, err))
				}
			}

			err = RunCreateCluster(f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().BoolVarP(&options.Yes, "yes", "y", options.Yes, "Specify --yes to immediately create the cluster")
	cmd.Flags().StringVar(&options.Target, "target", options.Target, fmt.Sprintf("Valid targets: %s, %s, %s. Set this flag to %s if you want kops to generate terraform", cloudup.TargetDirect, cloudup.TargetTerraform, cloudup.TargetCloudformation, cloudup.TargetTerraform))
	cmd.Flags().StringVar(&options.Models, "model", options.Models, "Models to apply (separate multiple models with commas)")

	// Configuration / state location
	if featureflag.EnableSeparateConfigBase.Enabled() {
		cmd.Flags().StringVar(&options.ConfigBase, "config-base", options.ConfigBase, "A cluster-readable location where we mirror configuration information, separate from the state store.  Allows for a state store that is not accessible from the cluster.")
	}

	cmd.Flags().StringVar(&options.Cloud, "cloud", options.Cloud, "Cloud provider to use - gce, aws, vsphere, openstack")

	cmd.Flags().StringSliceVar(&options.Zones, "zones", options.Zones, "Zones in which to run the cluster")
	cmd.Flags().StringSliceVar(&options.MasterZones, "master-zones", options.MasterZones, "Zones in which to run masters (must be an odd number)")

	cmd.Flags().StringVar(&options.KubernetesVersion, "kubernetes-version", options.KubernetesVersion, "Version of kubernetes to run (defaults to version in channel)")

	cmd.Flags().StringVar(&options.ContainerRuntime, "container-runtime", options.ContainerRuntime, "Container runtime to use: containerd, docker")

	cmd.Flags().StringVar(&sshPublicKey, "ssh-public-key", sshPublicKey, "SSH public key to use (defaults to ~/.ssh/id_rsa.pub on AWS)")

	cmd.Flags().StringVar(&options.NodeSize, "node-size", options.NodeSize, "Set instance size for nodes")

	cmd.Flags().StringVar(&options.MasterSize, "master-size", options.MasterSize, "Set instance size for masters")

	cmd.Flags().Int32Var(&options.MasterVolumeSize, "master-volume-size", options.MasterVolumeSize, "Set instance volume size (in GB) for masters")
	cmd.Flags().Int32Var(&options.NodeVolumeSize, "node-volume-size", options.NodeVolumeSize, "Set instance volume size (in GB) for nodes")

	cmd.Flags().StringVar(&options.VPCID, "vpc", options.VPCID, "Set to use a shared VPC")
	cmd.Flags().StringSliceVar(&options.SubnetIDs, "subnets", options.SubnetIDs, "Set to use shared subnets")
	cmd.Flags().StringSliceVar(&options.UtilitySubnetIDs, "utility-subnets", options.UtilitySubnetIDs, "Set to use shared utility subnets")
	cmd.Flags().StringVar(&options.NetworkCIDR, "network-cidr", options.NetworkCIDR, "Set to override the default network CIDR")
	cmd.Flags().BoolVar(&options.DisableSubnetTags, "disable-subnet-tags", options.DisableSubnetTags, "Set to disable automatic subnet tagging")

	cmd.Flags().Int32Var(&options.MasterCount, "master-count", options.MasterCount, "Set the number of masters.  Defaults to one master per master-zone")
	cmd.Flags().Int32Var(&options.NodeCount, "node-count", options.NodeCount, "Set the number of nodes")
	cmd.Flags().BoolVar(&options.EncryptEtcdStorage, "encrypt-etcd-storage", options.EncryptEtcdStorage, "Generate key in aws kms and use it for encrypt etcd volumes")
	cmd.Flags().StringVar(&options.EtcdStorageType, "etcd-storage-type", options.EtcdStorageType, "The default storage type for etc members")

	cmd.Flags().StringVar(&options.Image, "image", options.Image, "Image to use for all instances.")

	cmd.Flags().StringVar(&options.Networking, "networking", options.Networking, "Networking mode to use.  kubenet (default), classic, external, kopeio-vxlan (or kopeio), weave, flannel-vxlan (or flannel), flannel-udp, calico, canal, kube-router, romana, amazon-vpc-routed-eni, cilium, cni.")

	cmd.Flags().StringVar(&options.DNSZone, "dns-zone", options.DNSZone, "DNS hosted zone to use (defaults to longest matching zone)")
	cmd.Flags().StringVar(&options.OutDir, "out", options.OutDir, "Path to write any local output")
	cmd.Flags().StringSliceVar(&options.AdminAccess, "admin-access", options.AdminAccess, "Restrict API access to this CIDR.  If not set, access will not be restricted by IP.")
	cmd.Flags().StringSliceVar(&options.SSHAccess, "ssh-access", options.SSHAccess, "Restrict SSH access to this CIDR.  If not set, access will not be restricted by IP. (default [0.0.0.0/0])")

	// TODO: Can we deprecate this flag - it is awkward?
	cmd.Flags().BoolVar(&associatePublicIP, "associate-public-ip", false, "Specify --associate-public-ip=[true|false] to enable/disable association of public IP for master ASG and nodes. Default is 'true'.")

	cmd.Flags().StringSliceVar(&options.NodeSecurityGroups, "node-security-groups", options.NodeSecurityGroups, "Add precreated additional security groups to nodes.")
	cmd.Flags().StringSliceVar(&options.MasterSecurityGroups, "master-security-groups", options.MasterSecurityGroups, "Add precreated additional security groups to masters.")

	cmd.Flags().StringVar(&options.Channel, "channel", options.Channel, "Channel for default versions and configuration to use")

	// Network topology
	cmd.Flags().StringVarP(&options.Topology, "topology", "t", options.Topology, "Controls network topology for the cluster: public|private.")

	// Authorization
	cmd.Flags().StringVar(&options.Authorization, "authorization", options.Authorization, "Authorization mode to use: "+AuthorizationFlagAlwaysAllow+" or "+AuthorizationFlagRBAC)

	// DNS
	cmd.Flags().StringVar(&options.DNSType, "dns", options.DNSType, "DNS hosted zone to use: public|private.")

	// Bastion
	cmd.Flags().BoolVar(&options.Bastion, "bastion", options.Bastion, "Pass the --bastion flag to enable a bastion instance group. Only applies to private topology.")

	// Allow custom tags from the CLI
	cmd.Flags().StringVar(&options.CloudLabels, "cloud-labels", options.CloudLabels, "A list of KV pairs used to tag all instance groups in AWS (e.g. \"Owner=John Doe,Team=Some Team\").")

	// Master and Node Tenancy
	cmd.Flags().StringVar(&options.MasterTenancy, "master-tenancy", options.MasterTenancy, "The tenancy of the master group on AWS. Can either be default or dedicated.")
	cmd.Flags().StringVar(&options.NodeTenancy, "node-tenancy", options.NodeTenancy, "The tenancy of the node group on AWS. Can be either default or dedicated.")

	cmd.Flags().StringVar(&options.APILoadBalancerType, "api-loadbalancer-type", options.APILoadBalancerType, "Sets the API loadbalancer type to either 'public' or 'internal'")
	cmd.Flags().StringVar(&options.APISSLCertificate, "api-ssl-certificate", options.APISSLCertificate, "Currently only supported in AWS. Sets the ARN of the SSL Certificate to use for the API server loadbalancer.")

	// Allow custom public master name
	cmd.Flags().StringVar(&options.MasterPublicName, "master-public-name", options.MasterPublicName, "Sets the public master public name")

	// DryRun mode that will print YAML or JSON
	cmd.Flags().BoolVar(&options.DryRun, "dry-run", options.DryRun, "If true, only print the object that would be sent, without sending it. This flag can be used to create a cluster YAML or JSON manifest.")
	cmd.Flags().StringVarP(&options.Output, "output", "o", options.Output, "Output format. One of json|yaml. Used with the --dry-run flag.")

	if featureflag.SpecOverrideFlag.Enabled() {
		cmd.Flags().StringSliceVar(&options.Overrides, "override", options.Overrides, "Directly configure values in the spec")
	}

	// GCE flags
	cmd.Flags().StringVar(&options.Project, "project", options.Project, "Project to use (must be set on GCE)")
	cmd.Flags().StringVar(&options.GCEServiceAccount, "gce-service-account", options.GCEServiceAccount, "Service account with which the GCE VM runs. Warning: if not set, VMs will run as default compute service account.")

	if featureflag.VSphereCloudProvider.Enabled() {
		// vSphere flags
		cmd.Flags().StringVar(&options.VSphereServer, "vsphere-server", options.VSphereServer, "vsphere-server is required for vSphere. Set vCenter URL Ex: 10.192.10.30 or myvcenter.io (without https://)")
		cmd.Flags().StringVar(&options.VSphereDatacenter, "vsphere-datacenter", options.VSphereDatacenter, "vsphere-datacenter is required for vSphere. Set the name of the datacenter in which to deploy Kubernetes VMs.")
		cmd.Flags().StringVar(&options.VSphereResourcePool, "vsphere-resource-pool", options.VSphereDatacenter, "vsphere-resource-pool is required for vSphere. Set a valid Cluster, Host or Resource Pool in which to deploy Kubernetes VMs.")
		cmd.Flags().StringVar(&options.VSphereCoreDNSServer, "vsphere-coredns-server", options.VSphereCoreDNSServer, "vsphere-coredns-server is required for vSphere.")
		cmd.Flags().StringVar(&options.VSphereDatastore, "vsphere-datastore", options.VSphereDatastore, "vsphere-datastore is required for vSphere.  Set a valid datastore in which to store dynamic provision volumes.")
	}

	if featureflag.Spotinst.Enabled() {
		// Spotinst flags
		cmd.Flags().StringVar(&options.SpotinstProduct, "spotinst-product", options.SpotinstProduct, "Set the product description (valid values: Linux/UNIX, Linux/UNIX (Amazon VPC), Windows and Windows (Amazon VPC))")
		cmd.Flags().StringVar(&options.SpotinstOrientation, "spotinst-orientation", options.SpotinstOrientation, "Set the prediction strategy (valid values: balanced, cost, equal-distribution and availability)")
	}

	// Openstack flags
	cmd.Flags().StringVar(&options.OpenstackExternalNet, "os-ext-net", options.OpenstackExternalNet, "The name of the external network to use with the openstack router")
	cmd.Flags().StringVar(&options.OpenstackExternalSubnet, "os-ext-subnet", options.OpenstackExternalSubnet, "The name of the external floating subnet to use with the openstack router")
	cmd.Flags().StringVar(&options.OpenstackLbSubnet, "os-lb-floating-subnet", options.OpenstackLbSubnet, "The name of the external subnet to use with the kubernetes api")
	cmd.Flags().BoolVar(&options.OpenstackStorageIgnoreAZ, "os-kubelet-ignore-az", options.OpenstackStorageIgnoreAZ, "If true kubernetes may attach volumes across availability zones")
	cmd.Flags().BoolVar(&options.OpenstackLBOctavia, "os-octavia", options.OpenstackLBOctavia, "If true octavia loadbalancer api will be used")
	cmd.Flags().StringVar(&options.OpenstackDNSServers, "os-dns-servers", options.OpenstackDNSServers, "comma separated list of DNS Servers which is used in network")
	cmd.Flags().StringVar(&options.OpenstackNetworkID, "os-network", options.OpenstackNetworkID, "The ID of the existing OpenStack network to use")
	return cmd
}

func RunCreateCluster(f *util.Factory, out io.Writer, c *CreateClusterOptions) error {
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

	clusterName := c.ClusterName
	if clusterName == "" {
		return fmt.Errorf("--name is required")
	}

	// TODO: Reuse rootCommand stateStore logic?

	if c.OutDir == "" {
		if c.Target == cloudup.TargetTerraform {
			c.OutDir = "out/terraform"
		} else if c.Target == cloudup.TargetCloudformation {
			c.OutDir = "out/cloudformation"
		} else {
			c.OutDir = "out"
		}
	}

	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	cluster, err := clientset.GetCluster(clusterName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			cluster = nil
		} else {
			return err
		}
	}

	if cluster != nil {
		return fmt.Errorf("cluster %q already exists; use 'kops update cluster' to apply changes", clusterName)
	}

	cluster = &api.Cluster{}
	cluster.ObjectMeta.Name = clusterName

	channel, err := api.LoadChannel(c.Channel)
	if err != nil {
		return err
	}

	if channel.Spec.Cluster != nil {
		cluster.Spec = *channel.Spec.Cluster

		kubernetesVersion := api.RecommendedKubernetesVersion(channel, kops.Version)
		if kubernetesVersion != nil {
			cluster.Spec.KubernetesVersion = kubernetesVersion.String()
		}
	}
	cluster.Spec.Channel = c.Channel

	cluster.Spec.ConfigBase = c.ConfigBase
	configBase, err := clientset.ConfigBaseFor(cluster)
	if err != nil {
		return fmt.Errorf("error building ConfigBase for cluster: %v", err)
	}
	cluster.Spec.ConfigBase = configBase.Path()

	// In future we could change the default if the flag is not specified, e.g. in 1.7 maybe the default is RBAC?
	cluster.Spec.Authorization = &api.AuthorizationSpec{}
	if strings.EqualFold(c.Authorization, AuthorizationFlagAlwaysAllow) {
		cluster.Spec.Authorization.AlwaysAllow = &api.AlwaysAllowAuthorizationSpec{}
	} else if strings.EqualFold(c.Authorization, AuthorizationFlagRBAC) {
		cluster.Spec.Authorization.RBAC = &api.RBACAuthorizationSpec{}
	} else {
		return fmt.Errorf("unknown authorization mode %q", c.Authorization)
	}

	if c.Cloud != "" {
		cluster.Spec.CloudProvider = c.Cloud
	}

	allZones := sets.NewString()
	allZones.Insert(c.Zones...)
	allZones.Insert(c.MasterZones...)

	if c.VPCID != "" {
		cluster.Spec.NetworkID = c.VPCID
	} else if api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderAWS && len(c.SubnetIDs) > 0 {
		cloudTags := map[string]string{}
		awsCloud, err := awsup.NewAWSCloud(c.Zones[0][:len(c.Zones[0])-1], cloudTags)
		if err != nil {
			return fmt.Errorf("error loading cloud: %v", err)
		}
		res, err := awsCloud.EC2().DescribeSubnets(&ec2.DescribeSubnetsInput{
			SubnetIds: []*string{aws.String(c.SubnetIDs[0])},
		})
		if err != nil {
			return fmt.Errorf("error describing subnet %s: %v", c.SubnetIDs[0], err)
		}
		if len(res.Subnets) == 0 || res.Subnets[0].VpcId == nil {
			return fmt.Errorf("failed to determine VPC id of subnet %s", c.SubnetIDs[0])
		}
		cluster.Spec.NetworkID = *res.Subnets[0].VpcId
	}

	if api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderOpenstack {
		if cluster.Spec.CloudConfig == nil {
			cluster.Spec.CloudConfig = &api.CloudConfiguration{}
		}
		cluster.Spec.CloudConfig.Openstack = &api.OpenstackConfiguration{
			Router: &api.OpenstackRouter{
				ExternalNetwork: fi.String(c.OpenstackExternalNet),
			},
			BlockStorage: &api.OpenstackBlockStorageConfig{
				Version:  fi.String("v2"),
				IgnoreAZ: fi.Bool(c.OpenstackStorageIgnoreAZ),
			},
			Monitor: &api.OpenstackMonitor{
				Delay:      fi.String("1m"),
				Timeout:    fi.String("30s"),
				MaxRetries: fi.Int(3),
			},
		}

		if c.OpenstackNetworkID != "" {
			cluster.Spec.NetworkID = c.OpenstackNetworkID
		} else if len(c.SubnetIDs) > 0 {
			tags := make(map[string]string)
			tags[openstack.TagClusterName] = c.ClusterName
			osCloud, err := openstack.NewOpenstackCloud(tags, &cluster.Spec)
			if err != nil {
				return fmt.Errorf("error loading cloud: %v", err)
			}

			res, err := osCloud.FindNetworkBySubnetID(c.SubnetIDs[0])
			if err != nil {
				return fmt.Errorf("error finding network: %v", err)
			}
			cluster.Spec.NetworkID = res.ID
		}

		if c.OpenstackDNSServers != "" {
			cluster.Spec.CloudConfig.Openstack.Router.DNSServers = fi.String(c.OpenstackDNSServers)
		}
		if c.OpenstackExternalSubnet != "" {
			cluster.Spec.CloudConfig.Openstack.Router.ExternalSubnet = fi.String(c.OpenstackExternalSubnet)
		}
	}

	if cluster.Spec.CloudProvider == "" {
		for _, zone := range allZones.List() {
			cloud, known := fi.GuessCloudForZone(zone)
			if known {
				klog.Infof("Inferred --cloud=%s from zone %q", cloud, zone)
				cluster.Spec.CloudProvider = string(cloud)
				break
			}
		}
		if cluster.Spec.CloudProvider == "" {
			if allZones.Len() == 0 {
				return fmt.Errorf("must specify --zones or --cloud")
			}
			return fmt.Errorf("unable to infer CloudProvider from Zones (is there a typo in --zones?)")
		}
	}

	zoneToSubnetMap := make(map[string]*api.ClusterSubnetSpec)
	if len(c.Zones) == 0 {
		return fmt.Errorf("must specify at least one zone for the cluster (use --zones)")
	} else if api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderGCE {
		// On GCE, subnets are regional - we create one per region, not per zone
		for _, zoneName := range allZones.List() {
			region, err := gce.ZoneToRegion(zoneName)
			if err != nil {
				return err
			}

			// We create default subnets named the same as the regions
			subnetName := region

			subnet := model.FindSubnet(cluster, subnetName)
			if subnet == nil {
				subnet = &api.ClusterSubnetSpec{
					Name:   subnetName,
					Region: region,
				}
				cluster.Spec.Subnets = append(cluster.Spec.Subnets, *subnet)
			}
			zoneToSubnetMap[zoneName] = subnet
		}
	} else if api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderDO {
		if len(c.Zones) > 1 {
			return fmt.Errorf("digitalocean cloud provider currently only supports 1 region, expect multi-region support when digitalocean support is in beta")
		}

		// For DO we just pass in the region for --zones
		region := c.Zones[0]
		subnet := model.FindSubnet(cluster, region)

		// for DO, subnets are just regions
		subnetName := region

		if subnet == nil {
			subnet = &api.ClusterSubnetSpec{
				Name: subnetName,
				// region and zone are the same for DO
				Region: region,
				Zone:   region,
			}
			cluster.Spec.Subnets = append(cluster.Spec.Subnets, *subnet)
		}
		zoneToSubnetMap[region] = subnet
	} else if api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderALI {
		var zoneToSubnetSwitchID map[string]string
		if len(c.Zones) > 0 && len(c.SubnetIDs) > 0 && api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderALI {
			zoneToSubnetSwitchID, err = aliup.ZoneToVSwitchID(cluster.Spec.NetworkID, c.Zones, c.SubnetIDs)
			if err != nil {
				return err
			}
		}
		for _, zoneName := range allZones.List() {
			// We create default subnets named the same as the zones
			subnetName := zoneName

			subnet := model.FindSubnet(cluster, subnetName)
			if subnet == nil {
				subnet = &api.ClusterSubnetSpec{
					Name:   subnetName,
					Zone:   subnetName,
					Egress: c.Egress,
				}
				if vswitchID, ok := zoneToSubnetSwitchID[zoneName]; ok {
					subnet.ProviderID = vswitchID
				}
				cluster.Spec.Subnets = append(cluster.Spec.Subnets, *subnet)
			}
			zoneToSubnetMap[zoneName] = subnet
		}
	} else {
		var zoneToSubnetProviderID map[string]string
		if len(c.Zones) > 0 && len(c.SubnetIDs) > 0 {
			if api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderAWS {
				zoneToSubnetProviderID, err = getZoneToSubnetProviderID(cluster.Spec.NetworkID, c.Zones[0][:len(c.Zones[0])-1], c.SubnetIDs)
				if err != nil {
					return err
				}
			} else if api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderOpenstack {
				tags := make(map[string]string)
				tags[openstack.TagClusterName] = c.ClusterName
				zoneToSubnetProviderID, err = getSubnetProviderID(&cluster.Spec, allZones.List(), c.SubnetIDs, tags)
				if err != nil {
					return err
				}
			}
		}
		for _, zoneName := range allZones.List() {
			// We create default subnets named the same as the zones
			subnetName := zoneName

			subnet := model.FindSubnet(cluster, subnetName)
			if subnet == nil {
				subnet = &api.ClusterSubnetSpec{
					Name:   subnetName,
					Zone:   subnetName,
					Egress: c.Egress,
				}
				if subnetID, ok := zoneToSubnetProviderID[zoneName]; ok {
					subnet.ProviderID = subnetID
				}
				cluster.Spec.Subnets = append(cluster.Spec.Subnets, *subnet)
			}
			zoneToSubnetMap[zoneName] = subnet
		}
	}

	var masters []*api.InstanceGroup
	var nodes []*api.InstanceGroup
	var instanceGroups []*api.InstanceGroup
	cloudLabels, err := parseCloudLabels(c.CloudLabels)
	if err != nil {
		return fmt.Errorf("error parsing global cloud labels: %v", err)
	}
	if len(cloudLabels) != 0 {
		cluster.Spec.CloudLabels = cloudLabels
	}

	// Build the master subnets
	// The master zones is the default set of zones unless explicitly set
	// The master count is the number of master zones unless explicitly set
	// We then round-robin around the zones
	if len(masters) == 0 {
		masterCount := c.MasterCount
		masterZones := c.MasterZones
		if len(masterZones) != 0 {
			if c.MasterCount != 0 && c.MasterCount < int32(len(c.MasterZones)) {
				return fmt.Errorf("specified %d master zones, but also requested %d masters.  If specifying both, the count should match.", len(masterZones), c.MasterCount)
			}

			if masterCount == 0 {
				// If master count is not specified, default to the number of master zones
				masterCount = int32(len(c.MasterZones))
			}
		} else {
			// masterZones not set; default to same as node Zones
			masterZones = c.Zones

			if masterCount == 0 {
				// If master count is not specified, default to 1
				masterCount = 1
			}
		}

		if len(masterZones) == 0 {
			// Should be unreachable
			return fmt.Errorf("cannot determine master zones")
		}

		for i := 0; i < int(masterCount); i++ {
			zone := masterZones[i%len(masterZones)]
			name := zone
			if api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderDO {
				if int(masterCount) >= len(masterZones) {
					name += "-" + strconv.Itoa(1+(i/len(masterZones)))
				}
			} else {
				if int(masterCount) > len(masterZones) {
					name += "-" + strconv.Itoa(1+(i/len(masterZones)))
				}
			}

			g := &api.InstanceGroup{}
			g.Spec.Role = api.InstanceGroupRoleMaster
			g.Spec.MinSize = fi.Int32(1)
			g.Spec.MaxSize = fi.Int32(1)
			g.ObjectMeta.Name = "master-" + name

			subnet := zoneToSubnetMap[zone]
			if subnet == nil {
				klog.Fatalf("subnet not found in zoneToSubnetMap")
			}

			g.Spec.Subnets = []string{subnet.Name}
			if api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderGCE {
				g.Spec.Zones = []string{zone}
			}

			instanceGroups = append(instanceGroups, g)
			masters = append(masters, g)
		}
	}

	if len(cluster.Spec.EtcdClusters) == 0 {
		masterAZs := sets.NewString()
		duplicateAZs := false
		for _, ig := range masters {
			zones, err := model.FindZonesForInstanceGroup(cluster, ig)
			if err != nil {
				return err
			}
			for _, zone := range zones {
				if masterAZs.Has(zone) {
					duplicateAZs = true
				}

				masterAZs.Insert(zone)
			}
		}

		if duplicateAZs {
			klog.Warningf("Running with masters in the same AZs; redundancy will be reduced")
		}

		for _, etcdCluster := range cloudup.EtcdClusters {
			etcd := &api.EtcdClusterSpec{}
			etcd.Name = etcdCluster

			// if this is the main cluster, we use 200 millicores by default.
			// otherwise we use 100 millicores by default.  100Mi is always default
			// for event and main clusters.  This is changeable in the kops cluster
			// configuration.
			if etcd.Name == "main" {
				cpuRequest := resource.MustParse("200m")
				etcd.CPURequest = &cpuRequest
			} else {
				cpuRequest := resource.MustParse("100m")
				etcd.CPURequest = &cpuRequest
			}
			memoryRequest := resource.MustParse("100Mi")
			etcd.MemoryRequest = &memoryRequest

			var names []string
			for _, ig := range masters {
				name := ig.ObjectMeta.Name
				// We expect the IG to have a `master-` prefix, but this is both superfluous
				// and not how we named things previously
				name = strings.TrimPrefix(name, "master-")
				names = append(names, name)
			}

			names = trimCommonPrefix(names)

			for i, ig := range masters {
				m := &api.EtcdMemberSpec{}
				if c.EncryptEtcdStorage {
					m.EncryptedVolume = &c.EncryptEtcdStorage
				}
				if len(c.EtcdStorageType) > 0 {
					m.VolumeType = fi.String(c.EtcdStorageType)
				}
				m.Name = names[i]

				m.InstanceGroup = fi.String(ig.ObjectMeta.Name)
				etcd.Members = append(etcd.Members, m)
			}
			cluster.Spec.EtcdClusters = append(cluster.Spec.EtcdClusters, etcd)
		}
	}

	if len(nodes) == 0 {
		g := &api.InstanceGroup{}
		g.Spec.Role = api.InstanceGroupRoleNode
		g.ObjectMeta.Name = "nodes"

		subnetNames := sets.NewString()
		for _, zone := range c.Zones {
			subnet := zoneToSubnetMap[zone]
			if subnet == nil {
				klog.Fatalf("subnet not found in zoneToSubnetMap")
			}
			subnetNames.Insert(subnet.Name)
		}
		g.Spec.Subnets = subnetNames.List()

		if api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderGCE {
			g.Spec.Zones = c.Zones
		}

		instanceGroups = append(instanceGroups, g)
		nodes = append(nodes, g)
	}

	if c.NodeSize != "" {
		for _, group := range nodes {
			group.Spec.MachineType = c.NodeSize
		}
	}

	if c.Image != "" {
		for _, group := range instanceGroups {
			group.Spec.Image = c.Image
		}
	}

	if c.AssociatePublicIP != nil {
		for _, group := range instanceGroups {
			group.Spec.AssociatePublicIP = c.AssociatePublicIP
		}
	}

	if c.NodeCount != 0 {
		for _, group := range nodes {
			group.Spec.MinSize = fi.Int32(c.NodeCount)
			group.Spec.MaxSize = fi.Int32(c.NodeCount)
		}
	}

	if c.MasterTenancy != "" {
		for _, group := range masters {
			group.Spec.Tenancy = c.MasterTenancy
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

	if len(c.MasterSecurityGroups) > 0 {
		for _, group := range masters {
			group.Spec.AdditionalSecurityGroups = c.MasterSecurityGroups
		}
	}

	if c.MasterSize != "" {
		for _, group := range masters {
			group.Spec.MachineType = c.MasterSize
		}
	}

	if c.MasterVolumeSize != 0 {
		for _, group := range masters {
			group.Spec.RootVolumeSize = fi.Int32(c.MasterVolumeSize)
		}
	}

	if c.NodeVolumeSize != 0 {
		for _, group := range nodes {
			group.Spec.RootVolumeSize = fi.Int32(c.NodeVolumeSize)
		}
	}

	if c.DNSZone != "" {
		cluster.Spec.DNSZone = c.DNSZone
	}

	if c.Cloud != "" {
		if c.Cloud == "vsphere" {
			if !featureflag.VSphereCloudProvider.Enabled() {
				return fmt.Errorf("Feature flag VSphereCloudProvider is not set. Cloud vSphere will not be supported.")
			}

			if cluster.Spec.CloudConfig == nil {
				cluster.Spec.CloudConfig = &api.CloudConfiguration{}
			}

			if c.VSphereServer == "" {
				return fmt.Errorf("vsphere-server is required for vSphere. Set vCenter URL Ex: 10.192.10.30 or myvcenter.io (without https://)")
			}
			cluster.Spec.CloudConfig.VSphereServer = fi.String(c.VSphereServer)

			if c.VSphereDatacenter == "" {
				return fmt.Errorf("vsphere-datacenter is required for vSphere. Set the name of the datacenter in which to deploy Kubernetes VMs.")
			}
			cluster.Spec.CloudConfig.VSphereDatacenter = fi.String(c.VSphereDatacenter)

			if c.VSphereResourcePool == "" {
				return fmt.Errorf("vsphere-resource-pool is required for vSphere. Set a valid Cluster, Host or Resource Pool in which to deploy Kubernetes VMs.")
			}
			cluster.Spec.CloudConfig.VSphereResourcePool = fi.String(c.VSphereResourcePool)

			if c.VSphereCoreDNSServer == "" {
				return fmt.Errorf("A coredns server is required for vSphere.")
			}
			cluster.Spec.CloudConfig.VSphereCoreDNSServer = fi.String(c.VSphereCoreDNSServer)

			if c.VSphereDatastore == "" {
				return fmt.Errorf("vsphere-datastore is required for vSphere. Set a valid datastore in which to store dynamic provision volumes.")
			}
			cluster.Spec.CloudConfig.VSphereDatastore = fi.String(c.VSphereDatastore)
		}

		if featureflag.Spotinst.Enabled() {
			if cluster.Spec.CloudConfig == nil {
				cluster.Spec.CloudConfig = &api.CloudConfiguration{}
			}
			if c.SpotinstProduct != "" {
				cluster.Spec.CloudConfig.SpotinstProduct = fi.String(c.SpotinstProduct)
			}
			if c.SpotinstOrientation != "" {
				cluster.Spec.CloudConfig.SpotinstOrientation = fi.String(c.SpotinstOrientation)
			}
		}
	}

	// Populate project
	if c.Project != "" {
		cluster.Spec.Project = c.Project
	}
	if api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderGCE {
		if cluster.Spec.Project == "" {
			project, err := gce.DefaultProject()
			if err != nil {
				klog.Warningf("unable to get default google cloud project: %v", err)
			} else if project == "" {
				klog.Warningf("default google cloud project not set (try `gcloud config set project <name>`")
			} else {
				klog.Infof("using google cloud project: %s", project)
			}
			cluster.Spec.Project = project
		}
		if c.GCEServiceAccount != "" {
			klog.Infof("using GCE service account: %v", c.GCEServiceAccount)
			cluster.Spec.GCEServiceAccount = fi.String(c.GCEServiceAccount)
		} else {
			klog.Warning("using GCE default service account")
			cluster.Spec.GCEServiceAccount = fi.String("default")
		}
	}

	if c.KubernetesVersion != "" {
		cluster.Spec.KubernetesVersion = c.KubernetesVersion
	}

	if c.ContainerRuntime != "" {
		cluster.Spec.ContainerRuntime = c.ContainerRuntime
	}

	cluster.Spec.Networking = &api.NetworkingSpec{}
	switch c.Networking {
	case "classic":
		cluster.Spec.Networking.Classic = &api.ClassicNetworkingSpec{}
	case "kubenet":
		cluster.Spec.Networking.Kubenet = &api.KubenetNetworkingSpec{}
	case "external":
		cluster.Spec.Networking.External = &api.ExternalNetworkingSpec{}
	case "cni":
		cluster.Spec.Networking.CNI = &api.CNINetworkingSpec{}
	case "kopeio-vxlan", "kopeio":
		cluster.Spec.Networking.Kopeio = &api.KopeioNetworkingSpec{}
	case "weave":
		cluster.Spec.Networking.Weave = &api.WeaveNetworkingSpec{}

		if cluster.Spec.CloudProvider == "aws" {
			// AWS supports "jumbo frames" of 9001 bytes and weave adds up to 87 bytes overhead
			// sets the default to the largest number that leaves enough overhead and is divisible by 4
			jumboFrameMTUSize := int32(8912)
			cluster.Spec.Networking.Weave.MTU = &jumboFrameMTUSize
		}
	case "flannel", "flannel-vxlan":
		cluster.Spec.Networking.Flannel = &api.FlannelNetworkingSpec{
			Backend: "vxlan",
		}
	case "flannel-udp":
		klog.Warningf("flannel UDP mode is not recommended; consider flannel-vxlan instead")
		cluster.Spec.Networking.Flannel = &api.FlannelNetworkingSpec{
			Backend: "udp",
		}
	case "calico":
		cluster.Spec.Networking.Calico = &api.CalicoNetworkingSpec{
			MajorVersion: "v3",
		}
		// Validate to check if etcd clusters have an acceptable version
		if errList := validation.ValidateEtcdVersionForCalicoV3(cluster.Spec.EtcdClusters[0], cluster.Spec.Networking.Calico.MajorVersion, field.NewPath("spec", "networking", "calico")); len(errList) != 0 {

			// This is not a special version but simply of the 3 series
			for _, etcd := range cluster.Spec.EtcdClusters {
				etcd.Version = components.DefaultEtcd3Version_1_11
			}
		}
	case "canal":
		cluster.Spec.Networking.Canal = &api.CanalNetworkingSpec{}
	case "kube-router":
		cluster.Spec.Networking.Kuberouter = &api.KuberouterNetworkingSpec{}
	case "romana":
		cluster.Spec.Networking.Romana = &api.RomanaNetworkingSpec{}
	case "amazonvpc", "amazon-vpc-routed-eni":
		cluster.Spec.Networking.AmazonVPC = &api.AmazonVPCNetworkingSpec{}
	case "cilium":
		cluster.Spec.Networking.Cilium = &api.CiliumNetworkingSpec{}
	case "lyftvpc":
		cluster.Spec.Networking.LyftVPC = &api.LyftVPCNetworkingSpec{}
	case "gce":
		cluster.Spec.Networking.GCE = &api.GCENetworkingSpec{}
	default:
		return fmt.Errorf("unknown networking mode %q", c.Networking)
	}

	klog.V(4).Infof("networking mode=%s => %s", c.Networking, fi.DebugAsJsonString(cluster.Spec.Networking))

	if c.NetworkCIDR != "" {
		cluster.Spec.NetworkCIDR = c.NetworkCIDR
	}

	// Network Topology
	if c.Topology == "" {
		// The flag default should have set this, but we might be being called as a library
		klog.Infof("Empty topology. Defaulting to public topology")
		c.Topology = api.TopologyPublic
	}

	cluster.Spec.DisableSubnetTags = c.DisableSubnetTags

	switch c.Topology {
	case api.TopologyPublic:
		cluster.Spec.Topology = &api.TopologySpec{
			Masters: api.TopologyPublic,
			Nodes:   api.TopologyPublic,
			//Bastion: &api.BastionSpec{Enable: c.Bastion},
		}

		if c.Bastion {
			return fmt.Errorf("Bastion supports --topology='private' only.")
		}

		for i := range cluster.Spec.Subnets {
			cluster.Spec.Subnets[i].Type = api.SubnetTypePublic
		}

	case api.TopologyPrivate:
		if !supportsPrivateTopology(cluster.Spec.Networking) {
			return fmt.Errorf("Invalid networking option %s. Currently only '--networking kopeio-vxlan (or kopeio)', '--networking weave', '--networking flannel', '--networking calico', '--networking canal', '--networking kube-router', '--networking romana', '--networking amazon-vpc-routed-eni', '--networking cilium', '--networking lyftvpc', '--networking cni' are supported for private topologies", c.Networking)
		}
		cluster.Spec.Topology = &api.TopologySpec{
			Masters: api.TopologyPrivate,
			Nodes:   api.TopologyPrivate,
		}

		for i := range cluster.Spec.Subnets {
			cluster.Spec.Subnets[i].Type = api.SubnetTypePrivate
		}

		var utilitySubnets []api.ClusterSubnetSpec

		var zoneToSubnetProviderID map[string]string
		if len(c.Zones) > 0 && len(c.UtilitySubnetIDs) > 0 {
			if api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderAWS {
				zoneToSubnetProviderID, err = getZoneToSubnetProviderID(cluster.Spec.NetworkID, c.Zones[0][:len(c.Zones[0])-1], c.UtilitySubnetIDs)
				if err != nil {
					return err
				}
			} else if api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderOpenstack {
				tags := make(map[string]string)
				tags[openstack.TagClusterName] = c.ClusterName
				zoneToSubnetProviderID, err = getSubnetProviderID(&cluster.Spec, allZones.List(), c.UtilitySubnetIDs, tags)
				if err != nil {
					return err
				}
			}
		}

		for _, s := range cluster.Spec.Subnets {
			if s.Type == api.SubnetTypeUtility {
				continue
			}
			subnet := api.ClusterSubnetSpec{
				Name: "utility-" + s.Name,
				Zone: s.Zone,
				Type: api.SubnetTypeUtility,
			}
			if subnetID, ok := zoneToSubnetProviderID[s.Zone]; ok {
				subnet.ProviderID = subnetID
			}
			utilitySubnets = append(utilitySubnets, subnet)
		}
		cluster.Spec.Subnets = append(cluster.Spec.Subnets, utilitySubnets...)

		if c.Bastion {
			bastionGroup := &api.InstanceGroup{}
			bastionGroup.Spec.Role = api.InstanceGroupRoleBastion
			bastionGroup.ObjectMeta.Name = "bastions"
			bastionGroup.Spec.Image = c.Image
			instanceGroups = append(instanceGroups, bastionGroup)

			if !dns.IsGossipHostname(clusterName) {
				cluster.Spec.Topology.Bastion = &api.BastionSpec{
					BastionPublicName: "bastion." + clusterName,
				}
			}
		}

	default:
		return fmt.Errorf("Invalid topology %s.", c.Topology)
	}

	// DNS
	if c.DNSType == "" {
		// The flag default should have set this, but we might be being called as a library
		klog.Infof("Empty DNS. Defaulting to public DNS")
		c.DNSType = string(api.DNSTypePublic)
	}

	if cluster.Spec.Topology == nil {
		cluster.Spec.Topology = &api.TopologySpec{}
	}
	if cluster.Spec.Topology.DNS == nil {
		cluster.Spec.Topology.DNS = &api.DNSSpec{}
	}
	switch strings.ToLower(c.DNSType) {
	case "public":
		cluster.Spec.Topology.DNS.Type = api.DNSTypePublic
	case "private":
		cluster.Spec.Topology.DNS.Type = api.DNSTypePrivate
	default:
		return fmt.Errorf("unknown DNSType: %q", c.DNSType)
	}

	if c.MasterPublicName != "" {
		cluster.Spec.MasterPublicName = c.MasterPublicName
	}

	kv, err := k8sversion.Parse(cluster.Spec.KubernetesVersion)
	if err != nil {
		return fmt.Errorf("failed to parse kubernetes version: %s", err.Error())
	}

	// check if we should set anonymousAuth to false on k8s versions >=1.11
	if kv.IsGTE("1.11") {
		if cluster.Spec.Kubelet == nil {
			cluster.Spec.Kubelet = &api.KubeletConfigSpec{}
		}

		if cluster.Spec.Kubelet.AnonymousAuth == nil {
			cluster.Spec.Kubelet.AnonymousAuth = fi.Bool(false)
		}
	}

	// Populate the API access, so that it can be discoverable
	// TODO: This is the same code as in defaults - try to dedup?
	if cluster.Spec.API == nil {
		cluster.Spec.API = &api.AccessSpec{}
	}
	if cluster.Spec.API.IsEmpty() {
		if c.Cloud == "openstack" {
			initializeOpenstackAPI(c, cluster)
		} else if c.APILoadBalancerType != "" {
			cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}
		} else {
			switch cluster.Spec.Topology.Masters {
			case api.TopologyPublic:
				if dns.IsGossipHostname(cluster.Name) {
					// gossip DNS names don't work outside the cluster, so we use a LoadBalancer instead
					cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}
				} else {
					cluster.Spec.API.DNS = &api.DNSAccessSpec{}
				}

			case api.TopologyPrivate:
				cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}

			default:
				return fmt.Errorf("unknown master topology type: %q", cluster.Spec.Topology.Masters)
			}
		}
	}
	if cluster.Spec.API.LoadBalancer != nil && cluster.Spec.API.LoadBalancer.Type == "" {
		switch c.APILoadBalancerType {
		case "", "public":
			cluster.Spec.API.LoadBalancer.Type = api.LoadBalancerTypePublic
		case "internal":
			cluster.Spec.API.LoadBalancer.Type = api.LoadBalancerTypeInternal
		default:
			return fmt.Errorf("unknown api-loadbalancer-type: %q", c.APILoadBalancerType)
		}
	}

	if cluster.Spec.API.LoadBalancer != nil && c.APISSLCertificate != "" {
		cluster.Spec.API.LoadBalancer.SSLCertificate = c.APISSLCertificate
	}

	// Use Strict IAM policy and allow AWS ECR by default when creating a new cluster
	cluster.Spec.IAM = &api.IAMSpec{
		AllowContainerRegistry: true,
		Legacy:                 false,
	}

	if len(c.AdminAccess) != 0 {
		if len(c.SSHAccess) != 0 {
			cluster.Spec.SSHAccess = c.SSHAccess
		} else {
			cluster.Spec.SSHAccess = c.AdminAccess
		}
		cluster.Spec.KubernetesAPIAccess = c.AdminAccess
	} else if len(c.AdminAccess) == 0 {
		cluster.Spec.SSHAccess = c.SSHAccess
	}

	if err := commands.SetClusterFields(c.Overrides, cluster, instanceGroups); err != nil {
		return err
	}

	err = cloudup.PerformAssignments(cluster)
	if err != nil {
		return fmt.Errorf("error populating configuration: %v", err)
	}
	err = api.PerformAssignmentsInstanceGroups(instanceGroups)
	if err != nil {
		return fmt.Errorf("error populating configuration: %v", err)
	}

	if cluster.Spec.ExternalCloudControllerManager != nil && !featureflag.EnableExternalCloudController.Enabled() {
		klog.Warningf("Without setting the feature flag `+EnableExternalCloudController` the external cloud controller manager configuration will be discarded")
	}

	strict := false
	err = validation.DeepValidate(cluster, instanceGroups, strict)
	if err != nil {
		return err
	}

	assetBuilder := assets.NewAssetBuilder(cluster, "")
	fullCluster, err := cloudup.PopulateClusterSpec(clientset, cluster, assetBuilder)
	if err != nil {
		return err
	}

	var fullInstanceGroups []*api.InstanceGroup
	for _, group := range instanceGroups {
		fullGroup, err := cloudup.PopulateInstanceGroupSpec(fullCluster, group, channel)
		if err != nil {
			return err
		}
		fullGroup.AddInstanceGroupNodeLabel()
		if api.CloudProviderID(cluster.Spec.CloudProvider) == api.CloudProviderGCE {
			fullGroup.Spec.NodeLabels["cloud.google.com/metadata-proxy-ready"] = "true"
		}
		fullInstanceGroups = append(fullInstanceGroups, fullGroup)
	}

	err = validation.DeepValidate(fullCluster, fullInstanceGroups, true)
	if err != nil {
		return err
	}

	if c.DryRun {
		var obj []runtime.Object
		obj = append(obj, cluster)

		for _, group := range fullInstanceGroups {
			// Cluster name is not populated, and we need it
			group.ObjectMeta.Labels = make(map[string]string)
			group.ObjectMeta.Labels[api.LabelClusterName] = cluster.ObjectMeta.Name
			obj = append(obj, group)
		}
		switch c.Output {
		case OutputYaml:
			if err := fullOutputYAML(out, obj...); err != nil {
				return fmt.Errorf("error writing cluster yaml to stdout: %v", err)
			}
			return nil
		case OutputJSON:
			if err := fullOutputJSON(out, obj...); err != nil {
				return fmt.Errorf("error writing cluster json to stdout: %v", err)
			}
			return nil
		default:
			return fmt.Errorf("unsupported output type %q", c.Output)
		}
	}

	// Note we perform as much validation as we can, before writing a bad config
	err = registry.CreateClusterConfig(clientset, cluster, fullInstanceGroups)
	if err != nil {
		return fmt.Errorf("error writing updated configuration: %v", err)
	}

	err = registry.WriteConfigDeprecated(cluster, configBase.Join(registry.PathClusterCompleted), fullCluster)
	if err != nil {
		return fmt.Errorf("error writing completed cluster spec: %v", err)
	}

	if len(c.SSHPublicKeys) == 0 {
		autoloadSSHPublicKeys := true
		switch c.Cloud {
		case "gce":
			// We don't normally use SSH keys on GCE
			autoloadSSHPublicKeys = false
		}

		if autoloadSSHPublicKeys {
			// Load from default location, if found
			sshPublicKeyPath := "~/.ssh/id_rsa.pub"
			c.SSHPublicKeys, err = loadSSHPublicKeys(sshPublicKeyPath)
			if err != nil {
				// Don't wrap file-not-found
				if os.IsNotExist(err) {
					klog.V(2).Infof("ssh key not found at %s", sshPublicKeyPath)
				} else {
					return fmt.Errorf("error reading SSH key file %q: %v", sshPublicKeyPath, err)
				}
			}
		}
	}

	if len(c.SSHPublicKeys) != 0 {
		sshCredentialStore, err := clientset.SSHCredentialStore(cluster)
		if err != nil {
			return err
		}

		for k, data := range c.SSHPublicKeys {
			err = sshCredentialStore.AddSSHPublicKey(k, data)
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
		updateClusterOptions.Models = c.Models
		updateClusterOptions.OutDir = c.OutDir

		// SSHPublicKey has already been mapped
		updateClusterOptions.SSHPublicKey = ""

		// No equivalent options:
		//  updateClusterOptions.MaxTaskDuration = c.MaxTaskDuration
		//  updateClusterOptions.CreateKubecfg = c.CreateKubecfg

		_, err := RunUpdateCluster(f, clusterName, out, updateClusterOptions)
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
			fmt.Fprintf(&sb, " * edit this cluster with: kops edit cluster %s\n", clusterName)
			if len(nodes) > 0 {
				fmt.Fprintf(&sb, " * edit your node instance group: kops edit ig --name=%s %s\n", clusterName, nodes[0].ObjectMeta.Name)
			}
			if len(masters) > 0 {
				fmt.Fprintf(&sb, " * edit your master instance group: kops edit ig --name=%s %s\n", clusterName, masters[0].ObjectMeta.Name)
			}
			fmt.Fprintf(&sb, "\n")
			fmt.Fprintf(&sb, "Finally configure your cluster with: kops update cluster --name %s --yes\n", clusterName)
			fmt.Fprintf(&sb, "\n")

			_, err := out.Write(sb.Bytes())
			if err != nil {
				return fmt.Errorf("error writing to output: %v", err)
			}
		}
	}

	return nil
}

func supportsPrivateTopology(n *api.NetworkingSpec) bool {

	if n.CNI != nil || n.Kopeio != nil || n.Weave != nil || n.Flannel != nil || n.Calico != nil || n.Canal != nil || n.Kuberouter != nil || n.Romana != nil || n.AmazonVPC != nil || n.Cilium != nil || n.LyftVPC != nil || n.GCE != nil {
		return true
	}
	return false
}

func trimCommonPrefix(names []string) []string {
	// Trim shared prefix to keep the lengths sane
	// (this only applies to new clusters...)
	for len(names) != 0 && len(names[0]) > 1 {
		prefix := names[0][:1]
		allMatch := true
		for _, name := range names {
			if !strings.HasPrefix(name, prefix) {
				allMatch = false
			}
		}

		if !allMatch {
			break
		}

		for i := range names {
			names[i] = strings.TrimPrefix(names[i], prefix)
		}
	}

	return names
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
		return nil, fmt.Errorf("One or more key=value pairs are malformed:\n%s\n:%v", records, err)
	}

	m := make(map[string]string, len(kvPairs))
	for _, pair := range kvPairs {
		m[pair[0]] = pair[1]
	}
	return m, nil
}

func initializeOpenstackAPI(c *CreateClusterOptions, cluster *api.Cluster) {
	if c.APILoadBalancerType != "" {
		cluster.Spec.API.LoadBalancer = &api.LoadBalancerAccessSpec{}
		provider := "haproxy"
		if c.OpenstackLBOctavia {
			provider = "octavia"
		}

		cluster.Spec.CloudConfig.Openstack.Loadbalancer = &api.OpenstackLoadbalancerConfig{
			FloatingNetwork: fi.String(c.OpenstackExternalNet),
			Method:          fi.String("ROUND_ROBIN"),
			Provider:        fi.String(provider),
			UseOctavia:      fi.Bool(c.OpenstackLBOctavia),
		}

		if c.OpenstackLbSubnet != "" {
			cluster.Spec.CloudConfig.Openstack.Loadbalancer.FloatingSubnet = fi.String(c.OpenstackLbSubnet)
		}
	}
}

func getZoneToSubnetProviderID(VPCID string, region string, subnetIDs []string) (map[string]string, error) {
	res := make(map[string]string)
	cloudTags := map[string]string{}
	awsCloud, err := awsup.NewAWSCloud(region, cloudTags)
	if err != nil {
		return res, fmt.Errorf("error loading cloud: %v", err)
	}
	vpcInfo, err := awsCloud.FindVPCInfo(VPCID)
	if err != nil {
		return res, fmt.Errorf("error describing VPC: %v", err)
	}
	if vpcInfo == nil {
		return res, fmt.Errorf("VPC %q not found", VPCID)
	}
	subnetByID := make(map[string]*fi.SubnetInfo)
	for _, subnetInfo := range vpcInfo.Subnets {
		subnetByID[subnetInfo.ID] = subnetInfo
	}
	for _, subnetID := range subnetIDs {
		subnet, ok := subnetByID[subnetID]
		if !ok {
			return res, fmt.Errorf("subnet %s not found in VPC %s", subnetID, VPCID)
		}
		if res[subnet.Zone] != "" {
			return res, fmt.Errorf("subnet %s and %s have the same zone", subnetID, res[subnet.Zone])
		}
		res[subnet.Zone] = subnetID
	}
	return res, nil
}

func getSubnetProviderID(spec *api.ClusterSpec, zones []string, subnetIDs []string, tags map[string]string) (map[string]string, error) {
	res := make(map[string]string)
	osCloud, err := openstack.NewOpenstackCloud(tags, spec)
	if err != nil {
		return res, fmt.Errorf("error loading cloud: %v", err)
	}
	osCloud.UseZones(zones)

	networkInfo, err := osCloud.FindVPCInfo(spec.NetworkID)
	if err != nil {
		return res, fmt.Errorf("error describing Network: %v", err)
	}
	if networkInfo == nil {
		return res, fmt.Errorf("network %q not found", spec.NetworkID)
	}

	subnetByID := make(map[string]*fi.SubnetInfo)
	for _, subnetInfo := range networkInfo.Subnets {
		subnetByID[subnetInfo.ID] = subnetInfo
	}

	for _, subnetID := range subnetIDs {
		subnet, ok := subnetByID[subnetID]
		if !ok {
			return res, fmt.Errorf("subnet %s not found in network %s", subnetID, spec.NetworkID)
		}

		if res[subnet.Zone] != "" {
			return res, fmt.Errorf("subnet %s and %s have the same zone", subnetID, res[subnet.Zone])
		}
		res[subnet.Zone] = subnetID
	}
	return res, nil
}

func loadSSHPublicKeys(sshPublicKey string) (map[string][]byte, error) {
	sshPublicKeys := make(map[string][]byte)
	if sshPublicKey != "" {
		sshPublicKey = utils.ExpandPath(sshPublicKey)
		authorized, err := ioutil.ReadFile(sshPublicKey)
		if err != nil {
			return nil, err
		}
		sshPublicKeys[fi.SecretNameSSHPrimary] = authorized
		klog.Infof("Using SSH public key: %v\n", sshPublicKey)
	}
	return sshPublicKeys, nil
}
