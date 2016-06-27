package main

import (
	goflag "flag"
	"github.com/golang/glog"
	"github.com/spf13/pflag"
	"k8s.io/kube-deploy/upup/pkg/api"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
	"k8s.io/kube-deploy/upup/pkg/fi/vfs"
	"k8s.io/kubernetes/pkg/util/sets"
	"os"
	"os/exec"
	"path"
	"strings"
)

var EtcdClusters = []string{"main", "events"}

func main() {
	executableLocation, err := exec.LookPath(os.Args[0])
	if err != nil {
		glog.Fatalf("Cannot determine location of cloudup tool: %q.  Please report this problem!", os.Args[0])
	}

	modelsBaseDirDefault := path.Join(path.Dir(executableLocation), "models")

	dryrun := pflag.Bool("dryrun", false, "Don't create cloud resources; just show what would be done")
	target := pflag.String("target", "direct", "Target - direct, terraform")
	//configFile := pflag.String("conf", "", "Configuration file to load")
	modelsBaseDir := pflag.String("modelstore", modelsBaseDirDefault, "Source directory where models are stored")
	models := pflag.String("model", "proto,cloudup", "Models to apply (separate multiple models with commas)")
	nodeModel := pflag.String("nodemodel", "nodeup", "Model to use for node configuration")
	stateLocation := pflag.String("state", "", "Location to use to store configuration state")

	cloudProvider := pflag.String("cloud", "", "Cloud provider to use - gce, aws")

	zones := pflag.String("zones", "", "Zones in which to run the cluster")
	masterZones := pflag.String("master-zones", "", "Zones in which to run masters (must be an odd number)")

	project := pflag.String("project", "", "Project to use (must be set on GCE)")
	clusterName := pflag.String("name", "", "Name for cluster")
	kubernetesVersion := pflag.String("kubernetes-version", "", "Version of kubernetes to run (defaults to latest)")

	sshPublicKey := pflag.String("ssh-public-key", "~/.ssh/id_rsa.pub", "SSH public key to use")

	nodeSize := pflag.String("node-size", "", "Set instance size for nodes")

	masterSize := pflag.String("master-size", "", "Set instance size for masters")

	nodeCount := pflag.Int("node-count", 0, "Set the number of nodes")

	image := pflag.String("image", "", "Image to use")

	dnsZone := pflag.String("dns-zone", "", "DNS hosted zone to use (defaults to last two components of cluster name)")
	outDir := pflag.String("out", "", "Path to write any local output")

	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	pflag.Parse()
	goflag.CommandLine.Parse([]string{})

	isDryrun := false
	if *dryrun {
		isDryrun = true
		*target = "dryrun"
	}

	if *stateLocation == "" {
		glog.Errorf("--state is required")
		os.Exit(1)
	}

	statePath, err := vfs.Context.BuildVfsPath(*stateLocation)
	if err != nil {
		glog.Errorf("error building state location: %v", err)
		os.Exit(1)
	}

	if *outDir == "" {
		*outDir = "out"
	}

	stateStore, err := fi.NewVFSStateStore(statePath, isDryrun)
	if err != nil {
		glog.Errorf("error building state store: %v", err)
		os.Exit(1)
	}

	cluster, instanceGroups, err := api.ReadConfig(stateStore)
	if err != nil {
		glog.Errorf("error loading configuration: %v", err)
		os.Exit(1)
	}

	if *zones != "" {
		existingZones := make(map[string]*api.ClusterZoneSpec)
		for _, zone := range cluster.Spec.Zones {
			existingZones[zone.Name] = zone
		}

		for _, zone := range parseZoneList(*zones) {
			if existingZones[zone] == nil {
				cluster.Spec.Zones = append(cluster.Spec.Zones, &api.ClusterZoneSpec{
					Name: zone,
				})
			}
		}
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
	createEtcdCluster := false
	if *masterZones == "" {
		if len(masters) == 0 {
			// Default to putting into every zone
			// TODO: just the first 1 or 3 zones; or should we force users to declare?
			for _, zone := range cluster.Spec.Zones {
				g := &api.InstanceGroup{}
				g.Spec.Role = api.InstanceGroupRoleMaster
				g.Spec.Zones = []string{zone.Name}
				g.Spec.MinSize = fi.Int(1)
				g.Spec.MaxSize = fi.Int(1)
				g.Name = "master-" + zone.Name // Subsequent masters (if we support that) could be <zone>-1, <zone>-2
				instanceGroups = append(instanceGroups, g)
				masters = append(masters, g)
			}
			createEtcdCluster = true
		}
	} else {
		if len(masters) == 0 {
			for _, zone := range parseZoneList(*masterZones) {
				g := &api.InstanceGroup{}
				g.Spec.Role = api.InstanceGroupRoleMaster
				g.Spec.Zones = []string{zone}
				g.Spec.MinSize = fi.Int(1)
				g.Spec.MaxSize = fi.Int(1)
				g.Name = "master-" + zone
				instanceGroups = append(instanceGroups, g)
				masters = append(masters, g)
			}
			createEtcdCluster = true
		} else {
			// This is hard, because of the etcd cluster
			glog.Errorf("Cannot change master-zones from the CLI")
			os.Exit(1)
		}
	}

	if createEtcdCluster {
		zones := sets.NewString()
		for _, group := range instanceGroups {
			for _, zone := range group.Spec.Zones {
				zones.Insert(zone)
			}
		}
		etcdZones := zones.List()
		if (len(etcdZones) % 2) == 0 {
			// Not technically a requirement, but doesn't really make sense to allow
			glog.Errorf("There should be an odd number of master-zones, for etcd's quorum.  Hint: Use --zone and --master-zone to declare node zones and master zones separately.")
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

	if *nodeSize != "" {
		for _, group := range nodes {
			group.Spec.MachineType = *nodeSize
		}
	}

	if *image != "" {
		for _, group := range instanceGroups {
			group.Spec.Image = *image
		}
	}

	if *nodeCount != 0 {
		for _, group := range nodes {
			group.Spec.MinSize = nodeCount
			group.Spec.MaxSize = nodeCount
		}
	}

	if *masterSize != "" {
		for _, group := range masters {
			group.Spec.MachineType = *masterSize
		}
	}

	if *dnsZone != "" {
		cluster.Spec.DNSZone = *dnsZone
	}

	if *cloudProvider != "" {
		cluster.Spec.CloudProvider = *cloudProvider
	}

	if *project != "" {
		cluster.Spec.Project = *project
	}

	if *clusterName != "" {
		cluster.Name = *clusterName
	}

	if *kubernetesVersion != "" {
		cluster.Spec.KubernetesVersion = *kubernetesVersion
	}

	err = cluster.PerformAssignments()
	if err != nil {
		glog.Errorf("error populating configuration: %v", err)
		os.Exit(1)
	}
	err = api.PerformAssignmentsInstanceGroups(instanceGroups)
	if err != nil {
		glog.Errorf("error populating configuration: %v", err)
		os.Exit(1)
	}

	err = api.WriteConfig(stateStore, cluster, instanceGroups)
	if err != nil {
		glog.Errorf("error writing updated configuration: %v", err)
		os.Exit(1)
	}

	if *sshPublicKey != "" {
		*sshPublicKey = utils.ExpandPath(*sshPublicKey)
	}

	cmd := &cloudup.CreateClusterCmd{
		Cluster:        cluster,
		InstanceGroups: instanceGroups,
		ModelStore:     *modelsBaseDir,
		Models:         strings.Split(*models, ","),
		StateStore:     stateStore,
		Target:         *target,
		NodeModel:      *nodeModel,
		SSHPublicKey:   *sshPublicKey,
		OutDir:         *outDir,
	}

	//if *configFile != "" {
	//	//confFile := path.Join(cmd.StateDir, "kubernetes.yaml")
	//	err := cmd.LoadConfig(configFile)
	//	if err != nil {
	//		glog.Errorf("error loading config: %v", err)
	//		os.Exit(1)
	//	}
	//}

	err = cmd.Run()
	if err != nil {
		glog.Errorf("error running command: %v", err)
		os.Exit(1)
	}

	glog.Infof("Completed successfully")
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
