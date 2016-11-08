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
	"fmt"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/upup/pkg/kutil"
	"k8s.io/kubernetes/pkg/util/sets"
	"strings"
)

type CreateClusterOptions struct {
	Yes               bool
	Target            string
	Models            string
	Cloud             string
	Zones             string
	MasterZones       string
	NodeSize          string
	MasterSize        string
	NodeCount         int
	Project           string
	KubernetesVersion string
	OutDir            string
	Image             string
	SSHPublicKey      string
	VPCID             string
	NetworkCIDR       string
	DNSZone           string
	AdminAccess       string
	Networking        string
	AssociatePublicIP bool

	// Channel is the location of the api.Channel to use for our defaults
	Channel           string

	// The network topology to use
	Topology          string
}

func NewCmdCreateCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateClusterOptions{}

	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Create cluster",
		Long:  `Creates a k8s cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunCreateCluster(f, cmd, args, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().BoolVar(&options.Yes, "yes", false, "Specify --yes to immediately create the cluster")
	cmd.Flags().StringVar(&options.Target, "target", cloudup.TargetDirect, "Target - direct, terraform")
	cmd.Flags().StringVar(&options.Models, "model", "config,proto,cloudup", "Models to apply (separate multiple models with commas)")

	cmd.Flags().StringVar(&options.Cloud, "cloud", "", "Cloud provider to use - gce, aws")

	cmd.Flags().StringVar(&options.Zones, "zones", "", "Zones in which to run the cluster")
	cmd.Flags().StringVar(&options.MasterZones, "master-zones", "", "Zones in which to run masters (must be an odd number)")

	cmd.Flags().StringVar(&options.Project, "project", "", "Project to use (must be set on GCE)")
	cmd.Flags().StringVar(&options.KubernetesVersion, "kubernetes-version", "", "Version of kubernetes to run (defaults to version in channel)")

	cmd.Flags().StringVar(&options.SSHPublicKey, "ssh-public-key", "~/.ssh/id_rsa.pub", "SSH public key to use")

	cmd.Flags().StringVar(&options.NodeSize, "node-size", "", "Set instance size for nodes")

	cmd.Flags().StringVar(&options.MasterSize, "master-size", "", "Set instance size for masters")

	cmd.Flags().StringVar(&options.VPCID, "vpc", "", "Set to use a shared VPC")
	cmd.Flags().StringVar(&options.NetworkCIDR, "network-cidr", "", "Set to override the default network CIDR")

	cmd.Flags().IntVar(&options.NodeCount, "node-count", 0, "Set the number of nodes")

	cmd.Flags().StringVar(&options.Image, "image", "", "Image to use")

	cmd.Flags().StringVar(&options.Networking, "networking", "kubenet", "Networking mode to use.  kubenet (default), classic, external, cni.")

	cmd.Flags().StringVar(&options.DNSZone, "dns-zone", "", "DNS hosted zone to use (defaults to longest matching zone)")
	cmd.Flags().StringVar(&options.OutDir, "out", "", "Path to write any local output")
	cmd.Flags().StringVar(&options.AdminAccess, "admin-access", "", "Restrict access to admin endpoints (SSH, HTTPS) to this CIDR.  If not set, access will not be restricted by IP.")

	cmd.Flags().BoolVar(&options.AssociatePublicIP, "associate-public-ip", true, "Specify --associate-public-ip=[true|false] to enable/disable association of public IP for master ASG and nodes. Default is 'true'.")

	cmd.Flags().StringVar(&options.Channel, "channel", api.DefaultChannel, "Channel for default versions and configuration to use")

	// Network topology
	cmd.Flags().StringVarP(&options.Topology, "topology", "t", "public", "Controls network topology for the cluster. public|private. Default is 'public'.")

	return cmd
}

func RunCreateCluster(f *util.Factory, cmd *cobra.Command, args []string, out io.Writer, c *CreateClusterOptions) error {
	err := rootCommand.ProcessArgs(args)
	if err != nil {
		return err
	}

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

	clusterName := rootCommand.clusterName
	if clusterName == "" {
		return fmt.Errorf("--name is required")
	}

	// TODO: Reuse rootCommand stateStore logic?

	if c.OutDir == "" {
		if c.Target == cloudup.TargetTerraform {
			c.OutDir = "out/terraform"
		} else {
			c.OutDir = "out"
		}
	}

	clientset, err := rootCommand.Clientset()
	if err != nil {
		return err
	}

	cluster, err := clientset.Clusters().Get(clusterName)
	if err != nil {
		return err
	}

	if cluster != nil {
		return fmt.Errorf("cluster %q already exists; use 'kops update cluster' to apply changes", clusterName)
	}

	cluster = &api.Cluster{}

	channel, err := api.LoadChannel(c.Channel)
	if err != nil {
		return err
	}
	if channel.Spec.Cluster != nil {
		cluster.Spec = *channel.Spec.Cluster
	}
	cluster.Spec.Channel = c.Channel

	configBase, err := clientset.Clusters().(*vfsclientset.ClusterVFS).ConfigBase(clusterName)
	if err != nil {
		return fmt.Errorf("error building ConfigBase for cluster: %v", err)
	}
	cluster.Spec.ConfigBase = configBase.Path()

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
	default:
		return fmt.Errorf("unknown networking mode %q", c.Networking)
	}
	if c.Zones != "" {
		existingZones := make(map[string]*api.ClusterZoneSpec)
		for _, zone := range cluster.Spec.Zones {
			existingZones[zone.Name] = zone
		}

		for _, zone := range parseZoneList(c.Zones) {
			if existingZones[zone] == nil {
				cluster.Spec.Zones = append(cluster.Spec.Zones, &api.ClusterZoneSpec{
					Name: zone,
				})
			}
		}
	}

	if len(cluster.Spec.Zones) == 0 {
		return fmt.Errorf("must specify at least one zone for the cluster (use --zones)")
	}

	var masters []*api.InstanceGroup
	var nodes []*api.InstanceGroup
	var instanceGroups []*api.InstanceGroup

	if c.MasterZones == "" {
		if len(masters) == 0 {
			// We default to single-master (not HA), unless the user explicitly specifies it
			// HA master is a little slower, not as well tested yet, and requires more resources
			// Probably best not to make it the silent default!
			for _, zone := range cluster.Spec.Zones {
				g := &api.InstanceGroup{}
				g.Spec.Role = api.InstanceGroupRoleMaster
				g.Spec.Zones = []string{zone.Name}
				g.Spec.MinSize = fi.Int(1)
				g.Spec.MaxSize = fi.Int(1)
				g.Name = "master-" + zone.Name // Subsequent masters (if we support that) could be <zone>-1, <zone>-2
				instanceGroups = append(instanceGroups, g)
				masters = append(masters, g)

				// Don't force HA master
				break
			}
		}
	} else {
		if len(masters) == 0 {
			// Use the specified master zones (this is how the user gets HA master)
			for _, zone := range parseZoneList(c.MasterZones) {
				g := &api.InstanceGroup{}
				g.Spec.Role = api.InstanceGroupRoleMaster
				g.Spec.Zones = []string{zone}
				g.Spec.MinSize = fi.Int(1)
				g.Spec.MaxSize = fi.Int(1)
				g.Name = "master-" + zone
				instanceGroups = append(instanceGroups, g)
				masters = append(masters, g)
			}
		} else {
			// This is hard, because of the etcd cluster
			return fmt.Errorf("Cannot change master-zones from the CLI")
		}
	}

	if len(cluster.Spec.EtcdClusters) == 0 {
		zones := sets.NewString()
		for _, group := range masters {
			for _, zone := range group.Spec.Zones {
				zones.Insert(zone)
			}
		}
		etcdZones := zones.List()

		for _, etcdCluster := range cloudup.EtcdClusters {
			etcd := &api.EtcdClusterSpec{}
			etcd.Name = etcdCluster
			for _, zone := range etcdZones {
				m := &api.EtcdMemberSpec{}
				m.Name = zone
				m.Zone = fi.String(zone)
				etcd.Members = append(etcd.Members, m)
			}
			cluster.Spec.EtcdClusters = append(cluster.Spec.EtcdClusters, etcd)
		}
	}

	if len(nodes) == 0 {
		g := &api.InstanceGroup{}
		g.Spec.Role = api.InstanceGroupRoleNode
		g.Name = "nodes"
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

	for _, group := range instanceGroups {
		group.Spec.AssociatePublicIP = fi.Bool(c.AssociatePublicIP)
	}

	if c.NodeCount != 0 {
		for _, group := range nodes {
			group.Spec.MinSize = fi.Int(c.NodeCount)
			group.Spec.MaxSize = fi.Int(c.NodeCount)
		}
	}

	if c.MasterSize != "" {
		for _, group := range masters {
			group.Spec.MachineType = c.MasterSize
		}
	}

	if c.DNSZone != "" {
		cluster.Spec.DNSZone = c.DNSZone
	}

	if c.Cloud != "" {
		cluster.Spec.CloudProvider = c.Cloud
	}

	if c.Project != "" {
		cluster.Spec.Project = c.Project
	}

	if clusterName != "" {
		cluster.Name = clusterName
	}

	if c.KubernetesVersion != "" {
		cluster.Spec.KubernetesVersion = c.KubernetesVersion
	}

	if c.VPCID != "" {
		cluster.Spec.NetworkID = c.VPCID
	}

	if c.NetworkCIDR != "" {
		cluster.Spec.NetworkCIDR = c.NetworkCIDR
	}

	if cluster.SharedVPC() && cluster.Spec.NetworkCIDR == "" {
		return fmt.Errorf("Must specify NetworkCIDR when VPC is set")
	}

	if cluster.Spec.CloudProvider == "" {
		for _, zone := range cluster.Spec.Zones {
			cloud, known := fi.GuessCloudForZone(zone.Name)
			if known {
				glog.Infof("Inferred --cloud=%s from zone %q", cloud, zone.Name)
				cluster.Spec.CloudProvider = string(cloud)
				break
			}
		}
		if cluster.Spec.CloudProvider == "" {
			return fmt.Errorf("unable to infer CloudProvider from Zones (is there a typo in --zones?)")
		}
	}


	// Network Topology
	switch  c.Topology{
	case api.TopologyPublic:
		cluster.Spec.Topology = &api.TopologySpec{Masters: api.TopologyPublic, Nodes: api.TopologyPublic, BypassBastion: false}
	case api.TopologyPrivate:
		if c.Networking != "cni" {
			return fmt.Errorf("Invalid networking option %s. Currently only '--networking cni' is supported for private topologies", c.Networking)
		}
		cluster.Spec.Topology = &api.TopologySpec{Masters: api.TopologyPrivate, Nodes: api.TopologyPrivate, BypassBastion: false}
	case api.TopologyPrivateMasters:
		cluster.Spec.Topology = &api.TopologySpec{Masters: api.TopologyPrivate, Nodes: api.TopologyPublic, BypassBastion: false}
	case "":
		glog.Warningf("Empty topology. Defaulting to public topology.")
		cluster.Spec.Topology = &api.TopologySpec{Masters: api.TopologyPublic, Nodes: api.TopologyPublic, BypassBastion: false}
	default:
		return fmt.Errorf("Invalid topology %s.", c.Topology)
	}

	sshPublicKeys := make(map[string][]byte)
	if c.SSHPublicKey != "" {
		c.SSHPublicKey = utils.ExpandPath(c.SSHPublicKey)
		authorized, err := ioutil.ReadFile(c.SSHPublicKey)
		if err != nil {
			return fmt.Errorf("error reading SSH key file %q: %v", c.SSHPublicKey, err)
		}
		sshPublicKeys[fi.SecretNameSSHPrimary] = authorized
	}

	if c.AdminAccess != "" {
		cluster.Spec.AdminAccess = []string{c.AdminAccess}
	}

	err = cluster.PerformAssignments()
	if err != nil {
		return fmt.Errorf("error populating configuration: %v", err)
	}
	err = api.PerformAssignmentsInstanceGroups(instanceGroups)
	if err != nil {
		return fmt.Errorf("error populating configuration: %v", err)
	}

	strict := false
	err = api.DeepValidate(cluster, instanceGroups, strict)
	if err != nil {
		return err
	}

	fullCluster, err := cloudup.PopulateClusterSpec(cluster)
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

	err = api.DeepValidate(fullCluster, fullInstanceGroups, true)
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

	secretStore, err := registry.SecretStore(cluster)
	if err != nil {
		return err
	}

	err = registry.WriteConfig(configBase.Join(registry.PathClusterCompleted), fullCluster)
	if err != nil {
		return fmt.Errorf("error writing completed cluster spec: %v", err)
	}

	for k, data := range sshPublicKeys {
		err = keyStore.AddSSHPublicKey(k, data)
		if err != nil {
			return fmt.Errorf("error addding SSH public key: %v", err)
		}
	}

	if isDryrun {
		fmt.Print("Previewing changes that will be made:\n\n")
	}

	applyCmd := &cloudup.ApplyClusterCmd{
		Cluster:    fullCluster,
		Models:     strings.Split(c.Models, ","),
		Clientset:  clientset,
		TargetName: targetName,
		OutDir:     c.OutDir,
		DryRun:     isDryrun,
	}

	err = applyCmd.Run()
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
			fmt.Fprintf(&sb, " * edit your node instance group: kops edit ig --name=%s %s\n", clusterName, nodes[0].Name)
		}
		if len(masters) > 0 {
			fmt.Fprintf(&sb, " * edit your master instance group: kops edit ig --name=%s %s\n", clusterName, masters[0].Name)
		}
		fmt.Fprintf(&sb, "\n")
		fmt.Fprintf(&sb, "Finally configure your cluster with: kops update cluster %s --yes\n", clusterName)
		fmt.Fprintf(&sb, "\n")

		_, err := out.Write(sb.Bytes())
		if err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
	} else {
		glog.Infof("Exporting kubecfg for cluster")

		x := &kutil.CreateKubecfg{
			ContextName:  cluster.Name,
			KeyStore:     keyStore,
			SecretStore:  secretStore,
			KubeMasterIP: cluster.Spec.MasterPublicName,
		}

		err = x.WriteKubecfg()
		if err != nil {
			return err
		}
	}

	return nil
}

func parseZoneList(s string) []string {
	var filtered []string
	for _, v := range strings.Split(s, ",") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		v = strings.ToLower(v)
		filtered = append(filtered, v)
	}
	return filtered
}