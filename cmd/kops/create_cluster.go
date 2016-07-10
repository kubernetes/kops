package main

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kubernetes/pkg/util/sets"
	"os"
	"os/exec"
	"path"
	"strings"
)

type CreateClusterCmd struct {
	DryRun            bool
	Target            string
	ModelsBaseDir     string
	Models            string
	NodeModel         string
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
}

var createCluster CreateClusterCmd

func init() {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Create cluster",
		Long:  `Creates a k8s cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := createCluster.Run()
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	createCmd.AddCommand(cmd)

	executableLocation, err := exec.LookPath(os.Args[0])
	if err != nil {
		glog.Fatalf("Cannot determine location of kops tool: %q.  Please report this problem!", os.Args[0])
	}

	modelsBaseDirDefault := path.Join(path.Dir(executableLocation), "models")

	cmd.Flags().BoolVar(&createCluster.DryRun, "dryrun", false, "Don't create cloud resources; just show what would be done")
	cmd.Flags().StringVar(&createCluster.Target, "target", "direct", "Target - direct, terraform")
	//configFile := cmd.Flags().StringVar(&createCluster., "conf", "", "Configuration file to load")
	cmd.Flags().StringVar(&createCluster.ModelsBaseDir, "modeldir", modelsBaseDirDefault, "Source directory where models are stored")
	cmd.Flags().StringVar(&createCluster.Models, "model", "config,proto,cloudup", "Models to apply (separate multiple models with commas)")
	cmd.Flags().StringVar(&createCluster.NodeModel, "nodemodel", "nodeup", "Model to use for node configuration")

	cmd.Flags().StringVar(&createCluster.Cloud, "cloud", "", "Cloud provider to use - gce, aws")

	cmd.Flags().StringVar(&createCluster.Zones, "zones", "", "Zones in which to run the cluster")
	cmd.Flags().StringVar(&createCluster.MasterZones, "master-zones", "", "Zones in which to run masters (must be an odd number)")

	cmd.Flags().StringVar(&createCluster.Project, "project", "", "Project to use (must be set on GCE)")
	//cmd.Flags().StringVar(&createCluster.Name, "name", "", "Name for cluster")
	cmd.Flags().StringVar(&createCluster.KubernetesVersion, "kubernetes-version", "", "Version of kubernetes to run (defaults to latest)")

	cmd.Flags().StringVar(&createCluster.SSHPublicKey, "ssh-public-key", "~/.ssh/id_rsa.pub", "SSH public key to use")

	cmd.Flags().StringVar(&createCluster.NodeSize, "node-size", "", "Set instance size for nodes")

	cmd.Flags().StringVar(&createCluster.MasterSize, "master-size", "", "Set instance size for masters")

	cmd.Flags().StringVar(&createCluster.VPCID, "vpc", "", "Set to use a shared VPC")
	cmd.Flags().StringVar(&createCluster.NetworkCIDR, "network-cidr", "", "Set to override the default network CIDR")

	cmd.Flags().IntVar(&createCluster.NodeCount, "node-count", 0, "Set the number of nodes")

	cmd.Flags().StringVar(&createCluster.Image, "image", "", "Image to use")

	cmd.Flags().StringVar(&createCluster.DNSZone, "dns-zone", "", "DNS hosted zone to use (defaults to last two components of cluster name)")
	cmd.Flags().StringVar(&createCluster.OutDir, "out", "", "Path to write any local output")
}

var EtcdClusters = []string{"main", "events"}

func (c *CreateClusterCmd) Run() error {
	isDryrun := false
	if c.DryRun {
		isDryrun = true
		c.Target = "dryrun"
	}

	clusterName := rootCommand.clusterName
	if clusterName == "" {
		return fmt.Errorf("--name is required")
	}

	// TODO: Reuse rootCommand stateStore logic?

	if c.OutDir == "" {
		c.OutDir = "out"
	}

	clusterRegistry, err := rootCommand.ClusterRegistry()
	if err != nil {
		return err
	}

	cluster, err := clusterRegistry.Find(clusterName)
	if err != nil {
		return err
	}
	instanceGroupRegistry, err := clusterRegistry.InstanceGroups(clusterName)
	if err != nil {
		return err
	}

	var instanceGroups []*api.InstanceGroup

	// TODO: Rationalize this!
	creatingNewCluster := false

	if cluster == nil {
		cluster = &api.Cluster{}
		creatingNewCluster = true
	} else {
		groupNames, err := instanceGroupRegistry.List()
		if err != nil {
			return err
		}

		for _, groupName := range groupNames {
			instanceGroup, err := instanceGroupRegistry.Find(groupName)
			if err != nil {
				return err
			}
			if instanceGroup == nil {
				return fmt.Errorf("InstanceGroup %q was listed, but then could not be found", groupName)
			}
			instanceGroups = append(instanceGroups, instanceGroup)
		}
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

	for _, group := range instanceGroups {
		if group.IsMaster() {
			masters = append(masters, group)
		} else {
			nodes = append(nodes, group)
		}
	}

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
			glog.Errorf("Cannot change master-zones from the CLI")
			os.Exit(1)
		}
	}

	if len(cluster.Spec.EtcdClusters) == 0 {
		zones := sets.NewString()
		for _, group := range instanceGroups {
			for _, zone := range group.Spec.Zones {
				zones.Insert(zone)
			}
		}
		etcdZones := zones.List()
		if (len(etcdZones) % 2) == 0 {
			// Not technically a requirement, but doesn't really make sense to allow
			glog.Errorf("There should be an odd number of master-zones, for etcd's quorum.  Hint: Use --zones and --master-zones to declare node zones and master zones separately.")
			os.Exit(1)
		}

		for _, etcdCluster := range EtcdClusters {
			etcd := &api.EtcdClusterSpec{}
			etcd.Name = etcdCluster
			for _, zone := range etcdZones {
				m := &api.EtcdMemberSpec{}
				m.Name = zone
				m.Zone = zone
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
		glog.Errorf("Must specify NetworkCIDR when VPC is set")
		os.Exit(1)
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
	}

	if c.SSHPublicKey != "" {
		c.SSHPublicKey = utils.ExpandPath(c.SSHPublicKey)
	}

	err = cluster.PerformAssignments()
	if err != nil {
		return fmt.Errorf("error populating configuration: %v", err)
	}
	err = api.PerformAssignmentsInstanceGroups(instanceGroups)
	if err != nil {
		return fmt.Errorf("error populating configuration: %v", err)
	}

	if creatingNewCluster {
		err = api.CreateClusterConfig(clusterRegistry, cluster, instanceGroups)
	} else {
		err = api.UpdateClusterConfig(clusterRegistry, cluster, instanceGroups)
	}
	if err != nil {
		return fmt.Errorf("error writing updated configuration: %v", err)
	}

	cmd := &cloudup.CreateClusterCmd{
		Cluster:         cluster,
		InstanceGroups:  instanceGroups,
		ModelStore:      c.ModelsBaseDir,
		Models:          strings.Split(c.Models, ","),
		ClusterRegistry: clusterRegistry,
		Target:          c.Target,
		NodeModel:       c.NodeModel,
		SSHPublicKey:    c.SSHPublicKey,
		OutDir:          c.OutDir,
		DryRun:          isDryrun,
	}

	return cmd.Run()
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
