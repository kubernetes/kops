package main

import (
	goflag "flag"
	"github.com/golang/glog"
	"github.com/spf13/pflag"
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

	cluster, nodeSets, err := cloudup.ReadConfig(stateStore)
	if err != nil {
		glog.Errorf("error loading configuration: %v", err)
		os.Exit(1)
	}

	if *zones != "" {
		existingZones := make(map[string]*cloudup.ZoneConfig)
		for _, zone := range cluster.Zones {
			existingZones[zone.Name] = zone
		}

		for _, zone := range parseZoneList(*zones) {
			if existingZones[zone] == nil {
				cluster.Zones = append(cluster.Zones, &cloudup.ZoneConfig{
					Name: zone,
				})
			}
		}
	}

	createMasterVolumes := false
	if *masterZones == "" {
		if len(cluster.Masters) == 0 {
			// Default to putting into every zone
			// TODO: just the first 1 or 3 zones; or should we force users to declare?
			for _, zone := range cluster.Zones {
				m := &cloudup.MasterConfig{}
				m.Zone = zone.Name
				m.Name = zone.Name // Subsequent masters (if we support that) could be <zone>-1, <zone>-2
				cluster.Masters = append(cluster.Masters, m)
			}
			createMasterVolumes = true
		}
	} else {
		if len(cluster.Masters) == 0 {
			for _, zone := range parseZoneList(*masterZones) {
				m := &cloudup.MasterConfig{}
				m.Zone = zone
				m.Name = zone
				cluster.Masters = append(cluster.Masters, m)
			}
			createMasterVolumes = true
		} else {
			// This is hard, because of the etcd cluster
			glog.Errorf("Cannot change master-zones from the CLI")
			os.Exit(1)
		}
	}

	if createMasterVolumes {
		zones := sets.NewString()
		for _, m := range cluster.Masters {
			zones.Insert(m.Zone)
		}
		etcdZones := zones.List()
		if (len(etcdZones) % 2) == 0 {
			// Not technically a requirement, but doesn't really make sense to allow
			glog.Errorf("There should be an odd number of master-zones, for etcd's quorum.  Hint: Use --zone and --master-zone to declare node zones and master zones separately.")
			os.Exit(1)
		}

		for _, zone := range etcdZones {
			vol := &cloudup.VolumeConfig{}
			vol.Name = "etcd." + zone
			vol.Zone = zone
			vol.Roles = make(map[string]string)
			vol.Roles["etcd/main"] = zone + "/" + strings.Join(etcdZones, ",")
			vol.Roles["etcd/events"] = zone + "/" + strings.Join(etcdZones, ",")
			cluster.MasterVolumes = append(cluster.MasterVolumes, vol)
		}
	}

	if len(nodeSets) == 0 {
		nodeSets = append(nodeSets, &cloudup.NodeSetConfig{})
	}

	if *nodeSize != "" {
		for _, ns := range nodeSets {
			ns.MachineType = *nodeSize
		}
	}

	if *image != "" {
		for _, ns := range nodeSets {
			ns.Image = *image
		}
		for _, master := range cluster.Masters {
			master.Image = *image
		}
	}

	if *nodeCount != 0 {
		for _, ns := range nodeSets {
			ns.MinSize = nodeCount
			ns.MaxSize = nodeCount
		}
	}

	if *masterSize != "" {
		for _, master := range cluster.Masters {
			master.MachineType = *masterSize
		}
	}

	if *dnsZone != "" {
		cluster.DNSZone = *dnsZone
	}

	if *cloudProvider != "" {
		cluster.CloudProvider = *cloudProvider
	}

	if *project != "" {
		cluster.Project = *project
	}

	if *clusterName != "" {
		cluster.ClusterName = *clusterName
	}

	if *kubernetesVersion != "" {
		cluster.KubernetesVersion = *kubernetesVersion
	}

	err = cluster.PerformAssignments()
	if err != nil {
		glog.Errorf("error populating configuration: %v", err)
		os.Exit(1)
	}
	err = cloudup.PerformAssignmentsNodesets(nodeSets)
	if err != nil {
		glog.Errorf("error populating configuration: %v", err)
		os.Exit(1)
	}

	err = cloudup.WriteConfig(stateStore, cluster, nodeSets)
	if err != nil {
		glog.Errorf("error writing updated configuration: %v", err)
		os.Exit(1)
	}

	if *sshPublicKey != "" {
		*sshPublicKey = utils.ExpandPath(*sshPublicKey)
	}

	cmd := &cloudup.CreateClusterCmd{
		ClusterConfig: cluster,
		NodeSets:      nodeSets,
		ModelStore:    *modelsBaseDir,
		Models:        strings.Split(*models, ","),
		StateStore:    stateStore,
		Target:        *target,
		NodeModel:     *nodeModel,
		SSHPublicKey:  *sshPublicKey,
		OutDir:        *outDir,
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
