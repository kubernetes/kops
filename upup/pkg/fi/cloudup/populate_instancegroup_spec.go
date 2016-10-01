package cloudup

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/utils"
)

const DefaultNodeMachineTypeAWS = "t2.medium"
const DefaultNodeMachineTypeGCE = "n1-standard-2"

const DefaultMasterMachineTypeAWS = "m3.medium"
const DefaultMasterMachineTypeGCE = "n1-standard-1"

// PopulateInstanceGroupSpec sets default values in the InstanceGroup
// The InstanceGroup is simpler than the cluster spec, so we just populate in place (like the rest of k8s)
func PopulateInstanceGroupSpec(cluster *api.Cluster, input *api.InstanceGroup, channel *api.Channel) (*api.InstanceGroup, error) {
	err := input.Validate(false)
	if err != nil {
		return nil, err
	}

	ig := &api.InstanceGroup{}
	utils.JsonMergeStruct(ig, input)

	if ig.IsMaster() {
		if ig.Spec.MachineType == "" {
			ig.Spec.MachineType = defaultMasterMachineType(cluster)
		}
		if ig.Spec.MinSize == nil {
			ig.Spec.MinSize = fi.Int(1)
		}
		if ig.Spec.MaxSize == nil {
			ig.Spec.MaxSize = fi.Int(1)
		}
	} else {
		if ig.Spec.MachineType == "" {
			ig.Spec.MachineType = defaultNodeMachineType(cluster)
		}
		if ig.Spec.MinSize == nil {
			ig.Spec.MinSize = fi.Int(2)
		}
		if ig.Spec.MaxSize == nil {
			ig.Spec.MaxSize = fi.Int(2)
		}
	}

	if ig.Spec.AssociatePublicIP == nil {
		ig.Spec.AssociatePublicIP = fi.Bool(true)
	}

	if ig.Spec.Image == "" {
		ig.Spec.Image = defaultImage(cluster, channel)
	}

	if ig.IsMaster() {
		if len(ig.Spec.Zones) == 0 {
			return nil, fmt.Errorf("Master InstanceGroup %s did not specify any Zones", ig.Name)
		}
	} else {
		if len(ig.Spec.Zones) == 0 {
			for _, z := range cluster.Spec.Zones {
				ig.Spec.Zones = append(ig.Spec.Zones, z.Name)
			}
		}
	}

	return ig, nil
}

// defaultNodeMachineType returns the default MachineType for nodes, based on the cloudprovider
func defaultNodeMachineType(cluster *api.Cluster) string {
	switch cluster.Spec.CloudProvider {
	case "aws":
		return DefaultNodeMachineTypeAWS
	case "gce":
		return DefaultNodeMachineTypeGCE
	default:
		glog.V(2).Infof("Cannot set default MachineType for CloudProvider=%q", cluster.Spec.CloudProvider)
		return ""
	}
}

// defaultMasterMachineType returns the default MachineType for masters, based on the cloudprovider
func defaultMasterMachineType(cluster *api.Cluster) string {
	// TODO: We used to have logic like the following...
	//	{{ if gt .TotalNodeCount 500 }}
	//	MasterMachineType: n1-standard-32
	//	{{ else if gt .TotalNodeCount 250 }}
	//MasterMachineType: n1-standard-16
	//{{ else if gt .TotalNodeCount 100 }}
	//MasterMachineType: n1-standard-8
	//{{ else if gt .TotalNodeCount 10 }}
	//MasterMachineType: n1-standard-4
	//{{ else if gt .TotalNodeCount 5 }}
	//MasterMachineType: n1-standard-2
	//{{ else }}
	//MasterMachineType: n1-standard-1
	//{{ end }}
	//
	//{{ if gt TotalNodeCount 500 }}
	//MasterMachineType: c4.8xlarge
	//{{ else if gt TotalNodeCount 250 }}
	//MasterMachineType: c4.4xlarge
	//{{ else if gt TotalNodeCount 100 }}
	//MasterMachineType: m3.2xlarge
	//{{ else if gt TotalNodeCount 10 }}
	//MasterMachineType: m3.xlarge
	//{{ else if gt TotalNodeCount 5 }}
	//MasterMachineType: m3.large
	//{{ else }}
	//MasterMachineType: m3.medium
	//{{ end }}

	switch cluster.Spec.CloudProvider {
	case "aws":
		return DefaultMasterMachineTypeAWS
	case "gce":
		return DefaultMasterMachineTypeGCE
	default:
		glog.V(2).Infof("Cannot set default MachineType for CloudProvider=%q", cluster.Spec.CloudProvider)
		return ""
	}
}

// defaultImage returns the default Image, based on the cloudprovider
func defaultImage(cluster *api.Cluster, channel *api.Channel) string {
	if channel != nil {
		for _, image := range channel.Spec.Images {
			if image.ProviderID != cluster.Spec.CloudProvider {
				continue
			}

			return image.Name
		}
	}

	glog.Infof("Cannot set default Image for CloudProvider=%q", cluster.Spec.CloudProvider)
	return ""
}
