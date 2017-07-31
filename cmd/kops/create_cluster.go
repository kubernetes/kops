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

package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
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
	Project              string
	KubernetesVersion    string
	OutDir               string
	Image                string
	SSHPublicKey         string
	VPCID                string
	NetworkCIDR          string
	DNSZone              string
	AdminAccess          []string
	SSHAccess            []string
	Networking           string
	NodeSecurityGroups   []string
	MasterSecurityGroups []string
	AssociatePublicIP    *bool

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

	// vSphere options
	VSphereServer        string
	VSphereDatacenter    string
	VSphereResourcePool  string
	VSphereCoreDNSServer string
	// Note: We need open-vm-tools to be installed for vSphere Cloud Provider to work
	// We need VSphereDatastore to support Kubernetes vSphere Cloud Provider (v1.5.3)
	// We can remove this once we support higher versions.
	VSphereDatastore string
}

func (o *CreateClusterOptions) InitDefaults() {
	o.Yes = false
	o.Target = cloudup.TargetDirect
	o.Models = strings.Join(cloudup.CloudupModels, ",")
	o.SSHPublicKey = "~/.ssh/id_rsa.pub"
	o.Networking = "kubenet"
	o.Channel = api.DefaultChannel
	o.Topology = api.TopologyPublic
	o.DNSType = string(api.DNSTypePublic)
	o.Bastion = false

	// Default to open API & SSH access
	o.AdminAccess = []string{"0.0.0.0/0"}

	o.Authorization = AuthorizationFlagAlwaysAllow
}

var (
	create_cluster_long = templates.LongDesc(i18n.T(`
	Create a kubernetes cluster using command line flags.
	This command creates cloud based resources such as networks and virtual machines. Once
	the infrastructure is in place Kubernetes is installed on the virtual machines.

	These operations are done in parallel and rely on eventual consistency.
	`))

	create_cluster_example = templates.Examples(i18n.T(`
	# Create a cluster in AWS
	kops create cluster --name=kubernetes-cluster.example.com \
	--state=s3://kops-state-1234 --zones=eu-west-1a \
	--node-count=2

	# Create a cluster in AWS that has HA masters.  This cluster
	# will be setup with an internal networking in a private VPC.
	# A bastion instance will be setup to provide instance access.

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

	# Create cluster in GCE.
	# This is an alpha feature.
        export KOPS_STATE_STORE="gs://mybucket-kops"
        export ZONES=${MASTER_ZONES:-"us-east1-b,us-east1-c,us-east1-d"}
        export KOPS_FEATURE_FLAGS=AlphaAllowGCE

        kops create cluster kubernetes-k8s-gce.example.com
          --zones $ZONES \
          --master-zones $ZONES \
          --node-count 3
          --project my-gce-project \
          --image "ubuntu-os-cloud/ubuntu-1604-xenial-v20170202" \
          --yes
	`))

	create_cluster_short = i18n.T("Create a Kubernetes cluster.")
)

func NewCmdCreateCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateClusterOptions{}
	options.InitDefaults()

	associatePublicIP := false

	cmd := &cobra.Command{
		Use:     "cluster",
		Short:   create_cluster_short,
		Long:    create_cluster_long,
		Example: create_cluster_example,
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

			err = RunCreateCluster(f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().BoolVar(&options.Yes, "yes", options.Yes, "Specify --yes to immediately create the cluster")
	cmd.Flags().StringVar(&options.Target, "target", options.Target, "Target - direct, terraform, cloudformation")
	cmd.Flags().StringVar(&options.Models, "model", options.Models, "Models to apply (separate multiple models with commas)")

	cmd.Flags().StringVar(&options.Cloud, "cloud", options.Cloud, "Cloud provider to use - gce, aws, vsphere")

	cmd.Flags().StringSliceVar(&options.Zones, "zones", options.Zones, "Zones in which to run the cluster")
	cmd.Flags().StringSliceVar(&options.MasterZones, "master-zones", options.MasterZones, "Zones in which to run masters (must be an odd number)")

	cmd.Flags().StringVar(&options.Project, "project", options.Project, "Project to use (must be set on GCE)")
	cmd.Flags().StringVar(&options.KubernetesVersion, "kubernetes-version", options.KubernetesVersion, "Version of kubernetes to run (defaults to version in channel)")

	cmd.Flags().StringVar(&options.SSHPublicKey, "ssh-public-key", options.SSHPublicKey, "SSH public key to use")

	cmd.Flags().StringVar(&options.NodeSize, "node-size", options.NodeSize, "Set instance size for nodes")

	cmd.Flags().StringVar(&options.MasterSize, "master-size", options.MasterSize, "Set instance size for masters")

	cmd.Flags().Int32Var(&options.MasterVolumeSize, "master-volume-size", options.MasterVolumeSize, "Set instance volume size (in GB) for masters")
	cmd.Flags().Int32Var(&options.NodeVolumeSize, "node-volume-size", options.NodeVolumeSize, "Set instance volume size (in GB) for nodes")

	cmd.Flags().StringVar(&options.VPCID, "vpc", options.VPCID, "Set to use a shared VPC")
	cmd.Flags().StringVar(&options.NetworkCIDR, "network-cidr", options.NetworkCIDR, "Set to override the default network CIDR")

	cmd.Flags().Int32Var(&options.MasterCount, "master-count", options.MasterCount, "Set the number of masters.  Defaults to one master per master-zone")
	cmd.Flags().Int32Var(&options.NodeCount, "node-count", options.NodeCount, "Set the number of nodes")
	cmd.Flags().BoolVar(&options.EncryptEtcdStorage, "encrypt-etcd-storage", options.EncryptEtcdStorage, "Generate key in aws kms and use it for encrypt etcd volumes")

	cmd.Flags().StringVar(&options.Image, "image", options.Image, "Image to use for all instances.")

	cmd.Flags().StringVar(&options.Networking, "networking", "kubenet", "Networking mode to use.  kubenet (default), classic, external, kopeio-vxlan (or kopeio), weave, flannel, calico, canal, kube-router.")

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
	cmd.Flags().StringVarP(&options.Topology, "topology", "t", options.Topology, "Controls network topology for the cluster. public|private. Default is 'public'.")

	// Authorization
	cmd.Flags().StringVar(&options.Authorization, "authorization", options.Authorization, "Authorization mode to use: "+AuthorizationFlagAlwaysAllow+" or "+AuthorizationFlagRBAC)

	// DNS
	cmd.Flags().StringVar(&options.DNSType, "dns", options.DNSType, "DNS hosted zone to use: public|private. Default is 'public'.")

	// Bastion
	cmd.Flags().BoolVar(&options.Bastion, "bastion", options.Bastion, "Pass the --bastion flag to enable a bastion instance group. Only applies to private topology.")

	// Allow custom tags from the CLI
	cmd.Flags().StringVar(&options.CloudLabels, "cloud-labels", options.CloudLabels, "A list of KV pairs used to tag all instance groups in AWS (eg \"Owner=John Doe,Team=Some Team\").")

	// Master and Node Tenancy
	cmd.Flags().StringVar(&options.MasterTenancy, "master-tenancy", options.MasterTenancy, "The tenancy of the master group on AWS. Can either be default or dedicated.")
	cmd.Flags().StringVar(&options.NodeTenancy, "node-tenancy", options.NodeTenancy, "The tenancy of the node group on AWS. Can be either default or dedicated.")

	cmd.Flags().StringVar(&options.APILoadBalancerType, "api-loadbalancer-type", options.APILoadBalancerType, "Sets the API loadbalancer type to either 'public' or 'internal'")

	if featureflag.VSphereCloudProvider.Enabled() {
		// vSphere flags
		cmd.Flags().StringVar(&options.VSphereServer, "vsphere-server", options.VSphereServer, "vsphere-server is required for vSphere. Set vCenter URL Ex: 10.192.10.30 or myvcenter.io (without https://)")
		cmd.Flags().StringVar(&options.VSphereDatacenter, "vsphere-datacenter", options.VSphereDatacenter, "vsphere-datacenter is required for vSphere. Set the name of the datacenter in which to deploy Kubernetes VMs.")
		cmd.Flags().StringVar(&options.VSphereResourcePool, "vsphere-resource-pool", options.VSphereDatacenter, "vsphere-resource-pool is required for vSphere. Set a valid Cluster, Host or Resource Pool in which to deploy Kubernetes VMs.")
		cmd.Flags().StringVar(&options.VSphereCoreDNSServer, "vsphere-coredns-server", options.VSphereCoreDNSServer, "vsphere-coredns-server is required for vSphere.")
		cmd.Flags().StringVar(&options.VSphereDatastore, "vsphere-datastore", options.VSphereDatastore, "vsphere-datastore is required for vSphere.  Set a valid datastore in which to store dynamic provision volumes.")
	}
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

	configBase, err := clientset.ConfigBaseFor(cluster)
	if err != nil {
		return fmt.Errorf("error building ConfigBase for cluster: %v", err)
	}
	cluster.Spec.ConfigBase = configBase.Path()

	glog.V(4).Infof("networking mode=%s => %s", c.Networking, fi.DebugAsJsonString(cluster.Spec.Networking))

	// In future we could change the default if the flag is not specified, e.g. in 1.7 maybe the default is RBAC?
	cluster.Spec.Authorization = &api.AuthorizationSpec{}
	if strings.EqualFold(c.Authorization, AuthorizationFlagAlwaysAllow) {
		cluster.Spec.Authorization.AlwaysAllow = &api.AlwaysAllowAuthorizationSpec{}
	} else if strings.EqualFold(c.Authorization, AuthorizationFlagRBAC) {
		cluster.Spec.Authorization.RBAC = &api.RBACAuthorizationSpec{}
	} else {
		return fmt.Errorf("unknown authorization mode %q", c.Authorization)
	}

	if len(c.Zones) != 0 {
		existingSubnets := make(map[string]*api.ClusterSubnetSpec)
		for i := range cluster.Spec.Subnets {
			subnet := &cluster.Spec.Subnets[i]
			existingSubnets[subnet.Name] = subnet
		}
		for _, zoneName := range c.Zones {
			// We create default subnets named the same as the zones
			subnetName := zoneName
			if existingSubnets[subnetName] == nil {
				cluster.Spec.Subnets = append(cluster.Spec.Subnets, api.ClusterSubnetSpec{
					Name:   subnetName,
					Zone:   subnetName,
					Egress: c.Egress,
				})
			}
		}
	}

	if len(cluster.Spec.Subnets) == 0 {
		return fmt.Errorf("must specify at least one zone for the cluster (use --zones)")
	}

	var masters []*api.InstanceGroup
	var nodes []*api.InstanceGroup
	var instanceGroups []*api.InstanceGroup
	cloudLabels, err := parseCloudLabels(c.CloudLabels)
	if err != nil {
		return fmt.Errorf("error parsing global cloud labels: %v", err)
	}
	cluster.Spec.CloudLabels = cloudLabels

	// Build the master subnets
	// The master zones is the default set of zones unless explicitly set
	// The master count is the number of master zones unless explicitly set
	// We then round-robin around the zones
	if len(masters) == 0 {
		var masterSubnets []*api.ClusterSubnetSpec
		masterCount := c.MasterCount
		if len(c.MasterZones) != 0 {
			for _, subnetName := range c.MasterZones {
				subnet := findSubnet(cluster, subnetName)
				if subnet == nil {
					// Should have been caught already
					return fmt.Errorf("master-zone %q not included in zones", subnetName)
				}
				masterSubnets = append(masterSubnets, subnet)
			}

			if c.MasterCount != 0 && c.MasterCount < int32(len(masterSubnets)) {
				return fmt.Errorf("specified %d master zones, but also requested %d masters.  If specifying both, the count should match.", len(masterSubnets), c.MasterCount)
			}

			if masterCount == 0 {
				// If master count is not specified, default to the number of master zones
				masterCount = int32(len(c.MasterZones))
			}
		} else {
			for i := range cluster.Spec.Subnets {
				masterSubnets = append(masterSubnets, &cluster.Spec.Subnets[i])
			}

			if masterCount == 0 {
				// If master count is not specified, default to 1
				masterCount = 1
			}
		}

		if len(masterSubnets) == 0 {
			// Should be unreachable
			return fmt.Errorf("cannot determine master subnets")
		}

		for i := 0; i < int(masterCount); i++ {
			subnet := masterSubnets[i%len(masterSubnets)]
			name := subnet.Name
			if int(masterCount) > len(masterSubnets) {
				name += "-" + strconv.Itoa(1+(i/len(masterSubnets)))
			}

			g := &api.InstanceGroup{}
			g.Spec.Role = api.InstanceGroupRoleMaster
			g.Spec.Subnets = []string{subnet.Name}
			g.Spec.MinSize = fi.Int32(1)
			g.Spec.MaxSize = fi.Int32(1)
			g.ObjectMeta.Name = "master-" + name
			instanceGroups = append(instanceGroups, g)
			masters = append(masters, g)
		}
	}

	if len(cluster.Spec.EtcdClusters) == 0 {
		masterAZs := sets.NewString()
		duplicateAZs := false
		for _, ig := range masters {
			if len(ig.Spec.Subnets) != 1 {
				return fmt.Errorf("unexpected subnets for master instance group %q (expected exactly only, found %d)", ig.ObjectMeta.Name, len(ig.Spec.Subnets))
			}
			for _, subnetName := range ig.Spec.Subnets {
				subnet := findSubnet(cluster, subnetName)
				if subnet == nil {
					return fmt.Errorf("cannot find subnet %q (declared in instance group %q, not found in cluster)", subnetName, ig.ObjectMeta.Name)
				}

				if masterAZs.Has(subnet.Zone) {
					duplicateAZs = true
				}

				masterAZs.Insert(subnet.Zone)
			}
		}

		if duplicateAZs {
			glog.Warningf("Running with masters in the same AZs; redundancy will be reduced")
		}

		for _, etcdCluster := range cloudup.EtcdClusters {
			etcd := &api.EtcdClusterSpec{}
			etcd.Name = etcdCluster

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
		cluster.Spec.CloudProvider = c.Cloud

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
	}

	if c.Project != "" {
		cluster.Spec.Project = c.Project
	}

	if c.KubernetesVersion != "" {
		cluster.Spec.KubernetesVersion = c.KubernetesVersion
	}

	if cluster.Spec.CloudProvider == "" {
		for _, subnet := range cluster.Spec.Subnets {
			cloud, known := fi.GuessCloudForZone(subnet.Zone)
			if known {
				glog.Infof("Inferred --cloud=%s from zone %q", cloud, subnet.Zone)
				cluster.Spec.CloudProvider = string(cloud)
				break
			}
		}
		if cluster.Spec.CloudProvider == "" {
			return fmt.Errorf("unable to infer CloudProvider from Zones (is there a typo in --zones?)")
		}
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
	case "flannel":
		cluster.Spec.Networking.Flannel = &api.FlannelNetworkingSpec{}
	case "calico":
		cluster.Spec.Networking.Calico = &api.CalicoNetworkingSpec{}
	case "canal":
		cluster.Spec.Networking.Canal = &api.CanalNetworkingSpec{}
	case "kube-router":
		cluster.Spec.Networking.Kuberouter = &api.KuberouterNetworkingSpec{}
	default:
		return fmt.Errorf("unknown networking mode %q", c.Networking)
	}

	if c.VPCID != "" {
		cluster.Spec.NetworkID = c.VPCID
	}

	if c.NetworkCIDR != "" {
		cluster.Spec.NetworkCIDR = c.NetworkCIDR
	}

	// Network Topology
	if c.Topology == "" {
		// The flag default should have set this, but we might be being called as a library
		glog.Infof("Empty topology. Defaulting to public topology")
		c.Topology = api.TopologyPublic
	}

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
			return fmt.Errorf("Invalid networking option %s. Currently only '--networking kopeio-vxlan (or kopeio)', '--networking weave', '--networking flannel', '--networking calico', '--networking canal', '--networking kube-router' are supported for private topologies", c.Networking)
		}
		cluster.Spec.Topology = &api.TopologySpec{
			Masters: api.TopologyPrivate,
			Nodes:   api.TopologyPrivate,
		}

		for i := range cluster.Spec.Subnets {
			cluster.Spec.Subnets[i].Type = api.SubnetTypePrivate
		}

		var utilitySubnets []api.ClusterSubnetSpec
		for _, s := range cluster.Spec.Subnets {
			if s.Type == api.SubnetTypeUtility {
				continue
			}
			subnet := api.ClusterSubnetSpec{
				Name: "utility-" + s.Name,
				Zone: s.Zone,
				Type: api.SubnetTypeUtility,
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

			cluster.Spec.Topology.Bastion = &api.BastionSpec{
				BastionPublicName: "bastion." + clusterName,
			}

		}

	default:
		return fmt.Errorf("Invalid topology %s.", c.Topology)
	}

	// DNS
	if c.DNSType == "" {
		// The flag default should have set this, but we might be being called as a library
		glog.Infof("Empty DNS. Defaulting to public DNS")
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

	// Populate the API access, so that it can be discoverable
	// TODO: This is the same code as in defaults - try to dedup?
	if cluster.Spec.API == nil {
		cluster.Spec.API = &api.AccessSpec{}
	}
	if cluster.Spec.API.IsEmpty() {
		if c.APILoadBalancerType != "" {
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

	sshPublicKeys := make(map[string][]byte)
	if c.SSHPublicKey != "" {
		c.SSHPublicKey = utils.ExpandPath(c.SSHPublicKey)
		authorized, err := ioutil.ReadFile(c.SSHPublicKey)
		if err != nil {
			return fmt.Errorf("error reading SSH key file %q: %v", c.SSHPublicKey, err)
		}
		sshPublicKeys[fi.SecretNameSSHPrimary] = authorized

		glog.Infof("Using SSH public key: %v\n", c.SSHPublicKey)
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

	err = cloudup.PerformAssignments(cluster)
	if err != nil {
		return fmt.Errorf("error populating configuration: %v", err)
	}
	err = api.PerformAssignmentsInstanceGroups(instanceGroups)
	if err != nil {
		return fmt.Errorf("error populating configuration: %v", err)
	}

	strict := false
	err = validation.DeepValidate(cluster, instanceGroups, strict)
	if err != nil {
		return err
	}

	assetBuilder := assets.NewAssetBuilder()
	fullCluster, err := cloudup.PopulateClusterSpec(cluster, assetBuilder)
	if err != nil {
		return err
	}

	var fullInstanceGroups []*api.InstanceGroup
	for _, group := range instanceGroups {
		fullGroup, err := cloudup.PopulateInstanceGroupSpec(fullCluster, group, channel)
		if err != nil {
			return err
		}
		fullInstanceGroups = append(fullInstanceGroups, fullGroup)
	}

	err = validation.DeepValidate(fullCluster, fullInstanceGroups, true)
	if err != nil {
		return err
	}

	// Note we perform as much validation as we can, before writing a bad config
	err = registry.CreateClusterConfig(clientset, cluster, fullInstanceGroups)
	if err != nil {
		return fmt.Errorf("error writing updated configuration: %v", err)
	}

	keyStore, err := registry.KeyStore(cluster)
	if err != nil {
		return err
	}

	err = registry.WriteConfigDeprecated(configBase.Join(registry.PathClusterCompleted), fullCluster)
	if err != nil {
		return fmt.Errorf("error writing completed cluster spec: %v", err)
	}

	for k, data := range sshPublicKeys {
		err = keyStore.AddSSHPublicKey(k, data)
		if err != nil {
			return fmt.Errorf("error addding SSH public key: %v", err)
		}
	}

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

		err := RunUpdateCluster(f, clusterName, out, updateClusterOptions)
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
			fmt.Fprintf(&sb, "Finally configure your cluster with: kops update cluster %s --yes\n", clusterName)
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

	if n.CNI != nil || n.Kopeio != nil || n.Weave != nil || n.Flannel != nil || n.Calico != nil || n.Canal != nil || n.Kuberouter != nil {
		return true
	}
	return false
}

func findSubnet(c *api.Cluster, subnetName string) *api.ClusterSubnetSpec {
	for i := range c.Spec.Subnets {
		if c.Spec.Subnets[i].Name == subnetName {
			return &c.Spec.Subnets[i]
		}
	}
	return nil
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
