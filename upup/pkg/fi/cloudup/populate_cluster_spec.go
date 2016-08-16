package cloudup

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/upup/pkg/fi/vfs"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
)

var EtcdClusters = []string{"main", "events"}

type populateClusterSpec struct {
	// InputCluster is the api object representing the whole cluster, as input by the user
	// We build it up into a complete config, but we write the values as input
	InputCluster *api.Cluster

	// ModelStore is the location where models are found
	ModelStore string
	// Models is a list of cloudup models to apply
	Models []string

	// ClusterRegistry manages the cluster configuration storage
	ClusterRegistry *api.ClusterRegistry

	// fullCluster holds the built completed cluster spec
	fullCluster *api.Cluster
}

func findModelStore() (string, error) {
	executableLocation, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", fmt.Errorf("Cannot determine location of kops tool: %q.  Please report this problem!", os.Args[0])
	}

	var locations []string
	locations = append(locations, path.Join(path.Dir(executableLocation), "models"))

	gopath := os.Getenv("GOPATH")
	locations = append(locations, path.Join(gopath, "src/k8s.io/kops/upup/models"))

	for _, location := range locations {
		_, err = os.Stat(location)
		if err == nil {
			return location, nil
		}

		if os.IsNotExist(err) {
			continue
		} else {
			glog.Warningf("error accessing models directory %q: %v", location, err)
		}
	}

	return "", fmt.Errorf("models directory not found at %q.  Please report this problem!", strings.Join(locations, ","))
}

// PopulateClusterSpec takes a user-specified cluster spec, and computes the full specification that should be set on the cluster.
// We do this so that we don't need any real "brains" on the node side.
func PopulateClusterSpec(cluster *api.Cluster, clusterRegistry *api.ClusterRegistry) (*api.Cluster, error) {
	modelStore, err := findModelStore()
	if err != nil {
		return nil, err
	}

	c := &populateClusterSpec{
		InputCluster:    cluster,
		ModelStore:      modelStore,
		Models:          []string{"config"},
		ClusterRegistry: clusterRegistry,
	}
	err = c.run()
	if err != nil {
		return nil, err
	}
	return c.fullCluster, nil
}

func (c *populateClusterSpec) run() error {
	err := c.InputCluster.Validate(false)
	if err != nil {
		return err
	}

	// Copy cluster & instance groups, so we can modify them freely
	cluster := &api.Cluster{}
	utils.JsonMergeStruct(cluster, c.InputCluster)

	err = c.assignSubnets(cluster)
	if err != nil {
		return err
	}

	err = cluster.FillDefaults()
	if err != nil {
		return err
	}

	// TODO: Move to validate?
	// Check that instance groups are defined in valid zones
	{
		clusterZones := make(map[string]*api.ClusterZoneSpec)
		for _, z := range cluster.Spec.Zones {
			if clusterZones[z.Name] != nil {
				return fmt.Errorf("Zones contained a duplicate value: %v", z.Name)
			}
			clusterZones[z.Name] = z
		}

		// Check etcd configuration
		{
			for i, etcd := range cluster.Spec.EtcdClusters {
				if etcd.Name == "" {
					return fmt.Errorf("EtcdClusters #%d did not specify a Name", i)
				}

				for i, m := range etcd.Members {
					if m.Name == "" {
						return fmt.Errorf("EtcdMember #%d of etcd-cluster %s did not specify a Name", i, etcd.Name)
					}

					z := m.Zone
					if z == "" {
						return fmt.Errorf("EtcdMember %s:%s did not specify a Zone", etcd.Name, m.Name)
					}
				}

				etcdZones := make(map[string]*api.EtcdMemberSpec)
				etcdNames := make(map[string]*api.EtcdMemberSpec)

				for _, m := range etcd.Members {
					if etcdNames[m.Name] != nil {
						return fmt.Errorf("EtcdMembers found with same name %q in etcd-cluster %q", m.Name, etcd.Name)
					}

					if etcdZones[m.Zone] != nil {
						// Maybe this should just be a warning
						return fmt.Errorf("EtcdMembers are in the same zone %q in etcd-cluster %q", m.Zone, etcd.Name)
					}

					if clusterZones[m.Zone] == nil {
						return fmt.Errorf("EtcdMembers for %q is configured in zone %q, but that is not configured at the k8s-cluster level", etcd.Name, m.Zone)
					}
					etcdZones[m.Zone] = m
				}

				if (len(etcdZones) % 2) == 0 {
					// Not technically a requirement, but doesn't really make sense to allow
					return fmt.Errorf("There should be an odd number of master-zones, for etcd's quorum.  Hint: Use --zones and --master-zones to declare node zones and master zones separately.")
				}
			}
		}
	}

	if c.ClusterRegistry == nil {
		return fmt.Errorf("ClusterRegistry is required")
	}

	keyStore := c.ClusterRegistry.KeyStore(cluster.Name)
	// Always assume a dry run during this phase
	keyStore.(*fi.VFSCAStore).DryRun = true
	secretStore := c.ClusterRegistry.SecretStore(cluster.Name)

	if vfs.IsClusterReadable(secretStore.VFSPath()) {
		vfsPath := secretStore.VFSPath()
		cluster.Spec.SecretStore = vfsPath.Path()
		if s3Path, ok := vfsPath.(*vfs.S3Path); ok {
			if cluster.Spec.MasterPermissions == nil {
				cluster.Spec.MasterPermissions = &api.CloudPermissions{}
			}
			cluster.Spec.MasterPermissions.AddS3Bucket(s3Path.Bucket())
			if cluster.Spec.NodePermissions == nil {
				cluster.Spec.NodePermissions = &api.CloudPermissions{}
			}
			cluster.Spec.NodePermissions.AddS3Bucket(s3Path.Bucket())
		}
	} else {
		// We could implement this approach, but it seems better to get all clouds using cluster-readable storage
		return fmt.Errorf("secrets path is not cluster readable: %v", secretStore.VFSPath())
	}

	if vfs.IsClusterReadable(keyStore.VFSPath()) {
		vfsPath := keyStore.VFSPath()
		cluster.Spec.KeyStore = vfsPath.Path()
		if s3Path, ok := vfsPath.(*vfs.S3Path); ok {
			if cluster.Spec.MasterPermissions == nil {
				cluster.Spec.MasterPermissions = &api.CloudPermissions{}
			}
			cluster.Spec.MasterPermissions.AddS3Bucket(s3Path.Bucket())
			if cluster.Spec.NodePermissions == nil {
				cluster.Spec.NodePermissions = &api.CloudPermissions{}
			}
			cluster.Spec.NodePermissions.AddS3Bucket(s3Path.Bucket())
		}
	} else {
		// We could implement this approach, but it seems better to get all clouds using cluster-readable storage
		return fmt.Errorf("keyStore path is not cluster readable: %v", keyStore.VFSPath())
	}

	configPath, err := c.ClusterRegistry.ConfigurationPath(cluster.Name)
	if err != nil {
		return err
	}
	if vfs.IsClusterReadable(configPath) {
		cluster.Spec.ConfigStore = configPath.Path()
	} else {
		// We do support this...
	}

	// Normalize k8s version
	versionWithoutV := strings.TrimSpace(cluster.Spec.KubernetesVersion)
	if strings.HasPrefix(versionWithoutV, "v") {
		versionWithoutV = versionWithoutV[1:]
	}
	if cluster.Spec.KubernetesVersion != versionWithoutV {
		glog.Warningf("Normalizing kubernetes version: %q -> %q", cluster.Spec.KubernetesVersion, versionWithoutV)
		cluster.Spec.KubernetesVersion = versionWithoutV
	}

	cloud, err := BuildCloud(cluster)
	if err != nil {
		return err
	}

	if cluster.Spec.DNSZone == "" {
		dnsZone, err := cloud.FindDNSHostedZone(cluster.Name)
		if err != nil {
			return fmt.Errorf("Error determining default DNS zone; please specify --dns-zone: %v", err)
		}
		glog.Infof("Defaulting DNS zone to: %s", dnsZone)
		cluster.Spec.DNSZone = dnsZone
	}

	tags, err := buildClusterTags(cluster)
	if err != nil {
		return err
	}

	tf := &TemplateFunctions{
		cluster: cluster,
		tags:    tags,
	}

	templateFunctions := make(template.FuncMap)

	tf.AddTo(templateFunctions)

	specBuilder := &SpecBuilder{
		OptionsLoader: loader.NewOptionsLoader(templateFunctions),
		Tags:          tags,
	}
	completed, err := specBuilder.BuildCompleteSpec(&cluster.Spec, c.ModelStore, c.Models)
	if err != nil {
		return fmt.Errorf("error building complete spec: %v", err)
	}

	fullCluster := &api.Cluster{}
	*fullCluster = *cluster
	fullCluster.Spec = *completed
	tf.cluster = fullCluster

	err = fullCluster.Validate(true)
	if err != nil {
		return fmt.Errorf("Completed cluster failed validation: %v", err)
	}

	c.fullCluster = fullCluster

	return nil
}

func (c *populateClusterSpec) assignSubnets(cluster *api.Cluster) error {
	if cluster.Spec.NonMasqueradeCIDR == "" {
		glog.Warningf("NonMasqueradeCIDR not set; can't auto-assign dependent subnets")
		return nil
	}

	_, nonMasqueradeCIDR, err := net.ParseCIDR(cluster.Spec.NonMasqueradeCIDR)
	if err != nil {
		return fmt.Errorf("error parsing NonMasqueradeCIDR %q: %v", cluster.Spec.NonMasqueradeCIDR, err)
	}
	nmOnes, nmBits := nonMasqueradeCIDR.Mask.Size()

	if cluster.Spec.KubeControllerManager == nil {
		cluster.Spec.KubeControllerManager = &api.KubeControllerManagerConfig{}
	}

	if cluster.Spec.KubeControllerManager.ClusterCIDR == "" {
		// Allocate as big a range as possible: the NonMasqueradeCIDR mask + 1, with a '1' in the extra bit
		ip := nonMasqueradeCIDR.IP.Mask(nonMasqueradeCIDR.Mask)

		ip4 := ip.To4()
		if ip4 != nil {
			n := binary.BigEndian.Uint32(ip4)
			n += uint32(1 << uint(nmBits-nmOnes-1))
			ip = make(net.IP, len(ip4))
			binary.BigEndian.PutUint32(ip, n)
		} else {
			return fmt.Errorf("IPV6 subnet computations not yet implements")
		}

		cidr := net.IPNet{IP: ip, Mask: net.CIDRMask(nmOnes+1, nmBits)}
		cluster.Spec.KubeControllerManager.ClusterCIDR = cidr.String()
		glog.V(2).Infof("Defaulted KubeControllerManager.ClusterCIDR to %v", cluster.Spec.KubeControllerManager.ClusterCIDR)
	}

	if cluster.Spec.ServiceClusterIPRange == "" {
		// Allocate from the '0' subnet; but only carve off 1/4 of that (i.e. add 1 + 2 bits to the netmask)
		cidr := net.IPNet{IP: nonMasqueradeCIDR.IP.Mask(nonMasqueradeCIDR.Mask), Mask: net.CIDRMask(nmOnes+3, nmBits)}
		cluster.Spec.ServiceClusterIPRange = cidr.String()
		glog.V(2).Infof("Defaulted ServiceClusterIPRange to %v", cluster.Spec.ServiceClusterIPRange)
	}

	return nil
}
